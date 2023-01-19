package plan

import "github.com/mushiyu/go-mysql-server/sql"

// EmptyTable is a node representing an empty table.
var EmptyTable = new(emptyTable)

type emptyTable struct{}

func (emptyTable) Schema() sql.Schema   { return nil }
func (emptyTable) Children() []sql.Node { return nil }
func (emptyTable) Resolved() bool       { return true }
func (e *emptyTable) String() string    { return "EmptyTable" }

func (emptyTable) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// WithChildren implements the Node interface.
func (e *emptyTable) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(e, len(children), 0)
	}

	return e, nil
}
