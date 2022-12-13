package tracker

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/tracker/jira/cloud"
)

// Tracker types
const (
	JiraCloud = "jira-cloud"
)

type Metadata struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	ID         string      `json:"id"`
	Key        string      `json:"key"`
	Name       string      `json:"name"`
	IssueTypes []IssueType `json:"issueTypes"`
}

type IssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Connector is a connector for an external ticket tracker.
type Connector interface {
	// With updates the connector with the tracker model.
	With(t *model.Tracker)
	// Create a ticket in the external tracker.
	Create(t *model.Ticket) error
	// Refresh the status of a ticket.
	Refresh(t *model.Ticket) error
	//
	RefreshAll() error
	// GetMetadata from the tracker (ticket types, projects, etc)
	GetMetadata() (Metadata, error)
	// TestConnection to the external ticket tracker.
	TestConnection() (bool, error)
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
