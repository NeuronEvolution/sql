package generator

import "fmt"

func (g *Generator) genEntity(t *Table) {
	fmt.Println("genTableEntity", t.GoName)
	g.Pn("type %s struct{", t.GoName)
	for _, v := range t.ColumnList {
		if v.Size != "" {
			g.Pn("    %s %s//size=%s", v.GoName, v.GoType, v.Size)
		} else {
			g.Pn("    %s %s", v.GoName, v.GoType)
		}
	}
	g.Pn("}")
	g.Pn("")
}
