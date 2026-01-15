// Package methods using for userside methods
// Методы для бекенда
package methods

import (
	"backend_golang/types"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

func SendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func SendJSONResponse(w http.ResponseWriter, resp types.ResponseForAuth, statusCode int) {
	SendJSON(w, resp, statusCode)
}

func SendJSONResponseGeneric(w http.ResponseWriter, resp types.Response, statusCode int) {
	SendJSON(w, resp, statusCode)
}

func SendJSONError(w http.ResponseWriter, message string, statusCode int) {
	SendJSONResponse(w, types.ResponseForAuth{
		Success: false,
		Error:   message,
	}, statusCode)
}

func SendJSONErrorResponse(w http.ResponseWriter, message string, errorCode string, statusCode int) {
	SendJSONResponseGeneric(w, types.Response{
		Success: false,
		Message: message,
		Error:   errorCode,
	}, statusCode)
}

// SendJSONSuccessResponse отправляет успешный ответ без дополнительных данных
func SendJSONSuccessResponse(w http.ResponseWriter, message string, statusCode int) {
	SendJSONResponseGeneric(w, types.Response{
		Success: true,
		Message: message,
	}, statusCode)
}

// SendJSONSuccessResponseWithData отправляет успешный ответ с дополнительными данными
func SendJSONSuccessResponseWithData(w http.ResponseWriter, message string, data interface{}, statusCode int) {
	SendJSONResponseGeneric(w, types.Response{
		Success: true,
		Message: message,
		Data:    data,
	}, statusCode)
}
