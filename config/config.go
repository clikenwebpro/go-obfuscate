package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/vicdeo/go-obfuscate/faker"
)

// DatabaseConfig -- Database connection config
type DatabaseConfig struct {
	Hostname     string `yaml:"hostname"`
	Port         string `yaml:"port"`
	DatabaseName string `yaml:"databaseName"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
}

// OutputConfig -- dump-specific options
type OutputConfig struct {
	FileNameFormat string `yaml:"fileNameFormat"`
	Directory      string `yaml:"directory"`
}

// Config - global config
type Config struct {
	Database *DatabaseConfig        `yaml:"database"`
	Output   *OutputConfig          `yaml:"output"`
	Tables   map[string]interface{} `yaml:"tables"`
}

const (
	ignoreMarker   = "ignore"
	truncateMarker = "truncate"
)

// Create a new Config instance.
var (
	conf         *Config
	dumpFileName string
)

// IsIgnoredTable
func IsIgnoredTable(tableName string) bool {
	if val, ok := conf.Tables[tableName]; ok {
		return val == ignoreMarker
	}
	return false
}

// ShouldDumpData
func ShouldDumpData(tableName string) bool {
	if val, ok := conf.Tables[tableName]; ok {
		return val != truncateMarker && !IsIgnoredTable(tableName)
	}
	return false
}

// GetColumnFaker - get a proper data generator
func GetColumnFaker(tableName, columnName string) faker.FakeGenerator {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	if table, ok := conf.Tables[tableName]; ok {
		tableMap := table.(map[string]interface{})
		if column, ok := tableMap[columnName]; ok {
			columnMap := column.(map[string]interface{})
			return faker.New(columnMap)
		}
	}
	return nil
}

// GetConf - Read the config file and marshal into the conf Config struct.
func GetConf(configPath string) *Config {
	evaledPath, _ := filepath.EvalSymlinks(configPath)
	viper.AddConfigPath(filepath.Dir(evaledPath))
	viper.SetConfigName(filepath.Base(evaledPath))
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("%v", err)
	}

	conf = &Config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("unable to decode into Config struct, %v", err)
	}

	return conf
}

// GetDumpFileName -
func GetDumpFileName() string {
	if dumpFileName == "" {
		// Uses time.Time.Format (https://golang.org/pkg/time/#Time.Format). format appended with '.sql'.
		dumpFileName = time.Now().Format(conf.Output.FileNameFormat)
		dumpFileName = fmt.Sprintf(dumpFileName, conf.Database.DatabaseName) + ".sql"
	}
	return dumpFileName
}

func GetMysqlConfigDSN() string {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = conf.Database.User
	mysqlConfig.Passwd = conf.Database.Password
	mysqlConfig.DBName = conf.Database.DatabaseName
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = conf.Database.Hostname + ":" + conf.Database.Port
	return mysqlConfig.FormatDSN()
}
