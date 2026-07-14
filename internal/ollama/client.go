// Package ollama 提供对 Ollama HTTP API 的轻量封装。
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client 是 Ollama API 客户端。
type Client struct {
	Host    string
	Model   string
	Timeout time.Duration
	client  *http.Client
}

// NewClient 使用 host 和 model 创建客户端。
// host 为空时使用 http://localhost:11434，model 为空时使用 qwen2.5。
func NewClient(host, model string, timeout time.Duration) *Client {
	if host == "" {
		host = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen2.5"
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		Host:    strings.TrimRight(host, "/"),
		Model:   model,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// chatRequest 是 Ollama /api/chat 请求体。
type chatRequest struct {
	Model    string                 `json:"model"`
	Messages []chatMessage          `json:"messages"`
	Stream   bool                   `json:"stream"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse 是 Ollama /api/chat 响应体。
type chatResponse struct {
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
}

// Chat 向 Ollama 发送对话请求，返回模型生成的文本内容。
// systemPrompt 为系统提示词，userPrompt 为用户输入。启用 JSON 格式输出。
func (c *Client) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
		Format: "json",
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama: marshal request: %w", err)
	}

	u, err := url.Parse(c.Host + "/api/chat")
	if err != nil {
		return "", fmt.Errorf("ollama: invalid host: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("ollama: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: http %d: %s", resp.StatusCode, string(body))
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("ollama: decode response: %w", err)
	}
	return strings.TrimSpace(cr.Message.Content), nil
}
