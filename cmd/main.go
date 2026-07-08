package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"

	"github.com/tiufiakov/mediaCoder"
)

func main() {

	uuid := "7955bb19-b7cf-47d5-a7c2-33a997b34b72"
	inputPath := "/home/ilia/GolandProjects/mediaCoderProd/7.png"
	outputPath := "/home/ilia/GolandProjects/mediaCoderProd/6.png"

	mc := mediaCoder.MediaCoder{}

	mc.EmbedUUID(inputPath, outputPath, uuid)
	extractedUUID := mc.ExtractUUID(outputPath)
	fmt.Printf("🔍 %s\n", extractedUUID)

	if extractedUUID == uuid {
		fmt.Println("✅ UUID успешно извлечён!")
	} else {
		fmt.Println("❌ UUID не совпадает!")
	}
}
