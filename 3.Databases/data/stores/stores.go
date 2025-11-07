package stores

//go:generate mockgen -source=stores.go -destination=mock/stores.go

import gocql "github.com/apache/cassandra-gocql-driver/v2"

// gocql does not use interfaces, which makes mocking impossible.
// So pass our own wrapper implementation around.
// This allows us to use gomock in tests.

// SessionInterface allows gomock mock of gocql.Session
type SessionInterface interface {
	Query(string, ...any) QueryInterface
}

// QueryInterface allows gomock mock of gocql.Query
type QueryInterface interface {
	Bind(...any) QueryInterface
	Exec() error
	Iter() IterInterface
	Scan(...any) error
}

// IterInterface allows gomock mock of gocql.Iter
type IterInterface interface {
	Scan(...any) bool
}

// Session is a wrapper for a session for mockability.
type Session struct {
	session *gocql.Session
}

// Query is a wrapper for a query for mockability.
type Query struct {
	query *gocql.Query
}

// Iter is a wrapper for an iter for mockability.
type Iter struct {
	iter *gocql.Iter
}

// NewSession instantiates a new Session
func NewSession(session *gocql.Session) SessionInterface {
	return &Session{
		session,
	}
}

// NewQuery instantiates a new Query
func NewQuery(query *gocql.Query) QueryInterface {
	return &Query{
		query,
	}
}

// NewIter instantiates a new Iter
func NewIter(iter *gocql.Iter) IterInterface {
	return &Iter{
		iter,
	}
}

// Query wraps the session's query method
func (s *Session) Query(stmt string, values ...any) QueryInterface {
	return NewQuery(s.session.Query(stmt, values...))
}

// Bind wraps the query's Bind method
func (q *Query) Bind(v ...any) QueryInterface {
	return NewQuery(q.query.Bind(v...))
}

// Exec wraps the query's Exec method
func (q *Query) Exec() error {
	return q.query.Exec()
}

// Iter wraps the query's Iter method
func (q *Query) Iter() IterInterface {
	return NewIter(q.query.Iter())
}

// Scan wraps the query's Scan method
func (q *Query) Scan(dest ...any) error {
	return q.query.Scan(dest...)
}

// Scan is a wrapper for the iter's Scan method
func (i *Iter) Scan(dest ...any) bool {
	return i.iter.Scan(dest...)
}
