// Package api provides Todoist API client functionality.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// デフォルト設定
const (
	DefaultBaseURL = "https://api.todoist.com"
	DefaultTimeout = 30 * time.Second
	UserAgent      = "gotodoist/dev"

	// HTTPステータスコード
	httpStatusBadRequest = 400
)

// Client はTodoist APIクライアント
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
	userAgent  string
}

// NewClient は新しいAPIクライアントを作成する
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("API token is required")
	}

	baseURL, err := url.Parse(DefaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		token:     token,
		userAgent: UserAgent,
	}, nil
}

// SetBaseURL はベースURLを設定する（テスト用）
func (c *Client) SetBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	c.baseURL = u
	return nil
}

// SetTimeout はHTTPクライアントのタイムアウトを設定する
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// newRequest は新しいHTTPリクエストを作成する
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u := *c.baseURL
	u.Path += path

	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// ヘッダーを設定
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// do はHTTPリクエストを実行し、レスポンスをデコードする
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// エラーハンドリング
	if resp.StatusCode >= httpStatusBadRequest {
		return c.handleErrorResponse(resp)
	}

	// レスポンスボディを読み取り
	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// handleErrorResponse はエラーレスポンスを処理する
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response", resp.StatusCode)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    errorResp.Error,
	}
}

// ErrorResponse はAPIエラーレスポンスの構造体
type ErrorResponse struct {
	Error string `json:"error"`
}

// APIError はAPIエラーを表す
type APIError struct {
	StatusCode int
	Message    string
}

// Error はerrorインターフェースを実装する
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsNotFound はエラーが404かどうかを判定する
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized はエラーが401かどうかを判定する
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden はエラーが403かどうかを判定する
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsRateLimited はエラーが429かどうかを判定する
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// Sync はSync APIを実行する
func (c *Client) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	if req == nil {
		req = &SyncRequest{
			SyncToken:     "*",
			ResourceTypes: []string{ResourceAll},
		}
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, "/sync/v9/sync", req)
	if err != nil {
		return nil, err
	}

	var resp SyncResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("sync request failed: %w", err)
	}

	return &resp, nil
}
