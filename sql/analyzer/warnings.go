package analyzer

import (
	"github.com/mushiyu/go-mysql-server/sql"
	"github.com/mushiyu/go-mysql-server/sql/plan"
)

func clearWarnings(ctx *sql.Context, a *Analyzer, node sql.Node) (sql.Node, error) {
	children := node.Children()
	if len(children) == 0 {
		return node, nil
	}

	switch ch := children[0].(type) {
	case plan.ShowWarnings:
		return node, nil
	case *plan.Offset:
		clearWarnings(ctx, a, ch)
		return node, nil
	case *plan.Limit:
		clearWarnings(ctx, a, ch)
		return node, nil
	}

	ctx.ClearWarnings()
	return node, nil
}
