package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"strings"
)

func uuidToBytes(uuid string) []byte {
	uuidClean := strings.ReplaceAll(uuid, "-", "")
	bytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		_, err := fmt.Sscanf(uuidClean[i*2:i*2+2], "%02x", &bytes[i])
		if err != nil {
			log.Fatalf("Ошибка парсинга байта %d: %v", i, err)
		}
	}
	return bytes
}

func bytesToUUID(uuidBytes []byte) string {
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uuidBytes[0], uuidBytes[1], uuidBytes[2], uuidBytes[3],
		uuidBytes[4], uuidBytes[5],
		uuidBytes[6], uuidBytes[7],
		uuidBytes[8], uuidBytes[9],
		uuidBytes[10], uuidBytes[11], uuidBytes[12], uuidBytes[13], uuidBytes[14], uuidBytes[15])
}

func embedUUID(inputPath, outputPath, uuid string) {
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
	uuidBytesOrig := uuidToBytes(uuid)
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

func extractUUID(filePath string) string {
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

	return bytesToUUID(uuidBytes[:16])
}

func main() {
	uuid := "7955bb19-b7cf-47d5-a7c2-33a997b34b86"
	inputPath := "/home/ilia/GolandProjects/рабочий код для внедрения uuid в изображения/7.png"
	outputPath := "/home/ilia/GolandProjects/рабочий код для внедрения uuid в изображения/6.png"

	bytes := uuidToBytes(uuid)
	fmt.Printf("TEST uuidToBytes[15] = 0x%02x\n", bytes[15]) // Должно быть 0x86!
	fmt.Printf("TEST bytesToUUID = %s\n", bytesToUUID(bytes))
	embedUUID(inputPath, outputPath, uuid)
	extractedUUID := extractUUID(outputPath)
	fmt.Printf("🔍 %s\n", extractedUUID)

	if extractedUUID == uuid {
		fmt.Println("✅ UUID успешно извлечён!")
	} else {
		fmt.Println("❌ UUID не совпадает!")
	}
}
