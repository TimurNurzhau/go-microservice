package utils

import (
	"log"
	"time"
)

// LogAction - асинхронное логирование действий (goroutine)
func LogAction(action string, userID int) {
	go func() {
		// Имитация задержки записи в лог/БД
		time.Sleep(10 * time.Millisecond)
		log.Printf("[AUDIT] Action: %s, UserID: %d, Time: %s", action, userID, time.Now().Format(time.RFC3339))
	}()
}