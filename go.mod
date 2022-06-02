module github.com/vicdeo/go-obfuscate

go 1.15

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/manveru/faker v0.0.0-20171103152722-9fbc68a78c4d
	github.com/pioz/faker v1.7.2
	github.com/spf13/viper v1.12.0
	github.com/vicdeo/go-obfuscate/mysqldump v0.0.0-20220530211018-cb9d6aa79d40
)

replace github.com/vicdeo/go-obfuscate/mysqldump => ./mysqldump
replace github.com/vicdeo/go-obfuscate/config => ./config
