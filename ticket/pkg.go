package ticket

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/ticket/jira/cloud"
)

// Tracker types
const (
	JiraCloud = "jira-cloud"
)

// Connector is a connector for an external ticket tracker.
type Connector interface {
	// With updates the connector with the tracker model.
	With(t *model.Tracker)
	// Create a ticket in the external tracker.
	Create(t *model.Ticket) error
	// Refresh the status of a ticket.
	Refresh(t *model.Ticket) error
	// GetMetadata from the tracker (ticket types, projects, etc)
	GetMetadata() error
}

// NewConnector instantiates a connector for an external ticket tracker.
func NewConnector(t *model.Tracker) (conn Connector, err error) {
	switch t.Kind {
	case JiraCloud:
		conn = &cloud.Connector{}
		conn.With(t)
	default:
		err = liberr.New("not implemented")
	}
	return
}
