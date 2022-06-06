package mysqldump

import (
	"database/sql"
	"io"
	"os"

	"github.com/vicdeo/go-obfuscate/config"
)

/*
Register a new dumper.

	db: Database that will be dumped (https://golang.org/pkg/database/sql/#DB).
	conf: config read from the file
*/
func Register(db *sql.DB, conf *config.Config) (*Data, error) {
	// Create .sql file
	f, err := os.Create(conf.GetDumpFullPath())
	if err != nil {
		return nil, err
	}

	return &Data{
		Out:        f,
		Connection: db,
	}, nil
}

// Dump Creates a MYSQL dump from the connection to the stream.
// Seems to be unused.
func Dump(db *sql.DB, out io.Writer) error {
	return (&Data{
		Connection: db,
		Out:        out,
	}).Dump()
}

// Close the dumper.
// Will also close the database the dumper is connected to as well as the out stream if it has a Close method.
//
// Not required.
func (d *Data) Close() error {
	defer func() {
		d.Connection = nil
		d.Out = nil
	}()
	if out, ok := d.Out.(io.Closer); ok {
		out.Close()
	}
	return d.Connection.Close()
}
