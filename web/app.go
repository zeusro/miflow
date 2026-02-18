// Package web - App holds shared state for the web server.
package web

import (
	"errors"
	"strings"
	"time"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/device"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
	"github.com/zeusro/miflow/internal/minaservice"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/web/workflow"
)

var (
	errNoToken  = errors.New("no valid token, run login first")
	errNoDevice = errors.New("no device ID configured")
)

// App holds shared state for the web server.
type App struct {
	workflowStore *workflow.Store
	deviceAPI     *device.API
	miio          *miioservice.Service
	mina          *minaservice.Service
	defaultDID    string
}

// DeviceAPI returns the device API (nil if not logged in).
func (a *App) DeviceAPI() *device.API { return a.deviceAPI }

// WorkflowStore returns the workflow store.
func (a *App) WorkflowStore() *workflow.Store { return a.workflowStore }

// Miio returns the miio service (nil if not logged in).
func (a *App) Miio() *miioservice.Service { return a.miio }

// RunWorkflow executes a workflow asynchronously.
func (a *App) RunWorkflow(w *workflow.Workflow) {
	for _, step := range w.Steps {
		_ = a.runStep(step)
	}
}

// NewApp creates a new App instance.
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
	}, nil
}

func (a *App) resolveDID(step workflow.Step) string {
	if strings.TrimSpace(step.Device) != "" {
		return step.Device
	}
	return a.defaultDID
}

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
