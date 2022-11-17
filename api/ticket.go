package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	TicketsRoot = "/tickets"
	TicketRoot  = "/tickets" + "/:" + ID
)

//
// Params.
const (
	TrackerId = "tracker"
)

//
// TicketHandler handles ticket routes.
type TicketHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h TicketHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("tickets"))
	routeGroup.GET(TicketsRoot, h.List)
	routeGroup.GET(TicketsRoot+"/", h.List)
	routeGroup.POST(TicketsRoot, h.Create)
	routeGroup.GET(TicketRoot, h.Get)
	routeGroup.DELETE(TicketRoot, h.Delete)
}

// Get godoc
// @summary Get a ticket by ID.
// @description Get a ticket by ID.
// @tags get
// @produce json
// @success 200 {object} api.Ticket
// @router /tickets/{id} [get]
// @param id path string true "Ticket ID"
func (h TicketHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Ticket{}
	db := h.preLoad(h.DB, clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	resource := Ticket{}
	resource.With(m)
	ctx.JSON(http.StatusOK, resource)
}

// List godoc
// @summary List all tickets.
// @description List all tickets.
// @tags get
// @produce json
// @success 200 {object} []api.Ticket
// @router /tickets [get]
func (h TicketHandler) List(ctx *gin.Context) {
	var list []model.Ticket
	appId := ctx.Query(AppId)
	trackerId := ctx.Query(TrackerId)
	db := h.preLoad(h.DB, clause.Associations)
	if appId != "" {
		db = db.Where("ApplicationID = ?", appId)
	}
	if trackerId != "" {
		db = db.Where("TrackerID = ?", trackerId)
	}
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Ticket{}
	for i := range list {
		r := Ticket{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a ticket.
// @description Create a ticket.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Ticket
// @router /tickets [post]
// @param ticket body api.Ticket true "Ticket data"
func (h TicketHandler) Create(ctx *gin.Context) {
	r := &Ticket{}
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
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a ticket.
// @description Delete a ticket.
// @tags delete
// @success 204
// @router /tickets/{id} [delete]
// @param id path int true "Ticket id"
func (h TicketHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Ticket{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}
	result = h.DB.Delete(m)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Ticket API Resource
type Ticket struct {
	Resource
	Kind        string
	Reference   string
	Status      string
	Fields      Fields
	Application Ref
	Tracker     Ref
}

//
// With updates the resource with the model.
func (r *Ticket) With(m *model.Ticket) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Reference = m.Reference
	r.Status = m.Status
	r.Application = r.ref(m.ApplicationID, m.Application)
	r.Tracker = r.ref(m.TrackerID, m.Tracker)
	_ = json.Unmarshal(m.Fields, &r.Fields)
}

//
// Model builds a model.
func (r *Ticket) Model() (m *model.Ticket) {
	m = &model.Ticket{
		Kind:          r.Kind,
		ApplicationID: r.Application.ID,
		TrackerID:     r.Tracker.ID,
	}
	if r.Fields == nil {
		r.Fields = Fields{}
	}
	m.Fields, _ = json.Marshal(r.Fields)
	m.ID = r.ID

	return
}

type Fields map[string]interface{}
