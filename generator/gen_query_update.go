package generator

func (g *Generator) genQueryUpdate(t *Table) {
	// 更新字段
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		g.Pn("func (q *%sQuery)Set%s(v %s)*%sQuery{", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
		g.Pn("    q.updateFields=append(q.updateFields,\"%s\")", c.DbName)
		g.Pn("    q.updateParams=append(q.updateParams,v)")
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")
	}

	// 执行更新
	g.Pn("func (q *%sQuery)Update(ctx context.Context,tx *wrap.Tx)(result *wrap.Result,err error){", t.GoName)
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    var params []interface{}")
	g.Pn("    params=append(params,q.updateParams)")
	g.Pn("    query.WriteString(\"UPDATE %s SET \")", t.DbName)
	g.Pn("    updateItems:=make([]string,len(q.updateFields))")
	g.Pn("    for i,v:=range q.updateFields{")
	g.Pn("        updateItems[i]=v+\"=?\"")
	g.Pn("    }")
	g.Pn("    query.WriteString(strings.Join(updateItems,\",\"))")
	g.Pn("    where:=q.where.String()")
	g.Pn("    if where!=\"\"{")
	g.Pn("        query.WriteString(\" WHERE \")")
	g.Pn("        query.WriteString(where)")
	g.Pn("        params=append(params,q.whereParams)")
	g.Pn("    }")
	g.Pn("    ")
	g.Pn("    return q.dao.db.Exec(ctx,tx,query.String(),params...)")
	g.Pn("}")
	g.Pn("")
}
