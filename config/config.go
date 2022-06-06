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

type MainConfig interface {
	GetDumpFileName() string
	GetMysqlConfigDSN() string
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
	return true
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

func (config *Config) GetDumpFileName() string {
	if dumpFileName == "" {
		// Uses time.Time.Format (https://golang.org/pkg/time/#Time.Format). format appended with '.sql'.
		dumpFileName = time.Now().Format(config.Output.FileNameFormat)
		dumpFileName = fmt.Sprintf(dumpFileName, config.Database.DatabaseName) + ".sql"
	}
	return dumpFileName
}

func (config *Config) GetMysqlConfigDSN() string {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = config.Database.User
	mysqlConfig.Passwd = config.Database.Password
	mysqlConfig.DBName = config.Database.DatabaseName
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = config.Database.Hostname + ":" + config.Database.Port
	return mysqlConfig.FormatDSN()
}
