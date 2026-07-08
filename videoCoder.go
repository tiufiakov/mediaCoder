package mediaCoder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type VideoCoder struct {
	imageCoder *MediaCoder
}

func New(
	img *MediaCoder,
) *VideoCoder {

	return &VideoCoder{
		imageCoder: img,
	}
}
func (v *VideoCoder) EmbedUUID(
	input string,
	output string,
	uuid string,
) error {

	tmp := "frames"

	os.MkdirAll(tmp, 0755)

	// 1. Извлекаем ВСЕ кадры

	err := ffmpeg_go.Input(input).
		Output(
			fmt.Sprintf("%s/frame-%%05d.png", tmp),
			ffmpeg_go.KwArgs{},
		).
		Run()

	if err != nil {
		return err
	}

	// 2. Меняем первые 3

	for i := 1; i <= 3; i++ {

		frame := fmt.Sprintf(
			"%s/frame-%05d.png",
			tmp,
			i,
		)

		v.imageCoder.EmbedUUID(
			frame,
			frame,
			uuid,
		)

	}

	// 3. Собираем видео

	return ffmpeg_go.Input(
		fmt.Sprintf(
			"%s/frame-%%05d.png",
			tmp,
		),
	).
		Output(
			output,
			ffmpeg_go.KwArgs{
				"c:v":     "libx264",
				"pix_fmt": "yuv420p",
				"r":       "30",
			},
		).
		Run()
}

func (v *VideoCoder) ExtractUUID(
	videoPath string,
	imageExtractor *MediaCoder,
) (string, error) {

	tmpDir, err := os.MkdirTemp(
		"",
		"mediacoder-frames",
	)

	if err != nil {
		return "", err
	}

	defer os.RemoveAll(tmpDir)

	err = extractFrames(
		videoPath,
		tmpDir,
	)

	if err != nil {
		return "", err
	}

	votes := make(map[string]int)

	for i := 1; i <= 3; i++ {

		frame := filepath.Join(
			tmpDir,
			fmt.Sprintf(
				"frame-%03d.png",
				i,
			),
		)

		if _, err := os.Stat(frame); err != nil {
			continue
		}

		id := imageExtractor.ExtractUUID(
			frame,
		)

		// пропускаем мусор
		if id == "" ||
			id == "00000000-0000-0000-0000-000000000000" {
			continue
		}

		votes[id]++
	}

	if len(votes) == 0 {
		return "", fmt.Errorf(
			"uuid not found",
		)
	}

	var result string
	max := 0

	for id, count := range votes {

		if count > max {
			result = id
			max = count
		}
	}

	// хотя бы 2 из 3 кадров должны совпасть
	if max < 2 {
		return "", fmt.Errorf(
			"uuid not reliable",
		)
	}

	return result, nil
}

func extractFrames(
	video string,
	dir string,
) error {

	output := filepath.Join(
		dir,
		"frame-%03d.png",
	)

	cmd := exec.Command(
		"ffmpeg",
		"-i",
		video,
		"-vf",
		"select=eq(n\\,0)+eq(n\\,1)+eq(n\\,2)",
		"-vsync",
		"0",
		output,
		"-y",
	)

	return cmd.Run()
}
