package generator

func (g *Generator) genQuery(t *Table) {
	// 查询对象定义
	g.Pn("type %sQuery struct {", t.GoName)
	g.Pn("    QueryBase")
	g.Pn("    dao *%sDao", t.GoName)
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
			g.Pn("    q.where.WriteString(\" %s IN(\")", c.DbName)
			g.Pn("    q.where.WriteString(wrap.RepeatWithSeparator(\"?\",len(items),\",\"))")
			g.Pn("    q.where.WriteString(\")\")")
			g.Pn("    q.whereParams=append(q.whereParams,items)")
			g.Pn("    return q")
			g.Pn("}")
			g.Pn("")
		}
	}

	//分组
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

	//排序
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

	//按分组数量排序
	g.Pn("func (q *%sQuery)OrderByGroupCount(asc bool) *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.orderByFields=append(q.orderByFields,\"count(*)\")")
	g.Pn("    q.orderByOrders=append(q.orderByOrders,asc)")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	//分页
	g.Pn("func (q *%sQuery)Limit(startIncluded int64,count int64) *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.hasLimit=true")
	g.Pn("    q.limitStartIncluded=startIncluded")
	g.Pn("    q.limitCount=count")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	//写锁
	g.Pn("func (q *%sQuery)ForUpdate() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.forUpdate=true")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	//共享锁
	g.Pn("func (q *%sQuery)ForShare() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q.forShare=true")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	//更新字段
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

	//重复时更新字段
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		isUniqueIndex := false
		if t.UniqueIndexList != nil {
			for _, i := range t.UniqueIndexList {
				if i.Column == c {
					isUniqueIndex = true
					break
				}
			}
		}
		if isUniqueIndex {
			continue
		}

		if t.UniqueUnionIndexList != nil {
			for _, i := range t.UniqueUnionIndexList {
				for _, j := range i.ColumnList {
					if j == c {
						isUniqueIndex = true
						break
					}
				}
				if isUniqueIndex {
					break
				}
			}
		}
		if isUniqueIndex {
			continue
		}

		g.Pn("func (q *%sQuery)DuplicatedUpdate%s()*%sQuery{", t.GoName, c.GoName, t.GoName)
		g.Pn("    q.duplicatedUpdateFields=append(q.duplicatedUpdateFields,\"%s=VALUES(%s)\")", c.DbName, c.DbName)
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")
	}

	//返回指定字段
	for _, c := range t.ColumnList {
		g.Pn("func (q *%sQuery)Get%s()*%sQuery{", t.GoName, c.GoName, t.GoName)
		g.Pn("    q.getFields=append(q.getFields,\"%s\")", c.DbName)
		g.Pn("    return q")
		g.Pn("}")
		g.Pn("")
	}
}
