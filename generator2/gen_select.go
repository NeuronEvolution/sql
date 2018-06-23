package generator2

import "strings"

func (g *Generator) getSelectFieldsAndScanParams(t *Table) (fields []string, scanParams []string) {
	for _, c := range t.ColumnList {
		scanParams = append(scanParams, "&e."+c.GoName)
		fields = append(fields, c.DbName)
	}

	return fields, scanParams
}

func (g *Generator) genSelect(t *Table) {
	fields, scanParams := g.getSelectFieldsAndScanParams(t)

	// SELECT
	g.Pn("func (dao *%sDao)SelectById(ctx context.Context,tx *wrap.Tx,id int64)(e *%s,err error){",
		t.GoName, t.GoName)
	g.Pn("    query:=\"SELECT %s FROM %s WHERE id=?\"", strings.Join(fields, ","), t.DbName)
	g.Pn("    row:=dao.db.QueryRow(ctx,tx,query,id)")
	g.Pn("    e=&%s{}", t.GoName)
	g.Pn("    err=row.Scan(%s)", strings.Join(scanParams, ","))
	g.Pn("    if err==wrap.ErrNoRows{")
	g.Pn("        return nil,nil")
	g.Pn("    }")
	g.Pn("    return e,err")
	g.Pn("}")
	g.Pn("")
}
