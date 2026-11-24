package utils

import (
	"log"
	"time"
)

type auditEntry struct {
	action string
	userID int
	ts     time.Time
}

var auditCh = make(chan auditEntry, 10000)

func init() {
	go func() {
		for e := range auditCh {
			log.Printf("[AUDIT] Action: %s, UserID: %d, Time: %s",
				e.action, e.userID, e.ts.Format(time.RFC3339))
		}
	}()
}

func LogAction(action string, userID int) {
	select {
	case auditCh <- auditEntry{action, userID, time.Now()}:
	default:
	}
}