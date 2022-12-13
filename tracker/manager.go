package tracker

import (
	"context"
	"encoding/json"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/migration/v3/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
				time.Sleep(time.Second)
				m.reconnectTrackers()
				m.refreshTickets()
			}
		}
	}()
}

func (m *Manager) reconnectTrackers() {
	var list []model.Tracker
	result := m.DB.Preload(clause.Associations).Find(&list).Where("connected = ?", false)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query trackers.")
		return
	}
	for _, t := range list {
		ago := t.LastUpdated.Add(time.Second * 10)
		if ago.Before(time.Now()) {
			err := m.reconnect(&t)
			if err != nil {
				Log.Error(err, "Failed to update tracker", "tracker", t.ID)
			}
		}
	}

	return
}

// reconnect attempts to reestablish a connection to the external tracker.
func (m *Manager) reconnect(tracker *model.Tracker) (err error) {
	conn, err := NewConnector(tracker)
	if err != nil {
		return
	}
	connected, err := conn.TestConnection()
	if err != nil {
		err = nil
		Log.Error(err, "Connection test failed.", "tracker", tracker.ID)
		tracker.Message = err.Error()
	}
	if connected {
		metadata, mErr := conn.GetMetadata()
		if mErr != nil {
			Log.Error(mErr, "Could not retrieve metadata.", "tracker", tracker.ID)
			tracker.Message = mErr.Error()
			connected = false
		} else {
			marshalled, _ := json.Marshal(metadata)
			tracker.Metadata = marshalled
		}
	}

	tracker.Connected = connected
	tracker.LastUpdated = time.Now()

	result := m.DB.Save(tracker)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (m *Manager) refreshTickets() {
	var list []model.Tracker
	result := m.DB.Preload(clause.Associations).Where("connected = ?", true).Find(&list)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query trackers.")
		return
	}

}

//
//// Refresh all tickets..
//func (m *Manager) refreshTickets() {
//	var list []model.Ticket
//	result := m.DB.Preload("Tracker.Identity").Preload("Application").Where("created = ?", true).Find(&list)
//	if result.Error != nil {
//		Log.Error(result.Error, "Failed to query tickets.")
//		return
//	}
//	for _, t := range list {
//		err := m.refresh(&t)
//		if err != nil {
//			Log.Error(err, "Failed to refresh ticket.", "ticket", t.ID)
//		}
//	}
//}

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
	var list []model.Tracker
	result := m.DB.Preload(clause.Associations).Where("connected = ?", true).Find(&list)
	if result.Error != nil {
		Log.Error(result.Error, "Failed to query trackers.")
		return
	}
	for _, tracker := range list {
		for i := range tracker.Tickets {
			t := &tracker.Tickets[i]
			if t.Created {
				continue
			}
			err := m.create(t)
			if err != nil {
				Log.Error(err, "Failed to create ticket.", "ticket", t.ID)
			}
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
