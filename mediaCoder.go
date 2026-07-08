package mediaCoder

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/tiufiakov/mediaCoder/internal/uuidUtil"
)

type MediaCoder struct{}

func (*MediaCoder) EmbedUUID(inputPath, outputPath, uuid string) {
	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	pixels := bounds.Dx() * bounds.Dy()
	if pixels < 43 {
		log.Fatal("Изображение слишком маленькое!")
	}
	fmt.Printf("Размер: %dx%d = %d пикселей\n", bounds.Dx(), bounds.Dy(), pixels)

	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	uuidBytesOrig := uuidUtil.UUIDToBytes(uuid)
	uuidBytes := make([]byte, 17) // padding
	copy(uuidBytes, uuidBytesOrig)

	bitIndex := 0

outer:
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := rgba.At(x, y).RGBA() // a = альфа оригинала
			r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

			// R всегда
			bitPos := byte(bitIndex % 8)
			if bitIndex < 128 {
				bit := (uuidBytes[bitIndex/8] >> uint(bitPos)) & 1 // embed
				r8 = (r8 & 0xFE) | uint8(bit)
			}
			bitIndex++

			// G всегда
			bitPos = byte(bitIndex % 8)
			if bitIndex < 128 {
				bit := (uuidBytes[bitIndex/8] >> uint(bitPos)) & 1
				g8 = (g8 & 0xFE) | uint8(bit)
			}
			bitIndex++

			// B всегда
			bitPos = byte(bitIndex % 8)
			if bitIndex < 128 {
				bit := (uuidBytes[bitIndex/8] >> uint(bitPos)) & 1
				b8 = (b8 & 0xFE) | uint8(bit)
			}
			bitIndex++

			rgba.Set(x, y, color.RGBA{r8, g8, b8, a8})

			if bitIndex >= 130 {
				break outer
			}
		}
	}
	fmt.Printf("✅ Внедрено %d битов (цель: 128)\n", bitIndex)

	// Аналогично в extractUUID для извлечения
	fmt.Printf("✅ Внедрено %d битов\n", bitIndex)

	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	png.Encode(outFile, rgba)
}

func (*MediaCoder) ExtractUUID(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	uuidBytes := [17]byte{}

	bitIndex := 0
outer:
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			r, g, b, _ := rgba.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			bitPos := byte(bitIndex % 8)
			if bitIndex < 128 {
				bit := r8 & 1
				uuidBytes[bitIndex/8] |= byte(bit) << uint(bitPos)
			}
			bitIndex++

			// G
			bitPos = byte(bitIndex % 8)
			if bitIndex < 128 {
				bit := g8 & 1
				uuidBytes[bitIndex/8] |= byte(bit) << uint(bitPos)
			}
			bitIndex++

			// B
			bitPos = byte(bitIndex % 8)
			if bitIndex < 128 {
				bit := b8 & 1
				uuidBytes[bitIndex/8] |= byte(bit) << uint(bitPos)
			}
			bitIndex++

			if bitIndex >= 130 {
				break outer
			}
		}

	}
	fmt.Printf("Извлечено %d битов\n", bitIndex)

	return uuidUtil.BytesToUUID(uuidBytes[:16])
}
