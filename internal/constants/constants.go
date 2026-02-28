// Package constants 集中定义 miflow 的静态业务常量（URL、端口、OAuth 默认值等）。
package constants

// OAuth2 默认值（可由 config 覆盖）
const (
	OAuth2ClientID   = "2882303761520251711"
	OAuth2AuthURL    = "https://account.xiaomi.com/oauth2/authorize"
	OAuth2APIHost    = "ha.api.io.mi.com"
	OAuth2TokenPath  = "/app/v2/ha/oauth/get_token"
	DefaultCloudSvr  = "cn"
	TokenExpireRatio = 0.99
)

// 默认 OAuth 回调地址（需与 web.addr 端口一致）
const DefaultOAuthRedirectURI = "http://homeassistant.local:8123/callback"

// 服务默认端口
const (
	DefaultWebAddr  = ":8123"
	DefaultFlowAddr = ":18090"
	DefaultMp3Addr  = ":8090"
	DefaultMp3Port  = "8090"
)

// MiIO 默认回调端口
const DefaultCallbackPort = 8123
