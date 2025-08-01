package sync

import (
	"context"
	"log"
	"sync"
	"time"
)

// BackgroundSyncer はバックグラウンド同期を管理する
type BackgroundSyncer struct {
	manager  *Manager
	interval time.Duration
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// NewBackgroundSyncer は新しいBackgroundSyncerを作成する
func NewBackgroundSyncer(manager *Manager, interval time.Duration) *BackgroundSyncer {
	return &BackgroundSyncer{
		manager:  manager,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start はバックグラウンド同期を開始する
func (bs *BackgroundSyncer) Start(ctx context.Context) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.running {
		return // 既に実行中
	}

	bs.running = true
	bs.wg.Add(1)

	go bs.syncLoop(ctx)
}

// Stop はバックグラウンド同期を停止する
func (bs *BackgroundSyncer) Stop() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if !bs.running {
		return // 実行中でない
	}

	bs.running = false
	close(bs.stopCh)
	bs.wg.Wait()
}

// IsRunning はバックグラウンド同期が実行中かどうかを返す
func (bs *BackgroundSyncer) IsRunning() bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.running
}

// syncLoop はバックグラウンド同期のメインループ
func (bs *BackgroundSyncer) syncLoop(ctx context.Context) {
	defer bs.wg.Done()

	ticker := time.NewTicker(bs.interval)
	defer ticker.Stop()

	// 最初に一度同期を試行（起動時同期）
	bs.performSync(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Background sync stopped due to context cancellation")
			return
		case <-bs.stopCh:
			log.Println("Background sync stopped")
			return
		case <-ticker.C:
			bs.performSync(ctx)
		}
	}
}

// performSync は実際の同期処理を実行する
func (bs *BackgroundSyncer) performSync(ctx context.Context) {
	// タイムアウト付きコンテキストを作成（同期処理が長時間かかりすぎないように）
	syncCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := bs.manager.AutoSync(syncCtx, bs.interval); err != nil {
		log.Printf("Background sync failed: %v", err)
		// エラーが発生してもループは継続
	}
}

// TriggerSync は手動で同期をトリガーする（非同期）
func (bs *BackgroundSyncer) TriggerSync(ctx context.Context) {
	go func() {
		syncCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := bs.manager.IncrementalSync(syncCtx); err != nil {
			log.Printf("Manual sync failed: %v", err)
		}
	}()
}

// UpdateInterval は同期間隔を更新する
func (bs *BackgroundSyncer) UpdateInterval(newInterval time.Duration) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.interval = newInterval

	// 実行中の場合は再起動が必要
	// 実装を簡単にするため、ここでは新しい間隔は次回起動時に適用される
}

// GetInterval は現在の同期間隔を返す
func (bs *BackgroundSyncer) GetInterval() time.Duration {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.interval
}
