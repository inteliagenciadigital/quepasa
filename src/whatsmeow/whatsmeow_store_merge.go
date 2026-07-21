package whatsmeow

// Legacy store merge: whatsmeow sessions used to live in a standalone sqlite
// database (`whatsmeow.sqlite` / `whatsmeow.db`). Now the store shares a
// single database with the quepasa application tables. On startup, if a
// legacy file still exists and the target database has no whatsmeow tables
// yet, every table (schema + data + indexes) is copied over and the legacy
// file is renamed to `<file>.imported` so the merge runs only once.

import (
	"database/sql"
	"fmt"
	"os"

	library "github.com/inteliagenciadigital/quepasa/library"
	log "github.com/inteliagenciadigital/quepasa/qplog"
)

// MergeLegacyStore imports the tables of a legacy standalone whatsmeow sqlite
// database into the shared database. No-op for non-sqlite drivers, when no
// legacy file exists, or when the shared database already has whatsmeow tables.
func MergeLegacyStore(dbParameters library.DatabaseParameters, logentry log.Logger) error {
	if dbParameters.Driver != "sqlite3" {
		return nil
	}

	targetFile := library.GetSQLiteFilename(dbParameters.DataBase)

	var legacyFile string
	for _, candidate := range []string{"whatsmeow.db", "whatsmeow.sqlite"} {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		// when DBDATABASE points at the whatsmeow file itself, the store is
		// already living in the shared database: nothing to merge
		if library.SameFilePath(candidate, targetFile) {
			continue
		}
		legacyFile = candidate
		break
	}
	if len(legacyFile) == 0 {
		return nil
	}

	db, err := sql.Open("sqlite3", dbParameters.GetConnectionString())
	if err != nil {
		return fmt.Errorf("opening shared database: %w", err)
	}

	// if the shared database already has whatsmeow tables, the sessions were
	// either merged before or created directly there: never overwrite them
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='whatsmeow_version'").Scan(&count)
	db.Close()
	if err != nil {
		return fmt.Errorf("checking whatsmeow tables: %w", err)
	}
	if count > 0 {
		logentry.Warnf("legacy whatsmeow database found at %s but shared database already has whatsmeow tables, ignoring it (rename or delete the file to silence this warning)", legacyFile)
		return nil
	}

	logentry.Infof("importing legacy whatsmeow database from %s into shared database", legacyFile)
	return library.MergeSqliteDatabaseFile(dbParameters.GetConnectionString(), legacyFile, logentry)
}
