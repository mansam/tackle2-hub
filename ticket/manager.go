package ticket

import (
	"context"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/migration/v3/model"
	"gorm.io/gorm"
	"time"
)

var (
	Log = logging.WithName("tickets")
)

// Manager provides ticket management.
type Manager struct {
	// DB
	DB *gorm.DB
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	go func() {
		Log.Info("Started.")
		defer Log.Info("Died.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 10)
				m.createPending()
				m.refreshTickets()
			}
		}
	}()
}

// Refresh all tickets..
func (m *Manager) refreshTickets() {
	var list []model.Ticket
	result := m.DB.Preload("Tracker.Identity").Preload("Application").Where("created = ?", true).Find(&list)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query tickets.")
		return
	}
	for _, t := range list {
		err := m.refresh(&t)
		if err != nil {
			Log.Error(err, "Failed to refresh ticket.", "ticket", t.ID)
		}
	}
}

// Update the hub's representation of the ticket with fresh
// status information from the external tracker.
func (m *Manager) refresh(ticket *model.Ticket) (err error) {
	conn, err := NewConnector(&ticket.Tracker)
	if err != nil {
		return
	}
	err = conn.Refresh(ticket)
	if err != nil {
		Log.Error(err, "Failed to refresh ticket.", "ticket", ticket.ID)
	}
	result := m.DB.Save(ticket)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}

// Create pending tickets.
func (m *Manager) createPending() {
	var pending []model.Ticket
	result := m.DB.Preload("Tracker.Identity").Preload("Application").Where("created = ?", false).Find(&pending)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query pending tickets.")
		return
	}
	for _, t := range pending {
		err := m.create(&t)
		if err != nil {
			Log.Error(err, "Failed to create ticket.", "ticket", t.ID)
		}
	}
}

// Create the ticket in its tracker.
func (m *Manager) create(ticket *model.Ticket) (err error) {
	conn, err := NewConnector(&ticket.Tracker)
	if err != nil {
		return
	}
	err = conn.Create(ticket)
	if err != nil {
		return
	}
	result := m.DB.Save(ticket)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}
