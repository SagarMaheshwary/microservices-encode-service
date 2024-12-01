package video_encoder

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
	ffmpeglib "github.com/u2takey/ffmpeg-go"
)

type EncodeVideoToResolutionArgs struct {
	VideoCodec   string
	AudioCodec   string
	Resolution   string
	AudioBitRate string
	VideoBitRate string
}

type EncodeVideoToDashArgs struct {
	Copy            string
	SegmentDuration int
	UseTimeline     int
	UseTemplate     int
}

type VideoInfo struct {
	CodecName string `json:"codec_name"`
	BitRate   string `json:"bit_rate"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	Duration  string `json:"duration"`
}

type VideoEncodeOption struct {
	Width        int
	Height       int
	VideoCodec   string
	AudioCodec   string
	VideoBitRate string
	AudioBitRate string
	SegmentTime  int
	Format       string
}

var VideoEncodeOptions = []VideoEncodeOption{
	{
		Width:        1920,
		Height:       1080,
		VideoCodec:   "libx264",
		VideoBitRate: "1000k",
		AudioCodec:   "aac",
		AudioBitRate: "192k",
		SegmentTime:  3,
		Format:       "mp4",
	},
	{
		Width:        1280,
		Height:       720,
		VideoCodec:   "libx264",
		VideoBitRate: "750k",
		AudioCodec:   "aac",
		AudioBitRate: "160k",
		SegmentTime:  5,
		Format:       "mp4",
	},
	{
		Width:        854,
		Height:       480,
		VideoCodec:   "libx264",
		VideoBitRate: "500k",
		AudioCodec:   "aac",
		AudioBitRate: "128k",
		SegmentTime:  7,
		Format:       "mp4",
	},
	{
		Width:        640,
		Height:       360,
		VideoCodec:   "libx264",
		VideoBitRate: "250k",
		AudioCodec:   "aac",
		AudioBitRate: "96k",
		SegmentTime:  10,
		Format:       "mp4",
	},
	{
		Width:        320,
		Height:       180,
		VideoCodec:   "libx264",
		VideoBitRate: "150k",
		AudioCodec:   "aac",
		AudioBitRate: "48k",
		SegmentTime:  10,
		Format:       "mp4",
	},
}

func EncodeVideoToResolution(inPath string, outPath string, args *EncodeVideoToResolutionArgs) error {
	outArgs := ffmpeglib.KwArgs{
		"c:v": args.VideoCodec,
		"vf":  fmt.Sprintf("scale=%s", args.Resolution),
		"b:a": args.AudioBitRate,
		"b:v": args.VideoBitRate,
		"c:a": args.AudioCodec,
	}

	err := ffmpeglib.Input(inPath).Output(outPath, outArgs).OverWriteOutput().ErrorToStdOut().Run()

	if err != nil {
		log.Error("FFMPEG encode video to resolution failed %v", err)

		return err
	}

	return nil
}

func EncodeVideoToDash(inPath string, outPath string, args *EncodeVideoToDashArgs) error {
	outArgs := ffmpeglib.KwArgs{
		"c":            args.Copy,
		"f":            "dash",
		"seg_duration": args.SegmentDuration,
		"use_timeline": args.UseTimeline,
		"use_template": args.UseTemplate,
	}

	err := ffmpeglib.Input(inPath).Output(outPath, outArgs).OverWriteOutput().ErrorToStdOut().Run()

	if err != nil {
		log.Error("FFMPEG encode video to dash failed %v", err)
	}

	return nil
}

func GetVideoInfo(inPath string) (*VideoInfo, error) {
	args := []string{
		"-v",
		"error",
		"-select_streams",
		"v:0",
		"-show_entries",
		"stream=width,height,duration,codec_name,bit_rate",
		"-of",
		"json",
		inPath,
	}

	out, err := exec.Command("ffprobe", args...).Output()

	if err != nil {
		log.Error("Command failed %v", err)

		return nil, err
	}

	type FileOutput struct {
		Programs []any       `json:"programs"`
		Streams  []VideoInfo `json:"streams"`
	}

	m := new(FileOutput)

	json.Unmarshal([]byte(out), &m)

	log.Info("Video %q Info: %v", inPath, m)

	return &m.Streams[0], nil
}

func GetEncodingStartIndex(width int, height int) int {
	for i, v := range VideoEncodeOptions {
		if v.Width <= width && v.Height <= height {
			return i
		}
	}

	return 0
}
