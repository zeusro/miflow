package nlu

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/ollama"
)

// LLMClient 定义调用大模型所需的接口，便于测试替换。
type LLMClient interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Result 是 Execute 的返回结果。
type Result struct {
	Intent    *Intent                  `json:"intent"`
	Devices   []*device.Device         `json:"devices,omitempty"`
	Executed  bool                     `json:"executed"`
	Results   []map[string]interface{} `json:"results,omitempty"`
	Error     string                   `json:"error,omitempty"`
	Ambiguous bool                     `json:"ambiguous"`
}

// Service 自然语言控制服务。
type Service struct {
	api    *device.API
	client LLMClient
	cfg    config.OllamaConfig
}

// NewService 创建 NLU 服务。
func NewService(api *device.API, client LLMClient) *Service {
	if client == nil {
		cfg := config.Get().Ollama
		client = ollama.NewClient(cfg.Host, cfg.Model, 0)
	}
	return &Service{
		api:    api,
		client: client,
		cfg:    config.Get().Ollama,
	}
}

// Execute 解析并执行自然语言指令。
func (s *Service) Execute(ctx context.Context, text string) (*Result, error) {
	if s.api == nil {
		return nil, fmt.Errorf("nlu: device api not initialized")
	}
	if s.client == nil {
		return nil, fmt.Errorf("nlu: ollama client not initialized")
	}

	devices, err := s.api.List("", false, 0)
	if err != nil {
		return nil, fmt.Errorf("nlu: list devices: %w", err)
	}
	rooms, err := s.api.RoomsWithDevices()
	if err != nil {
		// 房间结构非必须，失败时继续使用设备列表
		rooms = nil
	}

	systemPrompt := s.cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt(devices, rooms)
	}

	content, err := s.client.Chat(ctx, systemPrompt, text)
	if err != nil {
		return nil, fmt.Errorf("nlu: llm chat: %w", err)
	}

	intent, err := parseIntent(content)
	if err != nil {
		return nil, fmt.Errorf("nlu: parse intent: %w", err)
	}

	res := &Result{Intent: intent}

	if intent.Action == "list_devices" {
		res.Devices = devices
		res.Executed = true
		return res, nil
	}
	if intent.Action == "unknown" {
		res.Error = "unknown intent"
		return res, nil
	}

	resolver := NewResolver(devices, rooms)
	matched, ambiguous, err := resolver.Resolve(intent)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.Ambiguous = ambiguous
	res.Devices = matched

	if len(matched) == 0 {
		res.Error = "no device matched"
		return res, nil
	}

	exec := NewExecutor(s.api)
	res.Results = make([]map[string]interface{}, 0, len(matched))
	for _, d := range matched {
		out, err := exec.Execute(d, intent)
		item := map[string]interface{}{
			"did":   d.DID,
			"name":  d.Name,
			"model": d.Model,
		}
		if err != nil {
			item["error"] = err.Error()
			res.Error = err.Error()
		} else {
			item["ok"] = true
			res.Executed = true
			if out != nil {
				item["data"] = out
			}
		}
		res.Results = append(res.Results, item)
	}

	return res, nil
}

func parseIntent(content string) (*Intent, error) {
	content = extractJSON(content)
	var intent Intent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, fmt.Errorf("invalid intent json: %w", err)
	}
	intent.Action = normalizeAction(intent.Action)
	return &intent, nil
}

func extractJSON(content string) string {
	// 若模型在 JSON 外加了 markdown 代码块，尝试提取。
	if idx := jsonStartIndex(content); idx >= 0 {
		content = content[idx:]
	}
	if idx := jsonEndIndex(content); idx >= 0 {
		content = content[:idx+1]
	}
	return content
}

func jsonStartIndex(s string) int {
	for i, r := range s {
		if r == '{' {
			return i
		}
	}
	return -1
}

func jsonEndIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' {
			return i
		}
	}
	return -1
}

func normalizeAction(action string) string {
	switch action {
	case "open", "on", "turn_on":
		return "turn_on"
	case "close", "off", "turn_off":
		return "turn_off"
	case "switch", "toggle":
		return "toggle"
	case "brightness", "set_brightness":
		return "set_brightness"
	case "volume", "set_volume":
		return "set_volume"
	case "mute", "set_mute":
		return "set_mute"
	case "channel", "set_channel":
		return "set_channel"
	case "speak", "say", "tts":
		return "tts"
	case "play":
		return "play"
	case "pause", "stop":
		return "pause"
	case "next":
		return "next"
	case "previous", "prev":
		return "previous"
	case "status", "query_status", "query":
		return "query_status"
	case "list", "list_devices":
		return "list_devices"
	case "unknown", "unsupported":
		return "unknown"
	}
	return action
}
