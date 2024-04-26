package model

type Platform struct {
	Model
	Name        string `gorm:"not null"`
	Kind        string `gorm:"not null"`
	URL         string `gorm:"not null"`
	Deployments []Deployment
}

type Deployment struct {
	Model
	PlatformID    uint `gorm:"index;not null"`
	Platform      *Platform
	ApplicationID uint `gorm:"uniqueIndex;not null"`
	Application   *Application
	IdentityID    uint `gorm:"index;not null"`
	Identity      *Identity
	Locator       string
}
