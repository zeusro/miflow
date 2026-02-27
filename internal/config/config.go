// Package config 提供从 YAML 文件加载配置，支持环境变量覆盖和默认值。
// 参考: https://github.com/zeusro/go-template
package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/zeusro/miflow/internal/constants"
	"gopkg.in/yaml.v3"
)

var (
	cached   *Config
	loadOnce sync.Once
)

// Get 返回缓存的配置，首次调用时从文件加载。
func Get() *Config {
	loadOnce.Do(func() { cached = Load() })
	return cached
}

// 默认配置路径（第一个存在的生效）。
var configPaths = []string{
	".config.yaml",
	"config.yaml",
	".miflow.yaml",
}

// init 将用户主目录下的配置路径加入默认路径列表。
func init() {
	if home, err := os.UserHomeDir(); err == nil {
		configPaths = append(configPaths,
			filepath.Join(home, ".config", "miflow", "config.yaml"),
			filepath.Join(home, ".miflow.yaml"),
		)
	}
}

// Config 保存 miflow 的全部配置。
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

	// Web 服务（cmd/web，OAuth 登录页）
	Web WebConfig `yaml:"web"`

	// Mp3 服务（cmd/mp3，HTTP 文件服务）
	Mp3 Mp3Config `yaml:"mp3"`

	// MiIO 相关
	MiIO MiIOConfig `yaml:"miio"`
}

// OAuthConfig 小米 OAuth 2.0 配置。
type OAuthConfig struct {
	ClientID    string `yaml:"client_id"`
	RedirectURI string `yaml:"redirect_uri"`
	CloudServer string `yaml:"cloud_server"` // 云服务器：cn, de, i2, ru, sg, us
	DeviceID    string `yaml:"device_id"`    // 可选，用于 OAuth device_id
	APIHost     string `yaml:"api_host"`
	TokenPath   string `yaml:"token_path"` // API 路径
	AuthURL     string `yaml:"auth_url"`
	// TokenExpireRatio 过期前多少比例时刷新，0-1
	TokenExpireRatio float64 `yaml:"token_expire_ratio"`
}

// HTTPConfig HTTP 客户端超时等配置。
type HTTPConfig struct {
	TimeoutSeconds int `yaml:"timeout_seconds"`
}

// FlowConfig flow 服务器配置。
type FlowConfig struct {
	Addr    string `yaml:"addr"`
	DataDir string `yaml:"data_dir"`
}

// WebConfig 网页服务器配置（OAuth 登录 UI + 设备管理）。
type WebConfig struct {
	Addr    string `yaml:"addr"`     // 默认 :8123，与 oauth.redirect_uri 一致
	DataDir string `yaml:"data_dir"` // SQLite 等数据目录，默认 ./webdata
}

// Mp3Config mp3 HTTP 文件服务配置。
type Mp3Config struct {
	Addr string `yaml:"addr"`
	Host string `yaml:"host"` // 本机 IP，供局域网访问，空则自动检测
}

// MiIOConfig MiIO 服务配置。
type MiIOConfig struct {
	SpecsCachePath string `yaml:"specs_cache_path"`
	CallbackPort   int    `yaml:"callback_port"` // OAuth 回调端口
}

// Load 从文件读取配置。若文件不存在，返回带默认值的配置。
// 环境变量可覆盖：MI_OAUTH_CLIENT_ID, MI_OAUTH_REDIRECT_URI, MI_CLOUD_SERVER, MI_DID, MI_DEBUG 等。
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

// defaultConfig 返回带默认值的配置。
func defaultConfig() *Config {
	tokenPath := ".mi.token"
	if home, err := os.UserHomeDir(); err == nil {
		tokenPath = filepath.Join(home, ".mi.token")
	}
	return &Config{
		Debug:     false,
		TokenPath: tokenPath,
		OAuth: OAuthConfig{
			ClientID:         constants.OAuth2ClientID,
			RedirectURI:      constants.DefaultOAuthRedirectURI,
			CloudServer:      constants.DefaultCloudSvr,
			APIHost:          constants.OAuth2APIHost,
			TokenPath:        constants.OAuth2TokenPath,
			AuthURL:          constants.OAuth2AuthURL,
			TokenExpireRatio: constants.TokenExpireRatio,
		},
		HTTP: HTTPConfig{
			TimeoutSeconds: 30,
		},
		Flow: FlowConfig{
			Addr:    constants.DefaultFlowAddr,
			DataDir: "./flowdata",
		},
		Web: WebConfig{
			Addr:    constants.DefaultWebAddr,
			DataDir: "./webdata",
		},
		Mp3: Mp3Config{
			Addr: constants.DefaultMp3Addr,
		},
		MiIO: MiIOConfig{
			SpecsCachePath: "",
			CallbackPort:   constants.DefaultCallbackPort,
		},
	}
}

// mergeConfig 将 src 的非零值合并到 dst。
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
	mergeWeb(&dst.Web, &src.Web)
	mergeMp3(&dst.Mp3, &src.Mp3)
	mergeMiIO(&dst.MiIO, &src.MiIO)
}

// mergeOAuth 合并 OAuth 配置。
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

// mergeHTTP 合并 HTTP 配置。
func mergeHTTP(dst, src *HTTPConfig) {
	if src.TimeoutSeconds > 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
}

// mergeFlow 合并 Flow 配置。
func mergeFlow(dst, src *FlowConfig) {
	if src.Addr != "" {
		dst.Addr = src.Addr
	}
	if src.DataDir != "" {
		dst.DataDir = src.DataDir
	}
}

// mergeWeb 合并 Web 配置。
func mergeWeb(dst, src *WebConfig) {
	if src.Addr != "" {
		dst.Addr = src.Addr
	}
	if src.DataDir != "" {
		dst.DataDir = src.DataDir
	}
}

// mergeMp3 合并 Mp3 配置。
func mergeMp3(dst, src *Mp3Config) {
	if src.Addr != "" {
		dst.Addr = src.Addr
	}
	if src.Host != "" {
		dst.Host = src.Host
	}
}

// mergeMiIO 合并 MiIO 配置。
func mergeMiIO(dst, src *MiIOConfig) {
	if src.SpecsCachePath != "" {
		dst.SpecsCachePath = src.SpecsCachePath
	}
	if src.CallbackPort > 0 {
		dst.CallbackPort = src.CallbackPort
	}
}

// expandPath 将 ~ 展开为用户主目录。
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

// applyEnvOverrides 用环境变量覆盖配置项。
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
