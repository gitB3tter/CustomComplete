package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"runtime"
)

//go:embed schema.sql
var mainSchema string

type ReadWrite struct{ *sql.DB }

type Read struct{ *sql.DB }

func ResetDB(path string) error {
	var errs []error
	for _, p := range []string{path, path + "-wal", path + "-shm", path + "-journal"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func GetReadWrite(path string) (ReadWrite, error) {
	//combining dsn styles, since unknown params are ignored, allows swapping out driver
	dsn := "file:" + path +
		"?_txlock=immediate" +
		"&_pragma=journal_mode(WAL)&_journal_mode=WAL" +
		"&_pragma=busy_timeout(5000)&_busy_timeout=5000" +
		"&_pragma=synchronous(NORMAL)&_sync=NORMAL" +
		"&_pragma=foreign_keys(1)&_fk=1" +
		"&_pragma=cache_size(-64000)&_cache_size=-64000"
	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return ReadWrite{}, err
	}
	if _, err = db.ExecContext(context.Background(), mainSchema); err != nil {
		db.Close()
		return ReadWrite{}, err
	}
	db.SetMaxOpenConns(1)
	return ReadWrite{db}, nil
}

func GetRead(path string) (Read, error) {
	dsn := "file:" + path +
		"?mode=ro" +
		"&_pragma=query_only(1)" +
		"&_pragma=busy_timeout(5000)&_busy_timeout=5000" +
		"&_pragma=synchronous(NORMAL)&_sync=NORMAL" +
		"&_pragma=foreign_keys(1)&_fk=1"
	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return Read{}, err
	}

	numberConnections := max(4, runtime.NumCPU())
	db.SetMaxOpenConns(numberConnections)
	db.SetMaxIdleConns(numberConnections)

	return Read{db}, nil
}
