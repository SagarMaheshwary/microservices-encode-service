package ffmpeg

import (
	"fmt"
	"os/exec"
	"strings"

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

func EncodeVideo(inPath string, outPath string, args *EncodeVideoArgs) error {
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

func GetResolution(in string) string {
	args := []string{"-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", in}
	out, err := exec.Command("ffprobe", args...).Output()

	if err != nil {
		log.Error("Command failed %v", err)
	}

	log.Info("Resolution: %s", out)

	return strings.Replace(string(out), "\n", "", 1)
}

// ffmpeglib.Input(path.Join(helper.RootDir(), "..", "..", "test-video.MP4")).
// 	Output(
// 		path.Join(helper.RootDir(), "..", "..", "out-video.webm"),
// 		ffmpeglib.KwArgs{
// 			"c:v":            "libvpx-vp9", //codec lib
// 			"keyint_min":     150,
// 			"g":              150,
// 			"tile-columns":   4,
// 			"frame-parallel": 1,
// 			"f":              "webm", //format
// 			"dash":           1,
// 			"an":             "",
// 			"vf":             "scale=1280:720", //resolution
// 			"b:v":            "250k",           //video bitrate
// 		},
// 	).
// 	ErrorToStdOut().
// 	Run()
