package schema

import (
	"github.com/FrankSantoso/service/internal/platform/database"
	"github.com/FrankSantoso/service/internal/schema/migrations"
	// "github.com/GuiaBolso/darwin"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres adapter
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(cfg database.Config, mFlag bool) error {
	dbURL := database.BuildDbURL(cfg)
	// Wrap assets into resources
	s := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
		})

	d, err := bindata.WithInstance(s)
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("go-bindata", d, dbURL)
	if err != nil {
		return err
	}

	if mFlag {
		err = m.Up()
		if err != nil {
			return err
		}
	} else {
		err = m.Down()
		if err != nil {
			return err
		}
	}
	return nil
}
