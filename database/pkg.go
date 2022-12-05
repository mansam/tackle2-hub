package database

import (
	"database/sql"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	model2 "github.com/konveyor/tackle2-hub/migration/v3/model"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var log = logging.WithName("db")

var Settings = &settings.Settings

const (
	ConnectionString = "file:%s?_journal=WAL"
	FKsOn            = "&_foreign_keys=yes"
	FKsOff           = "&_foreign_keys=no"
)

// Open and automigrate the DB.
func Open(enforceFKs bool) (db *gorm.DB, err error) {
	connStr := fmt.Sprintf(ConnectionString, Settings.DB.Path)
	if enforceFKs {
		connStr += FKsOn
	} else {
		connStr += FKsOff
	}
	db, err = gorm.Open(
		sqlite.Open(connStr),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	sqlDB.SetMaxOpenConns(1)
	err = db.AutoMigrate(model.Setting{}, model2.Tracker{}, model2.Ticket{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Close the DB.
func Close(db *gorm.DB) (err error) {
	var sqlDB *sql.DB
	sqlDB, err = db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = sqlDB.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
