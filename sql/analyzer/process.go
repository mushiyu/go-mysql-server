package analyzer

import (
	"github.com/mushiyu/go-mysql-server/sql"
	"github.com/mushiyu/go-mysql-server/sql/plan"
)

// trackProcess will wrap the query in a process node and add progress items
// to the already existing process.
func trackProcess(ctx *sql.Context, a *Analyzer, n sql.Node) (sql.Node, error) {
	if !n.Resolved() {
		return n, nil
	}

	if _, ok := n.(*plan.QueryProcess); ok {
		return n, nil
	}

	processList := a.Catalog.ProcessList

	var seen = make(map[string]struct{})
	n, err := plan.TransformUp(n, func(n sql.Node) (sql.Node, error) {
		switch n := n.(type) {
		case *plan.ResolvedTable:
			switch n.Table.(type) {
			case *plan.ProcessTable, *plan.ProcessIndexableTable:
				return n, nil
			}

			name := n.Table.Name()
			if _, ok := seen[name]; ok {
				return n, nil
			}

			var total int64 = -1
			if counter, ok := n.Table.(sql.PartitionCounter); ok {
				count, err := counter.PartitionCount(ctx)
				if err != nil {
					return nil, err
				}
				total = count
			}
			processList.AddProgressItem(ctx.Pid(), name, total)

			seen[name] = struct{}{}

			notify := func() {
				processList.UpdateProgress(ctx.Pid(), name, 1)
			}

			var t sql.Table
			switch table := n.Table.(type) {
			case sql.IndexableTable:
				t = plan.NewProcessIndexableTable(table, notify)
			default:
				t = plan.NewProcessTable(table, notify)
			}

			return plan.NewResolvedTable(t), nil
		default:
			return n, nil
		}
	})
	if err != nil {
		return nil, err
	}

	// Don't wrap CreateIndex in a QueryProcess, as it is a CreateIndexProcess.
	// CreateIndex will take care of marking the process as done on its own.
	if _, ok := n.(*plan.CreateIndex); ok {
		return n, nil
	}

	// Remove QueryProcess nodes from the subqueries. Otherwise, the process
	// will be marked as done as soon as a subquery finishes.
	node, err := plan.TransformUp(n, func(n sql.Node) (sql.Node, error) {
		if sq, ok := n.(*plan.SubqueryAlias); ok {
			if qp, ok := sq.Child.(*plan.QueryProcess); ok {
				return plan.NewSubqueryAlias(sq.Name(), qp.Child), nil
			}
		}
		return n, nil
	})
	if err != nil {
		return nil, err
	}

	return plan.NewQueryProcess(node, func() { processList.Done(ctx.Pid()) }), nil
}
