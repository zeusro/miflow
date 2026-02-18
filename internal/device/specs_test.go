package device

import (
	"os"
	"testing"

	"github.com/zeusro/miflow/internal/config"
	"github.com/zeusro/miflow/internal/miaccount"
	"github.com/zeusro/miflow/internal/miioservice"
)

func TestLoadAllModelSpecs(t *testing.T) {
	cfg := config.Get()
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = os.ExpandEnv("$HOME/.mi.token")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	token := store.LoadOAuth()
	if token == nil || !token.IsValid() {
		t.Skip("no valid OAuth token, run 'm login' first")
	}
	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		t.Fatalf("miioservice.New: %v", err)
	}
	api := NewAPI(ioSvc)
	specs, failed := api.LoadAllModelSpecs()
	if len(failed) > 0 && len(specs) == 0 {
		t.Fatalf("all models failed: %v", failed)
	}
	for model, spec := range specs {
		if spec == nil {
			t.Errorf("model %s: spec is nil", model)
			continue
		}
		if len(spec.Services) == 0 {
			t.Errorf("model %s: no services", model)
		}
		t.Logf("model %s: %s", model, spec.Summary())
	}
	for model, err := range failed {
		t.Logf("model %s failed: %v", model, err)
	}
}

func TestLoadSpec(t *testing.T) {
	cfg := config.Get()
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = os.ExpandEnv("$HOME/.mi.token")
	}
	store := &miaccount.TokenStore{Path: tokenPath}
	token := store.LoadOAuth()
	if token == nil || !token.IsValid() {
		t.Skip("no valid OAuth token, run 'm login' first")
	}
	ioSvc, err := miioservice.New(token, tokenPath)
	if err != nil {
		t.Fatalf("miioservice.New: %v", err)
	}
	api := NewAPI(ioSvc)
	spec, err := api.LoadSpec("bean.switch.bln31")
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	if spec == nil || len(spec.Services) == 0 {
		t.Fatal("spec has no services")
	}
	// Switch 服务应有 on 属性和 toggle 动作
	var hasSwitch, hasOn, hasToggle bool
	for _, svc := range spec.Services {
		if svc.Description == "Switch" {
			hasSwitch = true
			for _, p := range svc.Properties {
				if p.Description == "Switch Status" || p.Description == "On" {
					hasOn = true
					break
				}
			}
			for _, a := range svc.Actions {
				if a.Description == "Toggle" {
					hasToggle = true
					break
				}
			}
			break
		}
	}
	if !hasSwitch {
		t.Error("expected Switch service")
	}
	if !hasOn {
		t.Error("expected Switch Status/On property")
	}
	if !hasToggle {
		t.Error("expected Toggle action")
	}
	t.Logf("bean.switch.bln31: %s", spec.Summary())
}
