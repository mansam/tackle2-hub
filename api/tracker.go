package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/ticket"
	"gorm.io/gorm/clause"
	"net/http"
)

// Routes
const (
	TrackersRoot = "/trackers"
	TrackerRoot  = "/trackers" + "/:" + ID
)

// TrackerHandler handles ticket tracker routes.
type TrackerHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h TrackerHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("trackers"))
	routeGroup.GET(TrackersRoot, h.List)
	routeGroup.GET(TrackersRoot+"/", h.List)
	routeGroup.POST(TrackersRoot, h.Create)
	routeGroup.GET(TrackerRoot, h.Get)
	routeGroup.PUT(TrackerRoot, h.Update)
	routeGroup.DELETE(TrackerRoot, h.Delete)
}

// Get godoc
// @summary Get a tracker by ID.
// @description Get a tracker by ID.
// @tags get
// @produce json
// @success 200 {object} api.Tracker
// @router /trackers/{id} [get]
// @param id path string true "Tracker ID"
func (h TrackerHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
	db := h.preLoad(h.DB, clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	resource := Tracker{}
	resource.With(m)
	ctx.JSON(http.StatusOK, resource)
}

// List godoc
// @summary List all trackers.
// @description List all trackers.
// @tags get
// @produce json
// @success 200 {object} []api.Tracker
// @router /trackers [get]
func (h TrackerHandler) List(ctx *gin.Context) {
	var list []model.Tracker
	kind := ctx.Query(Kind)
	db := h.preLoad(h.DB, clause.Associations)
	if kind != "" {
		db = db.Where(Kind, kind)
	}
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Tracker{}
	for i := range list {
		r := Tracker{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a tracker.
// @description Create a tracker.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Tracker
// @router /trackers [post]
// @param tracker body api.Tracker true "Tracker data"
func (h TrackerHandler) Create(ctx *gin.Context) {
	r := &Tracker{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)

	// load the identity for the tracker so that
	// the connection can be tested and metadata
	// can be retrieved.
	i := model.Identity{}
	result := h.DB.First(&i, m.IdentityID)
	if result.Error != nil {
		h.createFailed(ctx, err)
		return
	}
	m.Identity = i

	connector, err := ticket.NewConnector(m)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	err = connector.GetMetadata()
	if err != nil {
		h.createFailed(ctx, err)
		return
	}

	result = h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a tracker.
// @description Delete a tracker.
// @tags delete
// @success 204
// @router /trackers/{id} [delete]
// @param id path int true "Tracker id"
func (h TrackerHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
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

// Update godoc
// @summary Update a tracker.
// @description Update a tracker.
// @tags update
// @accept json
// @success 204
// @router /trackers/{id} [put]
// @param id path int true "Tracker id"
// @param application body api.Tracker true "Tracker data"
func (h TrackerHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Tracker{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB.Model(m)
	db = db.Omit(clause.Associations)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Tracker API Resource
type Tracker struct {
	Resource
	Name     string `json:"name"`
	URL      string `json:"url"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Identity Ref    `json:"identity"`
}

// With updates the resource with the model.
func (r *Tracker) With(m *model.Tracker) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.URL = m.URL
	r.Kind = m.Kind
	r.Status = m.Status
	r.Identity = r.ref(m.IdentityID, &m.Identity)
}

// Model builds a model.
func (r *Tracker) Model() (m *model.Tracker) {
	m = &model.Tracker{
		Name:       r.Name,
		URL:        r.URL,
		Kind:       r.Kind,
		IdentityID: r.Identity.ID,
	}

	m.ID = r.ID

	return
}
