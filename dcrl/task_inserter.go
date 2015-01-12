package dcrl

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // sqlite3
)

type taskInserter struct {
	db    *sql.DB
	tx    *sql.Tx
	stmt  *sql.Stmt
	bufed int
}

const insertQuery = `insert into jobs
	(domain, output, result, err, log, id) values
	(?, ?, ?, ?, ?, ?)`

func newTaskInserter(db *sql.DB) (*taskInserter, error) {
	ret := new(taskInserter)
	ret.db = db

	e := ret.Prepare()
	if e != nil {
		return nil, e
	}

	return ret, nil
}

func (ins *taskInserter) Insert(t *task) error {
	_, e := ins.stmt.Exec(t.domain.String(),
		t.out, t.res, t.err, t.log, t.id,
	)

	if e != nil {
		return e
	}

	ins.bufed++
	if ins.bufed == 1000 {
		ins.Flush()
		ins.bufed = 0
	}

	return nil
}

func (ins *taskInserter) Prepare() error {
	var e error
	ins.tx, e = ins.db.Begin()
	if e != nil {
		return e
	}

	ins.stmt, e = ins.tx.Prepare(insertQuery)
	if e != nil {
		ins.tx.Rollback()
		return e
	}

	return nil
}

func (ins *taskInserter) Flush() error {
	e := ins.Close()
	if e != nil {
		return e
	}

	e = ins.Prepare()
	if e != nil {
		return e
	}

	return nil
}

func (ins *taskInserter) Close() error {
	err := ins.tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
