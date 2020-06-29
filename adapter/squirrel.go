package adapter

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/godzie44/d3/orm/query"
	"strings"
)

func QueryToSql(q *query.Query) (string, []interface{}, error) {
	sqQuery, err := toSquirrel(q)
	if err != nil {
		return "", nil, err
	}

	return sqQuery.PlaceholderFormat(squirrel.Dollar).ToSql()
}

func toSquirrel(q *query.Query) (*squirrel.SelectBuilder, error) {
	sb := squirrel.SelectBuilder{}

	var whereExpr squirrel.Sqlizer

	query.Visit(q, func(pred interface{}) {
		switch p := pred.(type) {
		case query.From:
			sb = sb.From(string(p))
		case query.Columns:
			for i := range p {
				p[i] = fmt.Sprintf("%s as \"%s\"", p[i], p[i])
			}
			sb = sb.Columns(p...)

		case *query.AndWhere:
			var expr = createWhereExpr(p.Where)
			if whereExpr == nil {
				whereExpr = expr
			} else {
				whereExpr = squirrel.And{whereExpr, expr}
			}
		case *query.OrWhere:
			var expr = createWhereExpr(p.Where)
			if whereExpr == nil {
				whereExpr = expr
			} else {
				whereExpr = squirrel.Or{whereExpr, expr}
			}
		case *query.Having:
			sb = sb.Having(p.Expr, p.Params...)
		case *query.Join:
			switch p.Type {
			case query.JoinLeft:
				sb = sb.LeftJoin(p.Join + " ON " + p.On)
			case query.JoinInner:
				sb = sb.Join(p.Join + " ON " + p.On)
			case query.JoinRight:
				sb = sb.RightJoin(p.Join + " ON " + p.On)
			}
		case query.Order:
			sb = sb.OrderBy(p...)
		case *query.Union:
			sqQuery, err := toSquirrel(p.Q)
			if err != nil {
				return
			}
			sql, args, err := sqQuery.PlaceholderFormat(squirrel.Question).ToSql()
			if err != nil {
				return
			}

			sb = sb.Suffix("UNION "+sql, args...)
		case query.GroupBy:
			sb = sb.GroupBy(string(p))
		case query.Limit:
			sb = sb.Limit(uint64(p))
		case query.Offset:
			sb = sb.Offset(uint64(p))
		}
	})
	sb = sb.Where(whereExpr)

	return &sb, nil
}

func createWhereExpr(where query.Where) squirrel.Sqlizer {
	switch len(where.Params) {
	case 0:
		return squirrel.Expr(strings.Join([]string{where.Field, where.Op}, " "))
	case 1:
		return squirrel.Expr(strings.Join([]string{where.Field, where.Op, "?"}, " "), where.Params[0])
	case 2:
		return squirrel.Expr(strings.Join([]string{where.Field, where.Op, "? AND ?"}, " "), where.Params...)
	default:
		return squirrel.Expr("")
	}
}
