.PHONY: all build test coverage lint fmt clean install help

# デフォルトターゲット
all: fmt lint test build

# ビルド設定
BINARY_NAME := gotodoist
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'github.com/kyokomi/gotodoist/cmd.version=$(VERSION)' \
           -X 'github.com/kyokomi/gotodoist/cmd.commit=$(COMMIT)' \
           -X 'github.com/kyokomi/gotodoist/cmd.date=$(DATE)'

# ビルド
build: ## アプリケーションをビルド
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

# インストール
install: ## アプリケーションをインストール
	go install -ldflags "$(LDFLAGS)" .

# テスト実行
test: ## テストを実行
	go test -v -race ./...

# カバレッジ付きテスト
coverage: ## カバレッジ付きでテストを実行
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ベンチマーク
bench: ## ベンチマークを実行
	go test -bench=. -benchmem ./...

# コードフォーマット
fmt: ## コードフォーマットを実行
	go fmt ./...

# Lint実行
lint: ## golangci-lintでlintを実行
	@if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.61.0; \
	fi
	golangci-lint run

# staticcheck実行（従来互換）
staticcheck: ## staticcheckでlintを実行
	@if ! which staticcheck > /dev/null; then \
		echo "Installing staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	staticcheck ./...

# 依存関係の更新
tidy: ## go mod tidyを実行
	go mod tidy

# Vulnチェック
vuln: ## 脆弱性チェックを実行
	@if ! which govulncheck > /dev/null; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	govulncheck ./...

# クリーンアップ
clean: ## ビルド成果物を削除
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# 開発サーバー（watchモード）
dev: ## ファイル変更を監視して自動ビルド
	@if ! which air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	air

# リリースビルド（複数プラットフォーム）
release: ## リリース用ビルドを作成
	@mkdir -p dist
	# macOS (Intel)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)_darwin_amd64 .
	# macOS (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)_darwin_arm64 .
	# Linux (amd64)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)_linux_amd64 .
	# Linux (arm64)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)_linux_arm64 .
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)_windows_amd64.exe .

# CI用のチェック（全部実行）
ci: fmt lint test ## CI環境で実行するチェック

# ヘルプ表示
help: ## このヘルプメッセージを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
