package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"log"
	"strconv"
	"time"
)

const (
	tableNameWithdrawals        = "withdrawals"
	queryCreateTableWithdrawals = `
			CREATE TABLE IF NOT EXISTS ` + tableNameWithdrawals + ` (
				order_id varchar NOT NULL UNIQUE PRIMARY KEY,
				user_id bigint NOT NULL,
				sum bigint NOT NULL,
				processed_at timestamp NOT NULL
			);
		`
	withdrawalsInsert     = "INSERT INTO " + tableNameWithdrawals + " (order_id, user_id, sum, processed_at) VALUES ($1, $2, $3, $4)"
	withdrawalsGetByID    = "SELECT * FROM " + tableNameWithdrawals + " WHERE order_id=$1"
	withdrawalsGetForUser = "SELECT * FROM " + tableNameWithdrawals + " WHERE user_id=$1 ORDER BY processed_at desc"
)

func (s *StorageDB) initWithdrawals(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "select * from "+tableNameWithdrawals+";")
	if err != nil {
		_, err = s.db.ExecContext(ctx, queryCreateTableWithdrawals)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableNameWithdrawals)
	}

	err = s.initWithdrawalsStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageDB) initWithdrawalsStatements() error {
	var err error
	var stmt *sql.Stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, withdrawalsInsert,
	)
	if err != nil {
		return err
	}
	s.stmts["withdrawalsInsert"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, withdrawalsGetByID,
	)
	if err != nil {
		return err
	}
	s.stmts["withdrawalsGetByID"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, withdrawalsGetForUser,
	)
	if err != nil {
		return err
	}
	s.stmts["withdrawalsGetForUser"] = stmt

	return nil
}

func (s *StorageDB) AddWithdraw(withdraw *gophermart.Withdraw) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txGetByID := tx.StmtContext(s.ctx, s.stmts["withdrawalsGetByID"])
	txInsertWithdrawal := tx.StmtContext(s.ctx, s.stmts["withdrawalsInsert"])
	txGetBalance := tx.StmtContext(s.ctx, s.stmts["balanceGet"])
	txUpdateBalance := tx.StmtContext(s.ctx, s.stmts["balanceUpdate"])

	var balance gophermart.Balance
	row := txGetBalance.QueryRowContext(s.ctx, withdraw.UserID)
	err = row.Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err == sql.ErrNoRows {
		return fmt.Errorf("user balance not found - %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to get user balance - %w", err)
	}
	if balance.Current < withdraw.Sum {
		return gophermart.ErrNotEnoughFunds
	}

	current := balance.Current - withdraw.Sum
	withdrawn := balance.Withdrawn + withdraw.Sum
	_, err = txUpdateBalance.ExecContext(s.ctx, withdraw.UserID, current, withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update user balance - %w", err)
	}

	var bw gophermart.Withdraw
	date := new(string)
	row = txGetByID.QueryRowContext(s.ctx, strconv.Itoa(int(withdraw.OrderID)))
	err = row.Scan(&bw.OrderID, &bw.UserID, &bw.Sum, date)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = txInsertWithdrawal.ExecContext(s.ctx, strconv.Itoa(int(withdraw.OrderID)), withdraw.UserID, withdraw.Sum, time.Now())
			if err != nil {
				return err
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("add order transaction failed - %w", err)
			}
			return nil
		}

		return err
	}
	return fmt.Errorf("withdraw already recorded by another user")
}

func (s *StorageDB) GetUserWithdrawals(userID uint64) ([]*gophermart.Withdraw, error) {
	var ws []*gophermart.Withdraw

	rows, err := s.stmts["withdrawalsGetForUser"].QueryContext(s.ctx, userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var w gophermart.Withdraw
		date := new(string)

		err = rows.Scan(&w.OrderID, &w.UserID, &w.Sum, date)
		if err != nil {
			return nil, err
		}

		if w.ProcessedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		ws = append(ws, &w)
	}

	return ws, nil
}

func (s *StorageDB) GetOrderWithdrawals(orderID uint64) (*gophermart.Withdraw, error) {
	var bw gophermart.Withdraw
	date := new(string)

	row := s.stmts["withdrawalsGetByID"].QueryRowContext(s.ctx, strconv.Itoa(int(orderID)))
	err := row.Scan(&bw.OrderID, &bw.UserID, &bw.Sum, date)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found - %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order - %w", err)
	}

	return &bw, nil
}
