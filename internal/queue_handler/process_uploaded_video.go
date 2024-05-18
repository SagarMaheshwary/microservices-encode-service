package queue_handler

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/helper"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/aws"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/ffmpeg"
	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
)

type ProcessUploadedVideoPayload struct {
	UploadId    string `json:"upload_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"`
}

// @TODO: If the uploaded video doesn't fit to any resolution scale then choose a lower resolution
var Resolutions = map[string][]map[string]string{
	"1920x1080": {
		{"resolution": "1920x1080", "bitrate": "750k", "segment_time": "3"},
		{"resolution": "1280x720", "bitrate": "500k", "segment_time": "5"},
		{"resolution": "854x480", "bitrate": "250k", "segment_time": "7"},
		{"resolution": "640x360", "bitrate": "100k", "segment_time": "10"},
		{"resolution": "320x180", "bitrate": "50k", "segment_time": "10"},
	},
	"1280x720": {
		{"resolution": "1280x720", "bitrate": "500k", "segment_time": "5"},
		{"resolution": "854x480", "bitrate": "250k", "segment_time": "7"},
		{"resolution": "640x360", "bitrate": "100k", "segment_time": "10"},
		{"resolution": "320x180", "bitrate": "50k", "segment_time": "10"},
	},
	"854x480": {
		// {"resolution": "854x480", "bitrate": "250k", "segment_time": "20"},
		// {"resolution": "640x360", "bitrate": "100k", "segment_time": "20"},
		{"resolution": "320x180", "bitrate": "50k", "segment_time": "20"},
	},
	"640x360": {
		{"resolution": "640x360", "bitrate": "100k", "segment_time": "10"},
		{"resolution": "320x180", "bitrate": "50k", "segment_time": "10"},
	},
	"320x180": {
		{"resolution": "320x180", "bitrate": "50k", "segment_time": "10"},
	},
}

func HandleProcessUploadedVideo(data *ProcessUploadedVideoPayload) {
	log.Info("Data To Encode %v", data)

	var err error
	videoDirPath := path.Join(helper.RootDir(), "assets", "videos", data.UploadId)

	err = os.Mkdir(videoDirPath, os.ModePerm)

	if err != nil {
		log.Error("Unable to create directory for video %s", videoDirPath)

		return
	}

	videoPath := path.Join(videoDirPath, "original")

	err = createFileFromS3(data.UploadId, videoPath)

	if err != nil {
		return
	}

	videoRes := ffmpeg.GetResolution(videoPath)

	resolutions, ok := Resolutions[videoRes]

	if !ok {
		log.Info("Video resolution does not match any supported resolutions.")

		return
	}

	for _, res := range resolutions {
		chunkDir := path.Join(videoDirPath, fmt.Sprintf("chunks-%s-%s", res["resolution"], res["bitrate"]))

		err = os.Mkdir(chunkDir, os.ModePerm)

		if err != nil {
			log.Error("Unable to create chunks directory for %s resolution", res["resolution"])

			continue
		}

		chunkPrefix := fmt.Sprintf("chunk-%s-%s", res["resolution"], res["bitrate"])
		chunkFile := path.Join(chunkDir, strings.Join([]string{chunkPrefix, "%03d.webm"}, "-"))

		ffmpeg.EncodeVideo(videoPath, chunkFile, &ffmpeg.EncodeVideoArgs{
			Codec:                "libvpx-vp9",
			Format:               "segment",
			KeyFramesIntervalMin: 150,
			GroupOfPictures:      150,
			TileColumns:          4,
			FrameParallel:        1,
			Resolution:           strings.Replace(res["resolution"], "x", ":", 1),
			DisableAudio:         true,
			Bitrate:              res["bitrate"],
			SegmentTime:          res["segment_time"],
			MOVFlags:             "+faststart",
		})

		files, err := os.ReadDir(chunkDir)

		if err != nil {
			log.Info("Unable to read chunks dir %s", chunkDir)

			return
		}

		for _, f := range files {
			log.Info("File Name %s", f.Name())

			aws.PutFileToS3(path.Join(chunkDir, f.Name()), path.Join(data.UploadId, f.Name()))
		}

		log.Info("Processed chunks for %s resolution", res["resolution"])
	}

	log.Info("ENCODING FINISHED")

	// return

	// pathPrefix := path.Join(helper.RootDir(), "..", "..")
	// uploadPath := path.Join(videoDirPath, "videos-temp", data.UploadId)
	// chunkDir := path.Join(videoDirPath, "videos", data.UploadId)

	//read video resolution
	//create chunks for each resolution

	//FILE CREATED ----

	//create dir for chunks
	// err = os.Mkdir(chunkDir, os.FileMode(fs.ModePerm))

	// if err != nil {
	// 	log.Error("Create dir failed %v", err)

	// 	return
	// }

	//Encode file into multiple formats: 480p, 720p, 1080p

	// in := uploadPath
	// out := path.Join(videoDirPath, data.UploadId, "video_640x360_250k_%03d.webm")

	// entries, err := os.ReadDir(path.Join(videoDirPath, data.UploadId))

	// if err != nil {
	// 	log.Error("Failed to read directory %v", err)
	// }

	// for _, e := range entries {
	// 	log.Info("NAME: %s", e.Name())
	// }

	//Encode videos into chunks
	//List directory where chunks are stored
	//Upload these files to s3
	//Publish a message via rabbitmq

	//Create chunks of each video
	//Upload them to S3
	//Publish video metadata to a queue (for video catalog service)
}

func createFileFromS3(key string, videoPath string) error {
	res, err := aws.GetFileFromS3(key)

	log.Info("S3 file read %v", res.ContentType)

	if err != nil {
		return err
	}

	s3objectBytes, err := io.ReadAll(res.Body)

	log.Info("Reading bytes")

	if err != nil {
		log.Error("Unable to read file bytes %v", err)
		return err
	}

	f, err := os.Create(videoPath)

	log.Info("File created %v", f)

	if err != nil {
		log.Error("Unable to create file %v", err)
		return err
	}

	_, err = f.Write(s3objectBytes)

	if err != nil {
		log.Error("Unable to write to file %v", err)
		return err
	}

	return nil
}
