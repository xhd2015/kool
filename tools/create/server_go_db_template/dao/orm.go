package dao

import (
	"context"
)

type ormAdaptor struct {
}

// Query implements engine.Engine.
func (ormAdaptor) Query(ctx context.Context, sql string, args []interface{}, result interface{}) error {
	return QuerySQL(ctx, sql, args, result)
}

// Exec implements engine.Engine.
func (ormAdaptor) Exec(ctx context.Context, sql string, args []interface{}) error {
	return ExecSQL(ctx, sql, args)
}

// ExecInsert implements engine.Engine.
func (ormAdaptor) ExecInsert(ctx context.Context, sql string, args []interface{}) (int64, error) {
	return InsertSQL(ctx, sql, args)
}
