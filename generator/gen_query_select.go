package generator

import "strings"

func (g *Generator) genQuerySelect(t *Table) {
	fields, scanParams := g.getSelectFieldsAndScanParams(t)

	// 分组
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" ||
			c.DbName == "update_time" || c.DbName == "update_version" ||
			c.IsUniqueIndex(t) {
			continue
		}

		g.Pn("func (q *%sQuery)GroupBy%s(asc bool) *%sQuery {", t.GoName, c.GoName, t.GoName)
		g.Pn("    q.groupByFields=append(q.groupByFields,\"%s\")", c.DbName)
		g.Pn("    q.groupByOrders=append(q.groupByOrders,asc)")
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")
	}

	// 排序
	for _, c := range t.ColumnList {
		if c.DbName == "update_version" {
			continue
		}

		g.Pn("func (q *%sQuery)OrderBy%s(asc bool) *%sQuery {", t.GoName, c.GoName, t.GoName)
		g.Pn("    q.orderByFields=append(q.orderByFields,\"%s\")", c.DbName)
		g.Pn("    q.orderByOrders=append(q.orderByOrders,asc)")
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")
	}

	// 按分组数量排序
	g.Pn("func (q *%sQuery)OrderByGroupCount(asc bool) *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.orderByFields=append(q.orderByFields,\"count(*)\")")
	g.Pn("    q.orderByOrders=append(q.orderByOrders,asc)")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 分页
	g.Pn("func (q *%sQuery)Limit(startIncluded int64,count int64) *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.hasLimit=true")
	g.Pn("    q.limitStartIncluded=startIncluded")
	g.Pn("    q.limitCount=count")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 写锁
	g.Pn("func (q *%sQuery)ForUpdate() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.forUpdate=true")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 共享锁
	g.Pn("func (q *%sQuery)ForShare() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.forShare=true")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 查询单条纪录
	g.Pn("func (q *%sQuery)Select(ctx context.Context,tx *wrap.Tx) (e *%s,err error) {",
		t.GoName, t.GoName)
	g.Pn("    if !q.hasLimit{")
	g.Pn("        q.limitCount=1")
	g.Pn("        q.hasLimit=true")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    queryString,params:=q.buildSelectQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT %s FROM %s \")", strings.Join(fields, ","), t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("    e=&%s{}", t.GoName)
	g.Pn("    row:=q.dao.db.QueryRow(ctx,tx,query.String(),params...)")
	g.Pn("    err=row.Scan(%s)", strings.Join(scanParams, ","))
	g.Pn("    if err==wrap.ErrNoRows{")
	g.Pn("        return nil,nil")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return e,err")
	g.Pn("}")
	g.Pn("")

	// 查询列表
	g.Pn("func (q *%sQuery)SelectList(ctx context.Context,tx *wrap.Tx) (list []*%s,err error) {",
		t.GoName, t.GoName)
	g.Pn("    queryString,params:=q.buildSelectQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT %s FROM %s \")", strings.Join(fields, ","), t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("    rows,err:=q.dao.db.Query(ctx,tx,query.String(),params...)")
	g.Pn("    if err!=nil{")
	g.Pn("        return nil,err")
	g.Pn("    }")
	g.Pn("    for rows.Next(){")
	g.Pn("        e:=%s{}", t.GoName)
	g.Pn("        err=rows.Scan(%s)", strings.Join(scanParams, ","))
	g.Pn("        if err!=nil{")
	g.Pn("            return nil,err")
	g.Pn("        }")
	g.Pn("        list=append(list,&e)")
	g.Pn("    }")
	g.Pn("    if rows.Err()!=nil{")
	g.Pn("        err=rows.Err()")
	g.Pn("        return nil,err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return list,nil")
	g.Pn("}")
	g.Pn("")

	// 查询数量
	g.Pn("func (q *%sQuery)SelectCount(ctx context.Context,tx *wrap.Tx) (count int64,err error) {",
		t.GoName)
	g.Pn("    queryString,params:=q.buildSelectQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT COUNT(*) FROM %s \")", t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("    row:=q.dao.db.QueryRow(ctx,tx,query.String(),params...)")
	g.Pn("    err=row.Scan(&count)")
	g.Pn("")
	g.Pn("    return count,err")
	g.Pn("}")
	g.Pn("")

	// 分组查询
	g.Pn("func (q *%sQuery)SelectGroupBy(ctx context.Context,tx *wrap.Tx,withCount bool) "+
		"(rows *wrap.Rows,err error) {", t.GoName)
	g.Pn("    queryString,params:=q.buildSelectQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT \")")
	g.Pn("    query.WriteString(strings.Join(q.groupByFields,\",\"))")
	g.Pn("    if withCount{")
	g.Pn("        query.WriteString(\",Count(*) \")")
	g.Pn("    }")
	g.Pn("    query.WriteString(\" FROM %s \")", t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("")
	g.Pn("    return q.dao.db.Query(ctx,tx,query.String(),params...)")
	g.Pn("}")
	g.Pn("")
}
