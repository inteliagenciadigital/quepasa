package models

// Legacy application database merge: when DBDATABASE points the shared
// database somewhere else than the historical `quepasa.sqlite` / `quepasa.db`
// file (for example DBDATABASE=whatsmeow), the application data would be left
// behind in the old file. On startup, if such a file still exists and the
// target database has no application tables yet, every table missing in the
// target (schema + data + indexes, including any whatsmeow_* tables of a
// previously unified file) is copied over and the file is renamed to
// `<file>.imported` so the merge runs only once.
//
// It runs before RenameLegacyTables, so imported tables may still carry
// legacy names: the rename step brings them to the current TablePrefix.

import (
	"fmt"
	"os"

	library "github.com/inteliagenciadigital/quepasa/library"
	log "github.com/inteliagenciadigital/quepasa/qplog"
)

// MergeLegacyApplicationDatabase imports the tables of a legacy standalone
// application sqlite database into the shared database. No-op for non-sqlite
// drivers, when no legacy file exists, when the legacy file is the shared
// database itself, or when the shared database already has application tables.
func MergeLegacyApplicationDatabase(logentry log.Logger) error {
	if dbParameters.Driver != "sqlite3" {
		return nil
	}

	targetFile := library.GetSQLiteFilename(dbParameters.DataBase)

	var legacyFile string
	for _, candidate := range []string{"quepasa.db", "quepasa.sqlite"} {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		// default setup: the shared database is the historical file itself,
		// RenameLegacyTables handles it in place
		if library.SameFilePath(candidate, targetFile) {
			continue
		}
		legacyFile = candidate
		break
	}
	if len(legacyFile) == 0 {
		return nil
	}

	// if the shared database already has application tables (any known
	// layout), the data was either merged before or created directly there:
	// never overwrite it
	db := GetDB()
	for _, name := range []string{"users", canonicalTablePrefix + "users", TablePrefix + "users"} {
		exists, err := tableExists(db, name)
		if err != nil {
			return fmt.Errorf("checking application tables: %w", err)
		}
		if exists {
			logentry.Warnf("legacy application database found at %s but shared database already has application tables, ignoring it (rename or delete the file to silence this warning)", legacyFile)
			return nil
		}
	}

	logentry.Infof("importing legacy application database from %s into shared database", legacyFile)
	return library.MergeSqliteDatabaseFile(dbParameters.GetConnectionString(), legacyFile, logentry)
}
