package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vicdeo/go-obfuscate/config"
	"github.com/vicdeo/go-obfuscate/mysqldump"
)

const (
	version                    = "0.9.0"
	errConfigFileNotFound      = 1
	errConfigFileInvalidMarkUp = 2
	errOutputDirectoryMissing  = 3
	errDBConnectionFailed      = 4
	errDumpFileIsNotWritable   = 5
)

var (
	conf *config.Config
)

func init() {
	loadConfig()
	prepareFS()
}

func main() {
	// Open connection to database
	db, err := sql.Open("mysql", conf.Database.GetMysqlConfigDSN())
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Please validate DB credentials.\n", err)
		os.Exit(errDBConnectionFailed)
	}

	// Register database with mysqldump
	dumper, err := mysqldump.Register(db, conf)
	if err != nil {
		fmt.Println("Error registering database:", err)
		os.Exit(errDumpFileIsNotWritable)
	}

	// TODO: add to config support of	dumper.LockTables = true
	// Dump database to file
	var err2 = dumper.Dump()
	if err2 != nil {
		fmt.Println("Error dumping:", err)
		return
	}
	fmt.Printf("File is saved to %s\n", conf.GetDumpFileName())

	// Close dumper, connected database and file stream.
	dumper.Close()
}

func loadConfig() {
	var configFilePath string
	var error error

	fmt.Println("go-obfuscate version", version)
	flag.StringVar(&configFilePath, "c", "./config.yaml", "MySQL connection details(./config.yaml)")
	flag.Parse()
	evaledPath, _ := filepath.EvalSymlinks(configFilePath)
	if _, err := os.Stat(evaledPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Config file does not exist:", configFilePath, "")
		os.Exit(errConfigFileNotFound)
	}

	fmt.Println("Using config file:", evaledPath)
	if conf, error = config.GetConf(filepath.Dir(evaledPath), filepath.Base(evaledPath)); error != nil {
		fmt.Println("Config file contains invalid YAML markup:")
		fmt.Printf("%v\n", error)
		os.Exit(errConfigFileInvalidMarkUp)
	}

	// TODO: validate config here after loading it
}

func prepareFS() {
	// Dump dir exists
	os.MkdirAll(conf.Output.Directory, 0777)
	if !isDir(conf.Output.Directory) {
		fmt.Println("Could not create directory ", conf.Output.Directory)
		os.Exit(errOutputDirectoryMissing)
	}

	// Dump file does not exist
	p := conf.GetDumpFullPath()
	if e, _ := exists(p); e {
		// TODO: recoverable error - just add an increasing posfix or whatever
		fmt.Println("Dump '" + p + "' already exists.")
	}
}

func isDir(p string) bool {
	if e, fi := exists(p); e {
		return fi.Mode().IsDir()
	}
	return false
}

func exists(p string) (bool, os.FileInfo) {
	f, err := os.Open(p)
	if err != nil {
		return false, nil
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return false, nil
	}
	return true, fi
}
