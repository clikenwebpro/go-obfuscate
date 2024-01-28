module github.com/vicdeo/go-obfuscate

go 1.15

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/manveru/faker v0.0.0-20171103152722-9fbc68a78c4d
	github.com/mibk/dupl v1.0.0 // indirect
	github.com/pioz/faker v1.7.2
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.1
)

replace github.com/vicdeo/go-obfuscate/mysqldump => ./mysqldump

replace github.com/vicdeo/go-obfuscate/config => ./config
