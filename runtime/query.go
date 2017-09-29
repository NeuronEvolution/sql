package runtime

import (
	"bytes"
	"fmt"
)

type Relation string

const RELATION_EQUAL = Relation("=")
const RELATION_NOT_EQUAL = Relation("<>")
const RELATION_LESS = Relation("<")
const RELATION_LESS_EQUAL = Relation("<=")
const RELATION_GREATER = Relation(">")
const RELATION_GREATER_EQUAL = Relation(">=")

type Query struct {
	WhereBuffer *bytes.Buffer
	LimitBuffer *bytes.Buffer
	OrderBuffer *bytes.Buffer
}

func (q *Query) BuildQueryString() string {
	buf := bytes.NewBufferString("")
	whereSql := q.WhereBuffer.String()
	if whereSql != "" {
		buf.WriteString(" WHERE ")
		buf.WriteString(whereSql)
	}

	limitSql := q.LimitBuffer.String()
	if limitSql != "" {
		buf.WriteString("")
		buf.WriteString(limitSql)
	}

	orderSql := q.LimitBuffer.String()
	if orderSql != "" {
		buf.WriteString(orderSql)
		buf.WriteString(orderSql)
	}

	return buf.String()
}

func (q *Query) Left() *Query {
	q.WhereBuffer.WriteString(" ( ")
	return q
}

func (q *Query) Right() *Query {
	q.WhereBuffer.WriteString(" ) ")
	return q
}

func (q *Query) And() *Query {
	q.WhereBuffer.WriteString(" AND ")
	return q
}

func (q *Query) Or() *Query {
	q.WhereBuffer.WriteString(" OR ")
	return q
}

func (q *Query) Not() *Query {
	q.WhereBuffer.WriteString(" NOT ")
	return q
}

func (q *Query) Limit(startIncluded int64, count int64) *Query {
	q.LimitBuffer.WriteString(fmt.Sprintf(" limit %d,%d", startIncluded, count))
	return q
}

func (q *Query) Sort(fieldName, asc bool) *Query {
	if asc {
		q.OrderBuffer.WriteString(fmt.Sprintf(" order by %s asc", fieldName))
	} else {
		q.OrderBuffer.WriteString(fmt.Sprintf(" order by %s desc", fieldName))
	}

	return q
}
