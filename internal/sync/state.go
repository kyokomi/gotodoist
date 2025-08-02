package sync

import (
	"fmt"
	"time"
)

// Status は同期状態を表す
type Status struct {
	InitialSyncDone bool      `json:"initial_sync_done"`
	LastSyncTime    time.Time `json:"last_sync_time"`
	SyncToken       string    `json:"sync_token"`
}

// String は同期状態を文字列として表現する
func (s *Status) String() string {
	status := "❌ Not initialized"
	if s.InitialSyncDone {
		if s.LastSyncTime.IsZero() {
			status = "✅ Initialized (never synced)"
		} else {
			status = fmt.Sprintf("✅ Last sync: %s", s.LastSyncTime.Format("2006-01-02 15:04:05"))
		}
	}

	// sync_tokenの表示を安全に処理
	tokenDisplay := s.SyncToken
	if len(tokenDisplay) > 8 {
		tokenDisplay = tokenDisplay[:8] + "..."
	} else if tokenDisplay == "" {
		tokenDisplay = "none"
	}

	return fmt.Sprintf("Sync Status: %s (token: %s)", status, tokenDisplay)
}
