package specs

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const InstanceURL = "http://miot-spec.org/miot-spec-v2/instance"

// FetchInstance 从 miot-spec.org 获取指定 URN 的完整规格 JSON（无需认证）。
func FetchInstance(urn string) (map[string]interface{}, error) {
	reqURL := InstanceURL + "?type=" + url.QueryEscape(urn)
	resp, err := http.DefaultClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
