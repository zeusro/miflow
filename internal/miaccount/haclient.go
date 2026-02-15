package miaccount

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeusro/miflow/internal/config"
)

// HAClient is HTTP client for ha.api.io.mi.com (OAuth Bearer token).
// Ref: https://github.com/XiaoMi/ha_xiaomi_home
type HAClient struct {
	Host        string
	BaseURL     string
	ClientID    string
	AccessToken string
	HTTP        *http.Client
	TokenStore  *TokenStore
	OAuthToken  *OAuthToken
}

// NewHAClient creates client for given OAuth token.
func NewHAClient(t *OAuthToken, tokenStore *TokenStore) *HAClient {
	cfg := config.Get()
	apiHost := cfg.OAuth.APIHost
	if apiHost == "" {
		apiHost = OAuth2APIHost
	}
	host := apiHost
	if t.CloudServer != "" && t.CloudServer != "cn" {
		host = t.CloudServer + "." + apiHost
	}
	clientID := t.OAuthClientID
	if clientID == "" {
		clientID = cfg.OAuth.ClientID
	}
	if clientID == "" {
		clientID = OAuth2ClientID
	}
	timeout := cfg.HTTP.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}
	return &HAClient{
		Host:        host,
		BaseURL:     "https://" + host,
		ClientID:    clientID,
		AccessToken: t.AccessToken,
		HTTP:        &http.Client{Timeout: time.Duration(timeout) * time.Second},
		TokenStore:  tokenStore,
		OAuthToken:  t,
	}
}

// ensureToken refreshes if expired.
func (c *HAClient) ensureToken() error {
	if c.OAuthToken.IsValid() {
		c.AccessToken = c.OAuthToken.AccessToken
		return nil
	}
	if c.OAuthToken.RefreshToken == "" {
		return fmt.Errorf("token expired, run 'm login' to re-authorize")
	}
	oc := NewOAuthClient()
	oc.CloudServer = c.OAuthToken.CloudServer
	oc.ClientID = c.OAuthToken.OAuthClientID
	oc.RedirectURI = c.OAuthToken.OAuthRedirect
	oc.DeviceID = c.OAuthToken.DeviceID
	oc.State = c.OAuthToken.State
	newT, err := oc.RefreshToken(c.OAuthToken.RefreshToken)
	if err != nil {
		return err
	}
	c.OAuthToken = newT
	c.AccessToken = newT.AccessToken
	if c.TokenStore != nil {
		_ = c.TokenStore.SaveOAuth(newT)
	}
	return nil
}

// Post sends POST to path with JSON body, returns parsed result.
func (c *HAClient) Post(path string, data interface{}) (map[string]interface{}, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}
	var body []byte
	if data != nil {
		var err error
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	url := c.BaseURL + path
	logHttpReq("POST", url, body)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
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
		// token invalid, clear and retry once
		c.OAuthToken.AccessToken = ""
		c.OAuthToken.ExpiresTS = 0
		if err := c.ensureToken(); err != nil {
			return nil, err
		}
		return c.Post(path, data)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(raw))
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if code, ok := out["code"].(float64); ok && code != 0 {
		return nil, fmt.Errorf("api error %.0f: %s", code, out["message"])
	}
	return out, nil
}

func (c *HAClient) setHeaders(req *http.Request) {
	req.Header.Set("Host", c.Host)
	req.Header.Set("X-Client-BizId", "haapi")
	req.Header.Set("Authorization", "Bearer"+c.AccessToken)
	req.Header.Set("X-Client-AppId", c.ClientID)
}
