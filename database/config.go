package database

import (
	"fmt"
	"os"

	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

type SqlDBConn interface {
	DSN() string
	Dialect() string
}

type PostgresConn struct {
	Host     string
	Port     string
	Password string
	User     string
	DBName   string
}

func (c *PostgresConn) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password, c.Host, c.Port, c.DBName)
}

func (c *PostgresConn) Dialect() string {
	return "pgx"
}

// Gets the DB connection information based on environment variables
func GetSqlDbConn() (SqlDBConn, error) {
	// Use postgres
	postgresDb := utils.GetEnv("POSTGRES_DB", "")
	postgresUser := utils.GetEnv("POSTGRES_USER", "")
	postgresPassword := utils.GetEnv("POSTGRES_PASSWORD", "")
	postgresHost := utils.GetEnv("POSTGRES_HOST", "127.0.0.1")
	postgresPort := utils.GetEnv("POSTGRES_PORT", "5432")

	if postgresDb == "" || postgresUser == "" || postgresPassword == "" {
		klog.Error("Postgres environment variables not set, not sure what to do? so exiting")
		os.Exit(1)
	}
	klog.V(3).Infof("Using PostgreSQL database %s@%s:%s", postgresUser, postgresHost, postgresPort)
	return &PostgresConn{
		Host:     postgresHost,
		Port:     postgresPort,
		Password: postgresPassword,
		User:     postgresUser,
		DBName:   postgresDb,
	}, nil
}
