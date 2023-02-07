package v3

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/migration/v3/model"
	"gorm.io/gorm"
)

var log = logging.WithName("migration|v3")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {

	err = db.Migrator().RenameTable(model.TagType{}, model.TagCategory{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.Migrator().RenameColumn(model.Tag{}, "TagTypeID", "CategoryID")
	if err != nil {
		err = liberr.Wrap(err)
	}

	// Create tables for Trackers and Tickets
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
