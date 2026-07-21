package models

// Table prefix transition, split in three steps for traceability:
//
//  1. renameBookkeepingTable: `migrations` -> MigrationsTableName. Cannot be a
//     tracked migration itself (the migrator needs the table to know what has
//     been applied), so it runs in Go before anything else. Metadata only.
//  2. ApplyPrefixRenameMigration: renames the unprefixed data tables to
//     TablePrefix. Recorded in the migrations ledger under
//     PrefixRenameMigrationId, so every database keeps an auditable record of
//     when the rename happened. It must run AFTER the historical migrations
//     (which reference the original unprefixed names, kept verbatim in
//     migrations/*.sql) and BEFORE any newer migration (written with the
//     canonical `quepasa_` prefix): MigrateToLatest runs the migrator in two
//     phases around it.
//  3. RenameCanonicalPrefixTables: `quepasa_*` -> TablePrefix when the
//     TablePrefix constant was customized. No-op on default builds; runs on
//     every startup because a prefix change can happen at any time after the
//     ledger already contains PrefixRenameMigrationId.
//
// For SQLite all renames run while PRAGMA foreign_keys is ON (the connection
// string enables it), so ALTER TABLE ... RENAME TO rewrites the REFERENCES
// clauses in dependent tables.

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	log "github.com/inteliagenciadigital/quepasa/qplog"
)

// PrefixRenameMigrationId is the ledger id of the table prefix rename.
// Migration files with greater ids must use the canonical `quepasa_` prefix;
// files with smaller ids are historical and keep the unprefixed names.
const PrefixRenameMigrationId = "202607061900"

// legacyTableNames lists every unprefixed table ever created by QuePasa,
// in dependency order (referenced tables first). `webhooks` only exists on
// databases that stopped before migration 202512151400 dropped it.
var legacyTableNames = []string{
	"users",
	"servers",
	"webhooks",
	"dispatching",
	"conversation_labels",
	"conversation_label_links",
	"spam_sections",
	"user_contexts",
	"app_settings",
}

func tableExists(db *sqlx.DB, name string) (bool, error) {
	var query string
	switch db.DriverName() {
	case "sqlite3":
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	case "postgres":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=current_schema() AND table_name=?"
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?"
	default:
		return false, fmt.Errorf("table existence check not supported for driver: %s", db.DriverName())
	}

	var count int
	err := db.Get(&count, db.Rebind(query), name)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// renameTable renames source to target when source exists and target does
// not. Returns whether the rename happened.
func renameTable(db *sqlx.DB, source string, target string, logentry log.Logger) (bool, error) {
	exists, err := tableExists(db, source)
	if err != nil {
		return false, fmt.Errorf("checking table %s: %w", source, err)
	}
	if !exists {
		return false, nil
	}

	targetExists, err := tableExists(db, target)
	if err != nil {
		return false, fmt.Errorf("checking table %s: %w", target, err)
	}
	if targetExists {
		logentry.Warnf("tables %s and %s both exist, keeping both (manual review required)", source, target)
		return false, nil
	}

	logentry.Infof("renaming table %s to %s", source, target)
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", source, target))
	if err != nil {
		return false, fmt.Errorf("renaming table %s: %w", source, err)
	}
	return true, nil
}

// renameBookkeepingTable brings the migrations ledger itself to the current
// prefixed name, preserving the applied-migrations history. Runs before the
// migrator, outside of the ledger by necessity.
func renameBookkeepingTable(db *sqlx.DB, logentry log.Logger) error {
	for _, legacy := range []string{"migrations", canonicalTablePrefix + "migrations"} {
		if legacy == MigrationsTableName {
			continue
		}
		if _, err := renameTable(db, legacy, MigrationsTableName, logentry); err != nil {
			return err
		}
	}
	return nil
}

// RenameCanonicalPrefixTables renames `quepasa_*` tables to a customized
// TablePrefix. No-op when TablePrefix is the canonical `quepasa_`.
func RenameCanonicalPrefixTables(db *sqlx.DB, logentry log.Logger) error {
	if TablePrefix == canonicalTablePrefix {
		return nil
	}
	for _, name := range legacyTableNames {
		if _, err := renameTable(db, canonicalTablePrefix+name, TablePrefix+name, logentry); err != nil {
			return err
		}
	}
	return nil
}

// ApplyPrefixRenameMigration renames the unprefixed data tables to
// TablePrefix and records PrefixRenameMigrationId in the migrations ledger.
// Skipped entirely when the ledger already contains the id.
func ApplyPrefixRenameMigration(db *sqlx.DB, logentry log.Logger) error {
	var found int
	err := db.Get(&found, db.Rebind("SELECT COUNT(*) FROM "+MigrationsTableName+" WHERE id=?"), PrefixRenameMigrationId)
	if err != nil {
		return fmt.Errorf("checking prefix rename migration: %w", err)
	}
	if found > 0 {
		return nil
	}

	logentry.Infof("running migration: %s (table prefix rename)", PrefixRenameMigrationId)
	for _, name := range legacyTableNames {
		if _, err = renameTable(db, name, TablePrefix+name, logentry); err != nil {
			return err
		}
	}

	_, err = db.Exec(db.Rebind("INSERT INTO "+MigrationsTableName+" (id) VALUES (?)"), PrefixRenameMigrationId)
	if err != nil {
		return fmt.Errorf("recording prefix rename migration: %w", err)
	}
	return nil
}
