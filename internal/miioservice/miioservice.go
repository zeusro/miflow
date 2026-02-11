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
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeusro/miflow/internal/miaccount"
)

const sidXiaomiIO = "xiaomiio"

// Service implements MiIO/MIoT API (xiaomiio).
type Service struct {
	Account *miaccount.Account
	Server  string
}

// New creates MiIO service. Region can be "" or "cn" for China.
func New(account *miaccount.Account, region string) *Service {
	server := "https://api.io.mi.com/app"
	if region != "" && region != "cn" {
		server = "https://" + region + ".api.io.mi.com/app"
	}
	return &Service{Account: account, Server: server}
}

// SignNonce computes key for request signing: base64(sha256(b64decode(ssecurity)+b64decode(nonce)))
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

// SignData builds _nonce, data, signature for MiIO request.
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

// MiIORequest sends signed request to MiIO server.
func (s *Service) MiIORequest(uri string, data interface{}) (map[string]interface{}, error) {
	if _, ok := s.Account.Token.Services[sidXiaomiIO]; !ok {
		if err := s.Account.Login(sidXiaomiIO); err != nil {
			return nil, err
		}
	}
	ssecurity := s.Account.Token.Services[sidXiaomiIO][0]
	var dataStr string
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		dataStr = string(b)
	}
	params, err := SignData(uri, dataStr, ssecurity)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("_nonce", params["_nonce"])
	form.Set("data", params["data"])
	form.Set("signature", params["signature"])
	req, err := http.NewRequest("POST", s.Server+uri, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "iOS-14.4-6.0.103-iPhone12,3--D7744744F7AF32F0544445285880DD63E47D9BE9-8816080-84A3F44E137B71AE-iPhone")
	req.Header.Set("x-xiaomi-protocal-flag-cli", "PROTOCAL-HTTP2")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "PassportDeviceId", Value: s.Account.Token.DeviceID})
	req.AddCookie(&http.Cookie{Name: "userId", Value: s.Account.Token.UserID})
	req.AddCookie(&http.Cookie{Name: "serviceToken", Value: s.Account.Token.Services[sidXiaomiIO][1]})
	resp, err := s.Account.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if result, ok := out["result"]; ok {
		if m, ok := result.(map[string]interface{}); ok {
			return m, nil
		}
		return nil, fmt.Errorf("unexpected result: %v", result)
	}
	return nil, fmt.Errorf("request error: %s", string(raw))
}

// HomeRequest calls /home/rpc/{did}.
func (s *Service) HomeRequest(did, method string, params interface{}) (map[string]interface{}, error) {
	return s.MiIORequest("/home/rpc/"+did, map[string]interface{}{
		"id":       1,
		"method":  method,
		"params":  params,
		"accessKey": "IOS00026747c5acafc2",
	})
}

// HomeGetProps returns property values.
func (s *Service) HomeGetProps(did string, props []string) ([]interface{}, error) {
	r, err := s.HomeRequest(did, "get_prop", props)
	if err != nil {
		return nil, err
	}
	var out []interface{}
	for _, p := range props {
		if v, ok := r[p]; ok {
			out = append(out, v)
		} else {
			out = append(out, nil)
		}
	}
	return out, nil
}

// HomeSetProp sets one property.
func (s *Service) HomeSetProp(did, prop string, value interface{}) (interface{}, error) {
	val := value
	if _, ok := value.([]interface{}); !ok {
		val = []interface{}{value}
	}
	r, err := s.HomeRequest(did, "set_"+prop, val)
	if err != nil {
		return nil, err
	}
	if len(r) > 0 {
		return r["0"], nil
	}
	return "ok", nil
}

// MiotRequest calls /miotspec/{cmd}.
func (s *Service) MiotRequest(cmd string, params interface{}) ([]map[string]interface{}, error) {
	r, err := s.MiIORequest("/miotspec/"+cmd, map[string]interface{}{"params": params})
	if err != nil {
		return nil, err
	}
	// result is array
	if arr, ok := r["result"].([]interface{}); ok {
		out := make([]map[string]interface{}, 0, len(arr))
		for _, it := range arr {
			if m, ok := it.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("miot result not array: %v", r)
}

// MiotGetProps gets MIoT properties. iids are [siid, piid] pairs.
func (s *Service) MiotGetProps(did string, iids [][2]int) ([]interface{}, error) {
	params := make([]map[string]interface{}, len(iids))
	for i, iid := range iids {
		params[i] = map[string]interface{}{"did": did, "siid": iid[0], "piid": iid[1]}
	}
	result, err := s.MiotRequest("prop/get", params)
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

// MiotSetProps sets MIoT properties. props are [siid, piid, value].
func (s *Service) MiotSetProps(did string, props [][3]interface{}) ([]int, error) {
	params := make([]map[string]interface{}, len(props))
	for i, p := range props {
		params[i] = map[string]interface{}{"did": did, "siid": p[0], "piid": p[1], "value": p[2]}
	}
	result, err := s.MiotRequest("prop/set", params)
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
	result, err := s.MiotRequest("action", map[string]interface{}{
		"did":  did,
		"siid": siid,
		"aiid": aiid,
		"in":   args,
	})
	if err != nil || len(result) == 0 {
		return -1, err
	}
	code, _ := result[0]["code"].(float64)
	return int(code), nil
}

// DeviceList returns devices. name: "" for all, "full" for full info, or keyword; getVirtualModel, getHuamiDevices.
func (s *Service) DeviceList(name string, getVirtualModel bool, getHuamiDevices int) ([]map[string]interface{}, error) {
	r, err := s.MiIORequest("/home/device_list", map[string]interface{}{
		"getVirtualModel": getVirtualModel,
		"getHuamiDevices": getHuamiDevices,
	})
	if err != nil {
		return nil, err
	}
	list, _ := r["list"].([]interface{})
	out := make([]map[string]interface{}, 0, len(list))
	for _, it := range list {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		if name == "full" {
			out = append(out, m)
			continue
		}
		if name != "" {
			did, _ := m["did"].(string)
			n, _ := m["name"].(string)
			if !strings.Contains(did, name) && !strings.Contains(n, name) {
				continue
			}
		}
		out = append(out, map[string]interface{}{
			"name":   m["name"],
			"model":  m["model"],
			"did":    m["did"],
			"token":  m["token"],
		})
	}
	return out, nil
}

// MiotSpec fetches MIoT spec for type (model keyword or URN). format: "text", "python", "json".
func (s *Service) MiotSpec(typ, format string) (interface{}, error) {
	specsPath := filepath.Join(os.TempDir(), "miservice_miot_specs.json")
	allSpecs := make(map[string]string)
	if data, err := os.ReadFile(specsPath); err == nil {
		json.Unmarshal(data, &allSpecs)
	}
	if len(allSpecs) == 0 {
		resp, err := s.Account.Client.Get("http://miot-spec.org/miot-spec-v2/instances?status=all")
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
	// filter by typ
	if typ != "" && !strings.HasPrefix(typ, "urn:") {
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
		} else {
			return filtered, nil
		}
	}
	if typ == "" {
		return allSpecs, nil
	}
	// fetch instance spec
	if !strings.HasPrefix(typ, "urn:") {
		for _, t := range allSpecs {
			if t == typ || strings.Contains(t, typ) {
				typ = t
				break
			}
		}
	}
	reqURL := "http://miot-spec.org/miot-spec-v2/instance?type=" + url.QueryEscape(typ)
	resp, err := s.Account.Client.Get(reqURL)
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

func formatMiotSpecText(result map[string]interface{}, format, url string) string {
	var buf bytes.Buffer
	buf.WriteString("# Generated by github.com/zeusro/miflow\n# ")
	buf.WriteString(url)
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
