package wrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/NeuronFramework/log"
	"go.uber.org/zap"
)

var ErrNoRows = errors.New("sql: no rows in result set")

type DB struct {
	logger *zap.Logger
	db     *sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db := &DB{}
	db.logger = log.TypedLogger(db)

	db.logger.Info("Open", zap.String("driverName", driverName), zap.String("dataSourceName", dataSourceName))

	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}
	db.db = sqlDB

	return db, err
}

func (db *DB) BeginReadCommittedTx(ctx context.Context, readonly bool) (*Tx, error) {
	db.logger.Info("DB.BeginReadCommittedTx", zap.Any("ctx", ctx), zap.Any("readonly", readonly))
	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: readonly})
	if err != nil {
		db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Tx{db: db, tx: tx}, nil
}

func (db *DB) Prepare(ctx context.Context, query string) (*Stmt, error) {
	db.logger.Info("DB.Prepare", zap.Any("ctx", ctx), zap.String("query", query))
	stmt, err := db.db.PrepareContext(ctx, query)
	if err != nil {
		db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Stmt{db: db, stmt: stmt, query: query}, nil
}

func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	db.logger.Info("DB.Query", zap.Any("ctx", ctx), zap.String("query", fmt.Sprintf(query, args...)))
	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Rows{db: db, rows: rows}, nil
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *Row {
	db.logger.Info("DB.QueryRow", zap.Any("ctx", ctx), zap.String("query", fmt.Sprintf(query, args...)))
	row := db.db.QueryRowContext(ctx, query, args...)
	return &Row{db: db, row: row}
}

func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (*Result, error) {
	db.logger.Info("DB.Exec", zap.Any("ctx", ctx), zap.String("query", fmt.Sprintf(query, args...)))
	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Result{db: db, result: result}, err
}

func (db *DB) Close() error {
	db.logger.Info("DB.Close")
	return db.db.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	db.logger.Info("DB.Ping")
	return db.db.PingContext(ctx)
}

type Tx struct {
	db       *DB
	tx       *sql.Tx
	commited bool
}

func (tx *Tx) Commit() error {
	tx.commited = true

	tx.db.logger.Info("Tx.Commit")
	err := tx.tx.Commit()
	if err != nil {
		tx.db.logger.Error("sqlDriver", zap.Error(err))
		return err
	}

	return nil
}

func (tx *Tx) Rollback() error {
	if tx.commited {
		return nil
	}

	tx.db.logger.Info("Tx.Rollback")
	err := tx.tx.Rollback()
	if err != nil {
		tx.db.logger.Error("sqlDriver", zap.Error(err))
		return err
	}

	return nil
}

func (tx *Tx) Prepare(ctx context.Context, query string) (*Stmt, error) {
	tx.db.logger.Info("Tx.Prepare", zap.Any("ctx", ctx), zap.String("query", query))
	stmt, err := tx.tx.PrepareContext(ctx, query)
	if err != nil {
		tx.db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Stmt{db: tx.db, stmt: stmt, query: query}, nil
}

func (tx *Tx) Stmt(ctx context.Context, stmt *Stmt) *Stmt {
	tx.db.logger.Info("Tx.Stmt", zap.Any("ctx", ctx), zap.Any("stmt", stmt))
	return &Stmt{db: tx.db, stmt: tx.tx.Stmt(stmt.stmt), query: stmt.query}
}

func (tx *Tx) Exec(ctx context.Context, query string, args ...interface{}) (*Result, error) {
	tx.db.logger.Info("Tx.Exec", zap.Any("ctx", ctx), zap.String("query", fmt.Sprintf(query, args...)))
	result, err := tx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		tx.db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Result{db: tx.db, result: result}, nil
}

func (tx *Tx) Query(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	tx.db.logger.Info("Tx.Query", zap.Any("ctx", ctx), zap.String("query", fmt.Sprintf(query, args...)))
	rows, err := tx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		tx.db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Rows{db: tx.db, rows: rows}, nil
}

func (tx *Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *Row {
	tx.db.logger.Info("Tx.QueryRow", zap.Any("ctx", ctx), zap.String("query", fmt.Sprintf(query, args...)))
	return &Row{db: tx.db, row: tx.tx.QueryRowContext(ctx, query, args...)}
}

type Stmt struct {
	db    *DB
	stmt  *sql.Stmt
	query string
}

func (s *Stmt) Close() error {
	return s.stmt.Close()
}

func (s *Stmt) Exec(ctx context.Context, args ...interface{}) (*Result, error) {
	s.db.logger.Info("Stmt.Exec", zap.Any("ctx", ctx), zap.String("stmt", s.query), zap.String("query", fmt.Sprint(args...)))
	result, err := s.stmt.ExecContext(ctx, args...)
	if err != nil {
		s.db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}

	return &Result{db: s.db, result: result}, nil
}

func (s *Stmt) Query(ctx context.Context, args ...interface{}) (*Rows, error) {
	s.db.logger.Info("Stmt.Query", zap.Any("ctx", ctx), zap.String("stmt", s.query), zap.String("query", fmt.Sprint(args...)))
	rows, err := s.stmt.QueryContext(ctx, args...)
	if err != nil {
		s.db.logger.Error("sqlDriver", zap.Error(err))
		return nil, err
	}
	return &Rows{db: s.db, rows: rows}, nil
}

func (s *Stmt) QueryRow(ctx context.Context, args ...interface{}) *Row {
	s.db.logger.Info("Stmt.QueryRow", zap.Any("ctx", ctx), zap.String("stmt", s.query), zap.String("query", fmt.Sprint(args...)))
	row := s.stmt.QueryRowContext(ctx, args...)
	return &Row{db: s.db, row: row}
}

type Rows struct {
	db   *DB
	rows *sql.Rows
}

func (r *Rows) Err() error {
	if r.rows.Err() != nil {
		r.db.logger.Error("sqlDriver", zap.Error(r.rows.Err()))
	}

	r.rows.Next()

	return r.rows.Err()
}

func (r *Rows) Next() bool {
	return r.rows.Next()
}

func (r *Rows) Scan(dest ...interface{}) error {
	err := r.rows.Scan(dest...)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNoRows
		}

		r.db.logger.Error("sqlDriver", zap.Error(err))

		return err
	}

	return nil
}

type Row struct {
	db  *DB
	row *sql.Row
}

func (r *Row) Scan(dest ...interface{}) error {
	err := r.row.Scan(dest...)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNoRows
		}

		r.db.logger.Error("sqlDriver", zap.Error(err))

		return err
	}

	return nil
}

type Result struct {
	db     *DB
	result sql.Result
}

func (r *Result) LastInsertId() (int64, error) {
	n, err := r.result.LastInsertId()
	if err != nil {
		r.db.logger.Info("LastInsertId", zap.Error(err))
		return 0, err
	}

	return n, nil
}
func (r *Result) RowsAffected() (int64, error) {
	n, err := r.result.RowsAffected()
	if err != nil {
		r.db.logger.Info("RowsAffected", zap.Error(err))
		return 0, err
	}

	return n, nil
}
