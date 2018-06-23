package generator2

func (g *Generator) genDatabase() {
	//def
	g.Pn("type DB struct{")
	g.Pn("    wrap.DB")
	for _, v := range g.TableList {
		g.Pn("    %s *%sDao", v.GoName, v.GoName)
	}
	g.Pn("}")
	g.Pn("")

	//new
	g.Pn("func NewDB() (d *DB, err error) {")
	g.Pn("    d = &DB{}")
	g.Pn("")
	g.Pn("    connectionString := os.Getenv(\"DB\")")
	g.Pn("    if connectionString == \"\" {")
	g.Pn("	      return nil, fmt.Errorf(\"DB env nil\")")
	g.Pn("    }")
	g.Pn("    connectionString+=\"/%s?parseTime=true\"", g.DbName)
	g.Pn("db, err := wrap.Open(\"mysql\", connectionString)")
	g.Pn("if err != nil {")
	g.Pn("	return nil, err")
	g.Pn("}")
	g.Pn("d.DB=*db")
	g.Pn("")
	g.Pn("err = d.Ping(context.Background())")
	g.Pn("if err != nil {")
	g.Pn("	return nil, err")
	g.Pn("}")
	g.Pn("")
	for _, v := range g.TableList {
		g.Pn("    d.%s,err=New%sDao(d)", v.GoName, v.GoName)
		g.Pn("    if err != nil {")
		g.Pn("	     return nil, err")
		g.Pn("    }")
		g.Pn("")
	}
	g.Pn("return d,nil")
	g.Pn("}")
	g.Pn("")
}
