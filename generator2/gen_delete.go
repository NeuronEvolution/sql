package generator2

func (g *Generator) genDelete(t *Table) {
	// DELETE by id
	g.Pn("func (dao *%sDao)DeleteById(ctx context.Context,tx *wrap.Tx,id int64)(result *wrap.Result,err error){",
		t.GoName)
	g.Pn("    query:=\"DELETE FROM %s WHERE id=?\"", t.GoName)
	g.Pn("    return dao.db.Exec(ctx,tx,query,id)")
	g.Pn("}")
	g.Pn("")
}
