package generator

import (
	"errors"
	"strings"
)

func goName(name string) (goName string) {
	goName = ""
	is := true
	for i := 0; i < len(name); i++ {
		if is {
			if name[i] == '_' {
				is = true
			} else {
				is = false
				goName += strings.ToUpper(string(name[i]))
			}
		} else {
			if name[i] == '_' {
				is = true
			} else {
				is = false
				goName += string(name[i])
			}
		}
	}
	return goName
}

func goType(typ string, notNull bool, unsigned bool) (goType string, goTypeReal string) {
	if typ == "bigint" {
		if notNull {
			if unsigned {
				return "uint64", "uint64"
			} else {
				return "int64", "int64"
			}
		} else {
			if unsigned {
				return "sql.NullUint64", "uint64"
			} else {
				return "sql.NullInt64", "int64"
			}
		}
	} else if typ == "int" || typ == "tinyint" {
		if notNull {
			if unsigned {
				return "uint32", "uint32"
			} else {
				return "int32", "int32"
			}
		} else {
			if unsigned {
				return "sql.NullUint64", "uint32"
			} else {
				return "sql.NullInt64", "int32"
			}
		}
	} else if typ == "varchar" || typ == "longtext" || typ == "char" {
		if notNull {
			return "string", "string"
		} else {
			return "sql.NullString", "string"
		}
	} else if typ == "datetime" || typ == "timestamp" {
		if notNull {
			return "time.Time", "time.Time"
		} else {
			return "mysql.NullTime", "time.Time"
		}
	} else if typ == "double" {
		if notNull {
			return "float64", "float64"
		} else {
			return "sql.NullFloat64", "float64"
		}
	} else if typ == "float" {
		if notNull {
			return "float32", "float32"
		} else {
			return "sql.NullFloat64", "float32"
		}
	} else {
		return typ, typ
	}
}

func (g *Generator) parseColumnLine(line string) (c *Column, err error) {
	c = &Column{}
	tokens := strings.Split(line, " ")
	c.DbName = strings.Trim(tokens[0], "`")
	c.GoName = goName(c.DbName)
	c.DbType = tokens[1]

	if strings.HasSuffix(c.DbType, ")") {
		typTokens := strings.Split(c.DbType, "(")
		c.DbType = typTokens[0]
		sizeString := strings.TrimRight(typTokens[1], ")")
		c.Size = sizeString
	}

	c.NotNull = strings.Contains(line, "NOT NULL")
	c.Unsigned = strings.Contains(line, "unsigned")

	c.GoType, c.GoTypeReal = goType(c.DbType, c.NotNull, c.Unsigned)
	c.AutoIncrement = strings.Contains(line, "AUTO_INCREMENT")

	return c, nil
}

func (g *Generator) parsePrimaryKey(line string) (primaryKeyName string, err error) {
	primaryKeyName = line[strings.Index(line, "`")+1 : strings.LastIndex(line, "`")]
	return primaryKeyName, nil
}

func (g *Generator) parseTable(lines []string, i *int) (t *Table, err error) {
	t = newTable()

	l := strings.TrimSpace(lines[*i])
	tokens := strings.Split(l, "`")
	t.DbName = tokens[1]
	t.GoName = goName(t.DbName)

	for ; *i < len(lines); *i++ {
		l = strings.TrimSpace(lines[*i])
		if strings.HasPrefix(l, ")") {
			return t, nil
		}

		if strings.HasPrefix(l, "`") {
			c, err := g.parseColumnLine(l)
			if err != nil {
				return nil, err
			}
			t.AddColumn(c)
		} else if strings.HasPrefix(l, "PRIMARY KEY") {
			primaryKeyName, err := g.parsePrimaryKey(l)
			if err != nil {
				return nil, err
			}
			for _, v := range t.ColumnList {
				if v.DbName == primaryKeyName {
					t.PrimaryColumn = v
					break
				}
			}
			if t.PrimaryColumn == nil {
				return nil, errors.New("primary key not found")
			}
		} else if strings.HasPrefix(l, "UNIQUE KEY") {
			names := l[strings.Index(l, "(`")+2 : strings.LastIndex(l, "`")]
			if strings.Contains(names, ",") {
				unionIndex := &UnionIndex{}
				names = strings.Replace(names, "`", "", -1)
				unionIndex.ColumnNameList = strings.Split(names, ",")
				for _, cName := range unionIndex.ColumnNameList {
					for _, v := range t.ColumnList {
						if v.DbName == cName {
							unionIndex.ColumnList = append(unionIndex.ColumnList, v)
							break
						}
					}
				}
				t.UniqueUnionIndexList = append(t.UniqueUnionIndexList, unionIndex)
			} else {
				index := &Index{}
				index.Name = names
				for _, v := range t.ColumnList {
					if v.DbName == index.Name {
						index.Column = v
						break
					}
				}
				t.UniqueIndexList = append(t.UniqueIndexList, index)
			}
		} else if strings.HasPrefix(l, "KEY") {
			names := l[strings.Index(l, "(`")+2 : strings.LastIndex(l, "`)")]
			if strings.Contains(names, ",") {
				unionIndex := &UnionIndex{}
				names = strings.Replace(names, "`", "", -1)
				unionIndex.ColumnNameList = strings.Split(names, ",")
				for _, cName := range unionIndex.ColumnNameList {
					for _, v := range t.ColumnList {
						if v.DbName == cName {
							unionIndex.ColumnList = append(unionIndex.ColumnList, v)
							break
						}
					}
				}
				t.UnionIndexList = append(t.UnionIndexList, unionIndex)
			} else {
				index := &Index{}
				index.Name = names
				for _, v := range t.ColumnList {
					if v.DbName == index.Name {
						index.Column = v
						break
					}
				}
				t.IndexList = append(t.IndexList, index)
			}
		} else {
			//nothing
		}
	}

	return nil, errors.New("table no end")
}

func (g *Generator) parse(sql string) error {
	lines := strings.Split(sql, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Database: ") {
			tokens := strings.Split(lines[i], " ")
			g.DbName = tokens[len(tokens)-1]
		} else if strings.HasPrefix(lines[i], "CREATE TABLE ") {
			table, err := g.parseTable(lines, &i)
			if err != nil {
				return err
			}
			g.TableList = append(g.TableList, table)
		}
	}
	return nil
}
