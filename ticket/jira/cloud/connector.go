package cloud

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/konveyor/tackle2-hub/model"
)

type Connector struct {
	tracker *model.Tracker
}

func (r *Connector) With(t *model.Tracker) {
	r.tracker = t

}

// Create the ticket in JIRA.
func (r *Connector) Create(t *model.Ticket) (err error) {
	client, err := r.connect()
	if err != nil {
		return
	}

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Summary:     fmt.Sprintf("Migrate %s", t.Application.Name),
			Description: "Created by Konveyor.",
			Type:        jira.IssueType{Name: t.Kind},
			Project:     jira.Project{Key: t.Parent},
		},
	}
	issue, _, err := client.Issue.Create(&i)
	if err != nil {
		t.Error = true
		t.Message = err.Error()
		return
	}
	t.Created = true
	t.Error = false
	t.Message = ""
	t.Reference = issue.Key
	t.Link = issue.Self

	return
}

func (r *Connector) Refresh(t *model.Ticket) (err error) {
	client, err := r.connect()
	if err != nil {
		return
	}

	opts := jira.GetQueryOptions{
		Expand: "status",
	}
	issue, _, err := client.Issue.Get(t.Reference, &opts)
	if err != nil {
		t.Error = true
		t.Message = err.Error()
		return
	}
	t.Error = false
	t.Message = ""
	if issue.Fields != nil && issue.Fields.Status != nil {
		t.Status = issue.Fields.Status.Name
	}

	return
}

func (r *Connector) GetMetadata() (err error) {
	client, err := r.connect()
	if err != nil {
		return
	}
	meta, _, err := client.Issue.GetCreateMetaWithOptions(nil)

}

func (r *Connector) connect() (client *jira.Client, err error) {
	err = r.tracker.Identity.Decrypt()
	if err != nil {
		return
	}
	transport := jira.BasicAuthTransport{
		Username: r.tracker.Identity.User,
		Password: r.tracker.Identity.Password,
	}
	client, err = jira.NewClient(transport.Client(), r.tracker.URL)
	if err != nil {
		return
	}
	return
}
