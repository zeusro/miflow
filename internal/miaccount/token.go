package miaccount

import (
	"encoding/json"
	"os"
)

// TokenStore persists and loads Xiaomi auth token.
type TokenStore struct {
	Path string
}

// Load reads token from file. Returns nil if file missing or invalid.
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


// Save writes token to file. If token is nil, removes file.
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
