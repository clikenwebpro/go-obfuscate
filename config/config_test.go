package config

import (
	//	"reflect"

	"strings"
	"testing"
)

type getDumpFileNamePair struct {
	fileNameFormat   string
	databaseName     string
	expectedFileName string
}

var getDumpFileNameTestcases = []getDumpFileNamePair{
	{"%s-2006-01-02T150405", "black_mamba", "black_mamba"},
}

func TestGetDumpFileName(t *testing.T) {
	for _, testcase := range getDumpFileNameTestcases {
		conf := &Config{
			Output:   &OutputConfig{FileNameFormat: testcase.fileNameFormat},
			Database: &DatabaseConfig{DatabaseName: testcase.databaseName},
		}
		fileName := conf.GetDumpFileName()

		if strings.HasPrefix(fileName, testcase.expectedFileName) == false {
			t.Error("File name should start with DB name, got ", fileName)
		}
		if strings.HasSuffix(fileName, ".sql") == false {
			t.Error("File name should end with '.sql' suffix, got ", fileName)
		}
	}
}

type getMysqlConfigDSNPair struct {
	expectedDSN  string
	user         string
	password     string
	databaseName string
	hostname     string
	port         string
}

var getMysqlConfigDSNTestcases = []getMysqlConfigDSNPair{
	{"dbuser:dbpass@tcp(127.0.0.1:3306)/black_mamba", "dbuser", "dbpass", "black_mamba", "127.0.0.1", "3306"},
}

func TestGetMysqlConfigDSN(t *testing.T) {
	for _, testcase := range getMysqlConfigDSNTestcases {
		conf := &Config{
			Database: &DatabaseConfig{
				User:         testcase.user,
				Password:     testcase.password,
				DatabaseName: testcase.databaseName,
				Hostname:     testcase.hostname,
				Port:         testcase.port,
			},
		}

		mysqlDSN := conf.GetMysqlConfigDSN()
		if testcase.expectedDSN != mysqlDSN {
			t.Error("Expected ", testcase.expectedDSN, " got ", mysqlDSN)
		}
	}
}
