package generator

import (
	"bytes"
	"fmt"
	"go/format"
)

type Generator struct {
	Namespace string
	DbName    string
	TableList []*Table

	buf *bytes.Buffer
}

func NewGenerator() *Generator {
	g := &Generator{}
	g.TableList = make([]*Table, 0)
	g.buf = bytes.NewBufferString("")

	return g
}

func (g *Generator) P(format string, a ...interface{}) {
	g.buf.WriteString(fmt.Sprintf(format, a...))
}

func (g *Generator) Pn(format string, a ...interface{}) {
	g.buf.WriteString(fmt.Sprintf(format+"\n", a...))
}

func (g *Generator) Gen(sql string, namespace string) (orm string, err error) {
	g.Namespace = namespace
	err = g.parse(sql)
	if err != nil {
		return "", nil
	}

	g.gen()

	data, err := format.Source(g.buf.Bytes())
	if err != nil {
		fmt.Println(err.Error())
		return g.buf.String(), nil
	}

	return string(data), nil
}

func (g *Generator) gen() {
	g.genImports()
	g.genCommonQuery()

	for _, t := range g.TableList {
		g.genTable(t)
	}

	g.genDatabase()
}

func (g *Generator) genTable(t *Table) {
	fmt.Println("genTable", t.DbName)

	g.genEntity(t)
	g.genQuery(t)
	g.genQuerySelect(t)
	g.genQueryInsert(t)
	g.genQueryUpdate(t)
	g.genQueryDelete(t)
	g.genDao(t)
}
