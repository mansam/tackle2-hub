package model

import "time"

type Tracker struct {
	Model
	Name        string
	URL         string
	Kind        string
	Identity    Identity
	IdentityID  uint
	Metadata    JSON
	Status      string
	Connected   bool
	LastUpdated time.Time
	Message     string
	Tickets     []Ticket
}

type Ticket struct {
	Model
	// Kind of ticket in the external tracker.
	Kind string
	// Parent resource that this ticket should belong to in the tracker. (e.g. Jira project)
	Parent string
	// Custom fields to send to the tracker when creating the ticket
	Fields JSON
	// Whether the last attempt to do something with the ticket reported an error
	Error bool
	// Error message, if any
	Message string
	// Whether the ticket was created in the external tracker
	Created bool
	// Reference id in external tracker
	Reference string
	// URL to ticket in external tracker
	Link string
	// Status of ticket in external tracker
	Status      string
	LastUpdated time.Time

	Application   Application
	ApplicationID uint
	Tracker       Tracker
	TrackerID     uint
}
