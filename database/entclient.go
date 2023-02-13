package database

import (
	"database/sql"

	"ariga.io/sqlcomment"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stablecog/sc-go/database/ent"
)

func NewEntClient(connInfo SqlDBConn) (*ent.Client, error) {
	db, err := sql.Open(connInfo.Dialect(), connInfo.DSN())
	if err != nil {
		return nil, err
	}

	// For some reason, ent doesn't recognize pgx as a valid dialect
	entDialect := connInfo.Dialect()
	if entDialect == "pgx" {
		entDialect = dialect.Postgres
	}
	driver := entsql.OpenDB(entDialect, db)
	sqlcommentDrv := sqlcomment.NewDriver(driver,
		sqlcomment.WithDriverVerTag(),
		sqlcomment.WithTags(sqlcomment.Tags{
			sqlcomment.KeyApplication: "stablecog",
		}),
	)

	return ent.NewClient(ent.Driver(sqlcommentDrv)), nil
}
