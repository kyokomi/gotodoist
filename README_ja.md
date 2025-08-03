# gotodoist

[![CI](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml/badge.svg)](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/kyokomi/gotodoist/graph/badge.svg?token=cGdi7YkLjv)](https://codecov.io/github/kyokomi/gotodoist)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyokomi/gotodoist)](https://goreportcard.com/report/github.com/kyokomi/gotodoist)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/kyokomi/gotodoist.svg)](https://pkg.go.dev/github.com/kyokomi/gotodoist)

Goで構築された、Todoistタスクを管理するための強力なコマンドラインツールです。

[English README is here](README.md)

## 特徴

- 📝 **タスク管理**: タスクの一覧表示、追加、更新、完了、削除
- 📁 **プロジェクト管理**: プロジェクトでタスクを整理
- 🏷️ **ラベル対応**: ラベルでタスクを分類
- 📅 **期限管理**: タスクの期限を設定・管理
- 🔄 **オフライン対応**: ローカル同期によるオフライン作業
- 🚀 **高速・軽量**: 速度と効率性を重視した設計

## インストール

### バイナリダウンロード

[リリースページ](https://github.com/kyokomi/gotodoist/releases)から、お使いのプラットフォーム用の最新バイナリをダウンロードしてください。

### Go Installを使用

```bash
go install github.com/kyokomi/gotodoist@latest
```

## クイックスタート

### 1. APIトークンの取得

1. [Todoist統合設定](https://todoist.com/prefs/integrations)にアクセス
2. 「APIトークン」セクションからAPIトークンをコピー

### 2. 認証設定

```bash
# 方法1: 環境変数（推奨）
export TODOIST_API_TOKEN="あなたのAPIトークン"

# 方法2: 設定ファイル
gotodoist config init
```

### 3. 使用開始

```bash
# Todoistと同期（初回使用時推奨）
gotodoist sync

# 全タスクを表示
gotodoist task list

# 新しいタスクを追加
gotodoist task add "食材を買う"

# 特定のプロジェクトにタスクを追加
gotodoist task add "レポート作成" -p "仕事"
```

## 主要コマンド

### タスク管理

```bash
# タスクの一覧表示
gotodoist task list                          # アクティブなタスク全て
gotodoist task list -p "仕事"                # "仕事"プロジェクトのタスク
gotodoist task list -f "p1"                  # 優先度1のタスク
gotodoist task list -f "@重要"               # "重要"ラベルのタスク
gotodoist task list -a                       # 全てのタスク（完了済みを含む）

# タスクの追加
gotodoist task add "タスクの内容"
gotodoist task add "重要なタスク" -P 1        # 優先度付き（1-4）
gotodoist task add "会議" -d "明日"           # 期限付き
gotodoist task add "クライアント電話" -p "仕事" -l "緊急,電話"  # プロジェクトとラベル付き

# タスクの更新
gotodoist task update <タスクID> -c "新しい内容"
gotodoist task update <タスクID> -P 2         # 優先度変更
gotodoist task update <タスクID> -d "来週月曜日"  # 期限変更

# タスクの完了/未完了
gotodoist task complete <タスクID>
gotodoist task uncomplete <タスクID>

# タスクの削除
gotodoist task delete <タスクID>
gotodoist task delete <タスクID> -f           # 確認をスキップ
```

### プロジェクト管理

```bash
# プロジェクトの一覧表示
gotodoist project list
gotodoist project list -v                    # 詳細表示（IDを含む）

# プロジェクトの追加
gotodoist project add "新しいプロジェクト"

# プロジェクトの更新
gotodoist project update <プロジェクトID> --name "更新された名前"

# プロジェクトの削除
gotodoist project delete <プロジェクトID>
gotodoist project delete <プロジェクトID> -f   # 確認をスキップ
```

### 同期

```bash
# Todoistとの同期
gotodoist sync                               # 全データの同期
gotodoist sync init                          # 初回フル同期
gotodoist sync status                        # 同期状況の確認
gotodoist sync reset -f                      # ローカルデータのリセット
```

### 設定

```bash
# 設定の初期化
gotodoist config init

# 現在の設定を表示
gotodoist config show

```

## 設定オプション

設定ファイルの場所:
- **Linux/macOS**: `~/.config/gotodoist/config.yaml`
- **Windows**: `%APPDATA%\gotodoist\config.yaml`

### 環境変数

- `TODOIST_API_TOKEN`: TodoistのAPIトークン

## 使用例とTips

### 優先度によるタスクフィルタリング
```bash
# 高優先度のタスク
gotodoist task list -f "p1"

# 中優先度・低優先度
gotodoist task list -f "p3 | p4"
```

### ラベルの活用
```bash
# 複数ラベル付きタスクの追加
gotodoist task add "PRレビュー" -l "コードレビュー,緊急"

# ラベルでフィルタリング
gotodoist task list -f "@コードレビュー"
```

### 期限の設定例
```bash
# 自然言語での期限設定
gotodoist task add "レポート提出" -d "来週金曜日"
gotodoist task add "チーム会議" -d "毎週月曜日"

# 具体的な日付
gotodoist task add "誕生日パーティー" -d "2024-12-25"
```

### フィルタの組み合わせ
```bash
# 仕事プロジェクトの高優先度タスク
gotodoist task list -p "仕事" -f "p1"

# 今日期限の緊急タスク
gotodoist task list -f "今日 & @緊急"
```

## トラブルシューティング

### よくある問題

1. **「APIトークンが見つかりません」エラー**
   - `TODOIST_API_TOKEN`が設定されているか確認
   - または`gotodoist config init`を実行

2. **「プロジェクトが見つかりません」エラー**
   - `gotodoist project list`で利用可能なプロジェクトを確認
   - プロジェクト名は大文字小文字を区別します

3. **同期の問題**
   - `gotodoist sync`を実行してローカルデータを更新
   - データが破損している場合は`gotodoist sync reset -f`を使用

## 貢献

貢献を歓迎します！[issuesページ](https://github.com/kyokomi/gotodoist/issues)でお手伝いいただける分野をご確認ください。

## ライセンス

このプロジェクトはMITライセンスの下でライセンスされています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## 関連リンク

- [Todoist API ドキュメント](https://developer.todoist.com/rest/v2/)
- [問題を報告](https://github.com/kyokomi/gotodoist/issues)
- [リリース](https://github.com/kyokomi/gotodoist/releases)

## 作者

[@kyokomi](https://github.com/kyokomi)
