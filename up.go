package goose

import (
	"database/sql"
	"fmt"
	"log"
)

// UpTo migrates up to a specific version.
func UpTo(db *sql.DB, dir string, version int64) error {
	migrations, err := CollectMigrations(dir, minVersion, version)
	if err != nil {
		return err
	}

	for {
		current, err := GetDBVersion(db)
		if err != nil {
			return err
		}

		next, err := migrations.Next(current)
		if err != nil {
			if err == ErrNoNextVersion {
				log.Printf("goose: no migrations to run. current version: %d\n", current)
				return nil
			}
			return err
		}

		if err = next.Up(db); err != nil {
			return err
		}
	}
}

func UpToAll(db *sql.DB, dir string, version int64) error {
	migrations, err := CollectMigrations(dir, minVersion, version)
	if err != nil {
		return err
	}

	for _, v := range migrations {
		var row MigrationRecord
		q := fmt.Sprintf("SELECT version_id, tstamp, is_applied FROM goose_db_version WHERE version_id=%d ORDER BY tstamp DESC LIMIT 1", v.Version)
		e := db.QueryRow(q).Scan(&row.VersionID, &row.TStamp, &row.IsApplied)

		if e != nil && e != sql.ErrNoRows {
			log.Fatal(e)
		}

		if !row.IsApplied {
			if err = v.Up(db); err != nil {
				return err
			}
		}
	}
	return nil
}

// Up applies all available migrations.
func Up(db *sql.DB, dir string) error {
	return UpTo(db, dir, maxVersion)
}

// Up applies all available migrations.
func UpAll(db *sql.DB, dir string) error {
	return UpToAll(db, dir, maxVersion)
}

// UpByOne migrates up by a single version.
func UpByOne(db *sql.DB, dir string) error {
	migrations, err := CollectMigrations(dir, minVersion, maxVersion)
	if err != nil {
		return err
	}

	currentVersion, err := GetDBVersion(db)
	if err != nil {
		return err
	}

	next, err := migrations.Next(currentVersion)
	if err != nil {
		if err == ErrNoNextVersion {
			log.Printf("goose: no migrations to run. current version: %d\n", currentVersion)
		}
		return err
	}

	if err = next.Up(db); err != nil {
		return err
	}

	return nil
}
