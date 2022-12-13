package cloud

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/tracker"
	"strings"
	"time"
)

type Connector struct {
	tracker *model.Tracker
}

func (r *Connector) With(t *model.Tracker) {
	r.tracker = t

}

// Create the ticket in JIRA.
func (r *Connector) Create(t *model.Ticket) (err error) {
	client, err := r.client()
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
	t.LastUpdated = time.Now()

	return
}

func (r *Connector) RefreshAll() (tickets map[*model.Ticket]bool, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	tickets = make(map[*model.Ticket]bool)

	var keys []string
	for i := range r.tracker.Tickets {
		t := &r.tracker.Tickets[i]
		keys = append(keys, t.Reference)
	}
	jql := fmt.Sprintf("key in (%s)", strings.Join(keys, ","))
	issues, _, err := client.Issue.Search(jql, &jira.SearchOptions{Expand: "status"})
	if err != nil {
		return
	}
	issuesByKey := make(map[string]*jira.Issue)
	for i := range issues {
		issue := &issues[i]
		issuesByKey[issue.Key] = issue
	}
	lastUpdated := time.Now()
	for i := range r.tracker.Tickets {
		t := &r.tracker.Tickets[i]
		issue, found := issuesByKey[t.Reference]
		if !found {
			tickets[t] = false
			continue
		}
		t.LastUpdated = lastUpdated
		if issue.Fields != nil && issue.Fields.Status != nil {
			t.Status = issue.Fields.Status.Name
		}
		tickets[t] = true
	}
	return
}

func (r *Connector) QueryAll() (err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	var keys []string
	for i := range r.tracker.Tickets {
		t := &r.tracker.Tickets[i]
		keys = append(keys, t.Reference)
	}
	jql := fmt.Sprintf("key in (%s)", strings.Join(keys, ","))
	issues, _, err := client.Issue.Search(jql, &jira.SearchOptions{Expand: "status"})
	if err != nil {
		return
	}
	issuesByKey := make(map[string]*jira.Issue)
	for i := range issues {
		issue := &issues[i]
		issuesByKey[issue.Key] = issue
	}

	lastUpdated := time.Now()
	for i := range r.tracker.Tickets {
		t := &r.tracker.Tickets[i]
		issue, found := issuesByKey[t.Reference]
		if !found {
			continue
		}
		t.LastUpdated = lastUpdated
		if issue.Fields != nil && issue.Fields.Status != nil {
			t.Status = issue.Fields.Status.Name
		}
	}
	return
}

func (r *Connector) Refresh(t *model.Ticket) (err error) {
	client, err := r.client()
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
		t.LastUpdated = time.Now()
		return
	}
	t.Error = false
	t.Message = ""
	t.LastUpdated = time.Now()
	if issue.Fields != nil && issue.Fields.Status != nil {
		t.Status = issue.Fields.Status.Name
	}

	return
}

func (r *Connector) GetMetadata() (metadata tracker.Metadata, err error) {
	client, err := r.client()
	if err != nil {
		return
	}
	meta, _, err := client.Issue.GetCreateMetaWithOptions(nil)
	if err != nil {
		return
	}

	for _, p := range meta.Projects {
		project := tracker.Project{ID: p.Id, Key: p.Key, Name: p.Name}
		for _, it := range p.IssueTypes {
			issueType := tracker.IssueType{ID: it.Id, Name: it.Name}
			project.IssueTypes = append(project.IssueTypes, issueType)
		}
		metadata.Projects = append(metadata.Projects, project)
	}

	return
}

func (r *Connector) client() (client *jira.Client, err error) {
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

// TestConnection to Jira Cloud.
func (r *Connector) TestConnection() (connected bool, err error) {
	client, err := r.client()
	if err != nil {
		return
	}

	_, _, err = client.User.GetSelf()
	if err != nil {
		return
	}

	connected = true
	return
}
