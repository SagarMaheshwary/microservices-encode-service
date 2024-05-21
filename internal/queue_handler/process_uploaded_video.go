package queue_handler

import (
	"fmt"
	"os"
	"path"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/helper"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/aws"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/encode"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

type ProcessUploadedVideoPayload struct {
	UploadId    string `json:"upload_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"`
}

func Test() {
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
	videoDirPath := path.Join(helper.RootDir(), "assets", "videos", data.UploadId)

	err = os.Mkdir(videoDirPath, os.ModePerm)

	if err != nil {
		log.Error("Unable to create directory for video %s", videoDirPath)

		return err
	}

	videoPath := path.Join(videoDirPath, "original")

	err = aws.DownloadFileFromS3(data.UploadId, videoPath)

	if err != nil {
		return err
	}

	info := encode.GetVideoInfo(videoPath)

	encodeFrom := encode.GetVideoEncodeOptionsIndex(info.Width, info.Height)

	for i := encodeFrom; i < len(encode.VideoEncodeOptions); i++ {
		opt := encode.VideoEncodeOptions[i]

		encodedVideoPath := path.Join(videoDirPath, fmt.Sprintf("%dx%d-%s.%s", opt.Width, opt.Height, opt.VideoBitRate, opt.Format))
		chunkDir := path.Join(videoDirPath, fmt.Sprintf("chunks-%dx%d-%s", opt.Width, opt.Height, opt.VideoBitRate))

		err = encodeVideoToResolution(videoPath, encodedVideoPath, &opt)

		if err != nil {
			return err
		}

		err = os.Mkdir(chunkDir, os.ModePerm)

		if err != nil {
			log.Error("unable to create directory %s", chunkDir)

			return err
		}

		dashPath := path.Join(chunkDir, "output.mpd")

		err = encodeVideoToDash(encodedVideoPath, dashPath, &opt)

		if err != nil {
			return err
		}

		uploadPrefix := path.Join(data.UploadId, fmt.Sprintf("chunks-%dx%d-%s", opt.Width, opt.Height, opt.VideoBitRate))

		err = uploadChunksToS3(uploadPrefix, chunkDir)

		if err != nil {
			return err
		}

		log.Info("Processed chunks for %s", chunkDir)
	}

	log.Info("Video encoding completed")

	//publish message to video catalog service

	return nil
}

func encodeVideoToResolution(inPath string, outPath string, opt *encode.VideoEncodeOption) error {
	err := encode.EncodeVideoToResolution(inPath, outPath, &encode.EncodeVideoToResolutionArgs{
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

func encodeVideoToDash(inPath string, outPath string, opt *encode.VideoEncodeOption) error {
	err := encode.EncodeVideoToDash(inPath, outPath, &encode.EncodeVideoToDashArgs{
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
		log.Info("Unable to read chunks dir %s", chunkDir)

		return err
	}

	for _, f := range files {
		log.Info("File Name %s", f.Name())

		filePath := path.Join(chunkDir, f.Name())
		uploadId := path.Join(uploadPathPrefix, f.Name())

		aws.PutFileToS3(filePath, uploadId)
	}

	return nil
}
