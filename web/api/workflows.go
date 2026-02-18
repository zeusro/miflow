package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/zeusro/miflow/internal/web/workflow"
	"github.com/zeusro/miflow/web"
)

// WorkflowsList handles GET /api/workflows - list workflows
func WorkflowsList(a *web.App, r *ghttp.Request) {
	list, err := a.WorkflowStore().List()
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, list)
}

// WorkflowGet handles GET /api/workflows/:id - get workflow
func WorkflowGet(a *web.App, r *ghttp.Request) {
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, "workflow id required")
		return
	}
	w, err := a.WorkflowStore().Get(id)
	if err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	if w == nil {
		Err(r, http.StatusNotFound, "workflow not found")
		return
	}
	JSON(r, http.StatusOK, w)
}

// WorkflowCreate handles POST /api/workflows - create workflow
func WorkflowCreate(a *web.App, r *ghttp.Request) {
	var w workflow.Workflow
	if err := json.NewDecoder(r.Request.Body).Decode(&w); err != nil {
		Err(r, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(w.Name) == "" {
		Err(r, http.StatusBadRequest, "name required")
		return
	}
	if err := a.WorkflowStore().Upsert(&w); err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, w)
}

// WorkflowUpdate handles PUT /api/workflows/:id - update workflow
func WorkflowUpdate(a *web.App, r *ghttp.Request) {
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, "workflow id required")
		return
	}
	var w workflow.Workflow
	if err := json.NewDecoder(r.Request.Body).Decode(&w); err != nil {
		Err(r, http.StatusBadRequest, "invalid JSON")
		return
	}
	w.ID = id
	if err := a.WorkflowStore().Upsert(&w); err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(r, http.StatusOK, w)
}

// WorkflowDelete handles DELETE /api/workflows/:id - delete workflow
func WorkflowDelete(a *web.App, r *ghttp.Request) {
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, "workflow id required")
		return
	}
	if err := a.WorkflowStore().Delete(id); err != nil {
		Err(r, http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteStatus(http.StatusNoContent)
}

// WorkflowRun handles POST /api/workflows/:id/run - run workflow
func WorkflowRun(a *web.App, r *ghttp.Request) {
	if !RequireAuth(a, r) {
		return
	}
	id := r.GetRouter("id").String()
	if id == "" {
		Err(r, http.StatusBadRequest, "workflow id required")
		return
	}
	w, err := a.WorkflowStore().Get(id)
	if err != nil || w == nil {
		Err(r, http.StatusNotFound, "workflow not found")
		return
	}
	go a.RunWorkflow(w)
	JSON(r, http.StatusAccepted, map[string]string{"status": "started", "id": id})
}
