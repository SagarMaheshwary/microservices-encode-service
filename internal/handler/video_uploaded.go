package handler

import (
	"fmt"
	"os"
	"path"

	"strconv"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/helper"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/aws"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/publisher"
	ve "github.com/sagarmaheshwary/microservices-encode-service/internal/lib/video_encoder"
)

type VideoUploadedMessage struct {
	VideoId     string `json:"video_id"`
	ThumbnailId string `json:"thumbnail_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"`
	UserId      int    `json:"user_id"`
}

type VideoEncodingCompletedMessage struct {
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	PublishedAt        string              `json:"published_at"`
	Height             int                 `json:"height"`
	Width              int                 `json:"width"`
	Duration           int                 `json:"duration"`
	EncodedResolutions []EncodedResolution `json:"resolutions"`
	UserId             int                 `json:"user_id"`
	OriginalId         string              `json:"original_id"`
	Thumbnail          string              `json:"thumbnail"`
	Path               string              `json:"path"`
}

type EncodedResolution struct {
	Height int      `json:"height"`
	Width  int      `json:"width"`
	Chunks []string `json:"chunks"`
}

func ProcessVideoUploadedMessage(data *VideoUploadedMessage) error {
	var err error
	var encodedResolutions []EncodedResolution

	videoDirPath := path.Join(helper.RootDir(), constant.TempVideosDownloadDirectory, data.VideoId)
	objectKey := fmt.Sprintf("%s/%s", constant.S3RawVideosDirectory, data.VideoId)

	videoPath, err := downloadFileFromS3(objectKey, videoDirPath)

	if err != nil {
		return err
	}

	info, err := ve.GetVideoInfo(videoPath)

	if err != nil {
		log.Error("ve.GetVideoInfo failed.")

		return err
	}

	encodeFrom := ve.GetEncodingStartIndex(info.Width, info.Height)

	opt := ve.VideoEncodeOptions[encodeFrom]

	filePrefix := fmt.Sprintf("%dx%d", opt.Width, opt.Height)

	chunkDirectory := path.Join(videoDirPath, filePrefix)
	encodedVideo := path.Join(videoDirPath, fmt.Sprintf("%s.%s", filePrefix, opt.Format))

	err = encodeVideoToResolution(videoPath, encodedVideo, &opt)

	if err != nil {
		log.Error("encodeVideoToResolution failed.")

		return err
	}

	err = encodeVideoToDash(encodedVideo, chunkDirectory, &opt)

	if err != nil {
		log.Error("encodeVideoToDash failed.")

		return err
	}

	uploadPrefix := path.Join(constant.S3EncodedVideosDirectory, data.VideoId, filePrefix)

	chunks, err := uploadChunksToS3(uploadPrefix, chunkDirectory)

	if err != nil {
		log.Error("uploadChunksToS3 failed.")

		return err
	}

	encodedResolutions = append(encodedResolutions, EncodedResolution{
		Height: opt.Height,
		Width:  opt.Width,
		Chunks: chunks,
	})

	log.Info("Processed chunks for %s", chunkDirectory)

	log.Info("Video encoding %s completed", data.VideoId)

	duration, _ := strconv.ParseFloat(info.Duration, strconv.IntSize)

	err = publisher.P.Publish(constant.QueueVideoCatalogService, &broker.MessageType{
		Key: constant.MessageTypeVideoEncodingCompleted,
		Data: &VideoEncodingCompletedMessage{
			Title:              data.Title,
			Description:        data.Description,
			Height:             info.Height,
			Width:              info.Width,
			Duration:           int(duration),
			EncodedResolutions: encodedResolutions,
			UserId:             data.UserId,
			OriginalId:         data.VideoId,
			Thumbnail:          fmt.Sprintf("%s/%s", constant.S3ThumbnailsDirectory, data.ThumbnailId),
			PublishedAt:        data.PublishedAt,
			Path:               uploadPrefix,
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

	dashPath := path.Join(outPath, constant.MPEGDASHManifestFile)

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
		chunkPath := path.Join(chunkDir, f.Name())
		uploadId := path.Join(uploadPathPrefix, f.Name())

		log.Info(`Uploading chunk file: \"%s", upload path: "%s" (%d)`, chunkPath, uploadId, i+1)

		err := aws.UploadObjectToS3(chunkPath, uploadId)

		if err != nil {
			return chunks, err //@TODO: retry failed chunks
		}

		chunks[i] = uploadId
	}

	return chunks, nil
}

func downloadFileFromS3(filename string, downloadDirectory string) (string, error) {
	err := os.Mkdir(downloadDirectory, os.ModePerm)

	if err != nil {
		log.Error("Unable to create directory for video %s", downloadDirectory)

		return "", err
	}

	videoPath := path.Join(downloadDirectory, helper.UniqueString(8))

	err = aws.DownloadS3Object(filename, videoPath)

	if err != nil {
		return "", err
	}

	return videoPath, nil
}
