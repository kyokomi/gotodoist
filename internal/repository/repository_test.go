package repository

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindProjectIDByName_Logic はプロジェクト検索ロジックの単体テスト
// 実際のAPIやストレージを使わず、ロジックのみをテストする
func TestFindProjectIDByName_Logic(t *testing.T) {
	// テスト用のプロジェクトデータ
	testProjects := []api.Project{
		{ID: "project1", Name: "Test Project"},
		{ID: "project2", Name: "Another Project"},
		{ID: "project3", Name: "test project"}, // 小文字
		{ID: "project4", Name: "Partial Match Test"},
	}

	tests := []struct {
		name        string
		nameOrID    string
		expectedID  string
		expectError bool
	}{
		{
			name:       "ID完全一致",
			nameOrID:   "project2",
			expectedID: "project2",
		},
		{
			name:       "名前完全一致（大文字小文字無視）",
			nameOrID:   "TEST PROJECT",
			expectedID: "project1",
		},
		{
			name:       "名前完全一致（小文字）",
			nameOrID:   "test project",
			expectedID: "project1", // "Test Project"と大文字小文字を無視して一致
		},
		{
			name:       "名前部分一致",
			nameOrID:   "partial",
			expectedID: "project4",
		},
		{
			name:        "見つからない",
			nameOrID:    "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findProjectIDByNameLogic(testProjects, tt.nameOrID)

			if tt.expectError {
				assert.Error(t, err, "期待されたエラーが発生しませんでした")
				return
			}

			require.NoError(t, err, "予期しないエラーが発生しました")
			assert.Equal(t, tt.expectedID, result, "findProjectIDByNameLogic()結果が期待値と異なります")
		})
	}
}

// findProjectIDByNameLogic はテスト用に抽出したロジック関数
// Repository.FindProjectIDByNameと同じロジックを実装
func findProjectIDByNameLogic(projects []api.Project, nameOrID string) (string, error) {
	nameOrIDLower := strings.ToLower(nameOrID)

	// 1. ID完全一致（最優先・最高速）
	for _, project := range projects {
		if project.ID == nameOrID {
			return project.ID, nil
		}
	}

	// 2. 名前完全一致（大文字小文字を無視）
	for _, project := range projects {
		if strings.EqualFold(project.Name, nameOrID) {
			return project.ID, nil
		}
	}

	// 3. 名前部分一致（最後の手段）
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), nameOrIDLower) {
			return project.ID, nil
		}
	}

	return "", fmt.Errorf("project not found: %s", nameOrID)
}
