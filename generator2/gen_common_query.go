package generator2

func (g *Generator) genCommonQuery() {
	g.Pn("type QueryBase struct {")
	g.Pn("    where *bytes.Buffer")
	g.Pn("    whereParams []interface{}")
	g.Pn("}")
	g.Pn("")
}
