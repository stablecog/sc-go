package repository

import (
	"context"
	"fmt"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/shared"
)

// Repository is a package that contains all the database access functions

type Repository struct {
	DB             *ent.Client
	ConnInfo       database.SqlDBConn
	Redis          *database.RedisWrapper
	Ctx            context.Context
	Qdrant         *qdrant.QdrantClient
	QueueThrottler *shared.UserQueueThrottlerMap
}

// WithTx runs a function in a transaction
// Usage example:
//
//	if err := r.WithTx(func(tx *ent.Tx) error {
//		 Do stuff with tx
//		return nil
//	}); err != nil {
//
//		 Handle error
//	}
func (r *Repository) WithTx(fn func(tx *ent.Tx) error) error {
	tx, err := r.DB.Tx(r.Ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
