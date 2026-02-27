// Package i18n provides internationalization for miflow CLI and web.
package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func jsonUnmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

//go:embed locale/*.json
var localeFS embed.FS

var bundle *i18n.Bundle

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", jsonUnmarshal)
	bundle.LoadMessageFileFS(localeFS, "locale/active.en.json")
	bundle.LoadMessageFileFS(localeFS, "locale/active.zh-CN.json")
}

// T returns the localized string for the given message ID.
// lang can be "en", "zh-CN", "zh", or empty (uses LANG env, falls back to en).
func T(lang, msgID string, templateData map[string]interface{}) string {
	if lang == "" {
		lang = DefaultLang()
	}
	accept := lang
	if accept == "zh" {
		accept = "zh-CN"
	}
	loc := i18n.NewLocalizer(bundle, accept, "en")
	s, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: templateData,
	})
	if err != nil {
		return msgID
	}
	return s
}

// DefaultLang returns the default language from LANG/LC_ALL env, or "en".
func DefaultLang() string {
	for _, key := range []string{"LC_ALL", "LANG", "LANGUAGE"} {
		if v := os.Getenv(key); v != "" {
			// e.g. zh_CN.UTF-8 -> zh-CN, en_US.UTF-8 -> en
			v = strings.Split(v, ".")[0]
			v = strings.ReplaceAll(v, "_", "-")
			if strings.HasPrefix(v, "zh") {
				return "zh-CN"
			}
			if strings.HasPrefix(v, "en") {
				return "en"
			}
			return v
		}
	}
	return "en"
}

// AcceptLanguage parses Accept-Language header and returns preferred lang (e.g. "zh-CN", "en").
func AcceptLanguage(header string) string {
	// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
	parts := strings.Split(header, ",")
	for _, p := range parts {
		p = strings.TrimSpace(strings.Split(p, ";")[0])
		p = strings.ReplaceAll(p, "_", "-")
		if strings.HasPrefix(p, "zh") {
			return "zh-CN"
		}
		if strings.HasPrefix(p, "en") {
			return "en"
		}
	}
	return "en"
}
