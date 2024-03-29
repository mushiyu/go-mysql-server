package plan

import (
	"reflect"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/mushiyu/go-mysql-server/sql"
)

// TableAlias is a node that acts as a table with a given name.
type TableAlias struct {
	*UnaryNode
	name string
}

// NewTableAlias returns a new Table alias node.
func NewTableAlias(name string, node sql.Node) *TableAlias {
	return &TableAlias{UnaryNode: &UnaryNode{Child: node}, name: name}
}

// Name implements the Nameable interface.
func (t *TableAlias) Name() string {
	return t.name
}

// WithChildren implements the Node interface.
func (t *TableAlias) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(t, len(children), 1)
	}

	return NewTableAlias(t.name, children[0]), nil
}

// RowIter implements the Node interface.
func (t *TableAlias) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	var table string
	if tbl, ok := t.Child.(sql.Nameable); ok {
		table = tbl.Name()
	} else {
		table = reflect.TypeOf(t.Child).String()
	}

	span, ctx := ctx.Span("sql.TableAlias", opentracing.Tag{Key: "table", Value: table})

	iter, err := t.Child.RowIter(ctx)
	if err != nil {
		span.Finish()
		return nil, err
	}

	return sql.NewSpanIter(span, iter), nil
}

func (t TableAlias) String() string {
	pr := sql.NewTreePrinter()
	_ = pr.WriteNode("TableAlias(%s)", t.name)
	_ = pr.WriteChildren(t.Child.String())
	return pr.String()
}
