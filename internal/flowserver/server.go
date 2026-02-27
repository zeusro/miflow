// Package flowserver 提供 Flow 可视化控制流 HTTP 服务，供 cmd/flow 和 pkg/cli 共用。
package flowserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/pkg/i18n"
)

const defaultPrefix = "flow "

// FlowStepType 定义流程中的单步类型。
type FlowStepType string

const (
	StepTypeTTS     FlowStepType = "tts"
	StepTypePlayURL FlowStepType = "play_url"
	StepTypeMiIO    FlowStepType = "miio"
	StepTypeDelay   FlowStepType = "delay"
)

// FlowStep 描述流程中的一个动作。
type FlowStep struct {
	Type       FlowStepType `json:"type"`
	Label      string       `json:"label,omitempty"`
	Device     string       `json:"device,omitempty"`
	Text       string       `json:"text,omitempty"`
	URL        string       `json:"url,omitempty"`
	MiIOText   string       `json:"miio_text,omitempty"`
	DurationMS int          `json:"duration_ms,omitempty"`
}

// Flow 是由步骤组成的简单线性流程。
type Flow struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Steps       []FlowStep `json:"steps"`
}

// Store 将流程以单个 JSON 文件形式保存在磁盘。
type Store struct {
	mu      sync.RWMutex
	path    string
	flows   []Flow
	loaded  bool
	modTime time.Time
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) load() error {
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
	if info, err := os.Stat(s.path); err == nil {
		s.modTime = info.ModTime()
	}
	s.loaded = true
	return nil
}

func (s *Store) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.MarshalIndent(s.flows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) List() ([]Flow, error) {
	if err := s.load(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Flow, len(s.flows))
	copy(out, s.flows)
	return out, nil
}

func (s *Store) Upsert(f Flow) (Flow, error) {
	if f.ID == "" {
		f.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeID(f.Name))
	}
	if err := s.load(); err != nil {
		return Flow{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.flows {
		if existing.ID == f.ID {
			s.flows[i] = f
			if err := s.save(); err != nil {
				return Flow{}, err
			}
			return f, nil
		}
	}
	s.flows = append(s.flows, f)
	return f, s.save()
}

func (s *Store) Get(id string) (Flow, bool, error) {
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

func (s *Store) Delete(id string) error {
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

// Server 持有 Flow 服务器的全局状态。
type Server struct {
	Store      *Store
	Mina       *minaservice.Service
	Miio       *miioservice.Service
	DefaultDID string
}

// Run 启动 Flow HTTP 服务。
func Run(addr, dataDir string) error {
	lang := i18n.DefaultLang()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("%s", i18n.T(lang, "flow.mkdir_failed", map[string]interface{}{"Err": err}))
	}
	store := NewStore(filepath.Join(dataDir, "flows.json"))

	cfg := config.Get()
	did := cfg.DefaultDID
	tokenPath := cfg.TokenPath
	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		log.Println(i18n.T(lang, "flow.warn_not_logged_in", nil))
	}

	var minaSvc *minaservice.Service
	var miioSvc *miioservice.Service
	if token != nil && token.IsValid() {
		var err error
		miioSvc, err = miioservice.New(token, tokenPath)
		if err != nil {
			log.Printf("%s", i18n.T(lang, "flow.miio_init_failed", map[string]interface{}{"Err": err}))
		} else {
			minaSvc = minaservice.NewWithMinaAPI(miioSvc, token, tokenPath)
		}
	}

	srv := &Server{
		Store:      store,
		Mina:       minaSvc,
		Miio:       miioSvc,
		DefaultDID: did,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/flows", srv.handleFlows)
	mux.HandleFunc("/api/flows/", srv.handleFlowByID)

	log.Printf("%s", i18n.T(lang, "flow.server_listening", map[string]interface{}{"Addr": addr}))
	return http.ListenAndServe(addr, logRequest(mux))
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s in %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, BuildIndexHTML(lang))
}

func (s *Server) handleFlows(w http.ResponseWriter, r *http.Request) {
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
	switch r.Method {
	case http.MethodGet:
		flows, err := s.Store.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, flows)
	case http.MethodPost:
		var f Flow
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			http.Error(w, i18n.T(lang, "flow.invalid_json", map[string]interface{}{"Err": err.Error()}), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(f.Name) == "" {
			http.Error(w, i18n.T(lang, "flow.name_required", nil), http.StatusBadRequest)
			return
		}
		saved, err := s.Store.Upsert(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, saved)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, i18n.T(lang, "flow.method_not_allowed", nil), http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFlowByID(w http.ResponseWriter, r *http.Request) {
	lang := i18n.AcceptLanguage(r.Header.Get("Accept-Language"))
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
			f, ok, err := s.Store.Get(id)
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
			if err := s.Store.Delete(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "GET, DELETE")
			http.Error(w, i18n.T(lang, "flow.method_not_allowed", nil), http.StatusMethodNotAllowed)
		}
	case "run":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, i18n.T(lang, "flow.method_not_allowed", nil), http.StatusMethodNotAllowed)
			return
		}
		f, ok, err := s.Store.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		go s.runFlow(f)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"status":"started","id":%q}`, f.ID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) runFlow(f Flow) {
	log.Printf("Running flow %s (%s) with %d steps\n", f.ID, f.Name, len(f.Steps))
	for i, step := range f.Steps {
		if err := s.runStep(step); err != nil {
			log.Printf("%s", i18n.T(i18n.DefaultLang(), "flow.step_error", map[string]interface{}{"ID": f.ID, "Step": i, "Type": step.Type, "Err": err}))
		}
	}
}

func (s *Server) resolveDID(step FlowStep) string {
	if strings.TrimSpace(step.Device) != "" {
		return step.Device
	}
	return s.DefaultDID
}

func (s *Server) runStep(step FlowStep) error {
	switch step.Type {
	case StepTypeDelay:
		if step.DurationMS <= 0 {
			return nil
		}
		time.Sleep(time.Duration(step.DurationMS) * time.Millisecond)
		return nil
	case StepTypeTTS:
		if s.Mina == nil {
			return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "flow.mina_not_init", nil))
		}
		did := s.resolveDID(step)
		if did == "" {
			return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "flow.no_device_tts", nil))
		}
		deviceID, err := s.Mina.GetMinaDeviceID(did)
		if err != nil {
			return err
		}
		_, err = s.Mina.TextToSpeech(deviceID, step.Text)
		return err
	case StepTypePlayURL:
		if s.Mina == nil {
			return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "flow.mina_not_init", nil))
		}
		did := s.resolveDID(step)
		if did == "" {
			return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "flow.no_device_play", nil))
		}
		deviceID, err := s.Mina.GetMinaDeviceID(did)
		if err != nil {
			return err
		}
		_, err = s.Mina.PlayByURL(deviceID, step.URL, 2)
		return err
	case StepTypeMiIO:
		if s.Miio == nil {
			return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "flow.miio_not_init", nil))
		}
		text := strings.TrimSpace(step.MiIOText)
		if text == "" {
			return nil
		}
		did := s.resolveDID(step)
		_, err := miiocommand.Run(s.Miio, did, text, defaultPrefix)
		return err
	default:
		return fmt.Errorf("%s", i18n.T(i18n.DefaultLang(), "flow.unsupported_step", map[string]interface{}{"Type": step.Type}))
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
