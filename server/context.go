package server

import (
	"context"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/mushiyu/go-mysql-server/sql"
	"github.com/mushiyu/vitess/go/mysql"
)

// SessionBuilder creates sessions given a MySQL connection and a server address.
type SessionBuilder func(conn *mysql.Conn, addr string) sql.Session

// DoneFunc is a function that must be executed when the session is used and
// it can be disposed.
type DoneFunc func()

// DefaultSessionBuilder is a SessionBuilder that returns a base session.
func DefaultSessionBuilder(c *mysql.Conn, addr string) sql.Session {
	client := c.RemoteAddr().String()
	return sql.NewSession(addr, client, c.User, c.ConnectionID)
}

// SessionManager is in charge of creating new sessions for the given
// connections and keep track of which sessions are in each connection, so
// they can be cancelled if the connection is closed.
type SessionManager struct {
	addr     string
	tracer   opentracing.Tracer
	memory   *sql.MemoryManager
	mu       *sync.Mutex
	builder  SessionBuilder
	sessions map[uint32]sql.Session
	pid      uint64
}

// NewSessionManager creates a SessionManager with the given SessionBuilder.
func NewSessionManager(
	builder SessionBuilder,
	tracer opentracing.Tracer,
	memory *sql.MemoryManager,
	addr string,
) *SessionManager {
	return &SessionManager{
		addr:     addr,
		tracer:   tracer,
		memory:   memory,
		mu:       new(sync.Mutex),
		builder:  builder,
		sessions: make(map[uint32]sql.Session),
	}
}

func (s *SessionManager) nextPid() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pid++
	return s.pid
}

// NewSession creates a Session for the given connection and saves it to
// session pool.
func (s *SessionManager) NewSession(conn *mysql.Conn) {
	s.mu.Lock()
	s.sessions[conn.ConnectionID] = s.builder(conn, s.addr)
	s.mu.Unlock()
}

func (s *SessionManager) session(conn *mysql.Conn) sql.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[conn.ConnectionID]
}

// NewContext creates a new context for the session at the given conn.
func (s *SessionManager) NewContext(conn *mysql.Conn) *sql.Context {
	return s.NewContextWithQuery(conn, "")
}

// NewContextWithQuery creates a new context for the session at the given conn.
func (s *SessionManager) NewContextWithQuery(
	conn *mysql.Conn,
	query string,
) *sql.Context {
	s.mu.Lock()
	sess, ok := s.sessions[conn.ConnectionID]
	if !ok {
		sess = s.builder(conn, s.addr)
		s.sessions[conn.ConnectionID] = sess
	}
	s.mu.Unlock()

	context := sql.NewContext(
		opentracing.ContextWithSpan(
			context.Background(),
			s.tracer.StartSpan("query"),
		),
		sql.WithSession(sess),
		sql.WithTracer(s.tracer),
		sql.WithPid(s.nextPid()),
		sql.WithQuery(query),
		sql.WithMemoryManager(s.memory),
	)

	return context
}

// CloseConn closes the connection in the session manager and all its
// associated contexts, which are cancelled.
func (s *SessionManager) CloseConn(conn *mysql.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, conn.ConnectionID)
}
