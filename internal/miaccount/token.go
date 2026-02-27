package miaccount

import (
	"encoding/json"
	"os"
)

// TokenStore 持久化并加载小米认证 token。
type TokenStore struct {
	Path string
}

// Load 从文件读取 token。文件不存在或无效时返回 nil。
func (s *TokenStore) Load() *Token {
	if s == nil || s.Path == "" {
		return nil
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil
	}
	var t Token
	if err := json.Unmarshal(data, &t); err != nil {
		return nil
	}
	return &t
}


// Save 将 token 写入文件。token 为 nil 时删除文件。
func (s *TokenStore) Save(t *Token) error {
	if s == nil || s.Path == "" {
		return nil
	}
	if t == nil {
		_ = os.Remove(s.Path)
		return nil
	}
	data, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0600)
}

// SaveOAuth 将 OAuth token 写入文件。
func (s *TokenStore) SaveOAuth(t *OAuthToken) error {
	if s == nil || s.Path == "" {
		return nil
	}
	if t == nil {
		_ = os.Remove(s.Path)
		return nil
	}
	data, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0600)
}

// LoadOAuth 从文件读取 OAuth token。未找到或无效时返回 nil。
func (s *TokenStore) LoadOAuth() *OAuthToken {
	if s == nil || s.Path == "" {
		return nil
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil
	}
	var t OAuthToken
	if err := json.Unmarshal(data, &t); err != nil {
		return nil
	}
	if t.AccessToken == "" && t.RefreshToken == "" {
		return nil
	}
	return &t
}
