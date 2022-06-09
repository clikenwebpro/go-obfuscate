package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
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
	errConfigHasDuplicates     = 6

	statsTemplate = `Config parsed. Found tables count:
 - to dump as is: {{.keep}}
 - to ignore: {{.ignore}}
 - to truncate: {{.truncate}}
 - to obfuscate: {{.obfuscate}}
Total: {{.total}}
`

	validationTemplate = `Checking for duplicated table names...done
{{range $k, $v := .}}{{if $v}} - {{range $v}}{{.}}{{end}} spotted multiple times in the {{$k}} section
{{end}}{{end}}
`
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

	statsTmpl, err := template.New("statistics").Parse(statsTemplate)
	if err == nil {
		statsTmpl.Execute(os.Stdout, map[string]int{
			"keep":      len(conf.Tables.Keep),
			"ignore":    len(conf.Tables.Ignore),
			"truncate":  len(conf.Tables.Truncate),
			"obfuscate": len(conf.Tables.Obfuscate),
			"total":     len(conf.Tables.Keep) + len(conf.Tables.Ignore) + len(conf.Tables.Truncate) + len(conf.Tables.Obfuscate),
		})
	}
	// Sanity check 1: each table name should be unique across all lists
	messages, hasErrors := config.ValidateConfig()
	valTmpl, err := template.New("validation").Parse(validationTemplate)
	if err == nil {
		valTmpl.Execute(os.Stdout, messages)
	}
	if hasErrors {
		fmt.Println("Please fix the reported errors before proceeding")
		os.Exit(errConfigHasDuplicates)
	}
	// TODO: revalidate config after DB connection is done to make sure it has the same tables as DB and show missing/extra tables if any
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
		// TODO: recoverable error - just add an increasing postfix or whatever
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
