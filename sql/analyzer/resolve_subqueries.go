package analyzer

import (
	"github.com/mushiyu/go-mysql-server/sql"
	"github.com/mushiyu/go-mysql-server/sql/plan"
)

func resolveSubqueries(ctx *sql.Context, a *Analyzer, n sql.Node) (sql.Node, error) {
	span, ctx := ctx.Span("resolve_subqueries")
	defer span.Finish()

	a.Log("resolving subqueries")
	return plan.TransformUp(n, func(n sql.Node) (sql.Node, error) {
		switch n := n.(type) {
		case *plan.SubqueryAlias:
			a.Log("found subquery %q with child of type %T", n.Name(), n.Child)
			child, err := a.Analyze(ctx, n.Child)
			if err != nil {
				return nil, err
			}

			return plan.NewSubqueryAlias(n.Name(), child), nil
		default:
			return n, nil
		}
	})
}
