package config

import (
	"fmt"
	"net"
	"path"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/vicdeo/go-obfuscate/faker"
)

type (
	// DatabaseConfig -- Database connection config
	DatabaseConfig struct {
		Net          string `yaml:"net,omitempty"`
		Socket       string `yaml:"socket,omitempty"`
		Hostname     string `yaml:"hostname,omitempty"`
		Port         string `yaml:"port,omitempty"`
		DatabaseName string `yaml:"databaseName,omitempty"`
		User         string `yaml:"user,omitempty"`
		Password     string `yaml:"password,omitempty"`
	}

	// OutputConfig -- dump-specific options
	OutputConfig struct {
		FileNameFormat string `yaml:"fileNameFormat"`
		Directory      string `yaml:"directory"`
	}

	TableConfig struct {
		Keep      []string               `yaml:"kept"`
		Ignore    []string               `yaml:"kept"`
		Truncate  []string               `yaml:"kept"`
		Obfuscate map[string]interface{} `yaml:"tables"`
	}

	// Config - global config
	Config struct {
		Database *DatabaseConfig `yaml:"database"`
		Output   *OutputConfig   `yaml:"output"`
		Tables   *TableConfig    `yaml:"tables"`
		clock    func() time.Time
	}
)

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
func GetConf(configDir, configFileName string) (*Config, error) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	viper.AddConfigPath(configDir)
	viper.SetConfigName(configFileName)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	conf = &Config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// IsIgnoredTable
func IsIgnoredTable(tableName string) bool {
	return contains(conf.Tables.Ignore, tableName)
}

// ShouldDumpData
func ShouldDumpData(tableName string) bool {
	if IsIgnoredTable(tableName) {
		return false
	}
	return !contains(conf.Tables.Truncate, tableName)
}

// GetColumnFaker - get a proper data generator
func GetColumnFaker(tableName, columnName string) faker.FakeGenerator {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	if table, ok := conf.Tables.Obfuscate[tableName]; ok {
		tableMap := table.(map[string]interface{})
		if column, ok := tableMap[columnName]; ok {
			columnMap := column.(map[string]interface{})
			return faker.New(columnMap)
		}
	}
	return nil
}

func (config *Config) ValidateConfig() (map[string][]string, bool) {
	hasErrors := false
	messages := make(map[string][]string, 0)

	tablesToObfuscate := make([]string, 0)
	for t, _ := range config.Tables.Obfuscate {
		tablesToObfuscate = append(tablesToObfuscate, t)
	}

	allTables := config.GetAllUniqueTableNames()
	for sectionTitle, sectionValues := range map[string][]string{
		"obfuscate": tablesToObfuscate,
		"keep":      config.Tables.Keep,
		"ignore":    config.Tables.Ignore,
		"truncate":  config.Tables.Truncate,
		"overall":   allTables,
	} {
		dupesInSection := duplicatesInSlice(sectionValues)
		messages[sectionTitle] = dupesInSection
		if len(dupesInSection) > 0 {
			hasErrors = true
		}
	}

	return messages, hasErrors
}

func (config *Config) GetDumpFullPath() string {
	return path.Join(config.Output.Directory, config.GetDumpFileName())
}

func (config *Config) GetDumpFileName() string {
	if dumpFileName == "" {
		// Uses time.Time.Format (https://golang.org/pkg/time/#Time.Format). format appended with '.sql'.
		dumpFileName = config.now().Format(config.Output.FileNameFormat)
		dumpFileName = fmt.Sprintf(dumpFileName, config.Database.DatabaseName) + ".sql"
	}
	return dumpFileName
}

func (config *DatabaseConfig) GetMysqlConfigDSN() string {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.DBName = config.DatabaseName
	mysqlConfig.Net = config.Net
	mysqlConfig.User = config.User
	mysqlConfig.Passwd = config.Password

	switch config.Net {
	case "tcp":
		mysqlConfig.Addr = net.JoinHostPort(config.Hostname, config.Port)
	case "unix":
		mysqlConfig.Addr = config.Socket
	}
	return mysqlConfig.FormatDSN()
}

func (config *Config) GetAllUniqueTableNames() []string {
	allTables := make([]string, 0)
	allTables = append(unique(config.getObfuscatedTableNames()), unique(config.Tables.Keep)...)
	allTables = append(allTables, unique(config.Tables.Ignore)...)
	allTables = append(allTables, unique(config.Tables.Truncate)...)
	return allTables
}

func (config *Config) getObfuscatedTableNames() []string {
	tablesToObfuscate := make([]string, 0)
	for t, _ := range config.Tables.Obfuscate {
		tablesToObfuscate = append(tablesToObfuscate, t)
	}
	return tablesToObfuscate
}

func (config *Config) now() time.Time {
	if config.clock == nil {
		return time.Now()
	}
	return config.clock()
}
