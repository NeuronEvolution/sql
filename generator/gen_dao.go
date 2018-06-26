package generator

import "fmt"

func (g *Generator) genDao(t *Table) {
	fmt.Println("genDao", t.GoName)

	//定义
	g.Pn("type %sDao struct{", t.GoName)
	g.Pn("    logger *zap.Logger")
	g.Pn("    db *DB")
	g.Pn("}")
	g.Pn("")

	//构造函数
	g.Pn("func New%sDao(db *DB)(t *%sDao,err error){", t.GoName, t.GoName)
	g.Pn("    t=&%sDao{}", t.GoName)
	g.Pn("    t.logger=log.TypedLogger(t)")
	g.Pn("    t.db=db")
	g.Pn("    ")
	g.Pn("    return t,nil")
	g.Pn("}")
	g.Pn("")

	//新建查询
	g.Pn("func (dao *%sDao)Query() *%sQuery {", t.GoName, t.GoName)
	g.Pn("    q:= &%sQuery{}", t.GoName)
	g.Pn("    q.dao=dao")
	g.Pn("    q.tableName=\"%s\"", t.DbName)
	g.Pn("    q.where=bytes.NewBufferString(\"\")")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")
}
