package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Ollama.Host != "http://localhost:11434" {
		t.Errorf("ollama host = %s, want http://localhost:11434", cfg.Ollama.Host)
	}
	if cfg.Ollama.Model != "qwen2.5" {
		t.Errorf("ollama model = %s, want qwen2.5", cfg.Ollama.Model)
	}
	if cfg.Ollama.Enabled {
		t.Error("ollama enabled should be false by default")
	}
	if cfg.Ollama.TimeoutSeconds != 60 {
		t.Errorf("ollama timeout = %d, want 60", cfg.Ollama.TimeoutSeconds)
	}
}

func TestMergeOllama(t *testing.T) {
	dst := defaultConfig()
	src := &Config{
		Ollama: OllamaConfig{
			Enabled:        true,
			Host:           "http://192.168.1.10:11434",
			Model:          "llama3",
			TimeoutSeconds: 120,
			SystemPrompt:   "custom",
		},
	}
	mergeConfig(dst, src)
	if !dst.Ollama.Enabled {
		t.Error("ollama enabled should be true")
	}
	if dst.Ollama.Host != "http://192.168.1.10:11434" {
		t.Errorf("host = %s", dst.Ollama.Host)
	}
	if dst.Ollama.Model != "llama3" {
		t.Errorf("model = %s", dst.Ollama.Model)
	}
	if dst.Ollama.TimeoutSeconds != 120 {
		t.Errorf("timeout = %d", dst.Ollama.TimeoutSeconds)
	}
	if dst.Ollama.SystemPrompt != "custom" {
		t.Errorf("system_prompt = %s", dst.Ollama.SystemPrompt)
	}
}

func TestEnvOverridesOllama(t *testing.T) {
	os.Setenv("MI_OLLAMA_ENABLED", "true")
	os.Setenv("MI_OLLAMA_HOST", "http://ollama.local:11434")
	os.Setenv("MI_OLLAMA_MODEL", "mistral")
	defer func() {
		os.Unsetenv("MI_OLLAMA_ENABLED")
		os.Unsetenv("MI_OLLAMA_HOST")
		os.Unsetenv("MI_OLLAMA_MODEL")
	}()

	cfg := defaultConfig()
	applyEnvOverrides(cfg)
	if !cfg.Ollama.Enabled {
		t.Error("ollama enabled should be true from env")
	}
	if cfg.Ollama.Host != "http://ollama.local:11434" {
		t.Errorf("host = %s", cfg.Ollama.Host)
	}
	if cfg.Ollama.Model != "mistral" {
		t.Errorf("model = %s", cfg.Ollama.Model)
	}
}
