package generator2

func (g *Generator) genQuery(t *Table) {
	// 查询对象定义
	g.Pn("type %sQuery struct {", t.GoName)
	g.Pn("    QueryBase")
	g.Pn("    dao *%sDao", t.GoName)
	g.Pn("}")
	g.Pn("")

	// 构造函数
	g.Pn("func (dao *%sDao)NewQuery() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q:= &%sQuery{}", t.GoName)
	g.Pn("    q.dao=dao")
	g.Pn("    q.where=bytes.NewBufferString(\"\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 左括号
	g.Pn("func(q *%sQuery)Left() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.where.WriteString(\" (\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 右括号
	g.Pn("func(q *%sQuery)Right() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.where.WriteString(\" )\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 与
	g.Pn("func(q *%sQuery)And() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.where.WriteString(\" AND\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 或
	g.Pn("func(q *%sQuery)Or() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.where.WriteString(\" OR\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 非
	g.Pn("func(q *%sQuery)Not() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.where.WriteString(\" NOT\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	// 每个列相关的函数
	for _, c := range t.ColumnList {
		// 相等
		g.Pn("func (q *%sQuery)%sEqual(v %s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
		g.Pn("    q.where.WriteString(\" %s=?\")", c.DbName)
		g.Pn("    q.whereParams=append(q.whereParams,v)")
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")

		// 不等
		g.Pn("func (q *%sQuery)%sNotEqual(v %s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
		g.Pn("    q.where.WriteString(\" %s<>?\")", c.DbName)
		g.Pn("    q.whereParams=append(q.whereParams,v)")
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")

		// 非字符串生成比较函数
		if c.GoTypeReal != "string" {
			// 小于
			g.Pn("func (q *%sQuery)%sLess(v %s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
			g.Pn("    q.where.WriteString(\" %s<?\")", c.DbName)
			g.Pn("    q.whereParams=append(q.whereParams,v)")
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")

			// 小于等于
			g.Pn("func (q *%sQuery)%sLessEqual(v %s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
			g.Pn("    q.where.WriteString(\" %s<=?\")", c.DbName)
			g.Pn("    q.whereParams=append(q.whereParams,v)")
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")

			// 大于
			g.Pn("func (q *%sQuery)%sGreater(v %s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
			g.Pn("    q.where.WriteString(\" %s>?\")", c.DbName)
			g.Pn("    q.whereParams=append(q.whereParams,v)")
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")

			// 大于等于
			g.Pn("func (q *%sQuery)%sGreaterEqual(v %s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
			g.Pn("    q.where.WriteString(\" %s>=?\")", c.DbName)
			g.Pn("    q.whereParams=append(q.whereParams,v)")
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")
		}

		// 可空字段判空
		if !c.NotNull {
			g.Pn("func (q *%sQuery)%sIsNull() *%sQuery {", t.GoName, c.GoName, t.GoName)
			g.Pn("    q.where.WriteString(\" %s IS NULL\")", c.DbName)
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")

			g.Pn("func (q *%sQuery)%sIsNotNull() *%sQuery {", t.GoName, c.GoName, t.GoName)
			g.Pn("    q.where.WriteString(\" %s IS NOT NULL\")", c.DbName)
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")
		}

		// In查询
		if c.GoTypeReal != "float32" && c.GoTypeReal != "float64" && c.GoTypeReal != "time.Time" {
			g.Pn("func (q *%sQuery)%sIn(items []%s) *%sQuery {", t.GoName, c.GoName, c.GoTypeReal, t.GoName)
			g.Pn("    count:=len(items)")
			g.Pn("    if count==0{")
			g.Pn("        return q")
			g.Pn("    }")
			g.Pn("")
			g.Pn("    q.where.WriteString(\" %s IN(\")", c.DbName)
			g.Pn("    if count>=1{")
			g.Pn("        q.where.WriteString(\"?\")")
			g.Pn("        q.where.WriteString(strings.Repeat(\",?\",count-1))")
			g.Pn("    }else{")
			g.Pn("        q.where.WriteString(\"?\")")
			g.Pn("    }")
			g.Pn("    q.where.WriteString(\")\")")
			g.Pn("    q.whereParams=append(q.whereParams,items)")
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")
		}
	}
}
