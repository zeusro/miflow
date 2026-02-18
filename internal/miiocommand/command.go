package miiocommand

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miioservice"
)

// Run parses text and runs the appropriate MiIO/MIoT command. did can be device ID or name.
// prefix is used in help (e.g. "m ").
func Run(svc *miioservice.Service, did, text, prefix string) (interface{}, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Help(did, prefix), nil
	}
	parts := strings.SplitN(text, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}

	// MiIO raw: /uri [data]
	if strings.HasPrefix(cmd, "/") {
		var data interface{}
		if arg != "" {
			if err := json.Unmarshal([]byte(arg), &data); err != nil {
				return nil, fmt.Errorf("invalid JSON: %w", err)
			}
		}
		return svc.MiIORequest(cmd, data)
	}

	// MIoT raw: prop/get, prop/set, action + JSON
	if cmd == "prop" || strings.HasPrefix(cmd, "prop/") || cmd == "action" {
		var params interface{}
		if arg != "" {
			if err := json.Unmarshal([]byte(arg), &params); err != nil {
				return nil, fmt.Errorf("invalid JSON: %w", err)
			}
		}
		return svc.MiotRequest(cmd, params)
	}

	argv := strings.Fields(arg)
	argc := len(argv)

	if cmd == "list" {
		var getVirtual bool
		var getHuami int
		name := ""
		if argc > 0 {
			name = argv[0]
		}
		if argc > 1 {
			getVirtual = parseBool(argv[1])
		}
		if argc > 2 {
			getHuami, _ = strconv.Atoi(argv[2])
		}
		return svc.DeviceList(name, getVirtual, getHuami)
	}

	if cmd == "spec" {
		typ := ""
		format := "text"
		if argc > 0 {
			typ = argv[0]
		}
		if argc > 1 {
			format = argv[1]
		}
		return svc.MiotSpec(typ, format)
	}

	if cmd == "spec_all" || cmd == "spec-all" {
		api := device.NewAPI(svc)
		specs, failed := api.LoadAllModelSpecs()
		ok := make(map[string]string)
		for m, s := range specs {
			if s != nil {
				ok[m] = s.Summary()
			}
		}
		failedStr := make(map[string]string)
		for m, e := range failed {
			if e != nil {
				failedStr[m] = e.Error()
			}
		}
		return map[string]interface{}{
			"ok":     ok,
			"failed": failedStr,
		}, nil
	}

	if cmd == "decode" {
		if argc < 3 {
			return nil, fmt.Errorf("decode requires: ssecurity nonce data [gzip]")
		}
		gzip := argc > 3 && argv[3] == "gzip"
		return miioservice.MiotDecode(argv[0], argv[1], argv[2], gzip)
	}

	if cmd == "?" || cmd == "？" || cmd == "help" || cmd == "-h" || cmd == "--help" {
		return Help(did, prefix), nil
	}

	// Resolve did to numeric if it's a name
	if did != "" && !isDigits(did) {
		devs, err := svc.DeviceList(did, false, 0)
		if err != nil || len(devs) == 0 {
			return nil, fmt.Errorf("device not found: %s", did)
		}
		if d, ok := devs[0]["did"].(string); ok {
			did = d
		}
	}

	if did == "" || cmd == "" {
		return Help(did, prefix), nil
	}

	// Parse comma-separated items: 1,1-2,2=#60,5-4 Hello #1
	items := strings.Split(cmd, ",")
	var props [][3]interface{} // get: [siid, piid], set: [siid, piid, value], action: [siid, aiid] + args
	setMode := true
	miot := true
	for _, item := range items {
		key, val := splitTwins(item, "=", "")
		siidStr, piidStr := splitTwins(key, "-", "1")
		if isDigits(siidStr) && isDigits(piidStr) {
			siid, _ := strconv.Atoi(siidStr)
			piid, _ := strconv.Atoi(piidStr)
			if val == "" {
				setMode = false
				props = append(props, [3]interface{}{siid, piid, nil})
			} else {
				props = append(props, [3]interface{}{siid, piid, stringOrValue(val)})
			}
		} else {
			miot = false
			break
		}
	}

	if miot && argc > 0 {
		// Action: siid-aiid arg1 arg2 ...
		args := make([]interface{}, 0, argc)
		if arg != "#NA" {
			for _, a := range argv {
				args = append(args, stringOrValue(a))
			}
		}
		siid, _ := props[0][0].(int)
		aiid, _ := props[0][1].(int)
		code, err := svc.MiotAction(did, siid, aiid, args)
		if err != nil {
			return nil, err
		}
		return code, nil
	}

	if setMode {
		if miot {
			setProps := make([][3]interface{}, 0, len(props))
			for _, p := range props {
				if p[2] != nil {
					setProps = append(setProps, p)
				}
			}
			return svc.MiotSetProps(did, setProps)
		}
		// Legacy home set
		var err error
		for _, p := range props {
			if p[2] != nil {
				_, err = svc.HomeSetProp(did, fmt.Sprintf("%v", p[0]), p[2])
				if err != nil {
					return nil, err
				}
			}
		}
		return "ok", nil
	}

	if miot {
		iids := make([][2]int, 0, len(props))
		for _, p := range props {
			siid, _ := p[0].(int)
			piid, _ := p[1].(int)
			iids = append(iids, [2]int{siid, piid})
		}
		return svc.MiotGetProps(did, iids)
	}
	// Legacy home get_prop
	propNames := make([]string, 0, len(props))
	for _, p := range props {
		propNames = append(propNames, fmt.Sprintf("%v", p[0]))
	}
	return svc.HomeGetProps(did, propNames)
}

func splitTwins(s, sep, defaultRight string) (string, string) {
	i := strings.Index(s, sep)
	if i < 0 {
		return s, defaultRight
	}
	return s[:i], s[i+len(sep):]
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func parseBool(s string) bool {
	return s == "true" || s == "1"
}

func stringOrValue(s string) interface{} {
	if s == "" {
		return s
	}
	if s[0] == '#' {
		switch s[1:] {
		case "null", "none":
			return nil
		case "false":
			return false
		case "true":
			return true
		default:
			n, _ := strconv.Atoi(s[1:])
			return n
		}
	}
	return s
}

// Help returns command help string.
func Help(did, prefix string) string {
	if did == "" {
		did = "267090026"
	}
	return fmt.Sprintf(`Get Props: %s [,...]
  %s1,1-2,1-3,2-1,2-2,3
Set Props: %s [,...]
  %s2=#60,2-2=#false,3=test
Do Action: %s [...] 
  %s2 #NA
  %s5 Hello
  %s5-4 Hello #1

Call MIoT: %s prop/get|prop/set|action <params>
  %saction '{"did":"%s","siid":5,"aiid":1,"in":["Hello"]}'

Call MiIO: %s/ <uri> <data>
  %s/home/device_list '{"getVirtualModel":false,"getHuamiDevices":1}'

Devs List: %slist [name=full|name_keyword] [getVirtualModel=false|true] [getHuamiDevices=0|1]
  %slist Light true 0

MIoT Spec: %sspec [model_keyword|type_urn] [format=text|python|json]
  %sspec speaker
  %sspec xiaomi.wifispeaker.lx04
  %sspec_all  获取 m list 中所有型号的 SPEC（按 docs/spec.md 流程）

MIoT Decode: %sdecode <ssecurity> <nonce> <data> [gzip]
`,
		prefix, prefix, prefix, prefix, prefix, prefix, prefix, prefix,
		prefix, prefix, did, prefix, prefix, prefix, prefix, prefix, prefix, prefix, prefix)
}
