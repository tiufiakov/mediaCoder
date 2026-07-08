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

	err := os.MkdirAll(
		tmp,
		0755,
	)

	if err != nil {
		return err
	}

	// извлекаем первые 3 кадра

	err = ffmpeg_go.Input(input).
		Output(
			fmt.Sprintf("%s/frame-%%03d.png", tmp),
			ffmpeg_go.KwArgs{
				"vframes": 3,
			},
		).
		Run()

	if err != nil {
		return err
	}

	// изменяем кадры

	for i := 1; i <= 3; i++ {

		filename := fmt.Sprintf(
			"%s/frame-%03d.png",
			tmp,
			i,
		)

		v.imageCoder.EmbedUUID(
			filename,
			filename,
			uuid,
		)

	}

	// собираем видео обратно

	err = ffmpeg_go.Input(
		fmt.Sprintf(
			"%s/frame-%%03d.png",
			tmp,
		),
	).
		Output(
			output,
			ffmpeg_go.KwArgs{
				"c:v":     "libx264",
				"pix_fmt": "yuv420p",
			},
		).
		Run()

	return err
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

	// Извлекаем первые 3 кадра
	err = extractFrames(
		videoPath,
		tmpDir,
	)

	if err != nil {
		return "", err
	}

	var found []string

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

		findUUID := imageExtractor.ExtractUUID(
			frame,
		)

		found = append(
			found,
			findUUID,
		)
	}

	if len(found) == 0 {
		return "", fmt.Errorf(
			"uuid not found",
		)
	}

	// Проверяем совпадение
	first := found[0]

	for _, id := range found[1:] {

		if id != first {
			return "", fmt.Errorf(
				"uuid mismatch",
			)
		}
	}

	return first, nil
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
		"-frames:v",
		"3",
		output,
		"-y",
	)

	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}
