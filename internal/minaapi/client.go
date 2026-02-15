// Package minaapi implements MiNA API client for api2.mina.mi.com.
// Ref: https://github.com/yihong0618/MiService/blob/main/miservice/minaservice.py
package minaapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/zeusro/miflow/internal/miaccount"
)

const (
	minaBaseURL = "https://api2.mina.mi.com"
	userAgent   = "MiHome/6.0.103 (com.xiaomi.mihome; build:6.0.103.1; iOS 14.4.0) Alamofire/6.0.103 MICO/iOSApp/appStore/6.0.103"
)

// Client calls MiNA API. Supports OAuth Bearer token (from m login).
// Note: MiNA API may require micoapi cookie auth; OAuth is experimental.
type Client struct {
	HTTP        *http.Client
	TokenStore  *miaccount.TokenStore
	OAuthToken  *miaccount.OAuthToken
	AccessToken string
}

// New creates client with OAuth token.
func New(t *miaccount.OAuthToken, tokenPath string) *Client {
	store := &miaccount.TokenStore{Path: tokenPath}
	token := t
	if token == nil {
		token = store.LoadOAuth()
	}
	accessToken := ""
	if token != nil {
		accessToken = token.AccessToken
	}
	return &Client{
		HTTP:        &http.Client{Timeout: 30 * time.Second},
		TokenStore:  store,
		OAuthToken:  token,
		AccessToken: accessToken,
	}
}

func (c *Client) ensureToken() error {
	if c.OAuthToken == nil || c.OAuthToken.AccessToken == "" {
		return fmt.Errorf("no OAuth token, run 'm login' first")
	}
	if !c.OAuthToken.IsValid() && c.OAuthToken.RefreshToken != "" {
		oc := miaccount.NewOAuthClient()
		oc.CloudServer = c.OAuthToken.CloudServer
		oc.ClientID = c.OAuthToken.OAuthClientID
		if oc.ClientID == "" {
			oc.ClientID = miaccount.OAuth2ClientID
		}
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
	} else {
		c.AccessToken = c.OAuthToken.AccessToken
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// MinaRequest calls MiNA API. uri should start with /. Uses GET when data is nil.
func (c *Client) MinaRequest(uri string, data map[string]interface{}) (map[string]interface{}, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}
	requestID := "app_ios_" + randString(30)
	if data != nil {
		data["requestId"] = requestID
	} else {
		if strings.Contains(uri, "?") {
			uri += "&requestId=" + requestID
		} else {
			uri += "?requestId=" + requestID
		}
	}

	var body []byte
	var err error
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	reqURL := minaBaseURL + uri
	method := "POST"
	if data == nil {
		method = "GET"
	}
	req, err := http.NewRequest(method, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer"+c.AccessToken)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// API 可能返回 HTML 错误页（如 401/403），非 JSON
	trimmed := bytes.TrimLeft(raw, " \t\r\n\xef\xbb\xbf") // 含 UTF-8 BOM
	if len(trimmed) > 0 && trimmed[0] == '<' {
		snippet := string(trimmed)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return nil, fmt.Errorf("mina api: http %d, 返回 HTML 非 JSON（OAuth 可能不被支持，需 micoapi 认证）: %s", resp.StatusCode, snippet)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		bodyStr := truncate(string(raw), 150)
		if bytes.Contains(raw, []byte("<")) {
			return nil, fmt.Errorf("mina api: http %d, 返回 HTML 非 JSON（OAuth 可能不被 api2.mina.mi.com 支持，需 micoapi 认证）: %s", resp.StatusCode, bodyStr)
		}
		return nil, fmt.Errorf("mina api: %w (http %d, body: %s)", err, resp.StatusCode, bodyStr)
	}
	if code, ok := out["code"].(float64); ok && code != 0 {
		msg, _ := out["message"].(string)
		return nil, fmt.Errorf("mina api error %.0f: %s", code, msg)
	}
	return out, nil
}

// DeviceList returns mina devices. master=0 for all. Uses GET per MiService.
func (c *Client) DeviceList(master int) ([]map[string]interface{}, error) {
	uri := fmt.Sprintf("/admin/v2/device_list?master=%d", master)
	res, err := c.MinaRequest(uri, nil)
	if err != nil {
		return nil, err
	}
	data, _ := res["data"].([]interface{})
	if data == nil {
		return nil, fmt.Errorf("mina device_list: no data")
	}
	out := make([]map[string]interface{}, 0, len(data))
	for _, it := range data {
		if m, ok := it.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

// UbusRequest sends ubus RPC to device.
func (c *Client) UbusRequest(deviceID, method, path string, message interface{}) (map[string]interface{}, error) {
	msgJSON, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"deviceId": deviceID,
		"message":  string(msgJSON),
		"method":   method,
		"path":     path,
	}
	return c.MinaRequest("/remote/ubus", data)
}

// Hardware models that require play_by_music_url instead of player_play_url.
// Ref: MiService minaservice._USE_PLAY_MUSIC_API
var usePlayMusicAPI = map[string]bool{
	"LX04": true, "LX05": true, "L05B": true, "L05C": true, "L06": true,
	"L06A": true, "X08A": true, "X10A": true, "X08C": true, "X08E": true,
	"X8F": true, "X4B": true, "OH2": true, "OH2P": true, "X6A": true,
}

// PlayByURL plays audio from URL. Uses play_by_music_url for L06A etc., else player_play_url.
func (c *Client) PlayByURL(deviceID, url string, typ int) (map[string]interface{}, error) {
	// Resolve hardware from device list to choose API
	devices, err := c.DeviceList(0)
	if err != nil {
		return nil, err
	}
	hardware := ""
	for _, d := range devices {
		devID, _ := d["deviceID"].(string)
		if devID == "" {
			devID, _ = d["deviceId"].(string) // camelCase variant
		}
		if devID == deviceID {
			hardware, _ = d["hardware"].(string)
			break
		}
	}
	if usePlayMusicAPI[strings.ToUpper(hardware)] {
		return c.PlayByMusicURL(deviceID, url, typ)
	}
	msg := map[string]interface{}{
		"url":   url,
		"type":  typ,
		"media": "app_ios",
	}
	return c.UbusRequest(deviceID, "player_play_url", "mediaplayer", msg)
}

// PlayByMusicURL uses player_play_music for L06A/LX05 etc. Ref: MiService play_by_music_url.
func (c *Client) PlayByMusicURL(deviceID, url string, typ int) (map[string]interface{}, error) {
	audioID := "1582971365183456177"
	id := "355454500"
	audioType := ""
	if typ == 1 {
		audioType = "MUSIC"
	}
	music := map[string]interface{}{
		"payload": map[string]interface{}{
			"audio_type": audioType,
			"audio_items": []map[string]interface{}{
				{
					"item_id": map[string]interface{}{
						"audio_id": audioID,
						"cp": map[string]interface{}{
							"album_id": "-1", "episode_index": 0, "id": id,
							"name": "xiaowei",
						},
					},
					"stream": map[string]interface{}{"url": url},
				},
			},
			"list_params": map[string]interface{}{
				"listId": "-1", "loadmore_offset": 0, "origin": "xiaowei", "type": "MUSIC",
			},
		},
		"play_behavior": "REPLACE_ALL",
	}
	musicJSON, err := json.Marshal(music)
	if err != nil {
		return nil, err
	}
	msg := map[string]interface{}{
		"startaudioid": audioID,
		"music":        string(musicJSON),
	}
	return c.UbusRequest(deviceID, "player_play_music", "mediaplayer", msg)
}
