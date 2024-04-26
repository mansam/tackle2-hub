package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Routes
const (
	PlatformsRoot = "/platforms"
	PlatformRoot  = PlatformsRoot + "/:" + ID
)

// PlatformHandler handles Platform resource routes.
type PlatformHandler struct {
	BaseHandler
}

func (h PlatformHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("platforms"))
	routeGroup.GET(PlatformsRoot, h.List)
	routeGroup.GET(PlatformsRoot+"/", h.List)
	routeGroup.POST(PlatformsRoot, h.Create)
	routeGroup.GET(PlatformRoot, h.Get)
	routeGroup.PUT(PlatformRoot, h.Update)
	routeGroup.DELETE(PlatformRoot, h.Delete)
}

// Get godoc
// @summary Get a platform by ID.
// @description Get a platform by ID.
// @tags platforms
// @produce json
// @success 200 {object} Platform
// @router /platforms/{id} [get]
// @param id path int true "Platform ID"
func (h PlatformHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	platform := &model.Platform{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(platform, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Platform{}
	r.With(platform)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all platforms.
// @description List all platforms.
// @tags platforms
// @produce json
// @success 200 {object} []Platform
// @router /platforms [get]
func (h PlatformHandler) List(ctx *gin.Context) {
	var list []model.Platform
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
	resources := []Platform{}
	for i := range list {
		r := Platform{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a platform.
// @description Create a platform.
// @tags platforms
// @accept json
// @produce json
// @success 201 {object} Platform
// @router /platforms [post]
// @param platform body Platform true "Platform data"
func (h PlatformHandler) Create(ctx *gin.Context) {
	platform := &Platform{}
	err := h.Bind(ctx, platform)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := platform.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	platform.With(m)

	h.Respond(ctx, http.StatusCreated, platform)
}

// Delete godoc
// @summary Delete a platform.
// @description Delete a platform.
// @tags platforms
// @success 204
// @router /Platforms/{id} [delete]
// @param id path int true "Platform ID"
func (h PlatformHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	platform := &model.Platform{}
	result := h.DB(ctx).First(platform, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(platform, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a platform.
// @description Update a platform.
// @tags platforms
// @accept json
// @success 204
// @router /platforms/{id} [put]
// @param id path int true "Platform ID"
// @param platform body Platform true "Platform data"
func (h PlatformHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Platform{}
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

// Platform REST resource.
type Platform struct {
	Resource    `yaml:",inline"`
	Name        string `json:"name" binding:"required"`
	Kind        string `json:"kind" binding:"oneof=kubernetes"`
	URL         string `json:"url" binding:"required"`
	Deployments []Ref  `json:"deployments"`
}

// With updates the resource with the model.
func (r *Platform) With(m *model.Platform) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.URL = m.URL
	r.Deployments = []Ref{}
	for _, d := range m.Deployments {
		ref := Ref{}
		ref.With(d.ID, "")
		r.Deployments = append(r.Deployments, ref)
	}
}

// Model builds a model.
func (r *Platform) Model() (m *model.Platform) {
	m = &model.Platform{
		Kind: r.Kind,
		Name: r.Name,
		URL:  r.URL,
	}
	m.ID = r.ID
	for _, d := range r.Deployments {
		deployment := model.Deployment{}
		deployment.ID = d.ID
		m.Deployments = append(m.Deployments, deployment)
	}
	return
}
