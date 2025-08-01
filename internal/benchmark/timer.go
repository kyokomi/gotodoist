// Package benchmark はパフォーマンス測定用のユーティリティを提供する
package benchmark

import (
	"fmt"
	"time"
)

// Timer はパフォーマンス測定のためのタイマー
type Timer struct {
	enabled   bool
	startTime time.Time
	steps     []Step
}

// Step は個別の処理ステップの測定結果
type Step struct {
	Name     string
	Duration time.Duration
}

// NewTimer は新しいタイマーを作成する
func NewTimer(enabled bool) *Timer {
	return &Timer{
		enabled:   enabled,
		startTime: time.Now(),
		steps:     make([]Step, 0),
	}
}

// Step は処理ステップの時間を記録する
func (t *Timer) Step(name string) {
	if !t.enabled {
		return
	}

	now := time.Now()
	var duration time.Duration

	if len(t.steps) == 0 {
		// 最初のステップは開始からの時間
		duration = now.Sub(t.startTime)
	} else {
		// 前のステップからの時間
		lastStepEnd := t.startTime
		for _, s := range t.steps {
			lastStepEnd = lastStepEnd.Add(s.Duration)
		}
		duration = now.Sub(lastStepEnd)
	}

	t.steps = append(t.steps, Step{
		Name:     name,
		Duration: duration,
	})
}

// PrintResults はベンチマーク結果を出力する
func (t *Timer) PrintResults() {
	if !t.enabled {
		return
	}

	totalDuration := time.Since(t.startTime)

	fmt.Printf("🔍 Performance Benchmark:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for i, step := range t.steps {
		percentage := float64(step.Duration) / float64(totalDuration) * 100
		fmt.Printf("%2d. %-30s %8s (%5.1f%%)\n",
			i+1, step.Name, formatDuration(step.Duration), percentage)
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("    %-30s %8s (100.0%%)\n", "Total", formatDuration(totalDuration))
	fmt.Printf("\n")
}

// formatDuration は時間を読みやすい形式でフォーマットする
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d)/float64(time.Microsecond))
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond))
	} else {
		return fmt.Sprintf("%.2fs", float64(d)/float64(time.Second))
	}
}

// GetTotalDuration は総実行時間を返す
func (t *Timer) GetTotalDuration() time.Duration {
	return time.Since(t.startTime)
}

// FormatDuration は外部から使用可能な時間フォーマット関数
func FormatDuration(d time.Duration) string {
	return formatDuration(d)
}
