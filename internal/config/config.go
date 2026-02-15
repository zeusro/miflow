// Package config provides configuration loading from YAML file with env override and defaults.
// Ref: https://github.com/zeusro/go-template
package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	cached   *Config
	loadOnce sync.Once
)

// Get returns cached config, loading from file on first call.
func Get() *Config {
	loadOnce.Do(func() { cached = Load() })
	return cached
}

// Default config paths (first existing wins).
var configPaths = []string{
	".config.yaml",
	"config.yaml",
	".miflow.yaml",
}

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		configPaths = append(configPaths,
			filepath.Join(home, ".config", "miflow", "config.yaml"),
			filepath.Join(home, ".miflow.yaml"),
		)
	}
}

// Config holds all miflow configuration.
type Config struct {
	Debug bool `yaml:"debug"`

	// OAuth / 小米云
	OAuth OAuthConfig `yaml:"oauth"`

	// Token 存储路径
	TokenPath string `yaml:"token_path"`

	// 默认设备 ID（覆盖 MI_DID 环境变量）
	DefaultDID string `yaml:"default_did"`

	// HTTP 客户端
	HTTP HTTPConfig `yaml:"http"`

	// Flow 服务（cmd/flow）
	Flow FlowConfig `yaml:"flow"`

	// xiaomusic
	Xiaomusic XiaomusicConfig `yaml:"xiaomusic"`

	// MiIO 相关
	MiIO MiIOConfig `yaml:"miio"`
}

// OAuthConfig for Xiaomi OAuth 2.0.
type OAuthConfig struct {
	ClientID    string `yaml:"client_id"`
	RedirectURI string `yaml:"redirect_uri"`
	CloudServer string `yaml:"cloud_server"` // cn, de, i2, ru, sg, us
	DeviceID    string `yaml:"device_id"`   // 可选，用于 OAuth device_id
	APIHost     string `yaml:"api_host"`
	TokenPath   string `yaml:"token_path"`   // API path
	AuthURL     string `yaml:"auth_url"`
	// TokenExpireRatio 过期前多少比例时刷新，0-1
	TokenExpireRatio float64 `yaml:"token_expire_ratio"`
}

// HTTPConfig for HTTP client timeouts etc.
type HTTPConfig struct {
	TimeoutSeconds int `yaml:"timeout_seconds"`
}

// FlowConfig for flow server.
type FlowConfig struct {
	Addr    string `yaml:"addr"`
	DataDir string `yaml:"data_dir"`
}

// XiaomusicConfig for xiaomusic CLI.
type XiaomusicConfig struct {
	MusicDir string `yaml:"music_dir"`
	Addr    string `yaml:"addr"`
}

// MiIOConfig for MiIO service.
type MiIOConfig struct {
	SpecsCachePath string `yaml:"specs_cache_path"`
	CallbackPort   int    `yaml:"callback_port"` // OAuth 回调端口
}

// Load reads config from file. If file not found, returns config with defaults.
// Env vars override: MI_OAUTH_CLIENT_ID, MI_OAUTH_REDIRECT_URI, MI_CLOUD_SERVER, MI_DID, MI_DEBUG, etc.
func Load() *Config {
	cfg := defaultConfig()
	for _, p := range configPaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var file Config
		if err := yaml.Unmarshal(data, &file); err != nil {
			continue
		}
		mergeConfig(cfg, &file)
		break
	}
	applyEnvOverrides(cfg)
	cfg.TokenPath = expandPath(cfg.TokenPath)
	return cfg
}

func defaultConfig() *Config {
	tokenPath := ".mi.token"
	if home, err := os.UserHomeDir(); err == nil {
		tokenPath = filepath.Join(home, ".mi.token")
	}
	return &Config{
		Debug:     false,
		TokenPath: tokenPath,
		OAuth: OAuthConfig{
			ClientID:         "2882303761520251711",
			RedirectURI:      "http://homeassistant.local:8123/callback",
			CloudServer:      "cn",
			APIHost:          "ha.api.io.mi.com",
			TokenPath:        "/app/v2/ha/oauth/get_token",
			AuthURL:          "https://account.xiaomi.com/oauth2/authorize",
			TokenExpireRatio: 0.7,
		},
		HTTP: HTTPConfig{
			TimeoutSeconds: 30,
		},
		Flow: FlowConfig{
			Addr:    ":18090",
			DataDir: "./flowdata",
		},
		Xiaomusic: XiaomusicConfig{
			MusicDir: "./music",
			Addr:    ":8090",
		},
		MiIO: MiIOConfig{
			SpecsCachePath: "",
			CallbackPort:   8123,
		},
	}
}

func mergeConfig(dst, src *Config) {
	if src.Debug {
		dst.Debug = true
	}
	if src.TokenPath != "" {
		dst.TokenPath = src.TokenPath
	}
	if src.DefaultDID != "" {
		dst.DefaultDID = src.DefaultDID
	}
	mergeOAuth(&dst.OAuth, &src.OAuth)
	mergeHTTP(&dst.HTTP, &src.HTTP)
	mergeFlow(&dst.Flow, &src.Flow)
	mergeXiaomusic(&dst.Xiaomusic, &src.Xiaomusic)
	mergeMiIO(&dst.MiIO, &src.MiIO)
}

func mergeOAuth(dst, src *OAuthConfig) {
	if src.ClientID != "" {
		dst.ClientID = src.ClientID
	}
	if src.RedirectURI != "" {
		dst.RedirectURI = src.RedirectURI
	}
	if src.CloudServer != "" {
		dst.CloudServer = src.CloudServer
	}
	if src.DeviceID != "" {
		dst.DeviceID = src.DeviceID
	}
	if src.APIHost != "" {
		dst.APIHost = src.APIHost
	}
	if src.TokenPath != "" {
		dst.TokenPath = src.TokenPath
	}
	if src.AuthURL != "" {
		dst.AuthURL = src.AuthURL
	}
	if src.TokenExpireRatio > 0 {
		dst.TokenExpireRatio = src.TokenExpireRatio
	}
}

func mergeHTTP(dst, src *HTTPConfig) {
	if src.TimeoutSeconds > 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
}

func mergeFlow(dst, src *FlowConfig) {
	if src.Addr != "" {
		dst.Addr = src.Addr
	}
	if src.DataDir != "" {
		dst.DataDir = src.DataDir
	}
}

func mergeXiaomusic(dst, src *XiaomusicConfig) {
	if src.MusicDir != "" {
		dst.MusicDir = src.MusicDir
	}
	if src.Addr != "" {
		dst.Addr = src.Addr
	}
}

func mergeMiIO(dst, src *MiIOConfig) {
	if src.SpecsCachePath != "" {
		dst.SpecsCachePath = src.SpecsCachePath
	}
	if src.CallbackPort > 0 {
		dst.CallbackPort = src.CallbackPort
	}
}

// expandPath expands ~ to user home directory.
func expandPath(p string) string {
	if p == "" || p[0] != '~' {
		return p
	}
	if len(p) > 1 && p[1] != '/' && p[1] != '\\' {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return filepath.Join(home, p[1:])
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("MI_OAUTH_CLIENT_ID"); v != "" {
		cfg.OAuth.ClientID = v
	}
	if v := os.Getenv("MI_OAUTH_REDIRECT_URI"); v != "" {
		cfg.OAuth.RedirectURI = v
	}
	if v := os.Getenv("MI_CLOUD_SERVER"); v != "" {
		cfg.OAuth.CloudServer = v
	}
	if v := os.Getenv("MI_OAUTH_DEVICE_ID"); v != "" {
		cfg.OAuth.DeviceID = v
	}
	if v := os.Getenv("MI_DID"); v != "" {
		cfg.DefaultDID = v
	}
	if v := os.Getenv("MI_DEBUG"); v == "1" || v == "true" {
		cfg.Debug = true
	}
	if v := os.Getenv("MI_TOKEN_PATH"); v != "" {
		cfg.TokenPath = v
	}
}
