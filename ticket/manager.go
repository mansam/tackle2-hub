package ticket

import (
	"context"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"time"
)

var (
	Log = logging.WithName("tickets")
)

//
// Manager provides ticket management.
type Manager struct {
	// DB
	DB *gorm.DB
}

//
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
				time.Sleep(time.Second)
				m.createPending()
			}
		}
	}()
}

//
// Create pending tickets.
func (m *Manager) createPending() {
	pending := &[]model.Ticket{}
	result := m.DB.Where("status = ?", "").Find(pending)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query pending tickets.")
		return
	}

}

//
// Create the ticket in its tracker.
func (m *Manager) create(ticket *model.Ticket) (err error) {
	conn, err := ticket.Tracker.Connection()
	if err != nil {
		return
	}
	err = conn.Create(ticket)
	if err != nil {
		return err
	}
	result := m.DB.Save(ticket)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}
