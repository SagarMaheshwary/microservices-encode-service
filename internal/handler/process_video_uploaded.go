package handler

import (
	"fmt"
	"os"
	"path"
	"strconv"

	cons "github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/helper"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/aws"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/publisher"
	ve "github.com/sagarmaheshwary/microservices-encode-service/internal/lib/video_encoder"
)

type VideoUploadedPayload struct {
	VideoId     string `json:"video_id"`
	ThumbnailId string `json:"thumbnail_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"`
	UserId      int    `json:"user_id"`
}

type EncodedVideo struct {
	Title        string                   `json:"title"`
	Description  string                   `json:"description"`
	PublishedAt  string                   `json:"published_at"`
	Height       int                      `json:"height"`
	Width        int                      `json:"width"`
	Duration     int                      `json:"duration"`
	Resolutions  []EncodedVideoResolution `json:"resolutions"`
	UserId       int                      `json:"user_id"`
	OriginalId   string                   `json:"original_id"`
	ThumbnailUrl string                   `json:"thumbnail_url"`
}

type EncodedVideoResolution struct {
	Height int      `json:"height"`
	Width  int      `json:"width"`
	Codec  string   `json:"codec"`
	Chunks []string `json:"chunks"`
}

func ProcessVideoUploaded(data *VideoUploadedPayload) error {
	var err error
	var encodedVideoResolutions []EncodedVideoResolution

	videoDirPath := path.Join(helper.RootDir(), cons.TempVideosDownloadDirectory, data.VideoId)
	objectKey := fmt.Sprintf("%s/%s", cons.S3RawVideosDirectory, data.VideoId)

	videoPath, err := downloadFileFromS3(videoDirPath, objectKey)

	if err != nil {
		return err
	}

	info, err := ve.GetVideoInfo(videoPath)

	if err != nil {
		return err
	}

	encodeFrom := ve.GetEncodingStartIndex(info.Width, info.Height)

	for i := encodeFrom; i < len(ve.VideoEncodeOptions); i++ {
		opt := ve.VideoEncodeOptions[i]

		filePrefix := fmt.Sprintf("%dx%d-%s", opt.Width, opt.Height, opt.VideoBitRate)

		encodedVideoPath := path.Join(videoDirPath, fmt.Sprintf("%s.%s", filePrefix, opt.Format))
		chunkDir := path.Join(videoDirPath, filePrefix)

		err = encodeVideoToResolution(videoPath, encodedVideoPath, &opt)

		if err != nil {
			return err
		}

		err = encodeVideoToDash(encodedVideoPath, chunkDir, &opt)

		if err != nil {
			return err
		}

		uploadPrefix := path.Join(cons.S3EncodedVideosDirectory, data.VideoId, filePrefix)

		chunks, err := uploadChunksToS3(uploadPrefix, chunkDir)

		if err != nil {
			return err
		}

		encodedVideoResolutions = append(encodedVideoResolutions, EncodedVideoResolution{
			Height: opt.Height,
			Width:  opt.Width,
			Codec:  opt.VideoCodec,
			Chunks: chunks,
		})

		log.Info("Processed chunks for %s", chunkDir)
	}

	log.Info("Video encoding %s completed", data.VideoId)

	duration, _ := strconv.ParseFloat(info.Duration, strconv.IntSize)

	err = publisher.P.Publish(cons.QueueVideoCatalogService, &broker.MessageType{
		Key: cons.MessageTypeVideoEncodingCompleted,
		Data: &EncodedVideo{
			Title:        data.Title,
			Description:  data.Description,
			PublishedAt:  data.PublishedAt,
			Height:       info.Height,
			Width:        info.Width,
			Duration:     int(duration),
			Resolutions:  encodedVideoResolutions,
			UserId:       data.UserId,
			OriginalId:   data.VideoId,
			ThumbnailUrl: fmt.Sprintf("%s/%s", cons.S3ThumbnailsDirectory, data.ThumbnailId),
		},
	})

	if err != nil {
		log.Error("Unable to send data to video catalog service %v", err)

		return err
	}

	return nil
}

func encodeVideoToResolution(inPath string, outPath string, opt *ve.VideoEncodeOption) error {
	err := ve.EncodeVideoToResolution(inPath, outPath, &ve.EncodeVideoToResolutionArgs{
		VideoCodec:   opt.VideoCodec,
		VideoBitRate: opt.VideoBitRate,
		AudioCodec:   opt.AudioCodec,
		AudioBitRate: opt.AudioBitRate,
		Resolution:   fmt.Sprintf("%d:%d", opt.Width, opt.Height),
	})

	if err != nil {
		return err
	}

	return nil
}

func encodeVideoToDash(inPath string, outPath string, opt *ve.VideoEncodeOption) error {
	err := os.Mkdir(outPath, os.ModePerm)

	if err != nil {
		log.Error("unable to create directory %s", outPath)

		return err
	}

	dashPath := path.Join(outPath, fmt.Sprintf("%s.%s", helper.UniqueString(8), cons.ExtensionMPEGDASH))

	err = ve.EncodeVideoToDash(inPath, dashPath, &ve.EncodeVideoToDashArgs{
		Copy:            "copy",
		SegmentDuration: opt.SegmentTime,
		UseTimeline:     1,
		UseTemplate:     1,
	})

	if err != nil {
		return err
	}

	return nil
}

func uploadChunksToS3(uploadPathPrefix string, chunkDir string) ([]string, error) {
	files, err := os.ReadDir(chunkDir)
	chunks := make([]string, len(files))

	if err != nil {
		log.Info("Unable to read chunks dir %s", chunkDir)

		return chunks, err
	}

	for i, f := range files {
		log.Info("File Name %s", f.Name())

		filePath := path.Join(chunkDir, f.Name())
		uploadId := path.Join(uploadPathPrefix, f.Name())

		err := aws.UploadObjectToS3(filePath, uploadId)

		if err != nil {
			return chunks, err
		}

		chunks[i] = uploadId
	}

	return chunks, nil
}

func downloadFileFromS3(dirPath string, filename string) (string, error) {
	err := os.Mkdir(dirPath, os.ModePerm)

	if err != nil {
		log.Error("Unable to create directory for video %s", dirPath)

		return "", err
	}

	videoPath := path.Join(dirPath, helper.UniqueString(8))

	err = aws.DownloadS3Object(filename, videoPath)

	if err != nil {
		return "", err
	}

	return videoPath, nil
}
