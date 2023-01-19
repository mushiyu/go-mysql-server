package plan

import "github.com/mushiyu/go-mysql-server/sql"

// Having node is a filter that supports aggregate expressions. A having node
// is identical to a filter node in behaviour. The difference is that some
// analyzer rules work specifically on having clauses and not filters. For
// that reason, Having is a completely new node instead of using just filter.
type Having struct {
	UnaryNode
	Cond sql.Expression
}

var _ sql.Expressioner = (*Having)(nil)

// NewHaving creates a new having node.
func NewHaving(cond sql.Expression, child sql.Node) *Having {
	return &Having{UnaryNode{Child: child}, cond}
}

// Resolved implements the sql.Node interface.
func (h *Having) Resolved() bool { return h.Cond.Resolved() && h.Child.Resolved() }

// Expressions implements the sql.Expressioner interface.
func (h *Having) Expressions() []sql.Expression { return []sql.Expression{h.Cond} }

// WithChildren implements the Node interface.
func (h *Having) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(h, len(children), 1)
	}

	return NewHaving(h.Cond, children[0]), nil
}

// WithExpressions implements the Expressioner interface.
func (h *Having) WithExpressions(exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(h, len(exprs), 1)
	}

	return NewHaving(exprs[0], h.Child), nil
}

// RowIter implements the sql.Node interface.
func (h *Having) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	span, ctx := ctx.Span("plan.Having")
	iter, err := h.Child.RowIter(ctx)
	if err != nil {
		span.Finish()
		return nil, err
	}

	return sql.NewSpanIter(span, NewFilterIter(ctx, h.Cond, iter)), nil
}

func (h *Having) String() string {
	p := sql.NewTreePrinter()
	_ = p.WriteNode("Having(%s)", h.Cond)
	_ = p.WriteChildren(h.Child.String())
	return p.String()
}
