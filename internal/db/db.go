package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"time"
)

const initTimeOut = 60 * time.Second

type StorageDb struct {
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
	dsn    string
	stmts  map[string]*sql.Stmt
}

func New(dsn string) (*StorageDb, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN needed")
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &StorageDb{
		ctx:    ctx,
		cancel: cancel,
		dsn:    dsn,
		stmts:  make(map[string]*sql.Stmt),
	}

	err := s.init(s.dsn)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed - %w", err)
	}

	return s, nil
}

func (s *StorageDb) init(dsn string) error {
	var err error
	s.db, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(s.ctx, initTimeOut)
	defer cancel()

	err = s.initUsers(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'users' table - %w`, err)
	}

	err = s.initSessions(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'sessions' table - %w`, err)
	}

	err = s.initOrders(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'orders' table - %w`, err)
	}

	err = s.initBalance(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'balance' table - %w`, err)
	}

	err = s.initWithdrawals(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create 'withdrawals' table - %w`, err)
	}

	s.db.SetMaxOpenConns(40)
	s.db.SetMaxIdleConns(20)
	s.db.SetConnMaxIdleTime(time.Second * 60)

	return nil
}
