# gotodoist

[![CI](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml/badge.svg)](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyokomi/gotodoist)](https://goreportcard.com/report/github.com/kyokomi/gotodoist)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/kyokomi/gotodoist.svg)](https://pkg.go.dev/github.com/kyokomi/gotodoist)

Goで構築された、Todoistタスクを管理するための強力なコマンドラインインターフェースツールです。

[English README is here](README.md)

## 特徴

- 📝 **タスク管理**: タスクの一覧表示、追加、更新、削除
- 📁 **プロジェクト管理**: プロジェクトでタスクを整理
- ⚙️ **設定管理**: 簡単なセットアップと設定
- 🌍 **多言語対応**: 英語と日本語をサポート
- 🚀 **高速・軽量**: 最適なパフォーマンスのためにGoで構築
- 🔒 **セキュア**: APIトークン保護付きのローカル設定

## インストール

### バイナリのダウンロード

[リリースページ](https://github.com/kyokomi/gotodoist/releases)から最新版をダウンロードしてください。

### ソースからビルド

```bash
git clone https://github.com/kyokomi/gotodoist.git
cd gotodoist
make build
```

### Go Installを使用

```bash
go install github.com/kyokomi/gotodoist@latest
```

## クイックスタート

### 1. APIトークンの取得

1. [Todoist統合設定](https://todoist.com/prefs/integrations)にアクセス
2. APIトークンをコピー

### 2. 設定

#### 方法A: 環境変数
```bash
export TODOIST_API_TOKEN="あなたのAPIトークン"
```

#### 方法B: 設定ファイル
```bash
gotodoist config init
```

これにより、`~/.config/gotodoist/config.yaml` に設定ファイルが作成されます。

### 3. 使用開始

```bash
# 全タスクを一覧表示
gotodoist task list

# 新しいタスクを追加
gotodoist task add "食材を買う"

# 全プロジェクトを一覧表示
gotodoist project list

# 新しいプロジェクトを追加
gotodoist project add "仕事のプロジェクト"
```

## 使用方法

### タスクコマンド

```bash
# 全タスクを一覧表示
gotodoist task list

# 新しいタスクを追加
gotodoist task add "タスクの内容"

# タスクを更新
gotodoist task update <タスクID> --content "新しい内容"

# タスクを削除
gotodoist task delete <タスクID>

# タスクを完了
gotodoist task complete <タスクID>
```

### プロジェクトコマンド

```bash
# 全プロジェクトを一覧表示
gotodoist project list

# 新しいプロジェクトを追加
gotodoist project add "プロジェクト名"

# プロジェクトを更新
gotodoist project update <プロジェクトID> --name "新しい名前"

# プロジェクトを削除
gotodoist project delete <プロジェクトID>
```

### 設定コマンド

```bash
# 設定を初期化
gotodoist config init

# 現在の設定を表示
gotodoist config show

# 言語設定を変更
gotodoist config set language ja  # または en
```

### グローバルオプション

```bash
# 詳細出力を有効化
gotodoist --verbose task list

# デバッグモードを有効化
gotodoist --debug task list

# 単一コマンドで言語を設定
gotodoist --lang en task list
```

## 設定

### 設定ファイルの場所

- **Linux/macOS**: `~/.config/gotodoist/config.yaml`
- **Windows**: `%APPDATA%\gotodoist\config.yaml`

### 設定オプション

```yaml
api_token: "あなたのtodoist-apiトークン"
base_url: "https://api.todoist.com/rest/v2"
language: "ja"  # en または ja
```

### 環境変数

- `TODOIST_API_TOKEN`: あなたのTodoist APIトークン
- `GOTODOIST_LANG`: 言語設定 (en/ja)

## 開発

### 前提条件

- Go 1.24以降
- Make（Makefileコマンドを使用する場合、オプション）

### ビルド

```bash
# アプリケーションをビルド
make build

# テストを実行
make test

# カバレッジ付きでテストを実行
make coverage

# コードフォーマット
make fmt

# リンターを実行
make lint

# 脆弱性チェック
make vuln
```

### 利用可能なMakeコマンド

全ての利用可能なコマンドを確認するには `make help` を実行してください：

```bash
make help
```

### プロジェクト構成

```
gotodoist/
├── cmd/           # CLIコマンド定義
├── internal/      # 内部パッケージ
│   ├── api/       # Todoist APIクライアント
│   ├── config/    # 設定管理
│   └── i18n/      # 国際化
├── locales/       # 翻訳ファイル
├── .github/       # GitHub Actionsワークフロー
├── Makefile       # ビルド自動化
├── go.mod         # Goモジュール定義
└── main.go        # アプリケーションエントリーポイント
```

## 貢献

貢献を歓迎します！お気軽にPull Requestを送信してください。

### 開発プロセス

1. リポジトリをフォーク
2. 機能ブランチを作成: `git checkout -b feature/#issue番号-説明`
3. 変更を実施
4. テストを実行: `make test`
5. リンターを実行: `make lint`
6. 変更を説明的なメッセージでコミット
7. あなたのフォークにプッシュしてPull Requestを送信

### コミットメッセージ形式

```
<タイプ>: <説明> (#<issue番号>)

タイプ: feat, fix, docs, refactor, test, chore
```

## ライセンス

このプロジェクトはMITライセンスの下でライセンスされています - 詳細は[LICENSE](LICENSE)ファイルを参照してください。

## 関連リンク

- [Todoist API ドキュメント](https://developer.todoist.com/rest/v2/)
- [課題追跡](https://github.com/kyokomi/gotodoist/issues)
- [リリース](https://github.com/kyokomi/gotodoist/releases)

## 作者

[@kyokomi](https://github.com/kyokomi)