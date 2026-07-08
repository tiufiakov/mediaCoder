package uuidUtil

import (
	"fmt"
	"log"
	"strings"
)

func UUIDToBytes(uuid string) []byte {
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

func BytesToUUID(uuidBytes []byte) string {
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uuidBytes[0], uuidBytes[1], uuidBytes[2], uuidBytes[3],
		uuidBytes[4], uuidBytes[5],
		uuidBytes[6], uuidBytes[7],
		uuidBytes[8], uuidBytes[9],
		uuidBytes[10], uuidBytes[11], uuidBytes[12], uuidBytes[13], uuidBytes[14], uuidBytes[15])
}
