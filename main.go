package main

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/vicdeo/go-obfuscate/config"
	"github.com/vicdeo/go-obfuscate/mysqldump"

	"github.com/go-sql-driver/mysql"
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
}

func main() {
	// Open connection to database
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = conf.Database.User
	mysqlConfig.Passwd = conf.Database.Password
	mysqlConfig.DBName = conf.Database.DatabaseName
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = conf.Database.Hostname + ":" + conf.Database.Port

	dumpDir := conf.Output.Directory // TODO: create if not exists
	dumpFilenameFormat := fmt.Sprintf(conf.Output.FileNameFormat, mysqlConfig.DBName)

	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

	// Register database with mysqldump
	dumper, err := mysqldump.Register(db, conf, dumpDir, dumpFilenameFormat)
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
	fmt.Printf("File is saved to %s\n", dumpFilenameFormat)

	// Close dumper, connected database and file stream.
	dumper.Close()
}
