package main

import (
	"flag"
	"fmt"
	"github.com/NeuronFramework/sql/generator2"
	"io/ioutil"
)

func main() {
	sqlFileFlag := flag.String("sql_file", "", "sql file")
	ormFileFlag := flag.String("orm_file", "", "orm file")
	packageNameFlag := flag.String("package_name", "", "package name")
	flag.Parse()

	sqlFile := *sqlFileFlag
	ormFile := *ormFileFlag
	packageName := *packageNameFlag

	if sqlFile == "" {
		fmt.Println("sql file null")
		return
	}

	if ormFile == "" {
		fmt.Println("orm file null")
		return
	}

	sqlData, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	gen := generator2.NewGenerator()
	orm, err := gen.Gen(string(sqlData), packageName)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(ormFile, []byte(orm), 0)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("OK")
}
