package encode

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/sagarmaheshwary/microservices-encode-service/internal/lib/log"
	ffmpeglib "github.com/u2takey/ffmpeg-go"
)

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
	Width       int
	Height      int
	BitRate     string
	SegmentTime string
}

var VideoEncodeOptions = []VideoEncodeOption{
	{
		Width:       1920,
		Height:      1080,
		BitRate:     "750k",
		SegmentTime: "3",
	},
	{
		Width:       1280,
		Height:      720,
		BitRate:     "500k",
		SegmentTime: "5",
	},
	{
		Width:       854,
		Height:      480,
		BitRate:     "250k",
		SegmentTime: "7",
	},
	{
		Width:       640,
		Height:      360,
		BitRate:     "100k",
		SegmentTime: "10",
	},
	{
		Width:       320,
		Height:      180,
		BitRate:     "50k",
		SegmentTime: "10",
	},
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
