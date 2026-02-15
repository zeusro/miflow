package miaccount

import (
	"log"
	"math/rand"

	"github.com/zeusro/miflow/internal/config"
)

func httpDebug() bool {
	return config.Get().Debug
}

func logHttpReq(method, url string, reqBody []byte) {
	if !httpDebug() {
		return
	}
	log.Printf("[HTTP] %s %s", method, url)
	if len(reqBody) > 0 && len(reqBody) < 2048 {
		log.Printf("[HTTP] Request: %s", string(reqBody))
	} else if len(reqBody) >= 2048 {
		log.Printf("[HTTP] Request: %d bytes (truncated)", len(reqBody))
	}
}

func logHttpResp(status int, body []byte) {
	if !httpDebug() {
		return
	}
	log.Printf("[HTTP] Response: %d", status)
	if len(body) > 0 && len(body) < 4096 {
		log.Printf("[HTTP] Body: %s", string(body))
	} else if len(body) >= 4096 {
		log.Printf("[HTTP] Body: %d bytes (truncated) %s...", len(body), string(body[:200]))
	}
}

// RandString returns a random string of length n (for deviceId, requestId, etc.).
func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
