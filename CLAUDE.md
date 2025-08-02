# gotodoist 開発ガイドライン

## プロジェクト概要
Todoist API v1を使用したGo言語製のCLIツール

## 🚀 クイック開発フロー

### 1. 作業開始時
```bash
# Issue作成（または既存Issueを確認）
gh issue create --title "タイトル" --body "詳細"

# ブランチ作成
git checkout -b feature/#<issue番号>-<説明>
```

### 2. 開発中
```bash
# コード整形とチェック（必須）
make fmt      # go fmt実行
make lint     # staticcheck実行
make test     # テスト実行
```

### 3. コミット＆PR
```bash
# コミット（形式: <type>: 説明 (#issue番号)）
git commit -m "feat: 新機能を追加 (#32)"

# PR作成
git push -u origin <branch-name>
gh pr create --title "タイトル" --body "## Summary
- 変更内容

## Test plan
- [ ] make fmt/lint/test 実行済み"
```

## 📝 コミットタイプ
- `feat`: 新機能
- `fix`: バグ修正
- `refactor`: リファクタリング
- `docs`: ドキュメント
- `test`: テスト
- `chore`: その他

## 🏗️ コード構成ガイドライン

### コード構成の基本原則
1. **明確な並び順**: init → コマンド定義 → パラメータ構造体 → ハンドラー → ビジネスロジック → エクゼキューター → レシーバーメソッド → ヘルパー
2. **コマンドごとのグループ化**: 各コマンドに関連する要素（パラメータ型、getParams、runハンドラー）をまとめて配置
3. **統一された4ステップパターン**: すべてのrunメソッドは「セットアップ → パラメータ取得 → 実行 → 結果表示」の構成
4. **Context管理**: 構造体に保持せず第一引数で引き回し。cmdパッケージでは冒頭で`ctx := context.Background()`を宣言、internalパッケージでは受け取ったcontextをそのまま使用

## 🛠️ 主要コマンド

```bash
make help     # 利用可能なコマンド一覧
make build    # ビルド
make dev      # 開発モード（ファイル監視）
make ci       # CI用チェック（fmt, lint, test）
```

## 📁 ディレクトリ構成
```
gotodoist/
├── cmd/           # CLIコマンド定義
├── internal/      # 内部パッケージ
│   ├── api/       # Todoist APIクライアント
│   ├── config/    # 設定管理
│   ├── repository/# データアクセス層（ローカル優先）
│   ├── storage/   # SQLiteストレージ
│   └── sync/      # 同期処理
├── locales/       # 翻訳ファイル
└── .github/       # GitHub Actions設定
```

## ⚡ 便利な情報
- コメントは日本語OK
- 環境変数: `TODOIST_API_TOKEN`
- 設定ファイル: `~/.config/gotodoist/config.yaml`
- デバッグ: `-v` または `--debug` フラグ