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
	errShowTablesFailed        = 5
	errDumpFileIsNotWritable   = 6
	errConfigHasDuplicates     = 7
	errConfigIncomplete        = 8
	errConfigHasUnknownType    = 9

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

	dbValidationTemplate = `Checking for missing/extra tables...done
{{if .missedInDb}}Tables that are found in config but not in DB:
{{range $v := .missedInDb}} - {{.}}
{{end}}{{end -}}
{{if .missedInConfig}}Tables that are found in DB but not in config:
{{range .missedInConfig}} - {{.}}
{{end}}{{end}}
`

	fakerValidationTemplate = `Checking obfuscated columns type...done
{{range $v := .}} - Column {{index $v 1}} in the table {{index $v 0}} has unknown or missing type
{{end}}
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
	exitOnError(err != nil, errDBConnectionFailed, fmt.Sprintf("Please validate DB credentials.\n%v", err))

	allConfigTables := conf.GetAllUniqueTableNames()
	allDbTables, err := mysqldump.ShowTables(db)
	exitOnError(err != nil, errShowTablesFailed, fmt.Sprintf("Error getting database table list: %v", err))

	diff, diff2 := difference(allConfigTables, allDbTables)
	dbValTmpl, err := template.New("dbValidation").Parse(dbValidationTemplate)
	if err == nil {
		dbValTmpl.Execute(os.Stdout, map[string][]string{
			"missedInDb":     diff,
			"missedInConfig": diff2,
		})
	}
	exitOnError(len(diff) > 0 || len(diff2) > 0, errConfigIncomplete, "Please fix the reported errors in your config file before proceeding")

	// Register database with mysqldump
	dumper, err := mysqldump.Register(db, conf)
	exitOnError(err != nil, errDumpFileIsNotWritable, fmt.Sprintf("Error registering database: %v", err))

	// TODO: add to config support of	dumper.LockTables = true
	err = dumper.Dump()
	if err != nil {
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
	_, err := os.Stat(evaledPath)
	exitOnError(errors.Is(err, os.ErrNotExist), errConfigFileNotFound, fmt.Sprintf("Config file does not exist: %s", configFilePath))

	fmt.Println("Using config file:", evaledPath)
	conf, error = config.GetConf(filepath.Dir(evaledPath), filepath.Base(evaledPath))
	exitOnError(error != nil, errConfigFileInvalidMarkUp, fmt.Sprintf("Config file contains invalid YAML markup:\n%v\n", error))

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
	messages, hasErrors := conf.ValidateConfig()
	valTmpl, err := template.New("validation").Parse(validationTemplate)
	if err == nil {
		valTmpl.Execute(os.Stdout, messages)
	}
	exitOnError(hasErrors, errConfigHasDuplicates, "Please fix the reported errors in your config file before proceeding")

	// Sanity check 2: each obfuscated column type should be known
	unknown, hasErrors := conf.ValidateObfuscateSection()
	fakerTmpl, err := template.New("faker").Parse(fakerValidationTemplate)
	if err == nil {
		fakerTmpl.Execute(os.Stdout, unknown)
	}
	exitOnError(hasErrors, errConfigHasDuplicates, "Please fix the reported errors in your config file before proceeding")
}

func prepareFS() {
	// Dump dir exists
	os.MkdirAll(conf.Output.Directory, 0777)
	exitOnError(!isDir(conf.Output.Directory), errOutputDirectoryMissing, fmt.Sprintf("Could not create directory %s\n", conf.Output.Directory))

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

func difference(slice1 []string, slice2 []string) ([]string, []string) {
	diff := make(map[int][]string, 0)

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff[i] = append(diff[i], s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff[0], diff[1]
}

func exitOnError(hasErrors bool, exitCode int, message string) {
	if hasErrors {
		fmt.Println(message)
		os.Exit(exitCode)
	}
}
