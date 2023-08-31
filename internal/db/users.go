package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"log"
)

const (
	tableNameUsers        = "users"
	queryCreateTableUsers = `
			CREATE TABLE ` + tableNameUsers + ` (
				id serial PRIMARY KEY,
				login varchar NOT NULL, 
				password bytea NOT NULL
			);
		`
	usersInsert     = "INSERT INTO " + tableNameUsers + " (login, password) VALUES ($1, $2)"
	usersGetByLogin = "SELECT * FROM " + tableNameUsers + " WHERE login=$1"
	usersGetByID    = "SELECT * FROM " + tableNameUsers + " WHERE id=$1"
	usersDelete     = "DELETE FROM " + tableNameUsers + " WHERE login=$1"
)

func (s *StorageDB) initUsers(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "select * from "+tableNameUsers+";")
	if err != nil {
		_, err = s.db.ExecContext(ctx, queryCreateTableUsers)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableNameUsers)
	}

	err = s.initUsersStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageDB) initUsersStatements() error {
	var err error
	var stmt *sql.Stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, usersInsert,
	)
	if err != nil {
		return err
	}
	s.stmts["usersInsert"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, usersGetByLogin,
	)
	if err != nil {
		return err
	}
	s.stmts["usersGetByLogin"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, usersGetByID,
	)
	if err != nil {
		return err
	}
	s.stmts["usersGetByID"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, usersDelete,
	)
	if err != nil {
		return err
	}
	s.stmts["usersDelete"] = stmt

	return nil
}

func (s *StorageDB) AddUser(u *gophermart.User) (uint64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	txInsert := tx.StmtContext(s.ctx, s.stmts["usersInsert"])
	txGet := tx.StmtContext(s.ctx, s.stmts["usersGetByLogin"])
	txInsertBalance := tx.StmtContext(s.ctx, s.stmts["balanceInsert"])

	row := txGet.QueryRowContext(s.ctx, u.Login)
	blankUser := gophermart.User{}
	err = row.Scan(&blankUser.ID, &blankUser.Login, &blankUser.Password)
	if err == sql.ErrNoRows {
		_, err = txInsert.ExecContext(s.ctx, u.Login, u.Password)
		if err != nil {
			return 0, err
		}

		row = txGet.QueryRowContext(s.ctx, u.Login)
		err = row.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return 0, err
		}

		_, err = txInsertBalance.ExecContext(s.ctx, u.ID)
		if err != nil {
			return 0, err
		}

	} else if err != nil {
		return 0, err
	} else {
		return 0, gophermart.ErrLoginAlreadyTaken
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("add user transaction failed - %w", err)
	}

	return u.ID, nil
}

func (s *StorageDB) GetUser(byKey interface{}) (*gophermart.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txGetByLogin := tx.StmtContext(s.ctx, s.stmts["usersGetByLogin"])
	txGetByID := tx.StmtContext(s.ctx, s.stmts["usersGetByID"])

	var u gophermart.User
	var row *sql.Row

	switch key := byKey.(type) {
	case string:
		row = txGetByLogin.QueryRowContext(s.ctx, key)
	case uint64:
		row = txGetByID.QueryRowContext(s.ctx, key)
	default:
		return nil, fmt.Errorf("given type not implemented")
	}

	err = row.Scan(&u.ID, &u.Login, &u.Password)
	if err == sql.ErrNoRows {
		return nil, gophermart.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("get user transaction failed - %w", err)
	}

	return &u, nil
}

func (s *StorageDB) DeleteUser(login string) error {
	res, err := s.stmts["usersDelete"].ExecContext(s.ctx, login)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
