// Package web - App 持有网页服务器的共享状态。
package web

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/web/workflow"
)

const oauthPendingTTL = 10 * time.Minute

type oauthEntry struct {
	oc        *miaccount.OAuthClient
	createdAt time.Time
}

// OAuthStore 按 state 保存待回调的 OAuth 客户端。
type OAuthStore struct {
	mu    sync.Mutex
	items map[string]*oauthEntry
}

func (s *OAuthStore) Put(state string, oc *miaccount.OAuthClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items == nil {
		s.items = make(map[string]*oauthEntry)
	}
	for k, e := range s.items {
		if time.Since(e.createdAt) > oauthPendingTTL {
			delete(s.items, k)
		}
	}
	s.items[state] = &oauthEntry{oc: oc, createdAt: time.Now()}
}

func (s *OAuthStore) Pop(state string) *miaccount.OAuthClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[state]; ok {
		delete(s.items, state)
		return e.oc
	}
	return nil
}

var (
	errNoToken  = errors.New("no valid token, run login first")
	errNoDevice = errors.New("no device ID configured")
)

// App 持有网页服务器的共享状态。
type App struct {
	workflowStore *workflow.Store
	deviceAPI     *device.API
	miio          *miioservice.Service
	mina          *minaservice.Service
	defaultDID    string
	oauthStore    *OAuthStore
	tokenPath     string
}

// OAuthStore 返回用于登录/回调流程的 OAuth 存储。
func (a *App) OAuthStore() *OAuthStore { return a.oauthStore }

// TokenPath 返回存储 OAuth token 的路径。
func (a *App) TokenPath() string { return a.tokenPath }

// DeviceAPI 返回设备 API（未登录时为 nil）。
func (a *App) DeviceAPI() *device.API { return a.deviceAPI }

// WorkflowStore 返回工作流存储。
func (a *App) WorkflowStore() *workflow.Store { return a.workflowStore }

// Miio 返回 miio 服务（未登录时为 nil）。
func (a *App) Miio() *miioservice.Service { return a.miio }

// RunWorkflow 异步执行工作流。
func (a *App) RunWorkflow(w *workflow.Workflow) {
	for _, step := range w.Steps {
		_ = a.runStep(step)
	}
}

// RefreshToken 重新加载 token 并初始化 miio/deviceAPI/mina。登录成功后调用。
func (a *App) RefreshToken() error {
	token := (&miaccount.TokenStore{Path: a.tokenPath}).LoadOAuth()
	if token == nil || !token.IsValid() {
		a.deviceAPI = nil
		a.miio = nil
		a.mina = nil
		return nil
	}
	miio, err := miioservice.New(token, a.tokenPath)
	if err != nil {
		return err
	}
	a.miio = miio
	a.deviceAPI = device.NewAPI(miio)
	a.mina = minaservice.NewWithMinaAPI(miio, token, a.tokenPath)
	return nil
}

// NewApp 创建新的 App 实例。
func NewApp() (*App, error) {
	cfg := config.Get()
	dataDir := cfg.Web.DataDir
	if dataDir == "" {
		dataDir = "./webdata"
	}
	store, err := workflow.NewStore(dataDir)
	if err != nil {
		return nil, err
	}

	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = ".mi.token"
	}
	token := (&miaccount.TokenStore{Path: tokenPath}).LoadOAuth()

	var miio *miioservice.Service
	var deviceAPI *device.API
	var mina *minaservice.Service
	if token != nil && token.IsValid() {
		miio, err = miioservice.New(token, tokenPath)
		if err == nil {
			deviceAPI = device.NewAPI(miio)
			mina = minaservice.NewWithMinaAPI(miio, token, tokenPath)
		}
	}

	return &App{
		workflowStore: store,
		deviceAPI:     deviceAPI,
		miio:          miio,
		mina:          mina,
		defaultDID:    cfg.DefaultDID,
		oauthStore:    &OAuthStore{},
		tokenPath:     tokenPath,
	}, nil
}

// resolveDID 解析步骤中的设备 ID，空则使用默认 DID。
func (a *App) resolveDID(step workflow.Step) string {
	if strings.TrimSpace(step.Device) != "" {
		return step.Device
	}
	return a.defaultDID
}

// runStep 执行单个工作流步骤（delay/tts/play_url/miio）。
func (a *App) runStep(step workflow.Step) error {
	switch step.Type {
	case workflow.StepTypeDelay:
		if step.DurationMS <= 0 {
			return nil
		}
		time.Sleep(time.Duration(step.DurationMS) * time.Millisecond)
		return nil
	case workflow.StepTypeTTS:
		if a.mina == nil {
			return errNoToken
		}
		did := a.resolveDID(step)
		if did == "" {
			return errNoDevice
		}
		deviceID, err := a.mina.GetMinaDeviceID(did)
		if err != nil {
			return err
		}
		_, err = a.mina.TextToSpeech(deviceID, step.Text)
		return err
	case workflow.StepTypePlayURL:
		if a.mina == nil {
			return errNoToken
		}
		did := a.resolveDID(step)
		if did == "" {
			return errNoDevice
		}
		deviceID, err := a.mina.GetMinaDeviceID(did)
		if err != nil {
			return err
		}
		_, err = a.mina.PlayByURL(deviceID, step.URL, 2)
		return err
	case workflow.StepTypeMiIO:
		if a.miio == nil {
			return errNoToken
		}
		text := strings.TrimSpace(step.MiIOText)
		if text == "" {
			return nil
		}
		did := a.resolveDID(step)
		_, err := miiocommand.Run(a.miio, did, text, "web ")
		return err
	default:
		return nil
	}
}
