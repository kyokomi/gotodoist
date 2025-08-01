# gotodoist 開発ガイドライン

## プロジェクト概要
Todoist API v1を使用したGo言語製のCLIツール

## 開発フロー

### 1. Issue管理
- 各タスクはGitHub Issueとして管理
- Issueには以下を記載:
  - タスクの詳細説明
  - 受け入れ条件
  - 作業ログ（進捗や課題など）

### 2. ブランチ戦略
- mainブランチから機能ブランチを作成
- ブランチ名: `feature/#<issue番号>-<簡潔な説明>`
- 例: `feature/#1-initial-setup`

### 3. 開発時の注意事項
- コード上のコメントは日本語でOK
- go fmtでコードフォーマット必須
- staticcheckでlintチェック必須
- 関数やメソッドには必ずコメントを記載

### 4. コミット規則
- コミットメッセージは日本語でOK
- 形式: `<type>: <description> (#<issue番号>)`
- typeの種類:
  - feat: 新機能
  - fix: バグ修正
  - docs: ドキュメント更新
  - refactor: リファクタリング
  - test: テスト追加・修正
  - chore: その他の変更

### 5. Pull Request
- PRタイトル: `<type>: <description> (#<issue番号>)`
- PR本文には以下を含める:
  - 変更内容の概要
  - 関連Issue番号
  - テスト結果
  - 振り返り（良かった点、改善点など）

### 6. レビュー
- セルフレビュー後にレビュー依頼
- CI/CDが全て通っていることを確認
- レビューコメントには真摯に対応

## 技術スタック
- 言語: Go 1.21+
- CLIフレームワーク: cobra
- HTTP Client: 標準ライブラリ（net/http）
- 設定管理: viper
- 多言語対応: go-i18n

## ディレクトリ構成
```
gotodoist/
├── cmd/           # CLIコマンド定義
├── internal/      # 内部パッケージ
│   ├── api/       # Todoist APIクライアント
│   ├── config/    # 設定管理
│   └── i18n/      # 多言語対応
├── locales/       # 翻訳ファイル
├── .github/       # GitHub Actions設定
├── staticcheck.conf # staticcheck設定
├── go.mod
├── go.sum
├── main.go
├── README.md      # 英語版
├── README_ja.md   # 日本語版
└── CLAUDE.md      # このファイル
```

## 主要コマンド

### Makefileを使用した開発
```bash
# ヘルプ表示（利用可能なコマンド一覧）
make help

# ビルド
make build

# テスト実行
make test

# カバレッジ付きテスト
make coverage

# フォーマット
make fmt

# Lint実行
make lint

# 脆弱性チェック
make vuln

# 依存関係の更新
make tidy

# 開発モード（ファイル変更監視）
make dev

# CI用チェック（fmt, lint, test）
make ci

# リリースビルド（複数プラットフォーム）
make release

# クリーンアップ
make clean
```

### 直接実行する場合
```bash
# ビルド
go build -o gotodoist

# テスト実行
go test ./...

# フォーマット
go fmt ./...

# Lint実行
staticcheck ./...

# 依存関係の更新
go mod tidy
```

## 環境変数
- `TODOIST_API_TOKEN`: Todoist APIトークン
- `GOTODOIST_LANG`: 言語設定（en/ja）

## デバッグ
- `-v`または`--verbose`フラグで詳細ログ出力
- `--debug`フラグでAPIリクエスト/レスポンスを表示