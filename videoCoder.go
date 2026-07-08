package mediaCoder

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

const blockSize = 8
const videoStrength = 20

type VideoCoder struct {
}

func New(
	img *MediaCoder,
) *VideoCoder {

	return &VideoCoder{}
}

func (v *VideoCoder) EmbedUUID(
	input string,
	output string,
	id string,
) error {

	tmp := "frames"

	err := os.MkdirAll(tmp, 0755)
	if err != nil {
		return err
	}

	// достаём все кадры

	err = ffmpeg_go.Input(input).
		Output(
			fmt.Sprintf("%s/frame-%%05d.png", tmp),
			ffmpeg_go.KwArgs{},
		).
		Run()

	if err != nil {
		return err
	}

	// меняем первые 3 кадра

	for i := 1; i <= 3; i++ {

		frame := fmt.Sprintf(
			"%s/frame-%05d.png",
			tmp,
			i,
		)

		err := v.embedFrame(
			frame,
			id,
		)

		if err != nil {
			return err
		}
	}

	// собираем видео

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
) (string, error) {

	tmpDir, err := os.MkdirTemp(
		"",
		"frames",
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

	votes := map[string]int{}

	for i := 1; i <= 3; i++ {

		frame := filepath.Join(
			tmpDir,
			fmt.Sprintf(
				"frame-%05d.png",
				i,
			),
		)

		file, err := os.Open(frame)

		if err != nil {
			continue
		}

		img, _, err := image.Decode(file)
		file.Close()

		if err != nil {
			continue
		}

		id := v.ExtractVideoUUID(img)

		if id != "" {
			votes[id]++
		}
	}

	var result string
	max := 0

	for id, count := range votes {

		if count > max {
			result = id
			max = count
		}
	}

	if max < 2 {
		return "", fmt.Errorf(
			"uuid not reliable",
		)
	}

	return result, nil
}

func (v *VideoCoder) embedFrame(
	path string,
	id string,
) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(file)
	file.Close()

	if err != nil {
		return err
	}

	result := v.EmbedVideoUUID(
		img,
		id,
	)

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	defer out.Close()

	return png.Encode(out, result)
}

func extractFrames(
	video string,
	dir string,
) error {

	output := filepath.Join(
		dir,
		"frame-%05d.png",
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

func (v *VideoCoder) EmbedVideoUUID(
	img image.Image,
	id string,
) image.Image {

	u := uuid.MustParse(id)
	bits := uuidToBits(u)

	bounds := img.Bounds()

	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			out.Set(x, y, img.At(x, y))
		}
	}

	index := 0

	for y := bounds.Min.Y; y+blockSize < bounds.Max.Y && index < 128; y += blockSize {

		for x := bounds.Min.X; x+blockSize < bounds.Max.X && index < 128; x += blockSize {

			for by := 0; by < blockSize; by++ {
				for bx := 0; bx < blockSize; bx++ {

					r, g, b, a := out.At(x+bx, y+by).RGBA()

					value := int(r >> 8)

					if bits[index] == 1 {
						value += videoStrength
					} else {
						value -= videoStrength
					}

					if value < 0 {
						value = 0
					}

					if value > 255 {
						value = 255
					}

					out.Set(
						x+bx,
						y+by,
						color.RGBA{
							uint8(value),
							uint8(g >> 8),
							uint8(b >> 8),
							uint8(a >> 8),
						},
					)
				}
			}

			index++
		}
	}

	return out
}

func (v *VideoCoder) ExtractVideoUUID(
	img image.Image,
) string {

	bounds := img.Bounds()

	bits := make([]byte, 0, 128)

	for y := 0; y < bounds.Dy() && len(bits) < 128; y += 4 {

		for x := 0; x < bounds.Dx() && len(bits) < 128; x += 4 {

			r, _, _, _ := img.At(x, y).RGBA()

			if r>>8 > 128 {
				bits = append(bits, 1)
			} else {
				bits = append(bits, 0)
			}
		}
	}

	return bitsToUUID(bits)
}

func bitsToUUID(bits []byte) string {
	if len(bits) < 128 {
		return ""
	}

	bytes := make([]byte, 16)

	for i := 0; i < 128; i++ {
		if bits[i] == 1 {
			bytes[i/8] |= 1 << (7 - uint(i%8))
		}
	}

	return uuid.UUID(bytes).String()
}

func uuidToBits(u uuid.UUID) []byte {
	bits := make([]byte, 0, 128)

	for _, b := range u {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>uint(i))&1)
		}
	}

	return bits
}
