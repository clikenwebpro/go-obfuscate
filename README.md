# go-obfuscate

## Installation et usage
Ce projet a été forké et j'ai gardé le readme initial qui est toute la suite mais vous n'aurez pas besoin de lire toute la suite.
Il est utile surtout pour obfusquer une bdd locale de prod. de Domino afin de la partager.
Pour générer une bdd avec des données obfusquées : 
* installer go correspondant à votre OS
* cloner ce repos
* lancer votre instance dockerisée de domino avec une BDD de prod.
* récupérer le fichier de config.yaml avec hjacquir
* lancer la commande : ``go run .``
* le fichier de dump sera généré dans le répertoire `dumps`
* pour recréer votre bdd de Domino avec les données obfusquées utiliser la commande : ``mysql -u root -proot domino < dump.sql``
**Attention pour récréer la bdd si vous devez ajouter dans le fichier sql la directive : ``SET GLOBAL max_allowed_packet=3073741824;`` sinon vous aurez une erreur de typ : ERROR 2006 (HY000) at line 4860: MySQL server has gone away.**

Welcome to the world of a mysqldump-alike tool that allows to produce a customized dump files on-the-fly. This includes:
- skipping tables from dumping
- truncating tables after the dump is applied
- keeping relations but replacing sensitive data in the dump file with a pretty-looking garbage.

## Usage
```
go-obfuscate [-c /path/to/config/file.yaml]
```
You'll need a configuration file in the YAML format.
By default `config.yaml` is used in the current directory.

It's a good idea to start with copying `config.yaml.sample` into `config.yaml` and use its original content as a reference.

## Configuration file format
The file has three main sections:
- `database` - this section contains database connection parameters
- `output` - this section contains output file parameters
- `tables` - this section contains subsections where database table names are listed

`tables` section has four subsections:
- `keep`- all tables listed in this section are dumped as-is, like an ordinary `mysqldump` does
- `ignore` - all tables listed in this section are **not** dumped
- `truncate` - all tables listed in this section are dumped as a pair of `DROP TABLE table_name` + `CREATE TABLE table_name` MySQL queries. No data is dumped.
- `obfuscate` - tables that are listed in this section could have column names and column type. In case the table has no single column name specified it behaves just like it was listed in the `keep` section. Otherwise the fake data of a specified type is generated and written into the dump instead of real data of a target column.

## Sanity checks
Before the creation of the dump the following checks are done:
- each subsection of `tables` is checked separately for duplicated table names inside it to ensure that the same table is not listed in the subsection multiple times.
- all columns that are going to be obfuscated are checked to have a known type (name, email, address, etc)
- all subsections of `tables` are checked for duplicated table names to ensure that the same table is not listed in the multiple subsections
- all tables listed in the configuration file are checked for existence in DB to prevent typos  in the table names
- all tables that are available in the DB are checked for presence in the `tables` section of the configuration file to ensure that the strategy is clear
- **[this one is not implemented yet]** all columns that are going to be obfuscated are checked for existence to prevent typos in the column names

Failing **any** of the checks above stops the program execution until the config file is fixed.

## Exit codes
The program produces non-zero exit codes on error

## License
This work is provided by [Viktar Dubiniuk](https://github.com/VicDeo) under MIT License

## Credits
- [go-mysqldump](https://github.com/jamf/go-mysqldump)
- [faker library by manveru](https://github.com/manveru/faker)
- [faker library by pioz](https://github.com/pioz/faker)
