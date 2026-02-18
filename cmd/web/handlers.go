package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/miiocommand"
	"github.com/zeusro/miflow/internal/web/workflow"
)

// apiJSON writes JSON response.
func apiJSON(r *ghttp.Request, code int, v interface{}) {
	r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Response.WriteStatus(code)
	if v != nil {
		b, _ := json.Marshal(v)
		r.Response.Write(b)
	}
}

// apiErr writes error JSON.
func apiErr(r *ghttp.Request, code int, msg string) {
	apiJSON(r, code, map[string]string{"error": msg})
}

// handleAuthCheck returns 401 if no valid token.
func (a *app) requireAuth(r *ghttp.Request) bool {
	if a.deviceAPI == nil {
		apiErr(r, http.StatusUnauthorized, "请先登录 (run login or visit /login)")
		return false
	}
	return true
}

// GET /api/devices - list devices
func (a *app) handleDevicesList(r *ghttp.Request) {
	if !a.requireAuth(r) {
		return
	}
	name := r.Get("name").String()
	getVirtual := r.Get("getVirtual").Bool()
	getHuami := r.Get("getHuami").Int()
	list, err := a.deviceAPI.List(name, getVirtual, getHuami)
	if err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, list)
}

// GET /api/devices/:id - get device detail
func (a *app) handleDeviceGet(r *ghttp.Request) {
	if !a.requireAuth(r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "device id required")
		return
	}
	d, err := a.deviceAPI.Get(id)
	if err != nil {
		apiErr(r, http.StatusNotFound, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, d)
}

// POST /api/devices/:id/control - control device (miot command, e.g. "2=#60" or "1,1-2")
func (a *app) handleDeviceControl(r *ghttp.Request) {
	if !a.requireAuth(r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "device id required")
		return
	}
	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Request.Body).Decode(&body); err != nil {
		apiErr(r, http.StatusBadRequest, "invalid JSON")
		return
	}
	cmd := strings.TrimSpace(body.Command)
	if cmd == "" {
		apiErr(r, http.StatusBadRequest, "command required")
		return
	}
	_, err := miiocommand.Run(a.miio, id, cmd, "web ")
	if err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/devices/:id/spec - get device MIoT spec (for control UI)
func (a *app) handleDeviceSpec(r *ghttp.Request) {
	if !a.requireAuth(r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "device id required")
		return
	}
	d, err := a.deviceAPI.Get(id)
	if err != nil {
		apiErr(r, http.StatusNotFound, err.Error())
		return
	}
	spec, err := a.deviceAPI.SpecForDevice(d, "json")
	if err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, spec)
}

// GET /api/workflows - list workflows
func (a *app) handleWorkflowsList(r *ghttp.Request) {
	list, err := a.workflowStore.List()
	if err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, list)
}

// GET /api/workflows/:id - get workflow
func (a *app) handleWorkflowGet(r *ghttp.Request) {
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "workflow id required")
		return
	}
	w, err := a.workflowStore.Get(id)
	if err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	if w == nil {
		apiErr(r, http.StatusNotFound, "workflow not found")
		return
	}
	apiJSON(r, http.StatusOK, w)
}

// POST /api/workflows - create workflow
func (a *app) handleWorkflowCreate(r *ghttp.Request) {
	var w workflow.Workflow
	if err := json.NewDecoder(r.Request.Body).Decode(&w); err != nil {
		apiErr(r, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(w.Name) == "" {
		apiErr(r, http.StatusBadRequest, "name required")
		return
	}
	if err := a.workflowStore.Upsert(&w); err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, w)
}

// PUT /api/workflows/:id - update workflow
func (a *app) handleWorkflowUpdate(r *ghttp.Request) {
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "workflow id required")
		return
	}
	var w workflow.Workflow
	if err := json.NewDecoder(r.Request.Body).Decode(&w); err != nil {
		apiErr(r, http.StatusBadRequest, "invalid JSON")
		return
	}
	w.ID = id
	if err := a.workflowStore.Upsert(&w); err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	apiJSON(r, http.StatusOK, w)
}

// DELETE /api/workflows/:id - delete workflow
func (a *app) handleWorkflowDelete(r *ghttp.Request) {
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "workflow id required")
		return
	}
	if err := a.workflowStore.Delete(id); err != nil {
		apiErr(r, http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteStatus(http.StatusNoContent)
}

// POST /api/workflows/:id/run - run workflow
func (a *app) handleWorkflowRun(r *ghttp.Request) {
	if !a.requireAuth(r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		apiErr(r, http.StatusBadRequest, "workflow id required")
		return
	}
	w, err := a.workflowStore.Get(id)
	if err != nil || w == nil {
		apiErr(r, http.StatusNotFound, "workflow not found")
		return
	}
	go a.runWorkflow(w)
	apiJSON(r, http.StatusAccepted, map[string]string{"status": "started", "id": id})
}

func (a *app) runWorkflow(w *workflow.Workflow) {
	for i, step := range w.Steps {
		if err := a.runStep(step); err != nil {
			// log error, continue
			_ = i
		}
	}
}
