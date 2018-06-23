package generator2

import "fmt"

func (g *Generator) genDao(t *Table) {
	fmt.Println("genDao", t.GoName)

	// 定义
	g.Pn("type %sDao struct{", t.GoName)
	g.Pn("    logger *zap.Logger")
	g.Pn("    db *DB")
	g.Pn("}")
	g.Pn("")

	// 构造函数
	g.Pn("func New%sDao(db *DB)(t *%sDao,err error){", t.GoName, t.GoName)
	g.Pn("    t=&%sDao{}", t.GoName)
	g.Pn("    t.logger=log.TypedLogger(t)")
	g.Pn("    t.db=db")
	g.Pn("    ")
	g.Pn("    return t,nil")
	g.Pn("}")
	g.Pn("")
}
