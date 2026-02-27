package miioservice

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/mihomeapi"
	"github.com/zeusro/miflow/miiot/specs"
	"github.com/zeusro/miflow/pkg/i18n"
)

// Service 通过 ha.api.io.mi.com 实现 MiIO/MIoT API（OAuth 2.0）。
// 参考: https://github.com/XiaoMi/ha_xiaomi_home
type Service struct {
	ha *mihomeapi.Service
}

// New 使用 OAuth token 创建 MiIO 服务。token 为空时从 tokenPath 加载。
func New(token *miaccount.OAuthToken, tokenPath string) (*Service, error) {
	if token == nil {
		store := &miaccount.TokenStore{Path: tokenPath}
		token = store.LoadOAuth()
	}
	if token == nil || !token.IsValid() {
		return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.no_token", nil))
	}
	ha, err := mihomeapi.New(token, tokenPath)
	if err != nil {
		return nil, err
	}
	return &Service{ha: ha}, nil
}

// SignNonce 计算请求签名密钥（供 MiotDecode 使用）。
func SignNonce(ssecurity, nonce string) (string, error) {
	sb, err := base64.StdEncoding.DecodeString(ssecurity)
	if err != nil {
		return "", err
	}
	nb, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(append(sb, nb...))
	return base64.StdEncoding.EncodeToString(h[:]), nil
}

// SignData 构建 MiIO 请求的 _nonce、data、signature（旧版，保留用于 decode）。
func SignData(uri, dataStr, ssecurity string) (map[string]string, error) {
	nonceBytes := make([]byte, 12)
	rand.Read(nonceBytes[:8])
	binary.BigEndian.PutUint32(nonceBytes[8:12], uint32(time.Now().Unix()/60))
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)
	snonce, err := SignNonce(ssecurity, nonce)
	if err != nil {
		return nil, err
	}
	snonceBytes, _ := base64.StdEncoding.DecodeString(snonce)
	msg := uri + "&" + snonce + "&" + nonce + "&data=" + dataStr
	mac := hmac.New(sha256.New, snonceBytes)
	mac.Write([]byte(msg))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return map[string]string{
		"_nonce":    nonce,
		"data":      dataStr,
		"signature": sign,
	}, nil
}

// MiIORequest: OAuth 模式不支持原始 MiIO（ha.api.io.mi.com 使用不同 API）。
func (s *Service) MiIORequest(uri string, data interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.raw_not_supported", map[string]interface{}{"URI": uri}))
}

// HomeRequest: OAuth 模式不支持旧版 home/rpc。
func (s *Service) HomeRequest(did, method string, params interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.legacy_home_not_supported", nil))
}

// HomeGetProps 返回属性值。OAuth 模式使用 MIoT get_prop。
func (s *Service) HomeGetProps(did string, props []string) ([]interface{}, error) {
	// 将旧版 prop 名映射到 siid-piid 需要设备规格；暂返回 nil
	return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.legacy_get_prop", nil))
}

// HomeSetProp: 旧版不支持。
func (s *Service) HomeSetProp(did, prop string, value interface{}) (interface{}, error) {
	return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.legacy_set_prop", nil))
}

// MiotRequest 通过 HA API 调用 miotspec prop/get、prop/set 或 action。
func (s *Service) MiotRequest(cmd string, params interface{}) ([]map[string]interface{}, error) {
	switch cmd {
	case "prop/get":
		pm, ok := toParamsArray(params)
		if !ok {
			return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.prop_get_expects", nil))
		}
		res, err := s.ha.GetProps(pm)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "prop/set":
		pm, ok := toParamsArray(params)
		if !ok {
			return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.prop_set_expects", nil))
		}
		_, err := s.ha.SetProps(pm)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case "action":
		p, ok := params.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.action_expects", nil))
		}
		did, _ := p["did"].(string)
		siid, _ := toInt(p["siid"])
		aiid, _ := toInt(p["aiid"])
		inRaw, _ := p["in"].([]interface{})
		_, err := s.ha.Action(did, siid, aiid, inRaw)
		if err != nil {
			return nil, err
		}
		return []map[string]interface{}{{"code": float64(0)}}, nil
	default:
		return nil, fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "miio.unknown_cmd", map[string]interface{}{"Cmd": cmd}))
	}
}

// toParamsArray 将 interface 转为 []map[string]interface{}。
func toParamsArray(p interface{}) ([]map[string]interface{}, bool) {
	arr, ok := p.([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]map[string]interface{}, len(arr))
	for i, it := range arr {
		m, ok := it.(map[string]interface{})
		if !ok {
			return nil, false
		}
		out[i] = m
	}
	return out, true
}

// toInt 从 interface 提取 int（支持 float64/int）。
func toInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	}
	return 0, false
}

// MiotGetProps 获取 MIoT 属性。
func (s *Service) MiotGetProps(did string, iids [][2]int) ([]interface{}, error) {
	params := make([]map[string]interface{}, len(iids))
	for i, iid := range iids {
		params[i] = map[string]interface{}{"did": did, "siid": iid[0], "piid": iid[1]}
	}
	result, err := s.ha.GetProps(params)
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(result))
	for i, m := range result {
		if code, _ := m["code"].(float64); code == 0 {
			out[i] = m["value"]
		} else {
			out[i] = nil
		}
	}
	return out, nil
}

// MiotSetProps 设置 MIoT 属性。
func (s *Service) MiotSetProps(did string, props [][3]interface{}) ([]int, error) {
	params := make([]map[string]interface{}, len(props))
	for i, p := range props {
		params[i] = map[string]interface{}{"did": did, "siid": p[0], "piid": p[1], "value": p[2]}
	}
	result, err := s.ha.SetProps(params)
	if err != nil {
		return nil, err
	}
	out := make([]int, len(result))
	for i, m := range result {
		code, _ := m["code"].(float64)
		out[i] = int(code)
	}
	return out, nil
}

// MiotAction 执行 MIoT 动作。
func (s *Service) MiotAction(did string, siid, aiid int, args []interface{}) (int, error) {
	_, err := s.ha.Action(did, siid, aiid, args)
	if err != nil {
		return -1, err
	}
	return 0, nil
}

// DeviceList 返回设备列表。
func (s *Service) DeviceList(name string, getVirtualModel bool, getHuamiDevices int) ([]map[string]interface{}, error) {
	return s.ha.DeviceList(name, getVirtualModel, getHuamiDevices)
}

// MiotSpec 从 miot-spec.org 获取 MIoT 规格（公开，无需认证）。
func (s *Service) MiotSpec(typ, format string) (interface{}, error) {
	specsPath := config.Get().MiIO.SpecsCachePath
	if specsPath == "" {
		specsPath = filepath.Join(os.TempDir(), "miservice_miot_specs.json")
	}
	allSpecs := make(map[string]string)
	if data, err := os.ReadFile(specsPath); err == nil {
		json.Unmarshal(data, &allSpecs)
	}
	if len(allSpecs) == 0 {
		resp, err := http.DefaultClient.Get(specs.InstancesURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var inst struct {
			Instances []struct {
				Model string `json:"model"`
				Type  string `json:"type"`
			} `json:"instances"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&inst); err != nil {
			return nil, err
		}
		for _, i := range inst.Instances {
			allSpecs[i.Model] = i.Type
		}
		os.WriteFile(specsPath, mustJSON(allSpecs), 0644)
	}
	if typ != "" && !strings.HasPrefix(typ, "urn:") {
		// 精确匹配优先：若 typ 为完整 model 且存在于 allSpecs，直接取 URN
		if urn, ok := allSpecs[typ]; ok {
			typ = urn
		} else {
			filtered := make(map[string]string)
			for m, t := range allSpecs {
				if typ == m || strings.Contains(m, typ) {
					filtered[m] = t
				}
			}
			if len(filtered) == 1 {
				for _, t := range filtered {
					typ = t
					break
				}
			} else if len(filtered) > 1 {
				return filtered, nil
			}
		}
	}
	if typ == "" {
		return allSpecs, nil
	}
	if !strings.HasPrefix(typ, "urn:") {
		for _, t := range allSpecs {
			if t == typ || strings.Contains(t, typ) {
				typ = t
				break
			}
		}
	}
	reqURL := specs.InstanceURL + "?type=" + url.QueryEscape(typ)
	resp, err := http.DefaultClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if format == "json" {
		return result, nil
	}
	return formatMiotSpecText(result, format, reqURL), nil
}

// mustJSON 将 v 序列化为 JSON 字节。
func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// formatMiotSpecText 将 MIoT 规格格式化为 text 或 python 格式。
func formatMiotSpecText(result map[string]interface{}, format, reqURL string) string {
	var buf bytes.Buffer
	buf.WriteString("# Generated by github.com/zeusro/miflow\n# ")
	buf.WriteString(reqURL)
	buf.WriteString("\n\n")
	services, _ := result["services"].([]interface{})
	for _, s := range services {
		svc, _ := s.(map[string]interface{})
		siid, _ := svc["iid"].(float64)
		desc, _ := svc["description"].(string)
		svcName := strings.ReplaceAll(desc, " ", "_")
		if format == "python" {
			buf.WriteString(fmt.Sprintf("\nclass %s(tuple, Enum):\n", svcName))
		}
		for _, p := range toSlice(svc["properties"]) {
			prop, _ := p.(map[string]interface{})
			piid, _ := prop["iid"].(float64)
			pdesc, _ := prop["description"].(string)
			name, comment := parseDesc(pdesc)
			if format == "python" {
				buf.WriteString(fmt.Sprintf("  %s = (%d, %d)%s\n", name, int(siid), int(piid), comment))
			} else {
				buf.WriteString(fmt.Sprintf(" %s = %d%s\n", name, int(piid), comment))
			}
		}
		for _, a := range toSlice(svc["actions"]) {
			act, _ := a.(map[string]interface{})
			aiid, _ := act["iid"].(float64)
			adesc, _ := act["description"].(string)
			name, comment := parseDesc(adesc)
			if format == "python" {
				buf.WriteString(fmt.Sprintf("  %s = (%d, %d)%s\n", name, int(siid), int(aiid), comment))
			} else {
				buf.WriteString(fmt.Sprintf(" %s = %d%s\n", name, int(aiid), comment))
			}
		}
	}
	return buf.String()
}

// toSlice 将 interface 转为 []interface{}。
func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// parseDesc 解析描述字符串，提取名称和注释部分。
func parseDesc(desc string) (name, comment string) {
	for i, r := range desc {
		if r == '-' || r == '—' || r == '{' || r == '「' || r == '[' || r == '【' || r == '(' || r == '（' || r == '<' || r == '《' {
			return name, " # " + desc[i:]
		}
		if r == ' ' {
			name += "_"
		} else {
			name += string(r)
		}
	}
	return name, ""
}

// MiotDecode 使用 ssecurity 和 nonce 解密 MIoT 载荷。gzip 为 true 时解密后解压。
func MiotDecode(ssecurity, nonce, data string, gzip bool) (map[string]interface{}, error) {
	key, err := SignNonce(ssecurity, nonce)
	if err != nil {
		return nil, err
	}
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	cipher, err := newRC4Cipher(keyBytes)
	if err != nil {
		return nil, err
	}
	// 丢弃密钥流前 1024 字节（MiIO）
	discard := make([]byte, 1024)
	cipher.XORKeyStream(discard, discard)
	enc, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	dec := make([]byte, len(enc))
	cipher.XORKeyStream(dec, enc)
	if gzip {
		dec, err = gzipDecode(dec)
		if err != nil {
			return nil, err
		}
	}
	var out map[string]interface{}
	if err := json.Unmarshal(dec, &out); err != nil {
		return nil, err
	}
	return out, nil
}
