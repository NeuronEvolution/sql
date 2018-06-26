package generator

import (
	"fmt"
	"github.com/NeuronFramework/sql/wrap"
	"strings"
)

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

func (g *Generator) genQueryInsert(t *Table) {
	var fields []string
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		fields = append(fields, c.DbName)
	}

	var insertParams []string
	for _, c := range t.ColumnList {
		if c.AutoIncrement || c.DbName == "create_time" || c.DbName == "update_time" {
			continue
		}

		insertParams = append(insertParams, "e."+c.GoName)
	}

	batchPlaceHolder := wrap.RepeatWithSeparator("?", len(fields), ",")

	//INSERT
	g.Pn("func (q *%sQuery)Insert(ctx context.Context,tx *wrap.Tx,e *%s)(result *wrap.Result,err error){",
		t.GoName, t.GoName)
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"%s\")",
		fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t.DbName,
			strings.Join(fields, ","),
			wrap.RepeatWithSeparator("?", len(fields), ",")))
	g.Pn("    params:=[]interface{} {%s}", strings.Join(insertParams, ","))
	g.Pn("    return q.dao.db.Exec(ctx,tx,query.String(),params...)")
	g.Pn("}")
	g.Pn("")

	//BATCH INSERT
	g.Pn("func (q *%sQuery)BatchInsert(ctx context.Context,tx *wrap.Tx,list []*%s)"+
		"(result *wrap.Result,err error){", t.GoName, t.GoName)
	g.Pn("    query:=bytes.NewBufferString(\"\")")
	g.Pn("    query.WriteString(\"INSERT INTO %s (%s) VALUES \")", t.DbName, strings.Join(fields, ","))
	g.Pn("    query.WriteString(wrap.RepeatWithSeparator(\"(%s)\",len(list),\",\"))", batchPlaceHolder)
	g.Pn("    params:=make([]interface{},len(list)*%d)", len(insertParams))
	g.Pn("    offset:=0")
	g.Pn("    for _,e:=range list{")
	for i, p := range insertParams {
		g.Pn("        params[offset+%d]=%s", i, p)
	}
	g.Pn("    offset+=%d", len(insertParams))
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return q.dao.db.Exec(ctx,tx,query.String(),params...)")
	g.Pn("}")
	g.Pn("")

	genDuplicated := false
	if (t.UniqueIndexList != nil && len(t.UniqueIndexList) > 0) ||
		(t.UniqueUnionIndexList != nil && len(t.UniqueUnionIndexList) > 0) {
		genDuplicated = true
	}
	if genDuplicated {
		// INSERT DUPLICATED
		g.Pn("func (q *%sQuery)InsertOnDuplicatedKeyUpdate(ctx context.Context,tx *wrap.Tx,e *%s)"+
			"(result *wrap.Result,err error){", t.GoName, t.GoName)
		g.Pn("    query:=bytes.NewBufferString(\"\")")
		g.Pn("    query.WriteString(\"%s\")",
			fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t.DbName,
				strings.Join(fields, ","),
				wrap.RepeatWithSeparator("?", len(fields), ",")))
		g.Pn("    query.WriteString(\" ON DUPLICATED KEY UPDATE \")")
		g.Pn("    if len(q.duplicatedUpdateFields)>0{")
		g.Pn("        query.WriteString(strings.Join(q.duplicatedUpdateFields,\",\"))")
		g.Pn("        query.WriteString(\",\")")
		g.Pn("    }")
		g.Pn("    query.WriteString(\"update_time=now()\")")
		g.Pn("    params:=[]interface{} {%s}", strings.Join(insertParams, ","))
		g.Pn("    return q.dao.db.Exec(ctx,tx,query.String(),params...)")
		g.Pn("}")
		g.Pn("")

		// BATCH INSERT [DUPLICATED]
		g.Pn("func (q *%sQuery)BatchInsertOnDuplicatedKeyUpdate(ctx context.Context,tx *wrap.Tx,list []*%s)"+
			"(result *wrap.Result,err error){", t.GoName, t.GoName)
		g.Pn("    query:=bytes.NewBufferString(\"\")")
		g.Pn("    query.WriteString(\"INSERT INTO %s (%s) VALUES \")", t.DbName, strings.Join(fields, ","))
		g.Pn("    query.WriteString(wrap.RepeatWithSeparator(\"(%s)\",len(list),\",\"))", batchPlaceHolder)
		g.Pn("    query.WriteString(\" ON DUPLICATED KEY UPDATE\")")
		g.Pn("    if len(q.duplicatedUpdateFields)>0{")
		g.Pn("        query.WriteString(strings.Join(q.duplicatedUpdateFields,\",\"))")
		g.Pn("        query.WriteString(\",\")")
		g.Pn("    }")
		g.Pn("    query.WriteString(\"update_time=now()\")")
		g.Pn("    params:=make([]interface{},len(list)*%d)", len(insertParams))
		g.Pn("    offset:=0")
		g.Pn("    for _,e:=range list{")
		for i, p := range insertParams {
			g.Pn("        params[offset+%d]=%s", i, p)
		}
		g.Pn("    offset+=%d", len(insertParams))
		g.Pn("    }")
		g.Pn("")
		g.Pn("    return q.dao.db.Exec(ctx,tx,query.String(),params...)")
		g.Pn("}")
		g.Pn("")
	}
}
