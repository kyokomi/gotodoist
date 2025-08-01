//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectLifecycle プロジェクトのライフサイクル全体をテストし、データ整合性を確認する
// このテストはローカルストレージとTodoist APIサーバー間のデータ同期の整合性を検証する
func TestProjectLifecycle(t *testing.T) {
	// 環境変数チェック
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		t.Skip("TODOIST_API_TOKEN環境変数が設定されていません")
	}

	// バイナリのビルド
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// テスト用の一時設定ディレクトリ
	tmpDir := t.TempDir()
	env := setupTestEnvironment(tmpDir, token)

	// ベースラインデータA: sync resetとsync initでクリーンな状態からサーバーデータを取得
	var baselineProjectsA, baselineTasksA string
	t.Run("ベースラインデータA取得", func(t *testing.T) {
		// ローカルストレージをリセット
		cmd := exec.Command(binaryPath, "sync", "reset", "-f")
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "sync reset失敗")
		t.Logf("sync reset完了: %s", strings.TrimSpace(string(output)))

		// 初期同期でサーバーデータを取得
		cmd = exec.Command(binaryPath, "sync", "init")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "sync init失敗")
		t.Logf("sync init完了: %s", strings.TrimSpace(string(output)))

		// ベースラインとなるプロジェクト一覧を取得
		cmd = exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		projectOutput, err := cmd.Output()
		require.NoError(t, err, "ベースラインプロジェクト一覧の取得に失敗")
		baselineProjectsA = string(projectOutput)

		projectCount := countProjectsFromOutput(baselineProjectsA)
		t.Logf("ベースライン時のプロジェクト数: %d", projectCount)

		// Todoistの無料プランはプロジェクト数に制限がある（通常5個）
		assert.Less(t, projectCount, 4, "プロジェクト数が制限に近づいています。テスト実行前にプロジェクトを削除してください。")

		// ベースラインとなるタスク一覧を取得
		cmd = exec.Command(binaryPath, "task", "list")
		cmd.Env = env
		taskOutput, err := cmd.Output()
		require.NoError(t, err, "ベースラインタスク一覧の取得に失敗")
		baselineTasksA = string(taskOutput)

		t.Logf("ベースラインデータA取得完了")
		t.Logf("- プロジェクト数: %d", projectCount)
		t.Logf("- タスク一覧文字数: %d", len(baselineTasksA))
	})

	// テスト用のプロジェクト名（タイムスタンプ付きでユニークにする）
	timestamp := time.Now().Format("20060102-150405")
	projectName := fmt.Sprintf("E2E-Test-Project-%s", timestamp)
	updatedProjectName := fmt.Sprintf("E2E-Test-Project-Updated-%s", timestamp)

	// ステップ1: プロジェクトを作成する
	t.Run("1. プロジェクトを作成", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "project", "add", projectName)
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクト作成に失敗")
		t.Logf("プロジェクト作成結果: %s", strings.TrimSpace(string(output)))
	})

	// ステップ2: プロジェクト一覧を取得して作成したプロジェクトが存在することを確認
	t.Run("2. プロジェクト一覧で存在確認", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクト一覧取得に失敗")

		outputStr := string(output)
		// より正確な検証: "📁 プロジェクト名" の形式で存在するかチェック
		expectedLine := fmt.Sprintf("📁 %s", projectName)
		assert.Contains(t, outputStr, expectedLine, "作成したプロジェクトが一覧に存在しません")
		t.Logf("✓ プロジェクト '%s' が一覧に存在することを確認", projectName)
	})

	// ステップ3: 作成したプロジェクトを更新する
	t.Run("3. プロジェクトを更新", func(t *testing.T) {
		// プロジェクト名を更新
		cmd := exec.Command(binaryPath, "project", "update", projectName, "--name", updatedProjectName)
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクト更新に失敗")
		t.Logf("プロジェクト更新結果: %s", strings.TrimSpace(string(output)))

		// 更新後の一覧確認
		cmd = exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "プロジェクト一覧取得に失敗")

		outputStr := string(output)
		// より正確な検証: "📁 プロジェクト名" の形式で存在するかチェック
		expectedLine := fmt.Sprintf("📁 %s", updatedProjectName)
		assert.Contains(t, outputStr, expectedLine, "更新したプロジェクトが一覧に存在しません")
		t.Logf("✓ プロジェクト更新後 '%s' が一覧に存在することを確認", updatedProjectName)
		// 以降のテストでは更新後の名前を使用
		projectName = updatedProjectName
		t.Logf("プロジェクト名を '%s' に更新", projectName)
	})

	// ステップ4: 作成したプロジェクトにタスクを3つ追加
	taskContents := []string{
		fmt.Sprintf("Task-1-%s", timestamp),
		fmt.Sprintf("Task-2-%s", timestamp),
		fmt.Sprintf("Task-3-%s", timestamp),
	}

	t.Run("4. プロジェクトにタスクを3つ追加", func(t *testing.T) {
		for i, taskContent := range taskContents {
			cmd := exec.Command(binaryPath, "task", "add", taskContent, "-p", projectName)
			cmd.Env = env
			output, err := cmd.Output()
			if err != nil {
				t.Errorf("タスク%d '%s' の作成に失敗: %v", i+1, taskContent, err)
				continue
			}
			t.Logf("タスク%d作成結果: %s", i+1, strings.TrimSpace(string(output)))
		}
	})

	// ステップ5: プロジェクトのタスク一覧を取得して3つのタスクが存在することを確認
	t.Run("5. プロジェクトのタスク一覧で3つのタスク存在確認", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "task", "list", "-p", projectName)
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクトのタスク一覧取得に失敗")

		outputStr := string(output)
		foundTasks := 0
		for _, taskContent := range taskContents {
			assert.Contains(t, outputStr, taskContent, "タスクが一覧に存在しません")
			t.Logf("✓ タスク '%s' が一覧に存在", taskContent)
			if strings.Contains(outputStr, taskContent) {
				foundTasks++
			}
		}

		// タスク数の確認
		assert.Equal(t, len(taskContents), foundTasks, "期待したタスク数と異なります")
	})

	// ステップ6: タスクを1つ更新する
	t.Run("6. タスクを1つ更新", func(t *testing.T) {
		updatedTaskContent := fmt.Sprintf("Updated-Task-1-%s", timestamp)

		// まずタスクIDを取得（更新後のプロジェクト名を使用）
		t.Logf("タスクID取得: プロジェクト名='%s', タスク内容='%s'", projectName, taskContents[0])
		taskID, err := findTaskIDByContent(binaryPath, env, projectName, taskContents[0])
		require.NoError(t, err, "更新対象タスクのID取得に失敗")

		// タスクの内容を更新
		cmd := exec.Command(binaryPath, "task", "update", taskID, "--content", updatedTaskContent)
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "タスク更新に失敗")
		t.Logf("タスク更新結果: %s", strings.TrimSpace(string(output)))

		// 更新後の確認
		cmd = exec.Command(binaryPath, "task", "list", "-p", projectName)
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "タスク一覧取得に失敗")

		outputStr := string(output)
		assert.Contains(t, outputStr, updatedTaskContent, "更新したタスクが一覧に存在しません")
		t.Logf("✓ タスク更新後 '%s' が一覧に存在することを確認", updatedTaskContent)
	})

	// ステップ6.5: タスクの完了と未完了のテスト
	t.Run("6.5. タスクの完了と未完了テスト", func(t *testing.T) {
		// 残っているタスクの中から1つ選択（3番目のタスク）
		testTaskContent := taskContents[2] // "Task-3-timestamp"

		// タスクIDを取得
		taskID, err := findTaskIDByContent(binaryPath, env, projectName, testTaskContent)
		require.NoError(t, err, "テスト対象タスクのID取得に失敗")
		t.Logf("テスト対象タスクID: %s, 内容: %s", taskID, testTaskContent)

		// サブテスト: タスクを完了にする
		t.Run("complete", func(t *testing.T) {
			cmd := exec.Command(binaryPath, "task", "complete", taskID)
			cmd.Env = env
			output, err := cmd.Output()
			require.NoError(t, err, "タスク完了に失敗")
			t.Logf("タスク完了結果: %s", strings.TrimSpace(string(output)))

			// 完了後の確認: 通常のタスク一覧では表示されないはず
			cmd = exec.Command(binaryPath, "task", "list", "-p", projectName)
			cmd.Env = env
			output, err = cmd.Output()
			require.NoError(t, err, "タスク一覧取得に失敗")

			outputStr := string(output)
			assert.NotContains(t, outputStr, testTaskContent, "完了したタスクが通常の一覧に表示されています")
			t.Logf("✓ 完了したタスク '%s' が通常の一覧から除外されていることを確認", testTaskContent)

			// --allフラグで確認: 完了済みタスクも表示されるはず
			cmd = exec.Command(binaryPath, "task", "list", "-p", projectName, "--all")
			cmd.Env = env
			output, err = cmd.Output()
			require.NoError(t, err, "全タスク一覧取得に失敗")

			outputStr = string(output)
			assert.Contains(t, outputStr, testTaskContent, "完了したタスクが--allフラグ付き一覧に表示されません")
			t.Logf("✓ 完了したタスク '%s' が--allフラグ付き一覧に表示されていることを確認", testTaskContent)
		})

		// サブテスト: タスクを未完了に戻す
		t.Run("uncomplete", func(t *testing.T) {
			cmd := exec.Command(binaryPath, "task", "uncomplete", taskID)
			cmd.Env = env
			output, err := cmd.Output()
			require.NoError(t, err, "タスク未完了に失敗")
			t.Logf("タスク未完了結果: %s", strings.TrimSpace(string(output)))

			// 未完了後の確認: 再び通常のタスク一覧に表示されるはず
			cmd = exec.Command(binaryPath, "task", "list", "-p", projectName)
			cmd.Env = env
			output, err = cmd.Output()
			require.NoError(t, err, "タスク一覧取得に失敗")

			outputStr := string(output)
			assert.Contains(t, outputStr, testTaskContent, "未完了に戻したタスクが一覧に表示されません")
			t.Logf("✓ 未完了に戻したタスク '%s' が通常の一覧に再表示されていることを確認", testTaskContent)
		})
	})

	// ステップ7: タスクを1つ削除する
	t.Run("7. タスクを1つ削除", func(t *testing.T) {
		taskToDelete := taskContents[1] // 2番目のタスクを削除

		// まずタスクIDを取得
		taskID, err := findTaskIDByContent(binaryPath, env, projectName, taskToDelete)
		require.NoError(t, err, "タスクID取得に失敗")
		t.Logf("削除対象タスクID: %s", taskID)

		// タスクIDで削除を実行
		cmd := exec.Command(binaryPath, "task", "delete", taskID, "-f")
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "タスク削除に失敗")
		t.Logf("タスク削除結果: %s", strings.TrimSpace(string(output)))

		// 削除後の確認
		cmd = exec.Command(binaryPath, "task", "list", "-p", projectName)
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "タスク一覧取得に失敗")

		outputStr := string(output)
		assert.NotContains(t, outputStr, taskToDelete, "削除したはずのタスクがまだ一覧に存在しています")
		t.Logf("✓ タスク削除後 '%s' が一覧から削除されていることを確認", taskToDelete)
	})

	// ステップ8: プロジェクトをアーカイブする
	t.Run("8. プロジェクトをアーカイブ", func(t *testing.T) {
		// プロジェクトアーカイブを実行
		cmd := exec.Command(binaryPath, "project", "archive", projectName)
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクトアーカイブに失敗")
		t.Logf("プロジェクトアーカイブ結果: %s", strings.TrimSpace(string(output)))

		// アーカイブ後の一覧確認（アクティブなプロジェクト一覧には表示されないはず）
		cmd = exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "アーカイブ後のプロジェクト一覧取得に失敗")

		outputStr := string(output)
		// より正確な検証: "📁 プロジェクト名" の形式で存在しないかチェック
		expectedLine := fmt.Sprintf("📁 %s", projectName)
		assert.NotContains(t, outputStr, expectedLine, "アーカイブしたプロジェクトがアクティブ一覧にまだ表示されています")
		t.Logf("✓ アーカイブ後 '%s' がアクティブ一覧から削除されていることを確認", projectName)
	})

	// ステップ9: プロジェクトをアンアーカイブする
	t.Run("9. プロジェクトをアンアーカイブ", func(t *testing.T) {
		// プロジェクトアンアーカイブを実行
		cmd := exec.Command(binaryPath, "project", "unarchive", projectName)
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクトアンアーカイブに失敗")
		t.Logf("プロジェクトアンアーカイブ結果: %s", strings.TrimSpace(string(output)))

		// アンアーカイブ後の一覧確認（再度アクティブ一覧に表示されるはず）
		cmd = exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "アンアーカイブ後のプロジェクト一覧取得に失敗")

		outputStr := string(output)
		// より正確な検証: "📁 プロジェクト名" の形式で存在するかチェック
		expectedLine := fmt.Sprintf("📁 %s", projectName)
		assert.Contains(t, outputStr, expectedLine, "アンアーカイブしたプロジェクトがアクティブ一覧に表示されません")
		t.Logf("✓ アンアーカイブ後 '%s' がアクティブ一覧に復活していることを確認", projectName)

		// タスクもアンアーカイブ後に再表示されるか確認
		cmd = exec.Command(binaryPath, "task", "list", "-p", projectName)
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "アンアーカイブ後のタスク一覧取得に失敗")

		outputStr = string(output)
		visibleTasks := 0
		for _, taskContent := range taskContents {
			if strings.Contains(outputStr, taskContent) {
				visibleTasks++
				t.Logf("✓ アンアーカイブ後にタスク '%s' が再表示されている", taskContent)
			}
		}

		// 削除されたタスクを除いて、残りのタスクが表示されているか確認
		expectedVisibleTasks := len(taskContents) - 1 // ステップ7で1つ削除されている想定
		assert.Equal(t, expectedVisibleTasks, visibleTasks, "アンアーカイブ後に期待されるタスク数と異なります")
		t.Logf("✓ アンアーカイブ後に期待されるタスク数 (%d個) が表示されています", expectedVisibleTasks)
	})

	// ステップ10: プロジェクトを削除する
	t.Run("10. プロジェクトを削除", func(t *testing.T) {
		// プロジェクトを削除
		cmd := exec.Command(binaryPath, "project", "delete", projectName, "-f")
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクト削除に失敗")
		t.Logf("プロジェクト削除結果: %s", strings.TrimSpace(string(output)))
	})

	// ステップ11: プロジェクト一覧を取得して削除したプロジェクトが存在しないことを確認
	t.Run("11. プロジェクト削除後の一覧確認", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "プロジェクト一覧取得に失敗")

		outputStr := string(output)
		// より正確な検証: "📁 プロジェクト名" の形式で存在しないかチェック
		expectedLine := fmt.Sprintf("📁 %s", projectName)
		assert.NotContains(t, outputStr, expectedLine, "削除したはずのプロジェクトがまだ一覧に存在しています")
		t.Logf("✓ プロジェクト削除後 '%s' が一覧から削除されていることを確認", projectName)
	})

	// ステップ12: 全タスク一覧を取得してプロジェクトと一緒にタスクが削除されていることを確認
	t.Run("12. 全タスク一覧でカスケード削除確認", func(t *testing.T) {
		// プロジェクト削除時に既にローカルストレージからタスクが削除され、
		// さらにIncrementalSyncも実行されているため、追加のsyncは不要
		cmd := exec.Command(binaryPath, "task", "list")
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "全タスク一覧取得に失敗")

		outputStr := string(output)
		for _, taskContent := range taskContents {
			assert.NotContains(t, outputStr, taskContent, "プロジェクト削除後もタスクが残っています")
		}
		t.Logf("✓ プロジェクト削除に伴いすべてのタスクが削除されていることを確認")
	})

	// データ整合性確認: 再度sync resetとsync initでサーバーデータを取得し、ベースラインAと一致するか確認
	t.Run("データ整合性確認", func(t *testing.T) {
		// ローカルストレージを再リセット
		cmd := exec.Command(binaryPath, "sync", "reset", "-f")
		cmd.Env = env
		output, err := cmd.Output()
		require.NoError(t, err, "最終sync reset失敗")
		t.Logf("最終sync reset完了: %s", strings.TrimSpace(string(output)))

		// 再度初期同期でサーバーデータを取得
		cmd = exec.Command(binaryPath, "sync", "init")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "最終sync init失敗")
		t.Logf("最終sync init完了: %s", strings.TrimSpace(string(output)))

		// 最終状態のプロジェクト一覧を取得
		cmd = exec.Command(binaryPath, "project", "list")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "最終プロジェクト一覧の取得に失敗")
		finalProjectsA := string(output)

		// 最終状態のタスク一覧を取得
		cmd = exec.Command(binaryPath, "task", "list")
		cmd.Env = env
		output, err = cmd.Output()
		require.NoError(t, err, "最終タスク一覧の取得に失敗")
		finalTasksA := string(output)

		// データ整合性の確認
		projectsMatch := compareDataConsistency(baselineProjectsA, finalProjectsA)
		tasksMatch := compareDataConsistency(baselineTasksA, finalTasksA)

		// 結果のレポート
		baselineProjectCount := countProjectsFromOutput(baselineProjectsA)
		finalProjectCount := countProjectsFromOutput(finalProjectsA)

		t.Logf("データ整合性確認結果:")
		t.Logf("- ベースラインプロジェクト数: %d", baselineProjectCount)
		t.Logf("- 最終プロジェクト数: %d", finalProjectCount)
		t.Logf("- プロジェクトデータ一致: %t", projectsMatch)
		t.Logf("- タスクデータ一致: %t", tasksMatch)

		assert.True(t, projectsMatch, "プロジェクトデータが一致しません")
		if projectsMatch {
			t.Logf("✅ プロジェクトデータが一致しています")
		} else {
			t.Logf("ベースライン:\n%s", baselineProjectsA)
			t.Logf("最終状態:\n%s", finalProjectsA)
		}

		assert.True(t, tasksMatch, "タスクデータが一致しません")
		if tasksMatch {
			t.Logf("✅ タスクデータが一致しています")
		} else {
			t.Logf("ベースライン:\n%s", baselineTasksA)
			t.Logf("最終状態:\n%s", finalTasksA)
		}

		if projectsMatch && tasksMatch {
			t.Logf("🎉 データ整合性確認完了: ローカルストレージとTodoistサーバーが完全に同期されています")
		}
	})
}

// buildBinary はテスト用のバイナリをビルドする
func buildBinary(t *testing.T) string {
	t.Helper()

	// 一時ファイル作成
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "gotodoist")

	// ビルド実行
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = ".." // e2eディレクトリから一つ上のディレクトリ

	err := cmd.Run()
	require.NoError(t, err, "バイナリのビルドに失敗")

	return binaryPath
}

// setupTestEnvironment テスト用の環境変数を設定する
func setupTestEnvironment(tmpDir, token string) []string {
	configDir := filepath.Join(tmpDir, ".config", "gotodoist")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		panic(fmt.Sprintf("設定ディレクトリの作成に失敗: %v", err))
	}

	return append(os.Environ(),
		"TODOIST_API_TOKEN="+token,
		"HOME="+tmpDir,
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, ".config"),
	)
}

// 以下は参考用：後で実装される可能性がある機能のためのテストヘルパー

// findProjectID プロジェクト名からIDを取得する（JSON出力が実装された場合用）
func findProjectID(binaryPath string, env []string, projectName string) (string, error) {
	cmd := exec.Command(binaryPath, "project", "list", "--json")
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var projects []map[string]interface{}
	if err := json.Unmarshal(output, &projects); err != nil {
		return "", err
	}

	for _, project := range projects {
		if project["name"] == projectName {
			return project["id"].(string), nil
		}
	}
	return "", fmt.Errorf("プロジェクト '%s' が見つかりません", projectName)
}

// findTaskID タスク内容からIDを取得する（JSON出力が実装された場合用）
func findTaskID(binaryPath string, env []string, taskContent string) (string, error) {
	cmd := exec.Command(binaryPath, "task", "list", "--json")
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var tasks []map[string]interface{}
	if err := json.Unmarshal(output, &tasks); err != nil {
		return "", err
	}

	for _, task := range tasks {
		if task["content"] == taskContent {
			return task["id"].(string), nil
		}
	}
	return "", fmt.Errorf("タスク '%s' が見つかりません", taskContent)
}

// countProjectsFromOutput はプロジェクト一覧出力からプロジェクト数をカウントする
func countProjectsFromOutput(output string) int {
	// "📁 📁 Projects (N):" の形式から数字を抽出
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Projects (") && strings.Contains(line, "):") {
			// "📁 📁 Projects (5):" から "5" を抽出
			start := strings.Index(line, "(") + 1
			end := strings.Index(line, ")")
			if start > 0 && end > start {
				countStr := line[start:end]
				if count, err := strconv.Atoi(countStr); err == nil {
					return count
				}
			}
		}
	}

	// "📁 No projects found" の場合は0を返す
	if strings.Contains(output, "No projects found") {
		return 0
	}

	// パースできない場合は安全のため大きな値を返す
	return 999
}

// findTaskIDByContent はタスク内容からIDを取得する（verbose出力から）
func findTaskIDByContent(binaryPath string, env []string, projectName, taskContent string) (string, error) {
	// verbose出力でタスク一覧を取得（IDが表示される）
	cmd := exec.Command(binaryPath, "task", "list", "-p", projectName, "-v", "-f", taskContent)
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get task list: %w", err)
	}

	// 出力をパースしてタスクIDを抽出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "ID: ") {
			currentTaskID := strings.ReplaceAll(line, "ID: ", "") // "   ID: " を除去
			return currentTaskID, nil
		}
	}

	return "", fmt.Errorf("task '%s' not found in project '%s'", taskContent, projectName)
}

// compareDataConsistency は2つのデータ出力文字列を比較してデータ整合性を確認する
// 時刻情報などの変動要素を除外した上で内容の一致を確認する
func compareDataConsistency(baseline, final string) bool {
	// 改行で正規化
	baselineNormalized := strings.TrimSpace(baseline)
	finalNormalized := strings.TrimSpace(final)

	// 簡単なケース: 完全一致
	if baselineNormalized == finalNormalized {
		return true
	}

	// より詳細な比較: 行ごとに分割して内容を比較
	baselineLines := strings.Split(baselineNormalized, "\n")
	finalLines := strings.Split(finalNormalized, "\n")

	// フィルタリング: タイムスタンプや動的要素を除外した有意な行のみを抽出
	baselineSignificant := filterSignificantLines(baselineLines)
	finalSignificant := filterSignificantLines(finalLines)

	// 有意な行の数が異なる場合は不一致
	if len(baselineSignificant) != len(finalSignificant) {
		return false
	}

	// 各行を比較
	for i, baselineLine := range baselineSignificant {
		if i >= len(finalSignificant) {
			return false
		}
		if baselineLine != finalSignificant[i] {
			return false
		}
	}

	return true
}

// filterSignificantLines はデータ比較において有意な行のみを抽出する
// タイムスタンプや一時的な情報を除外し、プロジェクト/タスクの実体的な内容のみを残す
func filterSignificantLines(lines []string) []string {
	var significant []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 空行をスキップ
		if trimmed == "" {
			continue
		}

		// 統計情報行をスキップ（"Projects (2):" など）
		if strings.Contains(trimmed, "Projects (") && strings.Contains(trimmed, "):") {
			continue
		}

		// Sync Status行をスキップ
		if strings.Contains(trimmed, "Sync Status:") {
			continue
		}

		// "Last sync:"で始まる行をスキップ
		if strings.Contains(trimmed, "Last sync:") {
			continue
		}

		// "No projects found"や"No tasks found"は有意
		if strings.Contains(trimmed, "No projects found") || strings.Contains(trimmed, "No tasks found") {
			significant = append(significant, trimmed)
			continue
		}

		// プロジェクト名の行（📁で始まる）
		if strings.HasPrefix(trimmed, "📁") && !strings.Contains(trimmed, "Projects (") {
			significant = append(significant, trimmed)
			continue
		}

		// タスク内容の行（⚪で始まる）
		if strings.HasPrefix(trimmed, "⚪") {
			significant = append(significant, trimmed)
			continue
		}

		// その他の有意そうな行（プロジェクト名やタスク内容が含まれている）
		// ただし、ID行やメタデータ行は除外
		if !strings.HasPrefix(trimmed, "ID:") &&
			!strings.HasPrefix(trimmed, "Project ID:") &&
			!strings.HasPrefix(trimmed, "Created:") &&
			!strings.HasPrefix(trimmed, "Updated:") {
			significant = append(significant, trimmed)
		}
	}

	return significant
}
