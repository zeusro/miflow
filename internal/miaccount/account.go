package miaccount

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const accountBase = "https://account.xiaomi.com/pass/"

var userAgents = []string{
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15",
	"MiHome/6.0.103 (com.xiaomi.mihome; build:6.0.103.1; iOS 14.4.0)",
}

func init() { rand.Seed(time.Now().UnixNano()) }

// RandString returns a random string of length n (for deviceId, requestId, etc.).
func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Account performs Xiaomi account login and signed requests.
type Account struct {
	Client     *http.Client
	Username   string
	Password   string
	TokenStore *TokenStore
	Token      *Token
	UserAgent  string
}

// NewAccount creates an account with optional token store path.
func NewAccount(username, password, tokenPath string) *Account {
	a := &Account{
		Client:    &http.Client{Timeout: 30 * time.Second},
		Username:  username,
		Password:  password,
		UserAgent: userAgents[rand.Intn(len(userAgents))],
	}
	if tokenPath != "" {
		a.TokenStore = &TokenStore{Path: tokenPath}
		a.Token = a.TokenStore.Load()
	}
	if a.Token == nil {
		a.Token = &Token{DeviceID: strings.ToUpper(RandString(16)), Services: make(map[string][]string)}
	}
	return a
}

// Login for the given service id (e.g. "micoapi", "xiaomiio"). Uses token file if set.
func (a *Account) Login(sid string) error {
	resp, err := a.serviceLogin("serviceLogin?sid=" + sid + "&_json=true")
	if err != nil {
		return err
	}
	code, _ := resp["code"].(float64)
	if code != 0 {
		// Auth required
		hash := fmt.Sprintf("%X", md5.Sum([]byte(a.Password)))
		data := map[string]string{
			"_json":    "true",
			"qs":       str(resp["qs"]),
			"sid":      str(resp["sid"]),
			"_sign":    str(resp["_sign"]),
			"callback": str(resp["callback"]),
			"user":     a.Username,
			"hash":     hash,
		}
		resp2, err := a.serviceLoginPost("serviceLoginAuth2", data)
		if err != nil {
			return err
		}
		code2, _ := resp2["code"].(float64)
		if code2 != 0 {
			return fmt.Errorf("login: %v", resp2)
		}
		a.Token.UserID = str(resp2["userId"])
		a.Token.PassToken = str(resp2["passToken"])
		location := str(resp2["location"])
		nonce := resp2["nonce"]
		ssecurity := str(resp2["ssecurity"])
		serviceToken, err := a.securityTokenService(location, nonce, ssecurity)
		if err != nil {
			return err
		}
		if a.Token.Services == nil {
			a.Token.Services = make(map[string][]string)
		}
		a.Token.Services[sid] = []string{ssecurity, serviceToken}
		if a.TokenStore != nil {
			_ = a.TokenStore.Save(a.Token)
		}
	}
	return nil
}

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func (a *Account) serviceLogin(uri string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", accountBase+uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.UserAgent)
	req.AddCookie(&http.Cookie{Name: "sdkVersion", Value: "3.9"})
	req.AddCookie(&http.Cookie{Name: "deviceId", Value: a.Token.DeviceID})
	if a.Token.PassToken != "" {
		req.AddCookie(&http.Cookie{Name: "userId", Value: a.Token.UserID})
		req.AddCookie(&http.Cookie{Name: "passToken", Value: a.Token.PassToken})
	} else {
		req.AddCookie(&http.Cookie{Name: "passToken", Value: ""})
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Response may be JSONP: "&&&START&&&{...}"
	body := raw
	if len(body) > 11 && string(body[:11]) == "&&&START&&&" {
		body = body[11:]
	}
	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *Account) serviceLoginPost(uri string, data map[string]string) (map[string]interface{}, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	req, err := http.NewRequest("POST", accountBase+uri, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "sdkVersion", Value: "3.9"})
	req.AddCookie(&http.Cookie{Name: "deviceId", Value: a.Token.DeviceID})
	req.AddCookie(&http.Cookie{Name: "userId", Value: a.Token.UserID})
	req.AddCookie(&http.Cookie{Name: "passToken", Value: a.Token.PassToken})
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := raw
	if len(body) > 11 && string(body[:11]) == "&&&START&&&" {
		body = body[11:]
	}
	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *Account) securityTokenService(location string, nonce interface{}, ssecurity string) (string, error) {
	nonceStr := fmt.Sprint(nonce)
	nsec := "nonce=" + nonceStr + "&" + ssecurity
	h := sha1.Sum([]byte(nsec))
	clientSign := base64.StdEncoding.EncodeToString(h[:])
	u := location + "&clientSign=" + url.QueryEscape(clientSign)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", a.UserAgent)
	resp, err := a.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	for _, c := range resp.Cookies() {
		if c.Name == "serviceToken" {
			return c.Value, nil
		}
	}
	body, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("no serviceToken: %s", string(body))
}

// Request performs a signed request to a Xiaomi service. data can be nil (GET) or body bytes or a func that returns (body, nil).
// If relogin is true and response is 401, clears token and retries once after Login(sid).
func (a *Account) Request(sid, method, requestURL string, data interface{}, relogin bool) ([]byte, error) {
	if _, ok := a.Token.Services[sid]; !ok {
		if err := a.Login(sid); err != nil {
			return nil, err
		}
	}
	var body io.Reader
	switch d := data.(type) {
	case nil:
	case []byte:
		body = bytes.NewReader(d)
	case string:
		body = strings.NewReader(d)
	default:
		return nil, fmt.Errorf("unsupported data type")
	}
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", a.UserAgent)
	req.AddCookie(&http.Cookie{Name: "userId", Value: a.Token.UserID})
	req.AddCookie(&http.Cookie{Name: "serviceToken", Value: a.Token.Services[sid][1]})
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 && relogin {
		a.Token.Services[sid] = nil
		delete(a.Token.Services, sid)
		if a.TokenStore != nil {
			_ = a.TokenStore.Save(a.Token)
		}
		return a.Request(sid, method, requestURL, data, false)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(out))
	}
	return out, nil
}

// RequestJSON parses response as JSON and returns code; if code!=0 returns error.
func (a *Account) RequestJSON(sid, method, requestURL string, data interface{}, relogin bool) (map[string]interface{}, error) {
	raw, err := a.Request(sid, method, requestURL, data, relogin)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if code, ok := m["code"].(float64); ok && code != 0 {
		if msg, _ := m["message"].(string); msg != "" && strings.Contains(strings.ToLower(msg), "auth") {
			return nil, fmt.Errorf("auth error: %s", msg)
		}
		return nil, fmt.Errorf("api error %.0f: %v", code, m)
	}
	return m, nil
}
