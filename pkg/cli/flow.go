package cli

import (
	"github.com/spf13/cobra"
	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/flowserver"
	"github.com/zeusro/miflow/pkg/i18n"
)

// NewFlowCmd 创建 flow 子命令。
func NewFlowCmd() *cobra.Command {
	var addr, dataDir string
	cfg := config.Get()

	flow := &cobra.Command{
		Use:   "flow",
		Short: "Run Flow server (visual control flow UI)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return flowserver.Run(addr, dataDir)
		},
	}

	lang := i18n.DefaultLang()
	flow.Flags().StringVar(&addr, "addr", cfg.Flow.Addr, i18n.T(lang, "flow.addr_desc", nil))
	flow.Flags().StringVar(&dataDir, "data-dir", cfg.Flow.DataDir, i18n.T(lang, "flow.datadir_desc", nil))

	return flow
}
