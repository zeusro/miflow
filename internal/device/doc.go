// Package device 封装接入设备的 API，基于 m list 导出的设备列表与 docs/spec.md 的 SPEC 查询流程。
//
// 设备列表格式与 m list 输出一致：did、model、name、token。
// SPEC 查询遵循 miot-spec.org：先 instances 获取 model→URN，再 instance 获取完整规格。
package device
