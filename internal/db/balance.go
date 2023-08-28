package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"log"
)

const (
	tableNameBalance        = "balance"
	queryCreateTableBalance = `
			CREATE TABLE ` + tableNameBalance + ` (
				user_id bigint NOT NULL,
				current bigint NOT NULL,
				withdrawn bigint NOT NULL,
				PRIMARY KEY (user_id)
			);
		`
	balanceInsert = "INSERT INTO " + tableNameBalance + " (user_id, current, withdrawn) VALUES ($1, 0, 0)"
	balanceGet    = "SELECT * FROM " + tableNameBalance + " WHERE user_id=$1"
	balanceUpdate = "UPDATE " + tableNameBalance + " SET current = $2, withdrawn = $3 WHERE user_id = $1"
)

func (s *StorageDb) initBalance(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "select * from "+tableNameBalance+";")
	if err != nil {
		_, err = s.db.ExecContext(ctx, queryCreateTableBalance)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableNameBalance)
	}

	err = s.initBalanceStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageDb) initBalanceStatements() error {
	var err error
	var stmt *sql.Stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, balanceInsert,
	)
	if err != nil {
		return err
	}
	s.stmts["balanceInsert"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, balanceGet,
	)
	if err != nil {
		return err
	}
	s.stmts["balanceGet"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, balanceUpdate,
	)
	if err != nil {
		return err
	}
	s.stmts["balanceUpdate"] = stmt

	return nil
}

func (s *StorageDb) GetBalance(userID uint64) (*gophermart.Balance, error) {
	b := &gophermart.Balance{}

	row := s.stmts["balanceGet"].QueryRowContext(s.ctx, userID)
	err := row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user balance not found - %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance - %w", err)
	}

	return b, nil
}

func (s *StorageDb) UpdateBalance(b *gophermart.Balance) error {
	result, err := s.stmts["balanceUpdate"].ExecContext(s.ctx, b.UserID, b.Current, b.Withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update user balance - %w", err)
	}
	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to update user balance - %w", err)
	}

	return nil
}
