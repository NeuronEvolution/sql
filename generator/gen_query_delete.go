package generator

func (g *Generator) genQueryDelete(t *Table) {
	g.Pn("func (q *%sQuery)Delete(ctx context.Context,tx *wrap.Tx)(result *wrap.Result,err error){", t.GoName)
	g.Pn("    query:=\"DELETE FROM %s WHERE \"+q.where.String()", t.DbName)
	g.Pn("    return q.dao.db.Exec(ctx,tx,query,q.whereParams...)")
	g.Pn("}")
	g.Pn("")
}
