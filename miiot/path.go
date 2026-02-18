package miiot

import "strings"

// ModelToPath 将 model 转为文件路径，规则：vendor.category.suffix -> vendor/category/suffix.go
// 例如 xiaomi.wifispeaker.oh2 -> xiaomi/wifispeaker/oh2.go
func ModelToPath(model string) string {
	parts := strings.Split(model, ".")
	if len(parts) < 3 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/") + "/" + parts[len(parts)-1] + ".go"
}
