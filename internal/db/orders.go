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
	tableNameOrders  = "orders"
	queryCreateTable = `
			CREATE TABLE IF NOT EXISTS ` + tableNameOrders + ` (
				id varchar NOT NULL UNIQUE PRIMARY KEY,
				user_id bigint NOT NULL,
				status char(256) NOT NULL, 
				accrual bigint,
				uploaded_at timestamp NOT NULL
			);
		`
	ordersInsert     = "INSERT INTO " + tableNameOrders + " (id, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4)"
	orderGetByID     = "SELECT * FROM " + tableNameOrders + " WHERE id=$1"
	ordersUpdate     = "UPDATE " + tableNameOrders + " SET status = $2, accrual = $3 WHERE id = $1"
	ordersGetByID    = "SELECT * FROM " + tableNameOrders + " WHERE id=$1"
	ordersGetForUser = "SELECT * FROM " + tableNameOrders + " WHERE user_id=$1 order by uploaded_at"
	ordersGetForPool = "SELECT * FROM " + tableNameOrders + " WHERE status='NEW' or status='PROCESSING' order by uploaded_at LIMIT $1"
)

func (s *StorageDB) initOrders(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "select * from "+tableNameOrders+";")
	if err != nil {
		_, err = s.db.ExecContext(ctx, queryCreateTable)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableNameOrders)
	}

	err = s.initOrdersStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageDB) initOrdersStatements() error {
	var err error
	var stmt *sql.Stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, ordersInsert,
	)
	if err != nil {
		return err
	}
	s.stmts["ordersInsert"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, orderGetByID,
	)
	if err != nil {
		return err
	}
	s.stmts["orderGetByID"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, ordersUpdate,
	)
	if err != nil {
		return err
	}
	s.stmts["ordersUpdate"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, ordersGetByID,
	)
	if err != nil {
		return err
	}
	s.stmts["ordersGetByID"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, ordersGetForUser,
	)
	if err != nil {
		return err
	}
	s.stmts["ordersGetForUser"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, ordersGetForPool,
	)
	if err != nil {
		return err
	}
	s.stmts["ordersGetForPool"] = stmt

	return nil
}

func (s *StorageDB) AddOrder(o *gophermart.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txInsert := tx.StmtContext(s.ctx, s.stmts["ordersInsert"])
	txGetByID := tx.StmtContext(s.ctx, s.stmts["ordersGetByID"])

	var order gophermart.Order
	date := new(string)
	accrual := new(sql.NullInt64)

	row := txGetByID.QueryRowContext(s.ctx, strconv.Itoa(int(o.ID)))
	err = row.Scan(&order.ID, &order.UserID, &order.Status, accrual, date)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = txInsert.ExecContext(s.ctx, strconv.Itoa(int(o.ID)), o.UserID, o.Status, o.UploadedAt)
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
	return gophermart.ErrOrderAlreadyLoadedByAnotherUser
}

func (s *StorageDB) GetOrder(orderID uint64) (*gophermart.Order, error) {
	o := &gophermart.Order{}
	accrual := new(sql.NullInt64)
	date := new(string)

	row := s.stmts["orderGetByID"].QueryRowContext(s.ctx, strconv.Itoa(int(orderID)))
	err := row.Scan(&o.ID, &o.UserID, &o.Status, accrual, date)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found - %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order - %w", err)
	}

	if accrual.Valid {
		o.Accrual = uint64(accrual.Int64)
	}

	if o.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
		return nil, err
	}

	return o, nil
}

func (s *StorageDB) GetUserOrders(id uint64) ([]*gophermart.Order, error) {
	var orders []*gophermart.Order

	rows, err := s.stmts["ordersGetForUser"].QueryContext(s.ctx, id)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bo gophermart.Order
		accrual := new(sql.NullInt64)
		date := new(string)

		err = rows.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			bo.Accrual = uint64(accrual.Int64)
		}

		if bo.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		orders = append(orders, &bo)
	}

	return orders, nil
}

func (s *StorageDB) GetPullOrders(limit uint32) (map[uint64]*gophermart.Order, error) {
	orders := make(map[uint64]*gophermart.Order)

	rows, err := s.stmts["ordersGetForPool"].QueryContext(s.ctx, limit)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bo gophermart.Order
		accrual := new(sql.NullInt64)
		date := new(string)

		err = rows.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			bo.Accrual = uint64(accrual.Int64)
		}

		if bo.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		orders[bo.ID] = &bo
	}

	return orders, nil
}

func (s *StorageDB) UpdateOrder(o *gophermart.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txUpdateOrder := tx.StmtContext(s.ctx, s.stmts["ordersUpdate"])
	txUpdateBalance := tx.StmtContext(s.ctx, s.stmts["balanceUpdateCurrent"])
	//txGetBalance := tx.StmtContext(s.ctx, s.stmts["balanceGet"])

	if o.Status == gophermart.StatusProcessed {
		_, err = txUpdateOrder.ExecContext(s.ctx, strconv.Itoa(int(o.ID)), o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %w", err)
		}

		_, err = txUpdateBalance.ExecContext(s.ctx, o.UserID, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update user balance - %w", err)
		}
	} else {
		_, err = txUpdateOrder.ExecContext(s.ctx, strconv.Itoa(int(o.ID)), o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("update order transaction failed - %w", err)
	}

	return nil
}
