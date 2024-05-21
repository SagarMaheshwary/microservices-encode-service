package encode

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

type EncodeVideoArgs struct {
	Codec                string
	KeyFramesIntervalMin int
	GroupOfPictures      int
	TileColumns          int
	FrameParallel        int
	Format               string
	DisableAudio         bool
	Resolution           string
	Bitrate              string
	SegmentTime          string
	MOVFlags             string
}

type EncodeAudioArgs struct {
	Codec   string
	Bitrate string
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
		VideoCodec:   "libvpx-vp9",
		VideoBitRate: "1000k",
		AudioCodec:   "vp9",
		AudioBitRate: "192k",
		SegmentTime:  3,
		Format:       "webm",
	},
	{
		Width:        1280,
		Height:       720,
		VideoCodec:   "libvpx-vp9",
		VideoBitRate: "750k",
		AudioCodec:   "libvorbis",
		AudioBitRate: "160k",
		SegmentTime:  5,
		Format:       "webm",
	},
	{
		Width:        854,
		Height:       480,
		VideoCodec:   "libvpx-vp9",
		VideoBitRate: "500k",
		AudioCodec:   "libvorbis",
		AudioBitRate: "128k",
		SegmentTime:  7,
		Format:       "webm",
	},
	{
		Width:        640,
		Height:       360,
		VideoCodec:   "libvpx-vp9",
		VideoBitRate: "250k",
		AudioCodec:   "libvorbis",
		AudioBitRate: "96k",
		SegmentTime:  10,
		Format:       "webm",
	},
	{
		Width:        320,
		Height:       180,
		VideoCodec:   "libvpx-vp9",
		VideoBitRate: "150k",
		AudioCodec:   "libvorbis",
		AudioBitRate: "48k",
		SegmentTime:  10,
		Format:       "webm",
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

func CreateVideoChunks(inPath string, outPath string, args *EncodeVideoArgs) error {
	outArgs := ffmpeglib.KwArgs{
		"c:v":              args.Codec,
		"keyint_min":       args.KeyFramesIntervalMin,
		"g":                args.GroupOfPictures,
		"tile-columns":     args.TileColumns,
		"frame-parallel":   args.FrameParallel,
		"f":                args.Format,
		"vf":               fmt.Sprintf("scale=%s", args.Resolution),
		"b:v":              args.Bitrate,
		"dash":             1,
		"segment_time":     args.SegmentTime,
		"movflags":         args.MOVFlags,
		"reset_timestamps": 1,
	}

	if args.DisableAudio {
		outArgs["an"] = ""
	}

	err := ffmpeglib.Input(inPath).Output(outPath, outArgs).OverWriteOutput().ErrorToStdOut().Run()

	if err != nil {
		log.Error("FFMPEG video encode failed %v", err)
	}

	return err
}

func EncodeAudio(inPath string, outPath string, args *EncodeAudioArgs) {
	outArgs := ffmpeglib.KwArgs{
		"acodec": args.Codec,
		"ab":     args.Bitrate,
		"dash":   1,
	}

	ffmpeglib.Input(inPath).Output(outPath, outArgs).ErrorToStdOut().Run()
}

func GetVideoInfo(in string) *VideoInfo {
	args := []string{"-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height,duration,codec_name,bit_rate", "-of", "json", in}
	out, err := exec.Command("ffprobe", args...).Output()

	if err != nil {
		log.Error("Command failed %v", err)
	}

	log.Info("Resolution Raw: %s", out)

	type FileOutput struct {
		Programs []any       `json:"programs"`
		Streams  []VideoInfo `json:"streams"`
	}

	m := new(FileOutput)

	json.Unmarshal([]byte(out), &m)

	log.Info("Resolution: %v", m)

	return &m.Streams[0]
}

func GetVideoEncodeOptionsIndex(width int, height int) int {
	for i, v := range VideoEncodeOptions {
		if v.Width <= width && v.Height <= height {
			return i
		}
	}

	return 0
}
