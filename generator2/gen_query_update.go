package generator2

func (g *Generator) genQueryUpdate(t *Table) {
	// 构造器定义
	g.Pn("type %sUpdateBuilder struct {", t.GoName)
	g.Pn("    dao *%sDao", t.GoName)
	g.Pn("    query *%sQuery", t.GoName)
	g.Pn("    updateFields []string")
	g.Pn("    updateParams []interface{}")
	g.Pn("}")
	g.Pn("")

	// 创建构造器
	g.Pn("func (q *%sQuery)DoUpdate()*%sUpdateBuilder{", t.GoName, t.GoName)
	g.Pn("    b:=&%sUpdateBuilder{}", t.GoName)
	g.Pn("    b.dao=q.dao")
	g.Pn("    b.query=q")
	g.Pn("    return b")
	g.Pn("}")
	g.Pn("")

	// 执行更新
	g.Pn("func (b *%sUpdateBuilder)Update(ctx context.Context,tx *wrap.Tx)(result *wrap.Result,err error){", t.GoName)
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    var params []interface{}")
	g.Pn("    params=append(params,b.updateParams)")
	g.Pn("    query.WriteString(\"UPDATE %s SET \")", t.DbName)
	g.Pn("    updateItems:=make([]string,len(b.updateFields))")
	g.Pn("    for i,v:=range b.updateFields{")
	g.Pn("        updateItems[i]=v+\"=?\"")
	g.Pn("    }")
	g.Pn("    query.WriteString(strings.Join(updateItems,\",\"))")
	g.Pn("    where:=b.query.where.String()")
	g.Pn("    if where!=\"\"{")
	g.Pn("        query.WriteString(\" WHERE \")")
	g.Pn("        query.WriteString(where)")
	g.Pn("        params=append(params,b.query.whereParams)")
	g.Pn("    }")
	g.Pn("    ")
	g.Pn("    return b.dao.db.Exec(ctx,tx,query.String(),params)")
	g.Pn("}")
	g.Pn("")

	// 更新字段
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		g.Pn("func (b *%sUpdateBuilder)Set%s(v %s)*%sUpdateBuilder{", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
		g.Pn("    b.updateFields=append(b.updateFields,\"%s\")", c.DbName)
		g.Pn("    b.updateParams=append(b.updateParams,v)")
		g.Pn("    return b")
		g.Pn("}")
		g.Pn("")
	}
}
