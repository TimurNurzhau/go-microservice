package utils

import (
	"log"
	"time"
)

// LogAction — асинхронное логирование действий (goroutine).
// Без искусственной задержки: логирование не влияет на latency ответа.
func LogAction(action string, userID int) {
	go func() {
		log.Printf("[AUDIT] Action: %s, UserID: %d, Time: %s",
			action, userID, time.Now().Format(time.RFC3339))
	}()
}
