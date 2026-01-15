// Package methods using for userside methods
// Методы для бекенда
package methods

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateSecureSession генерирует безопасную сессию заданной длины
func GenerateSecureSession(length int) string {
	randomPart := make([]byte, length)
	rand.Read(randomPart)

	now := time.Now()
	seconds := now.Unix()
	micros := now.UnixMicro() % 1000000

	return fmt.Sprintf("%x_%d_%06d",
		randomPart,
		seconds,
		micros)
}
