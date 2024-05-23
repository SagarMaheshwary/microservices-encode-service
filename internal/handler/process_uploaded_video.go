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

type ProcessUploadedVideoPayload struct {
	UploadId    string `json:"upload_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"`
}

type UploadedVideo struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	PublishedAt string                  `json:"published_at"`
	Height      int                     `json:"height"`
	Width       int                     `json:"width"`
	Duration    int                     `json:"duration"`
	Encodings   []UploadedVideoEncoding `json:"encodings"`
}

type UploadedVideoEncoding struct {
	Height int      `json:"height"`
	Width  int      `json:"width"`
	Codec  string   `json:"codec"`
	Chunks []string `json:"chunks"`
}

func Test() {
	/*
		{
		  "id": "video-id",
		  "title": "Video Title",
		  "description": "This is description.",
		  "resolution": "",
		  "duration": "",
		  "published_at": "",
		  "encodings": [
		    {
		      "resolution": "854x480",
		      "encoding": "vp9",
		      "chunks": ["video-id/854x480-250k/chunk-file-name"]
		    }
		  ]
		}
	*/

	//create directory in assets/videos for current video process
	//download video from s3 to local
	//read video resolution

	//LOOP START (loop over map values by video resolution)
	//	create directory for video chunks
	//	create chunks using ffmpeg
	//	read chunks directory and list chunk files

	//	LOOP START (loop over chunks list)
	//		upload each chunk to s3
	//	LOOP END

	//LOOP END
}

func HandleProcessUploadedVideo(data *ProcessUploadedVideoPayload) error {
	log.Info("Data To Encode %v", data)

	var err error
	var uploadedVideoEncodings []UploadedVideoEncoding

	videoDirPath := path.Join(helper.RootDir(), "assets", "videos", data.UploadId)

	videoPath, err := downloadFileFromS3(videoDirPath, data.UploadId)

	if err != nil {
		return err
	}

	info, err := ve.GetVideoInfo(videoPath)

	if err != nil {
		return err
	}

	encodeFrom := ve.GetVideoEncodeOptionsIndex(info.Width, info.Height)

	for i := encodeFrom; i < len(ve.VideoEncodeOptions); i++ {
		opt := ve.VideoEncodeOptions[i]

		filePrefix := fmt.Sprintf("%dx%d-%s", opt.Width, opt.Height, opt.VideoBitRate)

		encodedVideoPath := path.Join(videoDirPath, fmt.Sprintf("%s.%s", filePrefix, opt.Format))
		chunkDir := path.Join(videoDirPath, fmt.Sprintf("chunks-%s", filePrefix))

		err = encodeVideoToResolution(videoPath, encodedVideoPath, &opt)

		if err != nil {
			return err
		}

		err = encodeVideoToDash(encodedVideoPath, chunkDir, &opt)

		if err != nil {
			return err
		}

		uploadPrefix := path.Join(data.UploadId, filePrefix)

		chunks, err := uploadChunksToS3(uploadPrefix, chunkDir)

		if err != nil {
			return err
		}

		uploadedVideoEncodings = append(uploadedVideoEncodings, UploadedVideoEncoding{
			Height: opt.Height,
			Width:  opt.Width,
			Codec:  opt.VideoCodec,
			Chunks: chunks,
		})

		log.Info("Processed chunks for %s", chunkDir)
	}

	log.Info("Video encoding completed")

	duration, _ := strconv.Atoi(info.Duration)

	err = publisher.P.Publish(cons.QueueEncodeService, &broker.MessageType{
		Key: cons.MessageTypeEncodeUploadedVideo,
		Data: &UploadedVideo{
			Title:       data.Title,
			Description: data.Description,
			PublishedAt: data.PublishedAt,
			Height:      info.Height,
			Width:       info.Width,
			Duration:    duration,
			Encodings:   uploadedVideoEncodings,
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

	dashPath := path.Join(outPath, "output.mpd")

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

	for _, f := range files {
		log.Info("File Name %s", f.Name())

		filePath := path.Join(chunkDir, f.Name())
		uploadId := path.Join(uploadPathPrefix, f.Name())

		err := aws.PutFileToS3(filePath, uploadId)

		if err != nil {
			return chunks, err
		}

		chunks = append(chunks, uploadId)
	}

	return chunks, nil
}

func downloadFileFromS3(dirPath string, filename string) (string, error) {
	err := os.Mkdir(dirPath, os.ModePerm)

	if err != nil {
		log.Error("Unable to create directory for video %s", dirPath)

		return "", err
	}

	videoPath := path.Join(dirPath, "original")

	err = aws.DownloadFileFromS3(filename, videoPath)

	if err != nil {
		return "", err
	}

	return videoPath, nil
}
