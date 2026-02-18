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
)

// Service implements MiIO/MIoT API via ha.api.io.mi.com (OAuth 2.0).
// Ref: https://github.com/XiaoMi/ha_xiaomi_home
type Service struct {
	ha *mihomeapi.Service
}

// New creates MiIO service with OAuth token. Token is loaded from tokenPath if empty.
func New(token *miaccount.OAuthToken, tokenPath string) (*Service, error) {
	if token == nil {
		store := &miaccount.TokenStore{Path: tokenPath}
		token = store.LoadOAuth()
	}
	if token == nil || !token.IsValid() {
		return nil, fmt.Errorf("no valid OAuth token, run 'm login' first")
	}
	ha, err := mihomeapi.New(token, tokenPath)
	if err != nil {
		return nil, err
	}
	return &Service{ha: ha}, nil
}

// SignNonce computes key for request signing (used by MiotDecode).
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

// SignData builds _nonce, data, signature for MiIO request (legacy, kept for decode).
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

// MiIORequest: raw MiIO not supported in OAuth mode (ha.api.io.mi.com uses different API).
func (s *Service) MiIORequest(uri string, data interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("raw MiIO (%s) not supported in OAuth mode, use MIoT commands", uri)
}

// HomeRequest: legacy home/rpc not supported in OAuth mode.
func (s *Service) HomeRequest(did, method string, params interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("legacy home/rpc not supported in OAuth mode")
}

// HomeGetProps returns property values. In OAuth mode, uses MIoT get_prop.
func (s *Service) HomeGetProps(did string, props []string) ([]interface{}, error) {
	// Map legacy prop names to siid-piid would need device spec; return nil for now
	return nil, fmt.Errorf("legacy get_prop not supported, use MIoT format (siid-piid)")
}

// HomeSetProp: legacy not supported.
func (s *Service) HomeSetProp(did, prop string, value interface{}) (interface{}, error) {
	return nil, fmt.Errorf("legacy set_prop not supported, use MIoT format")
}

// MiotRequest calls miotspec prop/get, prop/set, or action via HA API.
func (s *Service) MiotRequest(cmd string, params interface{}) ([]map[string]interface{}, error) {
	switch cmd {
	case "prop/get":
		pm, ok := toParamsArray(params)
		if !ok {
			return nil, fmt.Errorf("prop/get expects params array")
		}
		res, err := s.ha.GetProps(pm)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "prop/set":
		pm, ok := toParamsArray(params)
		if !ok {
			return nil, fmt.Errorf("prop/set expects params array")
		}
		_, err := s.ha.SetProps(pm)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case "action":
		p, ok := params.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("action expects params object with did, siid, aiid, in")
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
		return nil, fmt.Errorf("unknown miot cmd: %s", cmd)
	}
}

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

func toInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	}
	return 0, false
}

// MiotGetProps gets MIoT properties.
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

// MiotSetProps sets MIoT properties.
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

// MiotAction runs a MIoT action.
func (s *Service) MiotAction(did string, siid, aiid int, args []interface{}) (int, error) {
	_, err := s.ha.Action(did, siid, aiid, args)
	if err != nil {
		return -1, err
	}
	return 0, nil
}

// DeviceList returns devices.
func (s *Service) DeviceList(name string, getVirtualModel bool, getHuamiDevices int) ([]map[string]interface{}, error) {
	return s.ha.DeviceList(name, getVirtualModel, getHuamiDevices)
}

// MiotSpec fetches MIoT spec from miot-spec.org (public, no auth).
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
		resp, err := http.DefaultClient.Get("http://miot-spec.org/miot-spec-v2/instances?status=all")
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
	reqURL := "http://miot-spec.org/miot-spec-v2/instance?type=" + url.QueryEscape(typ)
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

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

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

func toSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

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

// MiotDecode decrypts MIoT payload with ssecurity and nonce. If gzip is true, decompress after decrypt.
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
	// Discard first 1024 bytes of keystream (MiIO)
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
