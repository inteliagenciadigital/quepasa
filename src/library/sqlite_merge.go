package library

// Generic sqlite database file merge, used on startup to consolidate legacy
// standalone databases (application tables and whatsmeow store tables) into
// the single shared database. Copies schema + data + indexes of every table
// missing in the target, never touching tables that already exist there, and
// renames the legacy file to `<file>.imported` afterwards so the merge runs
// only once.

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	log "github.com/inteliagenciadigital/quepasa/qplog"
)

// SameFilePath reports whether two paths point to the same existing file.
// Missing files never match.
func SameFilePath(first string, second string) bool {
	if first == second {
		return true
	}
	firstInfo, err := os.Stat(first)
	if err != nil {
		return false
	}
	secondInfo, err := os.Stat(second)
	if err != nil {
		return false
	}
	return os.SameFile(firstInfo, secondInfo)
}

// MergeSqliteDatabaseFile copies every table (schema + rows + indexes) that
// exists in the legacy sqlite file but not in the target database, inside a
// single transaction, then renames the legacy file to `<file>.imported`.
// sqlite internal tables (sqlite_sequence, ...) are skipped.
func MergeSqliteDatabaseFile(connectionString string, legacyFile string, logentry log.Logger) error {
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return fmt.Errorf("opening shared database: %w", err)
	}
	defer db.Close()

	// ATTACH and PRAGMA are per-connection: force the pool to a single
	// connection so every statement below sees them
	db.SetMaxOpenConns(1)

	// table copy order from sqlite_master is arbitrary, so foreign keys must
	// be off while copying; must be set outside of the transaction
	if _, err = db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disabling foreign keys: %w", err)
	}

	if _, err = db.Exec("ATTACH DATABASE ? AS legacy", legacyFile); err != nil {
		return fmt.Errorf("attaching legacy database: %w", err)
	}

	copied, err := copyMissingSqliteTables(db, logentry)

	if _, detachErr := db.Exec("DETACH DATABASE legacy"); detachErr != nil && err == nil {
		err = fmt.Errorf("detaching legacy database: %w", detachErr)
	}
	if err != nil {
		return err
	}

	// rename so the merge never runs twice
	importedFile := legacyFile + ".imported"
	err = os.Rename(legacyFile, importedFile)
	if err != nil {
		return fmt.Errorf("renaming legacy database file: %w", err)
	}

	logentry.Infof("legacy database %s imported (%v tables), original file kept at %s", legacyFile, copied, importedFile)
	return nil
}

type sqliteSchemaEntry struct{ name, parent, sqltext string }

func readLegacySchema(db *sql.DB, objectType string) (entries []sqliteSchemaEntry, err error) {
	rows, err := db.Query(
		"SELECT name, tbl_name, sql FROM legacy.sqlite_master WHERE type=? AND sql IS NOT NULL", objectType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry sqliteSchemaEntry
		if err = rows.Scan(&entry.name, &entry.parent, &entry.sqltext); err != nil {
			return nil, err
		}
		if strings.HasPrefix(entry.name, "sqlite_") {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func copyMissingSqliteTables(db *sql.DB, logentry log.Logger) (copied int, err error) {
	tables, err := readLegacySchema(db, "table")
	if err != nil {
		return 0, fmt.Errorf("reading legacy schema: %w", err)
	}
	if len(tables) == 0 {
		return 0, fmt.Errorf("legacy database has no tables")
	}

	indexes, err := readLegacySchema(db, "index")
	if err != nil {
		return 0, fmt.Errorf("reading legacy indexes: %w", err)
	}

	// everything must be resolved before the transaction starts: the pool is
	// limited to one connection, so a db query while the tx is open deadlocks
	missing := make(map[string]bool)
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM main.sqlite_master WHERE type='table' AND name=?", table.name).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("checking table %s: %w", table.name, err)
		}
		missing[table.name] = count == 0
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("starting merge transaction: %w", err)
	}
	defer tx.Rollback()

	for _, table := range tables {
		if !missing[table.name] {
			logentry.Debugf("table %s already exists in shared database, not imported", table.name)
			continue
		}
		// the stored CREATE TABLE statement has no schema qualifier, so it
		// executes against the main (shared) database
		if _, err = tx.Exec(table.sqltext); err != nil {
			return 0, fmt.Errorf("creating table %s: %w", table.name, err)
		}
		if _, err = tx.Exec(fmt.Sprintf("INSERT INTO main.%s SELECT * FROM legacy.%s", table.name, table.name)); err != nil {
			return 0, fmt.Errorf("copying table %s: %w", table.name, err)
		}
		logentry.Debugf("imported legacy table: %s", table.name)
		copied++
	}

	for _, index := range indexes {
		if !missing[index.parent] {
			continue
		}
		if _, err = tx.Exec(index.sqltext); err != nil {
			return 0, fmt.Errorf("creating index %s: %w", index.name, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return copied, nil
}
