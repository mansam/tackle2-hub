package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/migration/v2/model"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

//
// Routes
const (
	MigrationWavesRoot = "/migrationwaves"
	MigrationWaveRoot = "/migrationwave" + "/:" + ID
)

//
// MigrationWaveHandler handles Migration Wave resource routes.
type MigrationWaveHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h MigrationWaveHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("migrationwaves"))
}

// Get godoc
// @summary Get aa migration wave by ID.
// @description Get a migration wave by ID.
// @tags get
// @produce json
// @success 200 {object} api.MigrationWave
// @router /migrationwaves/{id} [get]
// @param id path int true "Migration Wave ID"
func (h MigrationWaveHandler) Get(ctx *gin.Context) {
	m := &model.MigrationWave{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB, clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := MigrationWave{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all migration waves.
// @description List all migration waves.
// @tags list
// @produce json
// @success 200 {object} []api.MigrationWave
// @router /migrationwaves [get]
func (h MigrationWaveHandler) List(ctx *gin.Context) {
	var list []model.MigrationWave
	db := h.preLoad(h.DB, clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []MigrationWave{}
	for i := range list {
		r := MigrationWave{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a migration wave.
// @description Create a migration wave.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.MigrationWave
// @router /migrationwaves [post]
// @param migrationwave body api.MigrationWave true "Migration Wave data"
func (h MigrationWaveHandler) Create(ctx *gin.Context) {
	r := &MigrationWave{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	err = h.DB.Model(m).Association("Applications").Replace("Applications", m.Applications)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	err = h.DB.Model(m).Association("Stakeholders").Replace("Stakeholders", m.Stakeholders)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	err = h.DB.Model(m).Association("StakeholderGroups").Replace("StakeholderGroups", m.StakeholderGroups)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

//
// MigrationWave REST Resource
type MigrationWave struct {
	Resource
	Name string `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate time.Time `json:"endDate"`
	Applications []Ref `json:"applications"`
	Stakeholders []Ref `json:"stakeholders"`
	StakeholderGroups []Ref `json:"stakeholderGroups"`
}

//
// With updates the resource using the model.
func (r *MigrationWave) With(m *model.MigrationWave) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.StartDate = m.StartDate
	r.EndDate = m.EndDate
	for _, app := range m.Applications {
		ref := Ref{}
		ref.With(app.ID, app.Name)
		r.Applications = append(r.Applications, ref)
	}
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
	for _, sg := range m.StakeholderGroups {
		ref := Ref{}
		ref.With(sg.ID, sg.Name)
		r.StakeholderGroups = append(r.StakeholderGroups, ref)
	}
}

//
// Model builds a model.
func (r *MigrationWave) Model() (m *model.MigrationWave) {
	m = &model.MigrationWave{
		Name: r.Name,
		StartDate: r.StartDate,
		EndDate: r.EndDate,
	}
	m.ID = r.ID
	for _, ref := range r.Applications {
		m.Applications = append(
			m.Applications,
			model.Application{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.Stakeholders {
		m.Stakeholders = append(
			m.Stakeholders,
			model.Stakeholder{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.StakeholderGroups {
		m.StakeholderGroups = append(
			m.StakeholderGroups,
			model.StakeholderGroup{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}