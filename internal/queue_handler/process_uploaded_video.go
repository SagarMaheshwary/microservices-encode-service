package queue_handler

import (
	"fmt"
	"os"
	"path"
	"strings"

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

		chunkDir := path.Join(videoDirPath, fmt.Sprintf("chunks-%dx%d-%s", opt.Width, opt.Height, opt.BitRate))

		err = createVideoChunksByResolution(videoPath, chunkDir, &opt)

		if err != nil {
			return err
		}

		err = uploadChunksToS3(data.UploadId, chunkDir)

		if err != nil {
			return err
		}

		log.Info("Processed chunks for %s", chunkDir)
	}

	log.Info("Video encoding completed")

	//publish message to video catalog service

	return nil
}

func createVideoChunksByResolution(videoPath string, chunkDir string, opt *encode.VideoEncodeOption) error {
	err := os.Mkdir(chunkDir, os.ModePerm)

	if err != nil {
		log.Error("Unable to create chunks directory for %dx%d resolution", opt.Width, opt.Height)

		return err
	}

	chunkPrefix := fmt.Sprintf("chunk-%dx%d-%s", opt.Width, opt.Height, opt.BitRate)
	chunkFile := path.Join(chunkDir, strings.Join([]string{chunkPrefix, "%03d.webm"}, "-"))

	encode.CreateVideoChunks(videoPath, chunkFile, &encode.EncodeVideoArgs{
		Codec:                "libvpx-vp9",
		Format:               "segment",
		KeyFramesIntervalMin: 150,
		GroupOfPictures:      150,
		TileColumns:          4,
		FrameParallel:        1,
		Resolution:           fmt.Sprintf("%d:%d", opt.Width, opt.Height),
		DisableAudio:         true,
		Bitrate:              opt.BitRate,
		SegmentTime:          opt.SegmentTime,
		MOVFlags:             "+faststart",
	})

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

		aws.PutFileToS3(path.Join(chunkDir, f.Name()), path.Join(uploadPathPrefix, f.Name()))
	}

	return nil
}
