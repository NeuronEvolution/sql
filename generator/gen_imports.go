package generator

import "strings"

func (g *Generator) genImports() {
	g.Pn("package %s", strings.Replace(g.Namespace, "-", "_", -1))
	g.Pn("")
	g.Pn("import(")
	g.Pn("    \"go.uber.org/zap\"")
	g.Pn("    \"github.com/NeuronFramework/log\"")
	g.Pn("    \"bytes\"")
	g.Pn("    \"fmt\"")
	g.Pn("    \"os\"")
	g.Pn("    \"time\"")
	g.Pn("    \"strings\"")
	g.Pn("    \"context\"")
	g.Pn("    \"database/sql\"")
	g.Pn("    \"github.com/NeuronFramework/sql/wrap\"")
	g.Pn("     \"github.com/go-sql-driver/mysql\"")
	g.Pn(")")
	g.Pn("")

	g.Pn("var _ =sql.ErrNoRows")
	g.Pn("var _ =mysql.ErrOldProtocol")
}
