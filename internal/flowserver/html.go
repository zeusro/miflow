package flowserver

import (
	"embed"
	"strings"

	"github.com/zeusro/miflow/pkg/i18n"
)

//go:embed static/index.html
var indexHTMLFS embed.FS

// BuildIndexHTML 根据语言返回带 i18n 的首页 HTML。
func BuildIndexHTML(lang string) string {
	t := func(id string, data map[string]interface{}) string {
		return i18n.T(lang, id, data)
	}
	html, _ := indexHTMLFS.ReadFile("static/index.html")
	repl := map[string]string{
		"MiFlow 可视化控制流":     t("flow.ui.title", nil),
		"基于 <code>m</code> / MiNA 的可视化编排（后端 Go，前端极简版，可按需自定义）": t("flow.ui.subtitle", nil),
		"MiFlow · 设备工作控制流": t("flow.ui.title", nil),
		"新建 Flow":              t("flow.ui.new_flow", nil),
		"刷新":                   t("flow.ui.refresh", nil),
		"还没有配置任何 Flow。点击右上角「新建 Flow」开始。": t("flow.ui.no_flows_hint", nil),
		"名称":                   t("flow.ui.name", nil),
		"例如：早安流程 / 回家流程": t("flow.ui.name_placeholder", nil),
		"描述":                   t("flow.ui.description", nil),
		"这个 Flow 的用途说明":     t("flow.ui.desc_placeholder", nil),
		"步骤列表":                t("flow.ui.steps", nil),
		"添加步骤":                t("flow.ui.add_step", nil),
		"类型":                   t("flow.ui.type", nil),
		"设备 (可选)":             t("flow.ui.device_optional", nil),
		"参数":                   t("flow.ui.params", nil),
		"操作":                   t("flow.ui.actions", nil),
		"类型说明：":               t("flow.ui.type_help", nil),
		"运行 Flow":              t("flow.ui.run_flow", nil),
		"保存":                   t("flow.ui.save", nil),
		"暂无 Flow":              t("flow.ui.no_flow", nil),
		"(未命名 Flow)":          t("flow.ui.unnamed_flow", nil),
		"确认删除该 Flow？":        t("flow.ui.confirm_delete_flow", nil),
		"已删除 Flow":            t("flow.ui.deleted", nil),
		"删除失败":               t("flow.ui.delete_failed", nil),
		"加载失败":               t("flow.ui.load_failed", nil),
		"加载 Flow 失败":          t("flow.ui.load_failed_msg", nil),
		"已保存":                 t("flow.ui.saved", nil),
		"保存失败":               t("flow.ui.save_failed", nil),
		"请先保存 Flow 再运行":     t("flow.ui.save_first_run", nil),
		"已触发运行（在服务器日志中查看执行情况）": t("flow.ui.run_triggered_msg", nil),
		"运行失败":               t("flow.ui.run_failed", nil),
		"留空则使用 MI_DID":       t("flow.ui.device_optional_hint", nil),
		"等待毫秒数，例如 1000":     t("flow.ui.delay_hint", nil),
		"播报文本":               t("flow.ui.tts_hint", nil),
		"音频 URL":               t("flow.ui.audio_hint", nil),
		"等价 m 命令的参数，例如: 1,1-2=#60": t("flow.ui.miio_hint", nil),
	}
	s := string(html)
	for old, new := range repl {
		s = strings.ReplaceAll(s, old, new)
	}
	return s
}
