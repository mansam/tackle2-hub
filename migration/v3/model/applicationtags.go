package model

type ApplicationTags struct {
	ApplicationID uint `gorm:"primaryKey"`
	Application *Application
	TagID uint `gorm:"primaryKey"`
	Tag *Tag
	Source string `gorm:"primaryKey"`
}
