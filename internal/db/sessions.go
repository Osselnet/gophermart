package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"log"
)

const (
	tableNameSessions        = "sessions"
	queryCreateTableSessions = `
			CREATE TABLE ` + tableNameSessions + ` (
				user_id bigint NOT NULL,
				current bigint NOT NULL,
				withdrawn bigint NOT NULL,
				PRIMARY KEY (user_id)
			);
		`
	sessionsInsert = "INSERT INTO " + tableNameSessions + " (user_id, token, expiry) VALUES ($1, $2, $3)"
	sessionsGet    = "SELECT * FROM " + tableNameSessions + " WHERE token=$1"
	sessionsDelete = "DELETE FROM " + tableNameSessions + " WHERE token=$1"
)

func (s *StorageDB) initSessions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "select * from "+tableNameSessions+";")
	if err != nil {
		_, err = s.db.ExecContext(ctx, queryCreateTableSessions)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] table `%s` created", tableNameSessions)
	}

	err = s.initSessionsStatements()
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageDB) initSessionsStatements() error {
	stmt, err := s.db.PrepareContext(
		s.ctx, sessionsInsert,
	)
	if err != nil {
		return err
	}
	s.stmts["sessionsInsert"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, sessionsGet,
	)
	if err != nil {
		return err
	}
	s.stmts["sessionsGet"] = stmt

	stmt, err = s.db.PrepareContext(
		s.ctx, sessionsDelete,
	)
	if err != nil {
		return err
	}
	s.stmts["sessionsDelete"] = stmt

	return nil
}

func (s *StorageDB) AddSession(session *gophermart.Session) error {
	_, err := s.stmts["sessionsInsert"].ExecContext(s.ctx, session.UserID, session.Token, session.Expiry)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageDB) GetSession(token string) (*gophermart.Session, error) {
	session := &gophermart.Session{}
	row := s.stmts["sessionsGet"].QueryRowContext(s.ctx, token)
	err := row.Scan(&session.UserID, &session.Token, &session.Expiry)
	if err == sql.ErrNoRows {
		return nil, gophermart.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session - %w", err)
	}

	return session, nil
}

func (s *StorageDB) DeleteSession(token string) error {
	res, err := s.stmts["sessionsDelete"].ExecContext(s.ctx, token)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}
