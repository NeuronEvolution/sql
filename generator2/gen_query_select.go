package generator2

import "strings"

func (g *Generator) genQuerySelect(t *Table) {
	// 构造器定义
	g.Pn("type %sSelectBuilder struct {", t.GoName)
	g.Pn("    SelectBuilderBase")
	g.Pn("    dao *%sDao", t.GoName)
	g.Pn("}")
	g.Pn("")

	// 创建构造器
	g.Pn("func (q *%sQuery)DoSelect()*%sSelectBuilder{", t.GoName, t.GoName)
	g.Pn("    b:=&%sSelectBuilder{}", t.GoName)
	g.Pn("    b.dao=q.dao")
	g.Pn("    b.query=&q.QueryBase")
	g.Pn("    return b")
	g.Pn("}")
	g.Pn("")

	// 分组
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" ||
			c.DbName == "update_time" || c.DbName == "update_version" ||
			c.IsUniqueIndex(t) {
			continue
		}

		g.Pn("func (b *%sSelectBuilder)GroupBy%s(asc bool) *%sSelectBuilder {", t.GoName, c.GoName, t.GoName)
		g.Pn("    b.groupByFields=append(b.groupByFields,\"%s\")", c.DbName)
		g.Pn("    b.groupByOrders=append(b.groupByOrders,asc)")
		g.Pn("    return b")
		g.Pn("}")
		g.Pn("")
	}

	// 排序
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" ||
			c.DbName == "update_time" || c.DbName == "update_version" ||
			c.IsUniqueIndex(t) {
			continue
		}

		g.Pn("func (b *%sSelectBuilder)OrderBy%s(asc bool) *%sSelectBuilder {", t.GoName, c.GoName, t.GoName)
		g.Pn("    b.orderByFields=append(b.orderByFields,\"%s\")", c.DbName)
		g.Pn("    b.orderByOrders=append(b.orderByOrders,asc)")
		g.Pn("    return b")
		g.Pn("}")
		g.Pn("")
	}

	// 分页
	g.Pn("func (b *%sSelectBuilder)Limit(startIncluded int64,count int64) *%sSelectBuilder {", t.GoName, t.GoName)
	g.Pn("    b.hasLimit=true")
	g.Pn("    b.limitStartIncluded=startIncluded")
	g.Pn("    b.limitCount=count")
	g.Pn("    return b")
	g.Pn("}")
	g.Pn("")

	// 写锁
	g.Pn("func (b *%sSelectBuilder)ForUpdate() *%sSelectBuilder {", t.GoName, t.GoName)
	g.Pn("    b.forUpdate=true")
	g.Pn("    return b")
	g.Pn("}")
	g.Pn("")

	// 共享锁
	g.Pn("func (b *%sSelectBuilder)ForShare() *%sSelectBuilder {", t.GoName, t.GoName)
	g.Pn("    b.forShare=true")
	g.Pn("    return b")
	g.Pn("}")
	g.Pn("")

	fields, scanParams := g.getSelectFieldsAndScanParams(t)

	// 查询单条纪录
	g.Pn("func (b *%sSelectBuilder)Select(ctx context.Context,tx *wrap.Tx) (e *%s,err error) {",
		t.GoName, t.GoName)
	g.Pn("    if !b.hasLimit{")
	g.Pn("        b.limitCount=1")
	g.Pn("        b.hasLimit=true")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    queryString,params:=b.buildQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT %s FROM %s \")", strings.Join(fields, ","), t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("    e=&%s{}", t.GoName)
	g.Pn("    row:=b.dao.db.QueryRow(ctx,tx,query.String(),params)")
	g.Pn("    err=row.Scan(%s)", strings.Join(scanParams, ","))
	g.Pn("    if err==wrap.ErrNoRows{")
	g.Pn("        return nil,nil")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return e,err")
	g.Pn("}")
	g.Pn("")

	// 查询列表
	g.Pn("func (b *%sSelectBuilder)SelectList(ctx context.Context,tx *wrap.Tx) (list []*%s,err error) {",
		t.GoName, t.GoName)
	g.Pn("    queryString,params:=b.buildQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT %s FROM %s \")", strings.Join(fields, ","), t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("    rows,err:=b.dao.db.Query(ctx,tx,query.String(),params)")
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
	g.Pn("func (b *%sSelectBuilder)SelectCount(ctx context.Context,tx *wrap.Tx) (count int64,err error) {",
		t.GoName)
	g.Pn("    queryString,params:=b.buildQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT COUNT(*) FROM %s \")", t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("    row:=b.dao.db.QueryRow(ctx,tx,query.String(),params)")
	g.Pn("    err=row.Scan(&count)")
	g.Pn("")
	g.Pn("    return count,err")
	g.Pn("}")
	g.Pn("")

	// 分组查询
	g.Pn("func (b *%sSelectBuilder)SelectGroupBy(ctx context.Context,tx *wrap.Tx,withCount bool) "+
		"(rows *wrap.Rows,err error) {", t.GoName)
	g.Pn("    queryString,params:=b.buildQuery()")
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"SELECT \")")
	g.Pn("    query.WriteString(strings.Join(b.groupByFields,\",\"))")
	g.Pn("    if withCount{")
	g.Pn("        query.WriteString(\",Count(*) \")")
	g.Pn("    }")
	g.Pn("    query.WriteString(\" FROM %s \")", t.DbName)
	g.Pn("    query.WriteString(queryString)")
	g.Pn("")
	g.Pn("    return b.dao.db.Query(ctx,tx,query.String(),params)")
	g.Pn("}")
	g.Pn("")
}
