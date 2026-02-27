// Package login 实现 m login 子命令（OAuth 2.0 流程）。
package login

import (
	"fmt"
	"os"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
)

// Login 运行 OAuth 2.0 登录流程。
type Login struct {
	TokenPath string
}

// Run 执行登录命令。
func (l Login) Run() {
	oc := miaccount.NewOAuthClient()
	authURL := oc.GenAuthURL("", "", true)
	fmt.Fprintf(os.Stderr, "Open this URL in browser to login:\n%s\n\n", authURL)
	callbackPort := config.Get().MiIO.CallbackPort
	if callbackPort <= 0 {
		callbackPort = 8123
	}
	if err := miaccount.OpenAuthURL(authURL); err != nil {
		fmt.Fprintln(os.Stderr, "(Could not open browser, open the URL manually)")
	}
	code, err := miaccount.ServeCallback(callbackPort)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	token, err := oc.GetToken(code)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	store := &miaccount.TokenStore{Path: l.TokenPath}
	if err := store.SaveOAuth(token); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "Login successful. Token saved to", l.TokenPath)
}
