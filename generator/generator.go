package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
)

type Generator struct {
	Namespace string
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

func (g *Generator) genErrorLog(msg string) {
	g.Pn("        dao.logger.Error(\"%s\",zap.Error(err))", msg)
}

func (g *Generator) genDriverErrorLog() {
	g.genErrorLog("sqlDriver")
}

func (g *Generator) genConstants(t *Table) {
	g.Pn("const %s_TABLE_NAME = \"%s\"", strings.ToUpper(t.DbName), t.DbName)
	g.Pn("")

	var filedNameList []string
	for _, v := range t.ColumnList {
		g.Pn("const %s_FIELD_%s = \"%s\"", strings.ToUpper(t.DbName), strings.ToUpper(v.DbName), v.DbName)
		filedNameList = append(filedNameList, fmt.Sprintf("%s", v.DbName))
	}
	g.Pn("")
	g.Pn("const %s_ALL_FIELDS_STRING =\"%s\"", strings.ToUpper(t.DbName), strings.Join(filedNameList, ","))
	g.Pn("")
	g.Pn("var %s_ALL_FIELDS = []string{\n\"%s\",\n}", strings.ToUpper(t.DbName), strings.Join(filedNameList, "\",\n\""))
}

func (g *Generator) genEntity(t *Table) {
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

func (g *Generator) genPrepareInsertStmt(t *Table) {
	var fieldNames []string
	var fieldValues []string
	for _, v := range t.ColumnList {
		if v.AutoIncrement {
			continue
		}
		fieldNames = append(fieldNames, v.DbName)
		fieldValues = append(fieldValues, "?")
	}

	g.Pn("func (dao *%sDao) prepareInsertStmt() (err error){", t.GoName)
	g.Pn("    dao.insertStmt,err=dao.db.Prepare(context.Background(),\"INSERT INTO %s (%s) VALUES (%s)\")",
		t.DbName,
		strings.Join(fieldNames, ","),
		strings.Join(fieldValues, ","))
	g.Pn("    return err")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genPrepareUpdateStmt(t *Table) {
	var fieldNames []string
	for _, v := range t.ColumnList {
		if v == t.PrimaryColumn {
			continue
		}
		fieldNames = append(fieldNames, v.DbName+"=?")
	}

	g.Pn("func (dao *%sDao) prepareUpdateStmt() (err error){", t.GoName)
	g.Pn("    dao.updateStmt,err=dao.db.Prepare(context.Background(),\"UPDATE %s SET %s WHERE %s=?\")", t.DbName, strings.Join(fieldNames, ","), t.PrimaryColumn.DbName)
	g.Pn("    return err")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genPrepareDeleteStmt(t *Table) {
	g.Pn("func (dao *%sDao) prepareDeleteStmt() (err error){", t.GoName)
	g.Pn("    dao.deleteStmt,err=dao.db.Prepare(context.Background(),\"DELETE FROM %s WHERE %s=?\")", t.DbName, t.PrimaryColumn.DbName)
	g.Pn("    return err")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genDaoInitPrepareFunc(t *Table, funcName string) {
	g.Pn("    err=dao.%s()", funcName)
	g.Pn("    if err!=nil{")
	g.Pn("        return err")
	g.Pn("    }")
	g.Pn("    ")
}

func (g *Generator) genDaoInit(t *Table) {
	g.Pn("func (dao *%sDao) init() (err error){", t.GoName)

	g.genDaoInitPrepareFunc(t, "prepareInsertStmt")
	g.genDaoInitPrepareFunc(t, "prepareUpdateStmt")
	g.genDaoInitPrepareFunc(t, "prepareDeleteStmt")

	g.Pn("   return nil")
	g.Pn("}")
}

func (g *Generator) genDaoDef(t *Table) {
	g.Pn("type %sDao struct{", t.GoName)
	g.Pn("    logger *zap.Logger")
	g.Pn("    db *DB")
	g.Pn("    insertStmt *wrap.Stmt")
	g.Pn("    updateStmt *wrap.Stmt")
	g.Pn("    deleteStmt *wrap.Stmt")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genDaoNew(t *Table) {
	g.Pn("func New%sDao(db *DB)(t *%sDao,err error){", t.GoName, t.GoName)
	g.Pn("    t=&%sDao{}", t.GoName)
	g.Pn("    t.logger=log.TypedLogger(t)")
	g.Pn("    t.db=db")
	g.Pn("    err=t.init()")
	g.Pn("    if err!=nil{")
	g.Pn("        return nil,err")
	g.Pn("    }")
	g.Pn("    ")
	g.Pn("    return t,nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genInsert(t *Table) {
	var insertParams []string
	for _, v := range t.ColumnList {
		if v.AutoIncrement {
			continue
		}
		insertParams = append(insertParams, "e."+v.GoName)
	}

	g.Pn("func (dao *%sDao)Insert(ctx context.Context,tx *wrap.Tx,e *%s)(id int64,err error){", t.GoName, t.GoName)
	g.Pn("    stmt:=dao.insertStmt")
	g.Pn("    if tx!=nil{")
	g.Pn("        stmt=tx.Stmt(ctx,stmt)")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    result,err:=stmt.Exec(ctx,%s)", strings.Join(insertParams, ","))
	g.Pn("    if err!=nil{")
	g.Pn("        return 0,err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    id,err=result.LastInsertId()")
	g.Pn("    if err!=nil{")
	g.Pn("        return 0,err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return id,nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genUpdate(t *Table) {
	var updateParams []string
	for _, v := range t.ColumnList {
		if t.PrimaryColumn == v {
			continue
		}
		updateParams = append(updateParams, "e."+v.GoName)
	}

	g.Pn("func (dao *%sDao)Update(ctx context.Context,tx *wrap.Tx,e *%s)(err error){", t.GoName, t.GoName)
	g.Pn("    stmt:=dao.updateStmt")
	g.Pn("    if tx!=nil{")
	g.Pn("        stmt=tx.Stmt(ctx,stmt)")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    _,err=stmt.Exec(ctx,%s,e.%s)", strings.Join(updateParams, ","), t.PrimaryColumn.GoName)
	g.Pn("    if err!=nil{")
	g.Pn("        return err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genDelete(t *Table) {
	g.Pn("func (dao *%sDao)Delete(ctx context.Context,tx *wrap.Tx,%s %s)(err error){", t.GoName, t.PrimaryColumn.DbName, t.PrimaryColumn.GoType)
	g.Pn("    stmt:=dao.deleteStmt")
	g.Pn("    if tx!=nil{")
	g.Pn("        stmt=tx.Stmt(ctx,stmt)")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    _,err=stmt.Exec(ctx,%s)", t.PrimaryColumn.DbName)
	g.Pn("    if err!=nil{")
	g.Pn("        return err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genScanRow(t *Table) {
	var scanParams []string
	for _, v := range t.ColumnList {
		scanParams = append(scanParams, "&e."+v.GoName)
	}

	g.Pn("func (dao *%sDao)scanRow(row *wrap.Row)(*%s,error){", t.GoName, t.GoName)
	g.Pn("    e:=&%s{}", t.GoName)
	g.Pn("    err:=row.Scan(%s)", strings.Join(scanParams, ","))
	g.Pn("    if err!=nil{")
	g.Pn("        if err==wrap.ErrNoRows{")
	g.Pn("            return nil,nil")
	g.Pn("        }else{")
	g.Pn("            return nil,err")
	g.Pn("        }")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return e,nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genScanRows(t *Table) {
	var scanParams []string
	for _, v := range t.ColumnList {
		scanParams = append(scanParams, "&e."+v.GoName)
	}

	g.Pn("func (dao *%sDao)scanRows(rows *wrap.Rows)(list []*%s,err error){", t.GoName, t.GoName)
	g.Pn("    list=make([]*%s,0)", t.GoName)
	g.Pn("    for rows.Next(){")
	g.Pn("        e:=%s{}", t.GoName)
	g.Pn("        err=rows.Scan(%s)", strings.Join(scanParams, ","))
	g.Pn("        if err!=nil{")
	g.Pn("            return nil,err")
	g.Pn("        }")
	g.Pn("        list=append(list,&e)")
	g.Pn("    }")
	g.Pn("    if rows.Err()!=nil{")
	g.Pn("        err=rows.Err()")
	g.Pn("        return nil,err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return list,nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genSelect(t *Table) {
	g.Pn("func (dao *%sDao)doSelect(ctx context.Context,tx *wrap.Tx,query string)(*%s,error){", t.GoName, t.GoName)
	g.Pn("    querySql:=\"SELECT \"+%s+\" FROM %s \"+query", strings.ToUpper(t.DbName)+"_ALL_FIELDS_STRING", t.DbName)
	g.Pn("    var row *wrap.Row")
	g.Pn("    if tx==nil{")
	g.Pn("        row=dao.db.QueryRow(ctx,querySql)")
	g.Pn("    }else{")
	g.Pn("        row=tx.QueryRow(ctx,querySql)")
	g.Pn("    }")
	g.Pn("    return dao.scanRow(row)")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genSelectList(t *Table) {
	g.Pn("func (dao *%sDao)doSelectList(ctx context.Context,tx *wrap.Tx,query string)(list []*%s,err error){", t.GoName, t.GoName)
	g.Pn("    querySql:=\"SELECT \"+%s+\" FROM %s \"+query", strings.ToUpper(t.DbName)+"_ALL_FIELDS_STRING", t.DbName)
	g.Pn("    var rows *wrap.Rows")
	g.Pn("    if tx==nil{")
	g.Pn("        rows,err=dao.db.Query(ctx,querySql)")
	g.Pn("    }else {")
	g.Pn("        rows,err=tx.Query(ctx,querySql)")
	g.Pn("    }")
	g.Pn("    if err!=nil{")
	g.genDriverErrorLog()
	g.Pn("        return nil,err")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return dao.scanRows(rows)")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genQuery(t *Table) {
	g.Pn("type %sQuery struct{", t.GoName)
	g.Pn("    dao *%sDao", t.GoName)
	g.Pn("    forUpdate   bool")
	g.Pn("    forShare    bool")
	g.Pn("    whereBuffer *bytes.Buffer")
	g.Pn("    limitBuffer *bytes.Buffer")
	g.Pn("    orderBuffer *bytes.Buffer")
	g.Pn("}")
	g.Pn("")

	g.Pn("func New%sQuery(dao *%sDao)*%sQuery{", t.GoName, t.GoName, t.GoName)
	g.Pn("    q:=&%sQuery{}", t.GoName)
	g.Pn("    q.dao=dao")
	g.Pn("    q.whereBuffer=bytes.NewBufferString(\"\")")
	g.Pn("    q.limitBuffer=bytes.NewBufferString(\"\")")
	g.Pn("    q.orderBuffer=bytes.NewBufferString(\"\")")
	g.Pn("    ")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("")

	g.Pn("func (q *%sQuery) buildQueryString() string {", t.GoName)
	g.Pn("	buf := bytes.NewBufferString(\"\")")
	g.Pn("")
	g.Pn("	whereSql := q.whereBuffer.String()")
	g.Pn("	if whereSql != \"\" {")
	g.Pn("	buf.WriteString(\" WHERE \")")
	g.Pn("	buf.WriteString(whereSql)")
	g.Pn("}")
	g.Pn("")
	g.Pn("	orderSql := q.orderBuffer.String()")
	g.Pn("	if orderSql != \"\" {")
	g.Pn("	buf.WriteString(orderSql)")
	g.Pn("}")
	g.Pn("")
	g.Pn("	limitSql := q.limitBuffer.String()")
	g.Pn("	if limitSql != \"\" {")
	g.Pn("	buf.WriteString(limitSql)")
	g.Pn("}")
	g.Pn("")
	g.Pn("	if q.forUpdate {")
	g.Pn("	buf.WriteString(\" FOR UPDATE \")")
	g.Pn("}")
	g.Pn("")
	g.Pn("	if q.forShare {")
	g.Pn("	buf.WriteString(\" LOCK IN SHARE MODE \")")
	g.Pn("}")
	g.Pn("")
	g.Pn("	return buf.String()")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)Select(ctx context.Context)(*%s,error){", t.GoName, t.GoName)
	g.Pn("    return q.dao.doSelect(ctx,nil,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)SelectForUpdate(ctx context.Context,tx *wrap.Tx)(*%s,error){", t.GoName, t.GoName)
	g.Pn("    q.forUpdate=true")
	g.Pn("    return q.dao.doSelect(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)SelectForShare(ctx context.Context,tx *wrap.Tx)(*%s,error){", t.GoName, t.GoName)
	g.Pn("    q.forShare=true")
	g.Pn("    return q.dao.doSelect(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)SelectList(ctx context.Context)(list []*%s,err error){", t.GoName, t.GoName)
	g.Pn("    return q.dao.doSelectList(ctx,nil,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)SelectListForUpdate(ctx context.Context,tx *wrap.Tx)(list []*%s,err error){", t.GoName, t.GoName)
	g.Pn("    q.forUpdate=true")
	g.Pn("    return q.dao.doSelectList(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)SelectListForShare(ctx context.Context,tx *wrap.Tx)(list []*%s,err error){", t.GoName, t.GoName)
	g.Pn("    q.forShare=true")
	g.Pn("    return q.dao.doSelectList(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) Limit(startIncluded int64, count int64) *%sQuery {", t.GoName, t.GoName)
	g.Pn("	q.limitBuffer.WriteString(fmt.Sprintf(\" limit %%d,%%d\", startIncluded, count))")
	g.Pn("	return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) Sort(fieldName string, asc bool) *%sQuery {", t.GoName, t.GoName)
	g.Pn("	if asc {")
	g.Pn("	q.orderBuffer.WriteString(fmt.Sprintf(\" order by %%s asc\", fieldName))")
	g.Pn("} else {")
	g.Pn("	q.orderBuffer.WriteString(fmt.Sprintf(\" order by %%s desc\", fieldName))")
	g.Pn("}")
	g.Pn("")
	g.Pn("	return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func(q *%sQuery)where(format string,a ...interface{})*%sQuery{", t.GoName, t.GoName)
	g.Pn("    q.whereBuffer.WriteString(fmt.Sprintf(format,a...))")
	g.Pn("    return q")
	g.Pn("    }")
	g.Pn("")

	g.Pn("func(q *%sQuery)Left()*%sQuery{return q.where(\" ( \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)Right()*%sQuery{return q.where(\" ) \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)And()*%sQuery{return q.where(\" AND \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)Or()*%sQuery{return q.where(\" OR \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)Not()*%sQuery{return q.where(\" NOT \")}", t.GoName, t.GoName)
	g.Pn("")

	for _, c := range t.ColumnList {
		g.Pn("func (q *%sQuery)%s_Equal(v %s)*%sQuery{return q.where(\"%s='\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoType, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_NotEqual(v %s)*%sQuery{return q.where(\"%s<>'\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoType, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_Less(v %s)*%sQuery{return q.where(\"%s<'\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoType, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_LessEqual(v %s)*%sQuery{return q.where(\"%s<='\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoType, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_Greater(v %s)*%sQuery{return q.where(\"%s>'\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoType, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_GreaterEqual(v %s)*%sQuery{return q.where(\"%s>='\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoType, t.GoName, c.DbName)
	}
}

func (g *Generator) genDao(t *Table) {
	g.genConstants(t)

	g.genEntity(t)

	g.genQuery(t)

	g.genDaoDef(t)
	g.genDaoNew(t)
	g.genDaoInit(t)

	g.genPrepareInsertStmt(t)
	g.genPrepareUpdateStmt(t)
	g.genPrepareDeleteStmt(t)

	g.genInsert(t)
	g.genUpdate(t)
	g.genDelete(t)

	g.genScanRow(t)
	g.genScanRows(t)
	g.genSelect(t)
	g.genSelectList(t)

	g.Pn("func (dao *%sDao)GetQuery()*%sQuery{", t.GoName, t.GoName)
	g.Pn("    return New%sQuery(dao)", t.GoName)
	g.Pn("}")
	g.Pn("")
}

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
	g.Pn("func NewDB(connectionString string) (d *DB, err error) {")
	g.Pn("if connectionString == \"\" {")
	g.Pn("	return nil, fmt.Errorf(\"connectionString nil\")")
	g.Pn("}")
	g.Pn("")
	g.Pn("d = &DB{}")
	g.Pn("")
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

func (g *Generator) gen() {
	g.Pn("package %s", g.Namespace)
	g.Pn("")
	g.Pn("import(")
	g.Pn("    \"go.uber.org/zap\"")
	g.Pn("    \"github.com/NeuronFramework/log\"")
	g.Pn("    \"bytes\"")
	g.Pn("    \"fmt\"")
	g.Pn("    \"time\"")
	g.Pn("    \"context\"")
	g.Pn("    \"github.com/NeuronFramework/sql/wrap\"")
	g.Pn("    _ \"github.com/go-sql-driver/mysql\"")
	g.Pn(")")
	g.Pn("")
	for _, v := range g.TableList {
		g.genDao(v)
	}

	g.genDatabase()
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
