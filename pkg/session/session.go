package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type Store struct {
	name  string
	store sessions.Store
}

func NewCookieStore(name string, keypairs ...[]byte) *Store {
	return &Store{
		name:  name,
		store: sessions.NewCookieStore(keypairs...),
	}
}

func (s *Store) New(r *http.Request) (*sessions.Session, error) {
	return s.store.New(r, s.name)
}

func (s *Store) Get(r *http.Request) (*sessions.Session, error) {
	return s.store.Get(r, s.name)
}

func (s *Store) Save(r *http.Request, w http.ResponseWriter, a *sessions.Session) error {
	return s.store.Save(r, w, a)
}
