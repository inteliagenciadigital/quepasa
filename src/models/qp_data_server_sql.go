package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type QpDataServerSql struct {
	db *sqlx.DB
}

func (source QpDataServerSql) FindForUser(token string, user string) (response *QpServer, err error) {
	response = &QpServer{}
	err = source.db.Get(response, ApplyTablePrefix("SELECT token, wid, verified, devel, groups, broadcasts, readreceipts, deliveryreceipts, calls, readupdate, direct, store_retention_days, dispatch_types, user, contextid, metadata, timestamp FROM quepasa_servers WHERE token = ? AND user = ?"), token, user)
	if err != nil {
		response = nil
	}
	return
}

func (source QpDataServerSql) FindAll() (response []*QpServer) {
	_ = source.db.Select(&response, ApplyTablePrefix("SELECT token, wid, verified, devel, groups, broadcasts, readreceipts, deliveryreceipts, calls, readupdate, direct, store_retention_days, dispatch_types, user, contextid, metadata, timestamp FROM quepasa_servers"))
	return
}

func (source QpDataServerSql) Exists(token string) (bool, error) {
	sqlStmt := ApplyTablePrefix(`SELECT token FROM quepasa_servers WHERE token = ?`)
	err := source.db.QueryRow(sqlStmt, token).Scan(&token)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func (source QpDataServerSql) FindByToken(token string) (response *QpServer, err error) {
	response = &QpServer{}
	err = source.db.Get(response, ApplyTablePrefix("SELECT token, wid, verified, devel, groups, broadcasts, readreceipts, deliveryreceipts, calls, readupdate, direct, store_retention_days, dispatch_types, user, contextid, metadata, timestamp FROM quepasa_servers WHERE token = ?"), token)
	if err != nil {
		response = nil
	}
	return
}

func (source QpDataServerSql) Add(element *QpServer) error {
	query := ApplyTablePrefix(`INSERT INTO quepasa_servers (
		token, wid, verified, devel, groups, broadcasts, readreceipts, deliveryreceipts, calls, readupdate, direct, store_retention_days, dispatch_types, user, contextid, metadata
	) VALUES (
		:token, :wid, :verified, :devel, :groups, :broadcasts, :readreceipts, :deliveryreceipts, :calls, :readupdate, :direct, :store_retention_days, :dispatch_types, :user, :contextid, :metadata
	)`)
	params := map[string]any{
		"token":                element.Token,
		"wid":                  element.Wid,
		"verified":             element.Verified,
		"devel":                element.Devel,
		"groups":               element.Groups,
		"broadcasts":           element.Broadcasts,
		"readreceipts":         element.ReadReceipts,
		"deliveryreceipts":     element.DeliveryReceipts,
		"calls":                element.Calls,
		"readupdate":           element.ReadUpdate,
		"direct":               element.Direct,
		"store_retention_days": element.StoreRetentionDays,
		"dispatch_types":       element.DispatchTypes,
		"user":                 element.User,
		"contextid":            element.ContextId,
		"metadata":             element.Metadata,
	}
	_, err := source.db.NamedExec(query, params)
	return err
}

func (source QpDataServerSql) Update(element *QpServer) error {
	query := ApplyTablePrefix(`UPDATE quepasa_servers SET
		wid = :wid,
		verified = :verified,
		devel = :devel,
		groups = :groups,
		broadcasts = :broadcasts,
		readreceipts = :readreceipts,
		deliveryreceipts = :deliveryreceipts,
		calls = :calls,
		readupdate = :readupdate,
		direct = :direct,
		store_retention_days = :store_retention_days,
		dispatch_types = :dispatch_types,
		user = :user,
		contextid = :contextid,
		metadata = :metadata
	WHERE token = :token`)
	params := map[string]any{
		"token":                element.Token,
		"wid":                  element.Wid,
		"verified":             element.Verified,
		"devel":                element.Devel,
		"groups":               element.Groups,
		"broadcasts":           element.Broadcasts,
		"readreceipts":         element.ReadReceipts,
		"deliveryreceipts":     element.DeliveryReceipts,
		"calls":                element.Calls,
		"readupdate":           element.ReadUpdate,
		"direct":               element.Direct,
		"store_retention_days": element.StoreRetentionDays,
		"dispatch_types":       element.DispatchTypes,
		"user":                 element.User,
		"contextid":            element.ContextId,
		"metadata":             element.Metadata,
	}
	_, err := source.db.NamedExec(query, params)
	return err
}

func (source QpDataServerSql) Delete(token string) error {
	query := ApplyTablePrefix(`DELETE FROM quepasa_servers WHERE token = ?`)
	_, err := source.db.Exec(query, token)
	return err
}
