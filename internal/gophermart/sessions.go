package gophermart

import (
	"fmt"
	"sync"
	"time"
)

type Session struct {
	UserID uint64
	Token  string
	Expiry time.Time
}

type sessions struct {
	mu             sync.RWMutex
	storage        Storer
	bySessionToken map[string]*Session
}

func newSessions(st Storer) *sessions {
	return &sessions{
		storage:        st,
		bySessionToken: make(map[string]*Session),
	}
}

func (s Session) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}

func (sns *sessions) Add(session *Session) error {
	sns.mu.RLock()
	_, ok := sns.bySessionToken[session.Token]
	sns.mu.RUnlock()
	if ok {
		return fmt.Errorf("session already exists")
	}

	err := sns.storage.AddSession(session)
	if err != nil {
		return err
	}

	sns.mu.Lock()
	sns.bySessionToken[session.Token] = session
	sns.mu.Unlock()

	return nil
}

func (sns *sessions) Get(token string) (*Session, error) {
	var err error

	sns.mu.RLock()
	session, ok := sns.bySessionToken[token]
	sns.mu.RUnlock()
	if !ok {
		session, err = sns.storage.GetSession(token)
		if err != nil {
			return nil, fmt.Errorf("token session not found - %w", err)
		}
		sns.mu.Lock()
		sns.bySessionToken[session.Token] = session
		sns.mu.Unlock()
	}

	return session, nil
}

func (sns *sessions) Delete(token string) error {
	sns.mu.Lock()
	delete(sns.bySessionToken, token)
	sns.mu.Unlock()

	err := sns.storage.DeleteSession(token)
	if err != nil {
		return err
	}

	return nil
}
