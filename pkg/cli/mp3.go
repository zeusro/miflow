package cli

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/mp3server"
	"github.com/zeusro/miflow/pkg/i18n"
)

// NewMp3Cmd 创建 mp3 子命令。
func NewMp3Cmd() *cobra.Command {
	var addr, host string
	cfg := config.Get()

	mp3 := &cobra.Command{
		Use:   "mp3 [options] <file_path>",
		Short: "Map local music files to HTTP URLs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMp3(addr, host, args[0])
		},
	}

	lang := i18n.DefaultLang()
	mp3.Flags().StringVar(&addr, "addr", cfg.Mp3.Addr, i18n.T(lang, "mp3.addr_desc", nil))
	mp3.Flags().StringVar(&host, "host", cfg.Mp3.Host, i18n.T(lang, "mp3.host_desc", nil))

	return mp3
}

func runMp3(addr, host, filePath string) error {
	lang := i18n.DefaultLang()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	srv, err := mp3server.New(mp3server.Config{
		Addr:       addr,
		Host:       host,
		LogRequest: true,
	}, "/")
	if err != nil {
		return err
	}
	if err := srv.Start(); err != nil {
		return err
	}
	defer srv.Close()

	playURL, err := srv.PathToURL(filePath)
	if err != nil {
		return err
	}
	absPath, _ := filepath.Abs(filePath)
	log.Println(i18n.T(lang, "mp3.mapped", map[string]interface{}{"Local": absPath, "URL": playURL}))
	fmt.Println(playURL)
	if srv.Host() == "127.0.0.1" {
		fmt.Fprintln(os.Stderr, i18n.T(lang, "mp3.host_hint", nil))
	}

	if !srv.WaitReady(5 * time.Second) {
		return fmt.Errorf("%s", i18n.T(lang, "mp3.port_not_ready", map[string]interface{}{"Port": srv.Port()}))
	}
	log.Println(i18n.T(lang, "mp3.ready", map[string]interface{}{"Port": srv.Port()}))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	return nil
}
