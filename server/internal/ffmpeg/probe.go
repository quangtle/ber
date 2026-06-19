package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/anomalyco/ber/internal/library"
)

type ffprobeOutput struct {
	Format struct {
		Filename string `json:"filename"`
		Duration string `json:"duration"`
		Size     string `json:"size"`
		BitRate  string `json:"bit_rate"`
		FormatName string `json:"format_name"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func Probe(path string) (*library.Video, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	v := &library.Video{
		FilePath:  path,
		Container: probe.Format.FormatName,
	}

	title := strings.TrimSuffix(path, "."+probe.Format.FormatName)
	if idx := strings.LastIndex(title, "/"); idx >= 0 {
		title = title[idx+1:]
	}
	v.Title = title

	if d, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		v.Duration = d
	}

	if s, err := strconv.ParseInt(probe.Format.Size, 10, 64); err == nil {
		v.FileSize = s
	}

	if b, err := strconv.Atoi(probe.Format.BitRate); err == nil {
		v.Bitrate = b
	}

	for _, s := range probe.Streams {
		if s.CodecType == "video" {
			v.Codec = s.CodecName
			v.Width = s.Width
			v.Height = s.Height
			break
		}
	}

	return v, nil
}
