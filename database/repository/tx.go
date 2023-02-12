package repository

import (
	"context"

	"github.com/stablecog/sc-go/database/ent"
)

// Wrapper for *ent.Tx
// The purpose is to make it easier to pass around a transaction
// and to start a transaction if one isn't passed in
// If transaction is passed in we don't want to start one.
// We have to do this because *ent.Tx is not compatible with *ent.DB

// Example usage:
// tx := &DBTransaction{Tx: tx, DB: ent.DB}
// err = tx.Start(ctx)
// err = tx.Rollback()
// err = tx.Commit()

type DBTransaction struct {
	TX   *ent.Tx
	DB   *ent.Client
	noop bool
}

// Wrapper for ent.Tx that starts a tx if one isn't passed in
func (t *DBTransaction) Start(ctx context.Context) error {
	// Return TX if it's not nil
	if t.TX != nil {
		// Don't do commits or rollback
		t.noop = true
		return nil
	}

	// Start one if it's not
	tx, err := t.DB.Tx(ctx)
	if err != nil {
		return err
	}
	t.TX = tx
	t.noop = false
	return nil
}

func (t *DBTransaction) Rollback() error {
	if t.noop {
		return nil
	}
	return t.TX.Rollback()
}

func (t *DBTransaction) Commit() error {
	if t.noop {
		return nil
	}
	return t.TX.Commit()
}
