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
		return
	}

	err = db.Migrator().RenameColumn(model.ImportTag{}, "TagType", "Category")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	// Altering the primary key requires constructing a new table, so rename the old one,
	// create the new one, copy over the rows, and then drop the old one.
	err = db.Migrator().RenameTable("ApplicationTags", "ApplicationTags__old")
	if err != nil  {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(model.ApplicationTags{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	result := db.Exec("INSERT INTO ApplicationTags (ApplicationID, TagID) SELECT ApplicationID, TagID FROM ApplicationTags__old;")
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	err = db.Migrator().DropTable("ApplicationTags__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	// Create tables for Trackers, Tickets
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
