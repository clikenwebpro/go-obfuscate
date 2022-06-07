package config

import (
	//	"reflect"

	"testing"
	"time"
)

type getDumpFileNamePair struct {
	expectedFileName string
	config           Config
}

var getDumpFileNameTestcases = []getDumpFileNamePair{
	{
		"black_mamba-2022-06-01T010203.sql",
		Config{
			Output:   &OutputConfig{FileNameFormat: "%s-2006-01-02T150405"},
			Database: &DatabaseConfig{DatabaseName: "black_mamba"},
		},
	},
}

func TestGetDumpFileName(t *testing.T) {
	for _, testcase := range getDumpFileNameTestcases {

		testcase.config.clock = func() time.Time { return time.Date(2022, 06, 01, 01, 02, 03, 0, time.UTC) }
		fileName := testcase.config.GetDumpFileName()
		if testcase.expectedFileName != fileName {
			t.Error("Expected name is", testcase.expectedFileName, ", got", fileName)
		}
	}
}

type getMysqlConfigDSNPair struct {
	expectedDSN string
	config      DatabaseConfig
}

var getMysqlConfigDSNTestcases = []getMysqlConfigDSNPair{
	{
		"dbuser:dbpass@tcp(127.0.0.1:3306)/black_mamba",
		DatabaseConfig{Net: "tcp", DatabaseName: "black_mamba", User: "dbuser", Password: "dbpass", Hostname: "127.0.0.1", Port: "3306"},
	},
	{
		"unix(/tmp/mysql.sock)/black_mamba",
		DatabaseConfig{Net: "unix", DatabaseName: "black_mamba", Socket: "/tmp/mysql.sock"},
	},
	{
		"dbuser:dbpass@unix(/tmp/mysql.sock)/black_mamba",
		DatabaseConfig{Net: "unix", DatabaseName: "black_mamba", Socket: "/tmp/mysql.sock", User: "dbuser", Password: "dbpass"},
	},
}

func TestGetMysqlConfigDSN(t *testing.T) {
	for _, testcase := range getMysqlConfigDSNTestcases {
		mysqlDSN := testcase.config.GetMysqlConfigDSN()
		if testcase.expectedDSN != mysqlDSN {
			t.Error("Expected ", testcase.expectedDSN, " got ", mysqlDSN)
		}
	}
}
