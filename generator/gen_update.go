package generator

import "strings"

func (g *Generator) genDaoUpdate(t *Table) {
	// UPDATE by id=?
	var updateStmts []string
	var updateParams []string
	for _, v := range t.ColumnList {
		if t.PrimaryColumn == v {
			continue
		}

		if v.DbName == "create_time" || v.DbName == "update_time" {
			continue
		}

		updateStmts = append(updateStmts, v.DbName+"=?")
		updateParams = append(updateParams, "e."+v.GoName)
	}
	updateParams = append(updateParams, "e.Id")
	g.Pn("func (dao *%sDao)UpdateById(ctx context.Context,tx *wrap.Tx,e *%s)(result *wrap.Result,err error){",
		t.GoName, t.GoName)
	g.Pn("    query:=\"UPDATE %s SET %s WHERE id=?\"", t.DbName, strings.Join(updateStmts, ","))
	g.Pn("    params:=[]interface{} {%s}", strings.Join(updateParams, ","))
	g.Pn("    return dao.db.Exec(ctx,tx,query,params...)")
	g.Pn("}")
	g.Pn("")
}
