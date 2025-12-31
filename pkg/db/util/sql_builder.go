package util

import (
	"fmt"
	"github.com/misakacoder/kagome/str"
	"strings"
)

type QueryBuilder interface {
	Build() (string, []any)
}

type SQLBuilder struct {
	sql     string
	where   *str.Joiner
	args    []any
	orderBy string
}

func NewSQLBuilder(sql string) *SQLBuilder {
	return &SQLBuilder{
		sql:   sql,
		where: str.NewJoiner(" and ", "", ""),
	}
}

func (builder *SQLBuilder) Where(sql string, arg []any, condition bool) {
	if condition {
		builder.where.Append(sql)
		builder.args = append(builder.args, arg...)
	}
}

func (builder *SQLBuilder) OrderBy(orderBy string) {
	builder.orderBy = orderBy
}

func (builder *SQLBuilder) Build() (string, []any) {
	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString(strings.ReplaceAll(builder.sql, "\n", ""))
	orderBy := builder.orderBy
	if builder.where.Size() > 0 {
		sqlBuilder.WriteString(fmt.Sprintf(" where %s", builder.where.String()))
	}
	if orderBy != "" {
		sqlBuilder.WriteString(fmt.Sprintf(" order by %s", orderBy))
	}
	return sqlBuilder.String(), builder.args
}

type SqlBuilder struct {
	fragments []string
	args      []any
	hasWhere  bool
}

func NewSqlBuilder() *SqlBuilder {
	return &SqlBuilder{
		fragments: []string{"select"},
	}
}

func (builder *SqlBuilder) Select(field string) *SqlBuilder {
	builder.fragments = append(builder.fragments, fmt.Sprintf("%s,", field))
	return builder
}

func (builder *SqlBuilder) From(table string) *SqlBuilder {
	fragments := builder.fragments
	fragments[len(fragments)-1] = strings.TrimSuffix(fragments[len(fragments)-1], ",")
	return builder.addFragment(fmt.Sprintf("from %s", table))
}

func (builder *SqlBuilder) Join(table string, on string) *SqlBuilder {
	return builder.addFragment(fmt.Sprintf("join %s on %s", table, on))
}

func (builder *SqlBuilder) LeftJoin(table string, on string) *SqlBuilder {
	return builder.addFragment(fmt.Sprintf("left join %s on %s", table, on))
}

func (builder *SqlBuilder) RightJoin(table string, on string) *SqlBuilder {
	return builder.addFragment(fmt.Sprintf("right join %s on %s", table, on))
}

func (builder *SqlBuilder) Where(sql string, arg []any, condition bool) *SqlBuilder {
	if condition {
		fragment := fmt.Sprintf("and %s", sql)
		if !builder.hasWhere {
			builder.addFragment("where")
			fragment = sql
			builder.hasWhere = true
		}
		builder.addFragment(fragment)
		builder.args = append(builder.args, arg...)
	}
	return builder
}

func (builder *SqlBuilder) GroupBy(field string) *SqlBuilder {
	return builder.addFragment(fmt.Sprintf("group by %s", field))
}

func (builder *SqlBuilder) OrderBy(field string) *SqlBuilder {
	return builder.addFragment(fmt.Sprintf("order by %s", field))
}

func (builder *SqlBuilder) Build() (string, []any) {
	return strings.Join(builder.fragments, " "), builder.args
}

func (builder *SqlBuilder) addFragment(fragment string) *SqlBuilder {
	builder.fragments = append(builder.fragments, fragment)
	return builder
}
