package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/mushiyu/go-mysql-server/sql"
	"github.com/mushiyu/go-mysql-server/sql/plan"
	"gopkg.in/src-d/go-errors.v1"
)

func TestParseShowCreateTableQuery(t *testing.T) {
	testCases := []struct {
		query  string
		result sql.Node
		err    *errors.Kind
	}{
		{
			"SHOW CREATE",
			nil,
			errUnsupportedShowCreateQuery,
		},
		{
			"SHOW CREATE ANYTHING",
			nil,
			errUnsupportedShowCreateQuery,
		},
		{
			"SHOW CREATE ASDF foo",
			nil,
			errUnsupportedShowCreateQuery,
		},
		{
			"SHOW CREATE TABLE mytable",
			plan.NewShowCreateTable("", nil, "mytable"),
			nil,
		},
		{
			"SHOW CREATE TABLE `mytable`",
			plan.NewShowCreateTable("", nil, "mytable"),
			nil,
		},
		{
			"SHOW CREATE TABLE mydb.`mytable`",
			plan.NewShowCreateTable("mydb", nil, "mytable"),
			nil,
		},
		{
			"SHOW CREATE TABLE `mydb`.mytable",
			plan.NewShowCreateTable("mydb", nil, "mytable"),
			nil,
		},
		{
			"SHOW CREATE TABLE `mydb`.`mytable`",
			plan.NewShowCreateTable("mydb", nil, "mytable"),
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.query, func(t *testing.T) {
			require := require.New(t)

			result, err := parseShowCreate(tt.query)
			if tt.err != nil {
				require.Error(err)
				require.True(tt.err.Is(err))
			} else {
				require.NoError(err)
				require.Equal(tt.result, result)
			}
		})
	}
}
