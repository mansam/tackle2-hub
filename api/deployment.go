package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Routes
const (
	DeploymentsRoot = "/deployments"
	DeploymentRoot  = DeploymentsRoot + "/:" + ID
)

// DeploymentHandler handles Deployment resource routes.
type DeploymentHandler struct {
	BaseHandler
}

func (h DeploymentHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("deployments"))
	routeGroup.GET(DeploymentsRoot, h.List)
	routeGroup.GET(DeploymentsRoot+"/", h.List)
	routeGroup.POST(DeploymentsRoot, h.Create)
	routeGroup.GET(DeploymentRoot, h.Get)
	routeGroup.PUT(DeploymentRoot, h.Update)
	routeGroup.DELETE(DeploymentRoot, h.Delete)
}

// Get godoc
// @summary Get a deployment by ID.
// @description Get a deployment by ID.
// @tags deployments
// @produce json
// @success 200 {object} Deployment
// @router /deployments/{id} [get]
// @param id path int true "Deployment ID"
func (h DeploymentHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	deployment := &model.Deployment{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(deployment, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Deployment{}
	r.With(deployment)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all deployments.
// @description List all deployments.
// @tags deployments
// @produce json
// @success 200 {object} []Deployment
// @router /deployments [get]
func (h DeploymentHandler) List(ctx *gin.Context) {
	var list []model.Deployment
	kind := ctx.Query(Kind)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	if kind != "" {
		db = db.Where(Kind, kind)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Deployment{}
	for i := range list {
		r := Deployment{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a deployment.
// @description Create a deployment.
// @tags deployments
// @accept json
// @produce json
// @success 201 {object} Deployment
// @router /deployments [post]
// @param deployment body Deployment true "Deployment data"
func (h DeploymentHandler) Create(ctx *gin.Context) {
	deployment := &Deployment{}
	err := h.Bind(ctx, deployment)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := deployment.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	deployment.With(m)

	h.Respond(ctx, http.StatusCreated, deployment)
}

// Delete godoc
// @summary Delete a deployment.
// @description Delete a deployment.
// @tags deployments
// @success 204
// @router /Deployments/{id} [delete]
// @param id path int true "Deployment ID"
func (h DeploymentHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	deployment := &model.Deployment{}
	result := h.DB(ctx).First(deployment, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(deployment, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a deployment.
// @description Update a deployment.
// @tags deployments
// @accept json
// @success 204
// @router /deployments/{id} [put]
// @param id path int true "Deployment ID"
// @param deployment body Deployment true "Deployment data"
func (h DeploymentHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Deployment{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Deployment REST resource.
type Deployment struct {
	Resource    `yaml:",inline"`
	Platform    Ref    `json:"platform" binding:"required"`
	Application Ref    `json:"application" binding:"required"`
	Identity    Ref    `json:"identity" binding:"required"`
	Locator     string `json:"locator"`
}

// With updates the resource with the model.
func (r *Deployment) With(m *model.Deployment) {
	r.Resource.With(&m.Model)
	r.Platform = r.ref(m.PlatformID, m.Platform)
	r.Identity = r.ref(m.IdentityID, m.Identity)
	r.Application = r.ref(m.ApplicationID, m.Application)
	r.Locator = m.Locator
}

// Model builds a model.
func (r *Deployment) Model() (m *model.Deployment) {
	m = &model.Deployment{}
	m.ID = r.ID
	m.IdentityID = r.Identity.ID
	m.ApplicationID = r.Application.ID
	m.PlatformID = r.Platform.ID
	m.Locator = r.Locator

	return
}
