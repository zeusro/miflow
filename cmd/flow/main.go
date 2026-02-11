package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/minaservice"
)

// FlowStepType defines a single step kind inside a flow.
// This keeps the initial implementation simple but expressive enough
// for most "device work" control flows.
//
// Currently supported types:
//   - "tts"       : 小爱播报一段文字
//   - "play_url"  : 播放一个音频 URL
//   - "miio"      : 发送一条 miio/miot 文本命令（等价于 `m` 的参数）
//   - "delay"     : 等待一段时间（毫秒）
type FlowStepType string

const (
	StepTypeTTS     FlowStepType = "tts"
	StepTypePlayURL FlowStepType = "play_url"
	StepTypeMiIO    FlowStepType = "miio"
	StepTypeDelay   FlowStepType = "delay"

	defaultPrefix = "flow "
)

// FlowStep describes one action in a flow.
type FlowStep struct {
	Type       FlowStepType `json:"type"`
	Label      string       `json:"label,omitempty"`       // 简要说明，展示在 UI 上
	Device     string       `json:"device,omitempty"`      // 可选，覆盖环境变量 MI_DID
	Text       string       `json:"text,omitempty"`        // 用于 TTS
	URL        string       `json:"url,omitempty"`         // 用于 play_url
	MiIOText   string       `json:"miio_text,omitempty"`   // 用于 miio：等价于 `m` 的参数，如 "1,1-2=#60"
	DurationMS int          `json:"duration_ms,omitempty"` // 用于 delay
}

// Flow is a simple, linear flow made of steps.
type Flow struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Steps       []FlowStep `json:"steps"`
}

// FlowStore keeps flows on disk as a single JSON file.
type FlowStore struct {
	mu      sync.RWMutex
	path    string
	flows   []Flow
	loaded  bool
	modTime time.Time
}

func NewFlowStore(path string) *FlowStore {
	return &FlowStore{path: path}
}

func (s *FlowStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loaded {
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.flows = []Flow{}
			s.loaded = true
			return nil
		}
		return err
	}
	if err := json.Unmarshal(data, &s.flows); err != nil {
		return err
	}
	info, err := os.Stat(s.path)
	if err == nil {
		s.modTime = info.ModTime()
	}
	s.loaded = true
	return nil
}

func (s *FlowStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.flows, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return err
	}
	info, err := os.Stat(s.path)
	if err == nil {
		s.modTime = info.ModTime()
	}
	return nil
}

func (s *FlowStore) list() ([]Flow, error) {
	if err := s.load(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Flow, len(s.flows))
	copy(out, s.flows)
	return out, nil
}

func (s *FlowStore) upsert(f Flow) (Flow, error) {
	if f.ID == "" {
		// 简单的 ID 生成：时间戳 + 名称
		f.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeID(f.Name))
	}
	if err := s.load(); err != nil {
		return Flow{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, existing := range s.flows {
		if existing.ID == f.ID {
			s.flows[i] = f
			found = true
			break
		}
	}
	if !found {
		s.flows = append(s.flows, f)
	}
	if err := s.save(); err != nil {
		return Flow{}, err
	}
	return f, nil
}

func (s *FlowStore) get(id string) (Flow, bool, error) {
	if err := s.load(); err != nil {
		return Flow{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.flows {
		if f.ID == id {
			return f, true, nil
		}
	}
	return Flow{}, false, nil
}

func (s *FlowStore) delete(id string) error {
	if err := s.load(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	out := s.flows[:0]
	for _, f := range s.flows {
		if f.ID != id {
			out = append(out, f)
		}
	}
	s.flows = out
	return s.save()
}

func sanitizeID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "flow"
	}
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	return s
}

// app holds global state for the flow server.
type app struct {
	store      *FlowStore
	mina       *minaservice.Service
	miio       *miioservice.Service
	defaultDID string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	addr := flag.String("addr", ":18090", "HTTP 监听地址（用于可视化 Flow 配置）")
	dataDir := flag.String("data_dir", "./flowdata", "Flow 配置持久化目录")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}
	store := NewFlowStore(filepath.Join(*dataDir, "flows.json"))

	user := os.Getenv("MI_USER")
	pass := os.Getenv("MI_PASS")
	did := os.Getenv("MI_DID")
	if user == "" || pass == "" {
		log.Println("警告：未设置 MI_USER / MI_PASS，Flow 仍可编辑，但执行会失败")
	}

	var (
		account *miaccount.Account
		minaSvc *minaservice.Service
		miioSvc *miioservice.Service
	)
	if user != "" && pass != "" {
		tokenPath := filepath.Join(os.Getenv("HOME"), ".mi.token")
		account = miaccount.NewAccount(user, pass, tokenPath)
		minaSvc = minaservice.New(account)
		miioSvc = miioservice.New(account, "")
	}

	a := &app{
		store:      store,
		mina:       minaSvc,
		miio:       miioSvc,
		defaultDID: did,
	}

	mux := http.NewServeMux()
	// 简单的单页前端
	mux.HandleFunc("/", a.handleIndex)
	// RESTful API
	mux.HandleFunc("/api/flows", a.handleFlows)
	mux.HandleFunc("/api/flows/", a.handleFlowByID) // /api/flows/{id} 和 /api/flows/{id}/run

	log.Printf("Flow server listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, logRequest(mux)))
}

// logRequest wraps handler with basic logging.
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s in %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// --- HTTP Handlers ---

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

// /api/flows
func (a *app) handleFlows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		flows, err := a.store.list()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, flows)
	case http.MethodPost:
		var f Flow
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(f.Name) == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		saved, err := a.store.upsert(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, saved)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// /api/flows/{id} or /api/flows/{id}/run
func (a *app) handleFlowByID(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/flows/")
	if trimmed == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(trimmed, "/")
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		switch r.Method {
		case http.MethodGet:
			f, ok, err := a.store.get(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, f)
		case http.MethodDelete:
			if err := a.store.delete(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "GET, DELETE")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "run":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		f, ok, err := a.store.get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		go a.runFlow(f)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"status":"started","id":%q}`, f.ID)
	default:
		http.NotFound(w, r)
	}
}

// runFlow executes a flow sequentially.
func (a *app) runFlow(f Flow) {
	log.Printf("Running flow %s (%s) with %d steps\n", f.ID, f.Name, len(f.Steps))
	for i, step := range f.Steps {
		if err := a.runStep(step); err != nil {
			log.Printf("flow %s step %d (%s) error: %v", f.ID, i, step.Label, err)
			// 这里简单记录错误并继续后续步骤；也可以在未来支持“出错即停止”的选项
		}
	}
}

func (a *app) resolveDID(step FlowStep) string {
	if strings.TrimSpace(step.Device) != "" {
		return step.Device
	}
	return a.defaultDID
}

func (a *app) runStep(step FlowStep) error {
	switch step.Type {
	case StepTypeDelay:
		if step.DurationMS <= 0 {
			return nil
		}
		time.Sleep(time.Duration(step.DurationMS) * time.Millisecond)
		return nil
	case StepTypeTTS:
		if a.mina == nil {
			return fmt.Errorf("mina service not initialized (check MI_USER/MI_PASS)")
		}
		did := a.resolveDID(step)
		if did == "" {
			return fmt.Errorf("no device ID configured for TTS step")
		}
		deviceID, err := a.mina.GetMinaDeviceID(did)
		if err != nil {
			return err
		}
		_, err = a.mina.TextToSpeech(deviceID, step.Text)
		return err
	case StepTypePlayURL:
		if a.mina == nil {
			return fmt.Errorf("mina service not initialized (check MI_USER/MI_PASS)")
		}
		did := a.resolveDID(step)
		if did == "" {
			return fmt.Errorf("no device ID configured for play_url step")
		}
		deviceID, err := a.mina.GetMinaDeviceID(did)
		if err != nil {
			return err
		}
		_, err = a.mina.PlayByURL(deviceID, step.URL, 2)
		return err
	case StepTypeMiIO:
		if a.miio == nil {
			return fmt.Errorf("miio service not initialized (check MI_USER/MI_PASS)")
		}
		text := strings.TrimSpace(step.MiIOText)
		if text == "" {
			return nil
		}
		did := a.resolveDID(step)
		if did == "" {
			// 对于 list/spec 等命令可以为空，保持与 m 一致的行为
			did = ""
		}
		_, err := miiocommand.Run(a.miio, did, text, defaultPrefix)
		return err
	default:
		return fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// indexHTML is a tiny single-page "visual config" UI.
// 为了降低依赖，这里直接内嵌最简单的 HTML/JS，
// 提供：
//   - Flow 列表
//   - 基本的 Flow 编辑（名称/描述/步骤表格）
//   - 一键运行某个 Flow
//
// 后续可以随时替换为更复杂的前端，而无需改动后端 API。
const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>MiFlow 可视化控制流</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; padding: 0; background: #0f172a; color: #e5e7eb; }
    header { padding: 16px 24px; border-bottom: 1px solid #1f2937; display: flex; justify-content: space-between; align-items: center; }
    h1 { font-size: 18px; margin: 0; }
    main { display: flex; height: calc(100vh - 56px); }
    #sidebar { width: 260px; border-right: 1px solid #1f2937; padding: 12px; box-sizing: border-box; overflow-y: auto; }
    #content { flex: 1; padding: 16px 24px; box-sizing: border-box; overflow-y: auto; }
    button { background: #2563eb; color: white; border: none; border-radius: 4px; padding: 6px 10px; cursor: pointer; font-size: 13px; }
    button.secondary { background: #374151; }
    button.danger { background: #b91c1c; }
    button:disabled { opacity: .5; cursor: default; }
    input, textarea, select { background: #020617; border: 1px solid #1f2937; border-radius: 4px; color: #e5e7eb; padding: 4px 6px; font-size: 13px; width: 100%; box-sizing: border-box; }
    label { font-size: 12px; color: #9ca3af; display: block; margin-bottom: 4px; }
    .field { margin-bottom: 10px; }
    .flow-item { padding: 6px 8px; border-radius: 4px; cursor: pointer; margin-bottom: 4px; font-size: 13px; display: flex; justify-content: space-between; align-items: center; }
    .flow-item.active { background: #1f2937; }
    .flow-item:hover { background: #111827; }
    .flow-name { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .tag { font-size: 10px; background: #111827; padding: 2px 6px; border-radius: 999px; margin-left: 6px; }
    table { width: 100%; border-collapse: collapse; margin-top: 8px; font-size: 12px; }
    th, td { border-bottom: 1px solid #1f2937; padding: 4px 6px; text-align: left; vertical-align: middle; }
    th { font-weight: 500; color: #9ca3af; }
    tr:last-child td { border-bottom: none; }
    .steps-header { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; }
    .pill { display: inline-flex; align-items: center; gap: 4px; font-size: 11px; background: #111827; padding: 2px 6px; border-radius: 999px; color: #9ca3af; }
    .status { font-size: 12px; color: #9ca3af; margin-left: 8px; }
    .row-actions button { margin-left: 4px; }
    code { background: #020617; padding: 2px 4px; border-radius: 3px; }
  </style>
</head>
<body>
  <header>
    <div>
      <h1>MiFlow · 设备工作控制流</h1>
      <div style="font-size: 12px; color:#9ca3af;">基于 <code>m</code> / MiNA 的可视化编排（后端 Go，前端极简版，可按需自定义）</div>
    </div>
    <div>
      <button id="btn-new">新建 Flow</button>
      <span id="status" class="status"></span>
    </div>
  </header>
  <main>
    <aside id="sidebar">
      <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:8px;">
        <span style="font-size:12px; color:#9ca3af;">Flows</span>
        <button class="secondary" style="padding:2px 6px; font-size:11px;" id="btn-refresh">刷新</button>
      </div>
      <div id="flow-list"></div>
    </aside>
    <section id="content">
      <div id="empty-tip" style="font-size:13px; color:#9ca3af;">
        还没有配置任何 Flow。点击右上角「新建 Flow」开始。
      </div>
      <div id="editor" style="display:none; max-width:780px;">
        <div class="field">
          <label>名称</label>
          <input id="f-name" placeholder="例如：早安流程 / 回家流程">
        </div>
        <div class="field">
          <label>描述</label>
          <textarea id="f-desc" rows="2" placeholder="这个 Flow 的用途说明"></textarea>
        </div>
        <div class="steps-header">
          <div class="pill">
            步骤列表
            <span id="step-count" style="opacity:.7;"></span>
          </div>
          <div>
            <button class="secondary" id="btn-add-step">添加步骤</button>
          </div>
        </div>
        <table>
          <thead>
            <tr>
              <th style="width:90px;">类型</th>
              <th style="width:120px;">设备 (可选)</th>
              <th>参数</th>
              <th style="width:80px;">操作</th>
            </tr>
          </thead>
          <tbody id="steps-body"></tbody>
        </table>
        <div style="margin-top:12px; display:flex; justify-content:space-between; align-items:center;">
          <div style="font-size:11px; color:#6b7280;">
            类型说明：
            <code>delay</code> = 等待毫秒，
            <code>tts</code> = 小爱播报，
            <code>play_url</code> = 播放 URL，
            <code>miio</code> = 等价 <code>m</code> 的命令文本（如 <code>1,1-2=#60</code>）。
          </div>
          <div>
            <button id="btn-run" class="secondary">运行 Flow</button>
            <button id="btn-save">保存</button>
          </div>
        </div>
      </div>
    </section>
  </main>

  <script>
    const api = {
      async listFlows() {
        const res = await fetch('/api/flows');
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      },
      async getFlow(id) {
        const res = await fetch('/api/flows/' + encodeURIComponent(id));
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      },
      async saveFlow(flow) {
        const res = await fetch('/api/flows', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(flow),
        });
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      },
      async deleteFlow(id) {
        const res = await fetch('/api/flows/' + encodeURIComponent(id), { method: 'DELETE' });
        if (!res.ok && res.status !== 204) throw new Error(await res.text());
      },
      async runFlow(id) {
        const res = await fetch('/api/flows/' + encodeURIComponent(id) + '/run', { method: 'POST' });
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      },
    };

    const state = {
      flows: [],
      current: null,
    };

    const el = id => document.getElementById(id);

    function setStatus(msg, isError = false) {
      const s = el('status');
      s.textContent = msg || '';
      s.style.color = isError ? '#f97373' : '#9ca3af';
      if (msg) {
        setTimeout(() => { if (s.textContent === msg) s.textContent = ''; }, 4000);
      }
    }

    function renderFlowList() {
      const list = el('flow-list');
      list.innerHTML = '';
      if (!state.flows.length) {
        list.innerHTML = '<div style="font-size:12px; color:#6b7280;">暂无 Flow</div>';
        return;
      }
      for (const f of state.flows) {
        const div = document.createElement('div');
        div.className = 'flow-item' + (state.current && state.current.id === f.id ? ' active' : '');
        div.addEventListener('click', () => openFlow(f.id));

        const name = document.createElement('div');
        name.className = 'flow-name';
        name.textContent = f.name || '(未命名 Flow)';
        div.appendChild(name);

        const right = document.createElement('div');
        right.style.display = 'flex';
        right.style.alignItems = 'center';
        const tag = document.createElement('span');
        tag.className = 'tag';
        tag.textContent = (f.steps || []).length + ' steps';
        right.appendChild(tag);
        const del = document.createElement('button');
        del.className = 'secondary';
        del.style.marginLeft = '4px';
        del.style.padding = '2px 6px';
        del.style.fontSize = '11px';
        del.textContent = '删';
        del.addEventListener('click', (ev) => {
          ev.stopPropagation();
          if (confirm('确认删除该 Flow？')) {
            api.deleteFlow(f.id).then(() => {
              setStatus('已删除 Flow');
              if (state.current && state.current.id === f.id) {
                state.current = null;
                el('editor').style.display = 'none';
                el('empty-tip').style.display = 'block';
              }
              loadFlows();
            }).catch(e => setStatus(e.message || '删除失败', true));
          }
        });
        right.appendChild(del);
        div.appendChild(right);

        list.appendChild(div);
      }
    }

    function renderEditor() {
      const f = state.current;
      if (!f) return;
      el('empty-tip').style.display = 'none';
      el('editor').style.display = 'block';
      el('f-name').value = f.name || '';
      el('f-desc').value = f.description || '';

      const tbody = el('steps-body');
      tbody.innerHTML = '';
      (f.steps || []).forEach((step, idx) => {
        const tr = document.createElement('tr');

        const tdType = document.createElement('td');
        const sel = document.createElement('select');
        ['delay','tts','play_url','miio'].forEach(t => {
          const opt = document.createElement('option');
          opt.value = t;
          opt.textContent = t;
          sel.appendChild(opt);
        });
        sel.value = step.type || 'delay';
        sel.addEventListener('change', () => { step.type = sel.value; renderParamsCell(step, tdParams); });
        tdType.appendChild(sel);
        tr.appendChild(tdType);

        const tdDev = document.createElement('td');
        const inputDev = document.createElement('input');
        inputDev.placeholder = '留空则使用 MI_DID';
        inputDev.value = step.device || '';
        inputDev.addEventListener('input', () => { step.device = inputDev.value; });
        tdDev.appendChild(inputDev);
        tr.appendChild(tdDev);

        const tdParams = document.createElement('td');
        tr.appendChild(tdParams);
        function renderParamsCell(st, cell) {
          cell.innerHTML = '';
          if (st.type === 'delay') {
            const input = document.createElement('input');
            input.type = 'number';
            input.placeholder = '等待毫秒数，例如 1000';
            input.value = st.duration_ms || '';
            input.addEventListener('input', () => { st.duration_ms = Number(input.value) || 0; });
            cell.appendChild(input);
          } else if (st.type === 'tts') {
            const input = document.createElement('input');
            input.placeholder = '播报文本';
            input.value = st.text || '';
            input.addEventListener('input', () => { st.text = input.value; });
            cell.appendChild(input);
          } else if (st.type === 'play_url') {
            const input = document.createElement('input');
            input.placeholder = '音频 URL';
            input.value = st.url || '';
            input.addEventListener('input', () => { st.url = input.value; });
            cell.appendChild(input);
          } else if (st.type === 'miio') {
            const input = document.createElement('input');
            input.placeholder = '等价 m 命令的参数，例如: 1,1-2=#60';
            input.value = st.miio_text || '';
            input.addEventListener('input', () => { st.miio_text = input.value; });
            cell.appendChild(input);
          }
        }
        // 将 JS 字段映射到 JSON 字段名
        step.type = step.type || step.Type || 'delay';
        step.device = step.device || step.Device || '';
        step.duration_ms = step.duration_ms || step.DurationMS || 0;
        step.text = step.text || step.Text || '';
        step.url = step.url || step.URL || '';
        step.miio_text = step.miio_text || step.MiIOText || '';

        renderParamsCell(step, tdParams);

        const tdActions = document.createElement('td');
        tdActions.className = 'row-actions';
        const up = document.createElement('button');
        up.className = 'secondary';
        up.style.padding = '2px 4px';
        up.textContent = '↑';
        up.addEventListener('click', () => {
          if (idx === 0) return;
          const arr = f.steps;
          [arr[idx-1], arr[idx]] = [arr[idx], arr[idx-1]];
          renderEditor();
        });
        tdActions.appendChild(up);
        const down = document.createElement('button');
        down.className = 'secondary';
        down.style.padding = '2px 4px';
        down.textContent = '↓';
        down.addEventListener('click', () => {
          const arr = f.steps;
          if (idx >= arr.length - 1) return;
          [arr[idx+1], arr[idx]] = [arr[idx], arr[idx+1]];
          renderEditor();
        });
        tdActions.appendChild(down);
        const del = document.createElement('button');
        del.className = 'danger';
        del.style.padding = '2px 4px';
        del.textContent = '×';
        del.addEventListener('click', () => {
          f.steps.splice(idx, 1);
          renderEditor();
        });
        tdActions.appendChild(del);
        tr.appendChild(tdActions);

        tbody.appendChild(tr);
      });

      el('step-count').textContent = '(' + (f.steps || []).length + ')';
    }

    async function loadFlows() {
      try {
        const flows = await api.listFlows();
        state.flows = flows;
        renderFlowList();
      } catch (e) {
        setStatus(e.message || '加载失败', true);
      }
    }

    async function openFlow(id) {
      try {
        const f = await api.getFlow(id);
        state.current = f;
        renderFlowList();
        renderEditor();
      } catch (e) {
        setStatus(e.message || '加载 Flow 失败', true);
      }
    }

    function newFlow() {
      state.current = {
        id: '',
        name: '新建 Flow',
        description: '',
        steps: [
          { type: 'tts', text: 'MiFlow 已就绪', device: '', duration_ms: 0 },
        ],
      };
      renderFlowList();
      renderEditor();
    }

    async function saveCurrent() {
      if (!state.current) return;
      const f = state.current;
      f.name = el('f-name').value.trim() || f.name;
      f.description = el('f-desc').value.trim();
      try {
        const saved = await api.saveFlow(f);
        state.current = saved;
        await loadFlows();
        renderEditor();
        setStatus('已保存');
      } catch (e) {
        setStatus(e.message || '保存失败', true);
      }
    }

    async function runCurrent() {
      if (!state.current || !state.current.id) {
        setStatus('请先保存 Flow 再运行', true);
        return;
      }
      try {
        await api.runFlow(state.current.id);
        setStatus('已触发运行（在服务器日志中查看执行情况）');
      } catch (e) {
        setStatus(e.message || '运行失败', true);
      }
    }

    document.addEventListener('DOMContentLoaded', () => {
      el('btn-refresh').addEventListener('click', loadFlows);
      el('btn-new').addEventListener('click', () => { newFlow(); });
      el('btn-add-step').addEventListener('click', () => {
        if (!state.current) return;
        state.current.steps = state.current.steps || [];
        state.current.steps.push({ type: 'delay', duration_ms: 1000 });
        renderEditor();
      });
      el('btn-save').addEventListener('click', saveCurrent);
      el('btn-run').addEventListener('click', runCurrent);
      loadFlows();
    });
  </script>
</body>
</html>`

