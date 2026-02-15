package miaccount

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/zeusro/miflow/internal/config"
)

// OAuth2 constants (defaults, overridden by config)
const (
	OAuth2ClientID   = "2882303761520251711"
	OAuth2AuthURL    = "https://account.xiaomi.com/oauth2/authorize"
	OAuth2APIHost    = "ha.api.io.mi.com"
	OAuth2TokenPath  = "/app/v2/ha/oauth/get_token"
	DefaultCloudSvr  = "cn"
	TokenExpireRatio = 0.7
)

// OAuthToken holds OAuth 2.0 auth data (replaces password-based token).
type OAuthToken struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	ExpiresTS     int64  `json:"expires_ts"` // Unix timestamp when token is considered expired
	DeviceID      string `json:"device_id"`
	State         string `json:"state,omitempty"`
	CloudServer   string `json:"cloud_server,omitempty"`
	OAuthClientID string `json:"oauth_client_id,omitempty"`
	OAuthRedirect string `json:"oauth_redirect,omitempty"`
}

// IsValid returns true if access token exists and is not expired.
func (t *OAuthToken) IsValid() bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	if t.ExpiresTS > 0 && time.Now().Unix() >= t.ExpiresTS {
		return false
	}
	return true
}

// OAuthClient handles OAuth 2.0 flow for Xiaomi MIoT (白名单域名模式).
type OAuthClient struct {
	ClientID    string
	RedirectURI string
	DeviceID    string
	State       string
	CloudServer string
	HTTP        *http.Client
}

// NewOAuthClient creates client from config (file + env override), falls back to defaults.
func NewOAuthClient() *OAuthClient {
	cfg := config.Get()
	clientID := cfg.OAuth.ClientID
	if clientID == "" {
		clientID = OAuth2ClientID
	}
	redirectURI := cfg.OAuth.RedirectURI
	if redirectURI == "" {
		redirectURI = "http://homeassistant.local:8123/callback"
	}
	deviceID := "ha." + RandString(16)
	if d := cfg.OAuth.DeviceID; d != "" {
		deviceID = d
	}
	h := sha1.Sum([]byte("d=" + deviceID))
	state := hex.EncodeToString(h[:])
	cloud := cfg.OAuth.CloudServer
	if cloud == "" {
		cloud = DefaultCloudSvr
	}
	timeout := 30
	if cfg.HTTP.TimeoutSeconds > 0 {
		timeout = cfg.HTTP.TimeoutSeconds
	}
	return &OAuthClient{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		DeviceID:    deviceID,
		State:       state,
		CloudServer: cloud,
		HTTP:        &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// GenAuthURL returns the URL for user to authorize.
func (c *OAuthClient) GenAuthURL(redirectURI, state string, skipConfirm bool) string {
	if redirectURI == "" {
		redirectURI = c.RedirectURI
	}
	if state == "" {
		state = c.State
	}
	params := url.Values{
		"redirect_uri":  {redirectURI},
		"client_id":     {c.ClientID},
		"response_type": {"code"},
		"device_id":     {c.DeviceID},
		"state":         {state},
		"skip_confirm":  {"true"},
	}
	if !skipConfirm {
		params.Set("skip_confirm", "false")
	}
	authURL := config.Get().OAuth.AuthURL
	if authURL == "" {
		authURL = OAuth2AuthURL
	}
	return authURL + "?" + params.Encode()
}

// GetToken exchanges authorization code for access/refresh tokens.
func (c *OAuthClient) GetToken(code string) (*OAuthToken, error) {
	data := map[string]string{
		"client_id":    c.ClientID,
		"redirect_uri": c.RedirectURI,
		"code":         code,
		"device_id":    c.DeviceID,
	}
	return c.getToken(data)
}

// RefreshToken refreshes access token using refresh_token.
func (c *OAuthClient) RefreshToken(refreshToken string) (*OAuthToken, error) {
	data := map[string]string{
		"client_id":     c.ClientID,
		"redirect_uri":  c.RedirectURI,
		"refresh_token": refreshToken,
	}
	return c.getToken(data)
}

func (c *OAuthClient) getToken(data map[string]string) (*OAuthToken, error) {
	cfg := config.Get()
	apiHost := cfg.OAuth.APIHost
	if apiHost == "" {
		apiHost = OAuth2APIHost
	}
	host := apiHost
	if c.CloudServer != "" && c.CloudServer != "cn" {
		host = c.CloudServer + "." + apiHost
	}
	tokenPath := cfg.OAuth.TokenPath
	if tokenPath == "" {
		tokenPath = OAuth2TokenPath
	}
	payload, _ := json.Marshal(data)
	reqURL := "https://" + host + tokenPath + "?data=" + url.QueryEscape(string(payload))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if httpDebug() {
		logHttpReq("GET", reqURL, nil)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logHttpResp(resp.StatusCode, raw)
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("oauth: 401 unauthorized, token may be revoked")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("oauth: http %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		Code   float64 `json:"code"`
		Result struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("oauth: invalid response: %w", err)
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("oauth: code %.0f, %s", out.Code, string(raw))
	}
	r := &out.Result
	if r.AccessToken == "" || r.RefreshToken == "" {
		return nil, fmt.Errorf("oauth: missing access_token or refresh_token")
	}
	expireRatio := config.Get().OAuth.TokenExpireRatio
	if expireRatio <= 0 {
		expireRatio = TokenExpireRatio
	}
	expiresTS := time.Now().Unix() + int64(float64(r.ExpiresIn)*expireRatio)
	return &OAuthToken{
		AccessToken:   r.AccessToken,
		RefreshToken:  r.RefreshToken,
		ExpiresIn:     r.ExpiresIn,
		ExpiresTS:     expiresTS,
		DeviceID:      c.DeviceID,
		State:         c.State,
		CloudServer:   c.CloudServer,
		OAuthClientID: c.ClientID,
		OAuthRedirect: c.RedirectURI,
	}, nil
}

// ServeCallback starts HTTP server to receive OAuth callback and returns the auth code.
func ServeCallback(port int) (string, error) {
	ch := make(chan string, 1)
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		if authCode != "" {
			ch <- authCode
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<html><body><p>登录成功！可以关闭此页面。</p></body></html>`))
		} else {
			http.Error(w, "missing code", 400)
		}
	})
	go srv.ListenAndServe()
	select {
	case code := <-ch:
		srv.Close()
		return code, nil
	case <-time.After(120 * time.Second):
		srv.Close()
		return "", fmt.Errorf("oauth: timeout waiting for callback")
	}
}

// OpenAuthURL opens the auth URL in the default browser.
func OpenAuthURL(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}
