package wrap

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type Error struct {
	Err error
}

var ErrNoRows = ErrorWrap(fmt.Errorf("sql: no rows in result set"))
var ErrDuplicated = ErrorWrap(fmt.Errorf("sql: duplicated"))

func (e *Error) Error() string {
	return e.Err.Error()
}

func ErrorWrap(err error) *Error {
	return &Error{Err: err}
}

type DB struct {
	logger *zap.Logger
	db     *sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db := &DB{}
	db.logger = zap.L().Named("db")

	db.logger.Info("Open", zap.String("driverName", driverName), zap.String("dataSourceName", dataSourceName))
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		db.logger.Error("Open", zap.Error(err))
		return nil, ErrorWrap(err)
	}
	db.db = sqlDB

	return db, err
}

func (db *DB) transaction(
	ctx context.Context, readonly bool, f func(tx *Tx) (err error), isolation sql.IsolationLevel) (err error) {
	db.logger.Info("Begin")
	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{Isolation: isolation, ReadOnly: readonly})
	if err != nil {
		db.logger.Error("DB.transaction", zap.Error(err))
		return ErrorWrap(err)
	}

	defer func() {
		db.logger.Info("Rollback")
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			db.logger.Error("Rollback", zap.Error(rollbackErr))
		}
	}()

	err = f(&Tx{db: db, tx: tx})
	if err != nil {
		db.logger.Info("transaction exec failed", zap.Error(err))
		return ErrorWrap(err)
	}

	db.logger.Info("Commit")
	err = tx.Commit()
	if err != nil {
		db.logger.Error("Commit", zap.Error(err))
		return ErrorWrap(err)
	}

	return nil
}

func (db *DB) TransactionReadCommitted(ctx context.Context, readonly bool, f func(tx *Tx) (err error)) (err error) {
	return db.transaction(ctx, readonly, f, sql.LevelReadCommitted)
}

func (db *DB) TransactionRepeatableRead(ctx context.Context, readonly bool, f func(tx *Tx) (err error)) (err error) {
	return db.transaction(ctx, readonly, f, sql.LevelRepeatableRead)
}

func (db *DB) Prepare(ctx context.Context, query string) (*Stmt, error) {
	db.logger.Info("DB.Prepare", zap.Any("ctx", ctx.Err()), zap.String("query", query))
	stmt, err := db.db.PrepareContext(ctx, query)
	if err != nil {
		db.logger.Error("DB.Prepare", zap.Error(err))
		return nil, ErrorWrap(err)
	}

	return &Stmt{db: db, stmt: stmt, query: query}, nil
}

func (db *DB) Query(ctx context.Context, tx *Tx, query string, args ...interface{}) (*Rows, error) {
	db.logger.Info("DB.Query", zap.Any("ctx", ctx.Err()), zap.String("query", fmt.Sprintf(query, args...)))

	var rows *sql.Rows
	var err error

	if tx == nil {
		rows, err = db.db.QueryContext(ctx, query, args...)
	} else {
		rows, err = tx.tx.QueryContext(ctx, query, args)
	}

	if err != nil {
		db.logger.Error("DB.Query", zap.Error(err))
		return nil, ErrorWrap(err)
	}

	return &Rows{db: db, rows: rows}, nil
}

func (db *DB) QueryRow(ctx context.Context, tx *Tx, query string, args ...interface{}) *Row {
	db.logger.Info("DB.QueryRow", zap.Any("ctx", ctx.Err()), zap.String("query", fmt.Sprintf(query, args...)))

	var row *sql.Row
	if tx == nil {
		row = db.db.QueryRowContext(ctx, query, args...)
	} else {
		row = tx.tx.QueryRowContext(ctx, query, args...)
	}

	return &Row{db: db, row: row}
}

func (db *DB) Exec(ctx context.Context, tx *Tx, query string, args ...interface{}) (*Result, error) {
	db.logger.Info("DB.Exec", zap.Any("ctx", ctx.Err()), zap.String("query", fmt.Sprintf(query, args...)))

	var result sql.Result
	var err error

	if tx == nil {
		result, err = db.db.ExecContext(ctx, query, args...)
	} else {
		result, err = tx.tx.ExecContext(ctx, query, args...)
	}

	if err != nil {
		db.logger.Error("DB.Exec", zap.Error(err))
		switch err.(type) {
		case *mysql.MySQLError:
			mysqlErr := err.(*mysql.MySQLError)
			if mysqlErr.Number == 1062 {
				return nil, ErrDuplicated
			}
			return nil, ErrorWrap(mysqlErr)
		default:
			return nil, ErrorWrap(err)
		}
	}

	return &Result{db: db, result: result}, err
}

func (db *DB) Close() error {
	db.logger.Info("DB.Close")
	return db.db.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	db.logger.Info("DB.Ping")
	err := db.db.PingContext(ctx)
	if err != nil {
		db.logger.Error("DB.Ping", zap.Error(err))
		return ErrorWrap(err)
	}

	return nil
}

type Tx struct {
	db *DB
	tx *sql.Tx
}

type Stmt struct {
	db    *DB
	stmt  *sql.Stmt
	query string
}

func (s *Stmt) Close() error {
	err := s.stmt.Close()
	if err != nil {
		s.db.logger.Error("Stmt.Close", zap.Error(err))
		return ErrorWrap(err)
	}

	return nil
}

func (s *Stmt) Exec(ctx context.Context, args ...interface{}) (*Result, error) {
	buf := bytes.NewBufferString("")
	for _, v := range args {
		buf.WriteString(fmt.Sprint(v) + " ")
	}

	s.db.logger.Info("Stmt.Exec", zap.Any("ctx", ctx.Err()), zap.String("stmt", s.query), zap.String("query", buf.String()))
	result, err := s.stmt.ExecContext(ctx, args...)
	if err != nil {
		s.db.logger.Error("Stmt.Exec", zap.Error(err))
		return nil, ErrorWrap(err)
	}

	return &Result{db: s.db, result: result}, nil
}

func (s *Stmt) Query(ctx context.Context, args ...interface{}) (*Rows, error) {
	buf := bytes.NewBufferString("")
	for _, v := range args {
		buf.WriteString(fmt.Sprint(v) + ",")
	}

	s.db.logger.Info("Stmt.Query", zap.Any("ctx", ctx.Err()), zap.String("stmt", s.query), zap.String("query", buf.String()))
	rows, err := s.stmt.QueryContext(ctx, args...)
	if err != nil {
		s.db.logger.Error("Stmt.Query", zap.Error(err))
		return nil, ErrorWrap(err)
	}
	return &Rows{db: s.db, rows: rows}, nil
}

func (s *Stmt) QueryRow(ctx context.Context, args ...interface{}) *Row {
	buf := bytes.NewBufferString("")
	for _, v := range args {
		buf.WriteString(fmt.Sprint(v) + ",")
	}

	s.db.logger.Info("Stmt.QueryRow", zap.Any("ctx", ctx.Err()), zap.String("stmt", s.query), zap.String("query", buf.String()))
	row := s.stmt.QueryRowContext(ctx, args...)
	return &Row{db: s.db, row: row}
}

type Rows struct {
	db   *DB
	rows *sql.Rows
}

func (r *Rows) Err() error {
	err := r.rows.Err()
	if err != nil {
		r.db.logger.Error("Rows.Err", zap.Error(err))
		return ErrorWrap(err)
	}

	return nil
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

		r.db.logger.Error("Rows.Scan", zap.Error(err))
		return ErrorWrap(err)
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

		r.db.logger.Error("Row.Scan", zap.Error(err))
		return ErrorWrap(err)
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
		r.db.logger.Error("Result.LastInsertId", zap.Error(err))
		return 0, ErrorWrap(err)
	}

	return n, nil
}
func (r *Result) RowsAffected() (int64, error) {
	n, err := r.result.RowsAffected()
	if err != nil {
		r.db.logger.Error("Result.RowsAffected", zap.Error(err))
		return 0, ErrorWrap(err)
	}

	return n, nil
}
