package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/vicdeo/go-obfuscate/config"
	"github.com/vicdeo/go-obfuscate/mysqldump"
)

var (
	conf *config.Config
)

func init() {
	var mysqlConfigPath string
	flag.StringVar(&mysqlConfigPath, "c", "./config.yaml", "MySQL connection details(./config.yaml)")
	flag.Parse()
	fmt.Println(mysqlConfigPath)
	conf = config.GetConf(mysqlConfigPath)
	// TODO: validate config
	os.MkdirAll(conf.Output.Directory, 0777)
}

func main() {
	// Open connection to database
	db, err := sql.Open("mysql", config.GetMysqlConfigDSN())
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

	// Register database with mysqldump
	dumper, err := mysqldump.Register(db, conf)
	if err != nil {
		fmt.Println("Error registering databse:", err)
		return
	}

	// Dump database to file
	var err2 = dumper.Dump()
	if err2 != nil {
		fmt.Println("Error dumping:", err)
		return
	}
	fmt.Printf("File is saved to %s\n", config.GetDumpFileName())

	// Close dumper, connected database and file stream.
	dumper.Close()
}
