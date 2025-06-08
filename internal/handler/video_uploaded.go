package handler

import (
	"fmt"
	"os"
	"path"

	"strconv"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/constant"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/helper"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/aws"
	"golang.org/x/net/context"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/broker"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/logger"
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
	Title           string `json:"title"`
	Description     string `json:"description"`
	PublishedAt     string `json:"published_at"`
	Height          int    `json:"height"`
	Width           int    `json:"width"`
	DurationSeconds int    `json:"duration"`
	UserId          int    `json:"user_id"`
	OriginalId      string `json:"original_id"`
	Thumbnail       string `json:"thumbnail"`
	Path            string `json:"path"`
}

func ProcessVideoUploadedMessage(ctx context.Context, data *VideoUploadedMessage) error {
	var err error

	videoDirPath := path.Join(helper.GetRootDir(), "..", constant.TempVideosDownloadDirectory, data.VideoId)

	objectKey := fmt.Sprintf("%s/%s", constant.S3RawVideosDirectory, data.VideoId)

	videoPath, err := downloadFileFromS3(objectKey, videoDirPath)

	if err != nil {
		return err
	}

	info, err := ve.GetVideoInfo(videoPath)

	if err != nil {
		logger.Error("ve.GetVideoInfo failed! %v", err)

		return err
	}

	encodeFrom := ve.GetEncodingStartIndex(info.Width, info.Height)

	opt := ve.VideoEncodeOptions[encodeFrom]

	filePrefix := fmt.Sprintf("%dx%d", opt.Width, opt.Height)

	chunkDirectory := path.Join(videoDirPath, filePrefix)
	encodedVideo := path.Join(videoDirPath, fmt.Sprintf("%s.%s", filePrefix, opt.Format))

	err = encodeVideoToResolution(videoPath, encodedVideo, &opt)

	if err != nil {
		logger.Error("encodeVideoToResolution failed! %v", err)

		return err
	}

	err = encodeVideoToDash(encodedVideo, chunkDirectory, &opt)

	if err != nil {
		logger.Error("encodeVideoToDash failed! %v", err)

		return err
	}

	uploadPrefix := path.Join(constant.S3EncodedVideosDirectory, data.VideoId, filePrefix)

	err = uploadChunksToS3(uploadPrefix, chunkDirectory)

	if err != nil {
		logger.Error("uploadChunksToS3 failed! %v", err)

		return err
	}

	logger.Info("Processed chunks for %s", chunkDirectory)

	logger.Info("Video encoding %s completed", data.VideoId)

	duration, _ := strconv.ParseFloat(info.Duration, strconv.IntSize)

	err = publisher.P.Publish(ctx, constant.QueueVideoCatalogService, &broker.MessageType{
		Key: constant.MessageTypeVideoEncodingCompleted,
		Data: &VideoEncodingCompletedMessage{
			Title:           data.Title,
			Description:     data.Description,
			Height:          info.Height,
			Width:           info.Width,
			DurationSeconds: int(duration),
			UserId:          data.UserId,
			OriginalId:      data.VideoId,
			Thumbnail:       fmt.Sprintf("%s/%s", constant.S3ThumbnailsDirectory, data.ThumbnailId),
			PublishedAt:     data.PublishedAt,
			Path:            uploadPrefix,
		},
	})

	if err != nil {
		logger.Error("Unable to send data to video catalog service! %v", err)

		return err
	}

	return nil
}

func encodeVideoToResolution(in string, out string, opt *ve.VideoEncodeOption) error {
	err := ve.EncodeVideoToResolution(in, out, &ve.EncodeVideoToResolutionArgs{
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

func encodeVideoToDash(in string, out string, opt *ve.VideoEncodeOption) error {
	err := os.Mkdir(out, os.ModePerm)

	if err != nil {
		logger.Error("Unable to create directory! %s", out)

		return err
	}

	p := path.Join(out, constant.MPEGDASHManifestFile)

	err = ve.EncodeVideoToDash(in, p, &ve.EncodeVideoToDashArgs{
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

func uploadChunksToS3(uploadPathPrefix string, chunkDir string) error {
	files, err := os.ReadDir(chunkDir)

	if err != nil {
		logger.Info("Unable to read chunks directory! %s", chunkDir)

		return err
	}

	for i, f := range files {
		p := path.Join(chunkDir, f.Name())
		uploadId := path.Join(uploadPathPrefix, f.Name())

		logger.Info(`Uploading chunk file: \"%s", upload path: "%s" (%d)`, p, uploadId, i+1)

		err := aws.UploadObjectToS3(p, uploadId)

		if err != nil {
			return err //@TODO: retry failed chunks
		}
	}

	return nil
}

func downloadFileFromS3(filename string, downloadDirectory string) (string, error) {
	err := os.Mkdir(downloadDirectory, os.ModePerm)

	if err != nil {
		logger.Error("Unable to create directory for video %s", downloadDirectory)

		return "", err
	}

	p := path.Join(downloadDirectory, helper.UniqueString(8))

	err = aws.DownloadS3Object(filename, p)

	if err != nil {
		return "", err
	}

	return p, nil
}
