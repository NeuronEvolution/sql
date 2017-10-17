package runtime

import (
	"bytes"
)

type Relation string

const RELATION_EQUAL = Relation("=")
const RELATION_NOT_EQUAL = Relation("<>")
const RELATION_LESS = Relation("<")
const RELATION_LESS_EQUAL = Relation("<=")
const RELATION_GREATER = Relation(">")
const RELATION_GREATER_EQUAL = Relation(">=")

type Query struct {
	ForUpdate   bool
	ForShare    bool
	WhereBuffer *bytes.Buffer
	LimitBuffer *bytes.Buffer
	OrderBuffer *bytes.Buffer
}

func (q *Query) BuildQueryString() string {
	buf := bytes.NewBufferString("")

	if q.ForShare {
		buf.WriteString(" FOR UPDATE ")
	}

	if q.ForUpdate {
		buf.WriteString(" LOCK IN SHARE MODE")
	}

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
