package generator2

func (g *Generator) genQueryDelete(t *Table) {
	g.Pn("func (q *%sQuery)DoDelete(ctx context.Context,tx *wrap.Tx)(result *wrap.Result,err error){", t.GoName)
	g.Pn("    query:=\"DELETE FROM %s WHERE \"+q.where.String()", t.DbName)
	g.Pn("    return q.dao.db.Exec(ctx,tx,query,q.whereParams)")
	g.Pn("}")
	g.Pn("")
}
