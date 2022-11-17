package model

import liberr "github.com/konveyor/controller/pkg/error"

const (
	TrackerJIRACloud      = "jira-cloud"
	TrackerJIRAServer     = "jira-server"
	TrackerJIRADatacenter = "jira-datacenter"
)

type TrackerConnection interface {
	With(t *Tracker)
	Create(t *Ticket) error
	Query() ([]Ticket, error)
}

type Tracker struct {
	Model
	Name       string
	URL        string
	Kind       string
	Identity   *Identity
	IdentityID *uint
	Metadata   JSON
	Status     string
}

func (r *Tracker) Connection() (conn TrackerConnection, err error) {
	switch r.Kind {
	case TrackerJIRACloud:
		conn = &JiraCloudConnection{}
		conn.With(r)
	default:
		err = liberr.New("not implemented")
	}
	return
}

type Ticket struct {
	Model
	Kind          string
	Reference     string
	Status        string
	Fields        JSON
	Application   Application
	ApplicationID uint
	Tracker       Tracker
	TrackerID     uint
}

type JiraCloudConnection struct {
}

func (r *JiraCloudConnection) With(t *Tracker) {}
func (r *JiraCloudConnection) Create(t *Ticket) (err error) {
	return
}
func (r *JiraCloudConnection) Query() (tickets []Ticket, err error) {
	return
}
