package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
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

func (g *Generator) genErrorLog(msg string) {
	g.Pn("        dao.logger.Error(\"%s\",zap.Error(err))", msg)
}

func (g *Generator) genDriverErrorLog() {
	g.genErrorLog("sqlDriver")
}

func (g *Generator) genConstants(t *Table) {
	g.Pn("const %s_TABLE_NAME = \"%s\"", strings.ToUpper(t.DbName), t.DbName)
	g.Pn("")

	g.Pn("type %s_FIELD string", strings.ToUpper(t.DbName))
	var filedNameList []string
	for _, v := range t.ColumnList {
		g.Pn("const %s_FIELD_%s = %s_FIELD(\"%s\")", strings.ToUpper(t.DbName), strings.ToUpper(v.DbName), strings.ToUpper(t.DbName), v.DbName)
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

		if v.DbName == "create_time" || v.DbName == "update_time" {
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
	var where = fmt.Sprintf("WHERE %s=?", t.PrimaryColumn.DbName)

	var fieldNames []string
	for _, v := range t.ColumnList {
		if v == t.PrimaryColumn {
			continue
		}

		if v.DbName == "create_time" || v.DbName == "update_time" {
			continue
		}

		if v.DbName == "update_version" {
			fieldNames = append(fieldNames, "update_version=update_version+1")
			where = where + " AND update_version=?"
			continue
		}

		fieldNames = append(fieldNames, v.DbName+"=?")
	}

	g.Pn("func (dao *%sDao) prepareUpdateStmt() (err error){", t.GoName)
	g.Pn("    dao.updateStmt,err=dao.db.Prepare(context.Background(),\"UPDATE %s SET %s %s\")",
		t.DbName,
		strings.Join(fieldNames, ","),
		where)
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
	g.Pn("")
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

		if v.DbName == "create_time" || v.DbName == "update_time" {
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
	var updateVersionColume *Column
	var updateParams []string
	for _, v := range t.ColumnList {
		if t.PrimaryColumn == v {
			continue
		}

		if v.DbName == "create_time" || v.DbName == "update_time" {
			continue
		}

		if v.DbName == "update_version" {
			updateVersionColume = v
			continue
		}

		updateParams = append(updateParams, "e."+v.GoName+",")
	}

	g.Pn("func (dao *%sDao)Update(ctx context.Context,tx *wrap.Tx,e *%s)(err error){", t.GoName, t.GoName)
	g.Pn("    stmt:=dao.updateStmt")
	g.Pn("    if tx!=nil{")
	g.Pn("        stmt=tx.Stmt(ctx,stmt)")
	g.Pn("    }")
	g.Pn("")
	if updateVersionColume != nil {
		g.Pn("    _,err=stmt.Exec(ctx,%se.%s,e.%s)", strings.Join(updateParams, ""), t.PrimaryColumn.GoName, updateVersionColume.GoName)
	} else {
		g.Pn("    _,err=stmt.Exec(ctx,%se.%s)", strings.Join(updateParams, ""), t.PrimaryColumn.GoName)
	}
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
	g.Pn("func (dao *%sDao)QueryOne(ctx context.Context,tx *wrap.Tx,query string)(*%s,error){", t.GoName, t.GoName)
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
	g.Pn("func (dao *%sDao)QueryList(ctx context.Context,tx *wrap.Tx,query string)(list []*%s,err error){", t.GoName, t.GoName)
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

func (g *Generator) genSelectCount(t *Table) {
	g.Pn("func (dao *%sDao) QueryCount(ctx context.Context, tx *wrap.Tx, query string) (count int64, err error) {", t.GoName)
	g.Pn("querySql := \"SELECT COUNT(1) FROM %s \" + query", t.DbName)
	g.Pn("var row *wrap.Row")
	g.Pn("if tx == nil {")
	g.Pn("    row = dao.db.QueryRow(ctx, querySql)")
	g.Pn("} else {")
	g.Pn("    row = tx.QueryRow(ctx, querySql)")
	g.Pn("}")
	g.Pn("if err != nil {")
	g.Pn("    dao.logger.Error(\"sqlDriver\", zap.Error(err))")
	g.Pn("    return 0, err")
	g.Pn("}")
	g.Pn("")
	g.Pn("err = row.Scan(&count)")
	g.Pn("if err != nil {")
	g.Pn("    return 0, err")
	g.Pn("}")
	g.Pn("")
	g.Pn("    return count, nil")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genSelectGroupBy(t *Table) {
	g.Pn("func (dao *%sDao)QueryGroupBy(ctx context.Context,tx *wrap.Tx,groupByFields []string,query string)(rows *wrap.Rows,err error){", t.GoName)
	g.Pn("    querySql:=\"SELECT \"+strings.Join(groupByFields,\",\")+\",count(1) FROM %s \"+query", t.DbName)
	g.Pn("    if tx==nil{")
	g.Pn("        return dao.db.Query(ctx,querySql)")
	g.Pn("    }else {")
	g.Pn("        return tx.Query(ctx,querySql)")
	g.Pn("    }")
	g.Pn("}")
	g.Pn("")
}

func (g *Generator) genQuery(t *Table) {
	g.Pn("type %sQuery struct{", t.GoName)
	g.Pn("    BaseQuery")
	g.Pn("    dao *%sDao", t.GoName)
	g.Pn("}")
	g.Pn("")

	g.Pn("func New%sQuery(dao *%sDao)*%sQuery{", t.GoName, t.GoName, t.GoName)
	g.Pn("    q:=&%sQuery{}", t.GoName)
	g.Pn("    q.dao=dao")
	g.Pn("    ")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)QueryOne(ctx context.Context,tx *wrap.Tx)(*%s,error){", t.GoName, t.GoName)
	g.Pn("    return q.dao.QueryOne(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)QueryList(ctx context.Context,tx *wrap.Tx)(list []*%s,err error){", t.GoName, t.GoName)
	g.Pn("    return q.dao.QueryList(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery)QueryCount(ctx context.Context,tx *wrap.Tx)(count int64,err error){", t.GoName)
	g.Pn("    return q.dao.QueryCount(ctx,tx,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q* %sQuery)QueryGroupBy(ctx context.Context,tx *wrap.Tx)(rows *wrap.Rows,err error){", t.GoName)
	g.Pn("    return q.dao.QueryGroupBy(ctx,tx,q.groupByFields,q.buildQueryString())")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) ForUpdate()*%sQuery{", t.GoName, t.GoName)
	g.Pn("    q.forUpdate=true")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) ForShare()*%sQuery{", t.GoName, t.GoName)
	g.Pn("    q.forShare=true")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) GroupBy(fields ...%s_FIELD) *%sQuery {", t.GoName, strings.ToUpper(t.DbName), t.GoName)
	g.Pn("    q.groupByFields=make([]string,len(fields))")
	g.Pn("    for i,v:=range fields{")
	g.Pn("        q.groupByFields[i]=string(v)")
	g.Pn("    }")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) Limit(startIncluded int64, count int64) *%sQuery {", t.GoName, t.GoName)
	g.Pn("	q.limit=fmt.Sprintf(\" limit %%d,%%d\", startIncluded, count)")
	g.Pn("	return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) OrderBy(fieldName %s_FIELD, asc bool) *%sQuery {", t.GoName, strings.ToUpper(t.DbName), t.GoName)
	g.Pn("    if q.order!=\"\"{")
	g.Pn("        q.order+=\",\"")
	g.Pn("    }")
	g.Pn("    q.order+=string(fieldName)+\" \"")
	g.Pn("    if asc {")
	g.Pn("        q.order+=\"asc\"")
	g.Pn("    } else {")
	g.Pn("        q.order+=\"desc\"")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func (q *%sQuery) OrderByGroupCount(asc bool) *%sQuery {",
		t.GoName, t.GoName)
	g.Pn("    if q.order!=\"\"{")
	g.Pn("        q.order+=\",\"")
	g.Pn("    }")
	g.Pn("    q.order+=\"count(1) \"")
	g.Pn("    if asc {")
	g.Pn("        q.order+=\"asc\"")
	g.Pn("    } else {")
	g.Pn("        q.order+=\"desc\"")
	g.Pn("    }")
	g.Pn("")
	g.Pn("    return q")
	g.Pn("}")
	g.Pn("")

	g.Pn("func(q *%sQuery)w(format string,a ...interface{})*%sQuery{", t.GoName, t.GoName)
	g.Pn("    q.where+=fmt.Sprintf(format,a...)")
	g.Pn("    return q")
	g.Pn("    }")
	g.Pn("")

	g.Pn("func(q *%sQuery)Left()*%sQuery{return q.w(\" ( \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)Right()*%sQuery{return q.w(\" ) \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)And()*%sQuery{return q.w(\" AND \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)Or()*%sQuery{return q.w(\" OR \")}", t.GoName, t.GoName)
	g.Pn("func(q *%sQuery)Not()*%sQuery{return q.w(\" NOT \")}", t.GoName, t.GoName)
	g.Pn("")

	for _, c := range t.ColumnList {
		g.Pn("func (q *%sQuery)%s_Equal(v %s)*%sQuery{return q.w(\"%s='\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoTypeReal, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_NotEqual(v %s)*%sQuery{return q.w(\"%s<>'\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoTypeReal, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_Less(v %s)*%sQuery{return q.w(\"%s<'\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoTypeReal, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_LessEqual(v %s)*%sQuery{return q.w(\"%s<='\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoTypeReal, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_Greater(v %s)*%sQuery{return q.w(\"%s>'\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoTypeReal, t.GoName, c.DbName)
		g.Pn("func (q *%sQuery)%s_GreaterEqual(v %s)*%sQuery{return q.w(\"%s>='\"+fmt.Sprint(v)+\"'\")}", t.GoName, c.GoName, c.GoTypeReal, t.GoName, c.DbName)
		if !c.NotNull {
			g.Pn("func (q *%sQuery)%s_IsNull()*%sQuery{return q.w(\"%s IS NULL\")}", t.GoName, c.GoName, t.GoName, c.DbName)
			g.Pn("func (q *%sQuery)%s_NotNull()*%sQuery{return q.w(\"%s IS NOT NULL\")}", t.GoName, c.GoName, t.GoName, c.DbName)
		}
	}
	g.Pn("")
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
	g.genSelectCount(t)
	g.genSelectGroupBy(t)

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

func (g *Generator) genBaseQuery() {
	g.Pn("type BaseQuery struct {")
	g.Pn("	forUpdate     bool")
	g.Pn("	forShare      bool")
	g.Pn("	where         string")
	g.Pn("	limit         string")
	g.Pn("	order         string")
	g.Pn("	groupByFields []string")
	g.Pn("}")
	g.Pn("")
	g.Pn("func (q *BaseQuery) buildQueryString() string {")
	g.Pn("	buf := bytes.NewBufferString(\"\")")
	g.Pn("")
	g.Pn("	if q.where != \"\" {")
	g.Pn("	buf.WriteString(\" WHERE \")")
	g.Pn("	buf.WriteString(q.where)")
	g.Pn("}")
	g.Pn("")
	g.Pn("	if q.groupByFields != nil && len(q.groupByFields) > 0 {")
	g.Pn("	buf.WriteString(\" GROUP BY \")")
	g.Pn("	buf.WriteString(strings.Join(q.groupByFields, \",\"))")
	g.Pn("}")
	g.Pn("")
	g.Pn("	if q.order != \"\" {")
	g.Pn("    buf.WriteString(\" order by \")")
	g.Pn("    buf.WriteString(q.order)")
	g.Pn("}")
	g.Pn("")
	g.Pn("	if q.limit != \"\" {")
	g.Pn("	buf.WriteString(q.limit)")
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
}

func (g *Generator) gen() {
	g.Pn("package %s", g.Namespace)
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

	g.genBaseQuery()

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
