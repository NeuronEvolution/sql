package generator

type Column struct {
	DbName        string
	GoName        string
	DbType        string
	GoType        string //may be NULL
	GoTypeReal    string
	Size          string
	AutoIncrement bool
	NotNull       bool
	Unsigned      bool
}

type Index struct {
	Name   string
	Column *Column
}

type UnionIndex struct {
	Name           string
	ColumnNameList []string
	ColumnList     []*Column
}

type Table struct {
	DbName               string
	GoName               string
	ColumnList           []*Column
	PrimaryColumn        *Column
	CreateTimeColumn     *Column
	UpdateTimeColumn     *Column
	UpdateVersionColumn  *Column
	IndexList            []*Index
	UniqueIndexList      []*Index
	UnionIndexList       []*UnionIndex
	UniqueUnionIndexList []*UnionIndex
}

func newTable() (t *Table) {
	t = &Table{}
	t.ColumnList = make([]*Column, 0)

	return t
}

func (t *Table) AddColumn(c *Column) {
	t.ColumnList = append(t.ColumnList, c)

	if c.DbName == "create_time" {
		t.CreateTimeColumn = c
	}

	if c.DbName == "update_time" {
		t.UpdateTimeColumn = c
	}

	if c.DbName == "update_version" {
		t.UpdateVersionColumn = c
	}
}
