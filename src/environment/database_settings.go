package environment

import (
	library "github.com/inteliagenciadigital/quepasa/library"
)

// Database environment variable names used to build the SQL connection for
// the single shared database, holding both the QuePasa application tables
// (`quepasa_*`) and the Whatsmeow store tables (`whatsmeow_*`).
const (
	ENV_DBDRIVER   = "DBDRIVER"   // SQL driver: sqlite3, postgres, or mysql. Default: sqlite3.
	ENV_DBHOST     = "DBHOST"     // Hostname for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBDATABASE = "DBDATABASE" // Database name for postgres/mysql, or sqlite base file path/name. Default: quepasa.
	ENV_DBPORT     = "DBPORT"     // TCP port for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBUSER     = "DBUSER"     // Username for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBPASSWORD = "DBPASSWORD" // Password for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBSSLMODE  = "DBSSLMODE"  // PostgreSQL sslmode value. Usually unused by sqlite3 and mysql.
)

// DatabaseSettings holds the SQL connection settings for the single shared
// database (QuePasa application tables + Whatsmeow store tables).
//
// For sqlite3, only Driver and Database are normally relevant.
// If Database is empty, both the application and whatsmeow.Start() fall back
// to the default database name `quepasa`.
type DatabaseSettings struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Database string `json:"database"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// NewDatabaseSettings loads the shared database connection settings from
// environment variables.
func NewDatabaseSettings() DatabaseSettings {
	return DatabaseSettings{
		Driver:   getEnvOrDefaultString(ENV_DBDRIVER, "sqlite3"),
		Host:     getEnvOrDefaultString(ENV_DBHOST, ""),
		Database: getEnvOrDefaultString(ENV_DBDATABASE, ""),
		Port:     getEnvOrDefaultString(ENV_DBPORT, ""),
		User:     getEnvOrDefaultString(ENV_DBUSER, ""),
		Password: getEnvOrDefaultString(ENV_DBPASSWORD, ""),
		SSLMode:  getEnvOrDefaultString(ENV_DBSSLMODE, ""),
	}
}

// GetDBParameters converts the environment-facing settings into the shared
// database parameter struct consumed by whatsmeow.Start() and models.GetDB().
func (config DatabaseSettings) GetDBParameters() library.DatabaseParameters {
	return library.DatabaseParameters{
		Driver:   config.Driver,
		Host:     config.Host,
		DataBase: config.Database,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		SSL:      config.SSLMode,
	}
}
