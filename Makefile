.PHONY: build test lint fmt clean run help

# 変数定義
BINARY_NAME=gotodoist
MAIN_FILE=main.go
GO=go
GOLANGCI_LINT=golangci-lint

# デフォルトターゲット
default: build

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  make build    - バイナリをビルド"
	@echo "  make test     - テストを実行"
	@echo "  make lint     - golangci-lintを実行"
	@echo "  make fmt      - コードをフォーマット"
	@echo "  make clean    - ビルド成果物を削除"
	@echo "  make run      - アプリケーションを実行"
	@echo "  make install  - バイナリをインストール"

# ビルド
build:
	$(GO) build -o $(BINARY_NAME) $(MAIN_FILE)

# テスト実行
test:
	$(GO) test -v ./...

# カバレッジ付きテスト
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint実行
lint:
	@if ! which $(GOLANGCI_LINT) > /dev/null; then \
		echo "golangci-lintがインストールされていません。以下のコマンドでインストールしてください:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run

# フォーマット
fmt:
	$(GO) fmt ./...

# クリーン
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# 実行
run: build
	./$(BINARY_NAME)

# インストール
install:
	$(GO) install

# 依存関係の整理
tidy:
	$(GO) mod tidy

# ベンダー依存関係
vendor:
	$(GO) mod vendor

# すべての品質チェック
check: fmt lint test