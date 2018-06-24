package generator

import (
	"fmt"
	"github.com/NeuronFramework/sql/wrap"
	"strings"
)

func (g *Generator) buildInsertFields(t *Table) (fields []string) {
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		fields = append(fields, c.DbName)
	}

	return fields
}

func (g *Generator) buildInsertParams(t *Table) (params []string) {
	for _, c := range t.ColumnList {
		if c.AutoIncrement {
			continue
		}

		if c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		params = append(params, "e."+c.GoName)
	}

	return params
}

func (g *Generator) buildInsertOnDuplicatedKeyUpdateStmts(t *Table) []string {
	var updateStmtList []string
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		hasUniqueIndex := false
		if t.UniqueIndexList != nil {
			for _, i := range t.UniqueIndexList {
				if i.Column == c {
					hasUniqueIndex = true
					break
				}
			}
		}
		if hasUniqueIndex {
			continue
		}

		if t.UniqueUnionIndexList != nil {
			for _, i := range t.UniqueUnionIndexList {
				for _, j := range i.ColumnList {
					if j == c {
						hasUniqueIndex = true
						break
					}
				}
				if hasUniqueIndex {
					break
				}
			}
		}
		if hasUniqueIndex {
			continue
		}

		updateStmtList = append(updateStmtList, c.DbName+"=VALUES("+c.DbName+")")
	}

	return updateStmtList
}

func (g *Generator) buildInsertStmt(t *Table) string {
	fields := g.buildInsertFields(t)
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t.DbName,
		strings.Join(fields, ","),
		wrap.RepeatWithSeparator("?", len(fields), ","))
}

func (g *Generator) genDaoInsert(t *Table) {
	fields := g.buildInsertFields(t)
	insertParams := g.buildInsertParams(t)

	genDuplicated := false
	if (t.UniqueIndexList != nil && len(t.UniqueIndexList) > 0) ||
		(t.UniqueUnionIndexList != nil && len(t.UniqueUnionIndexList) > 0) {
		genDuplicated = true
	}

	// INSERT [DUPLICATED]
	if genDuplicated {
		g.Pn("func (dao *%sDao)Insert(ctx context.Context,tx *wrap.Tx,e *%s,onDuplicatedKeyUpdate bool)"+
			"(result *wrap.Result,err error){", t.GoName, t.GoName)
	} else {
		g.Pn("func (dao *%sDao)Insert(ctx context.Context,tx *wrap.Tx,e *%s)(result *wrap.Result,err error){",
			t.GoName, t.GoName)
	}
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"%s\")", g.buildInsertStmt(t))
	if genDuplicated {
		g.Pn("    if onDuplicatedKeyUpdate{")
		g.Pn("        query.WriteString(\" ON DUPLICATED KEY UPDATE %s\")",
			strings.Join(g.buildInsertOnDuplicatedKeyUpdateStmts(t), ","))
		g.Pn("    }")
	}
	g.Pn("    params:=[]interface{} {%s}", strings.Join(insertParams, ","))
	g.Pn("    return dao.db.Exec(ctx,tx,query.String(),params...)")
	g.Pn("}")
	g.Pn("")

	// BATCH INSERT [DUPLICATED]
	batchPlaceHolder := wrap.RepeatWithSeparator("?", len(fields), ",")
	if genDuplicated {
		g.Pn("func (dao *%sDao)BatchInsert(ctx context.Context,tx *wrap.Tx,list []*%s,onDuplicatedKeyUpdate bool)"+
			"(result *wrap.Result,err error){", t.GoName, t.GoName)
	} else {
		g.Pn("func (dao *%sDao)BatchInsert(ctx context.Context,tx *wrap.Tx,list []*%s)"+
			"(result *wrap.Result,err error){", t.GoName, t.GoName)
	}
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"INSERT INTO %s (%s) VALUES \")", t.DbName, strings.Join(fields, ","))
	g.Pn("    query.WriteString(wrap.RepeatWithSeparator(\"(%s)\",len(list),\",\"))", batchPlaceHolder)
	if genDuplicated {
		g.Pn("    if onDuplicatedKeyUpdate{")
		g.Pn("        query.WriteString(\" ON DUPLICATED KEY UPDATE %s\")",
			strings.Join(g.buildInsertOnDuplicatedKeyUpdateStmts(t), ","))
		g.Pn("    }")
	}
	g.Pn("    params:=make([]interface{},len(list)*%d)", len(insertParams))
	g.Pn("    offset:=0")
	g.Pn("    for _,e:=range list{")
	for i, p := range insertParams {
		g.Pn("        params[offset+%d]=%s", i, p)
	}
	g.Pn("    offset+=%d", len(insertParams))
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return dao.db.Exec(ctx,tx,query.String(),params...)")
	g.Pn("}")
	g.Pn("")
}
