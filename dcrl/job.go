package dcrl

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // sqlite3

	"lonnie.io/dig8/dns8"
)

// Progress reports how much a job is crawled
type Progress struct {
	Name    string
	Crawled int
	Total   int
	Done    bool
	Error   string
}

// Job is a crawler job that contains a list of domains
type Job struct {
	Name     string
	DB       string
	Archive  string
	Domains  []*dns8.Domain
	Progress func(p *Progress) // progress report function

	db *sql.DB
}

func (j *Job) reportProg(p *Progress) {
	if j.Progress == nil {
		return
	}
	j.Progress(p)
}

func (j *Job) prog(crawled int) {
	ret := new(Progress)
	ret.Name = j.Name
	ret.Total = len(j.Domains)
	ret.Crawled = crawled

	j.reportProg(ret)
}

func (j *Job) errProg(e error) {
	ret := new(Progress)
	ret.Name = j.Name
	ret.Total = len(j.Domains)
	ret.Done = true
	ret.Error = e.Error()

	j.reportProg(ret)
}

func (j *Job) doneProg() {
	ret := new(Progress)
	ret.Name = j.Name
	ret.Total = len(j.Domains)
	ret.Done = true

	j.reportProg(ret)
}

func (j *Job) createDB() error {
	dbpath := j.DB
	if dbpath == "" {
		dbpath = j.Name + ".db"
	}

	if _, e := os.Stat(dbpath); e == nil {
		log.Printf("[%s] job db exists, try to delete", j.Name)
		e = os.Remove(dbpath)
		if e != nil {
			log.Printf("[%s] db remove failed", j.Name)
			return e
		}
	}

	db, e := sql.Open("sqlite3", dbpath)
	if e != nil {
		return e
	}

	j.db = db
	q := func(sql string) error {
		_, e := db.Exec(sql)
		if e != nil {
			log.Printf("sql fail: %s\n", sql)
			log.Print(e)
			return e
		}

		return nil
	}

	if e = q(`create table jobs (
			id int not null primary key,
			domain text not null,
			output text not null,
			result text not null,
			err text not null,
			log text not null)`); e != nil {
		return e
	}

	if e = q(`create table doms (
			id int not null primary key,
			domain text not null
		)`); e != nil {
		return e
	}

	tx, e := j.db.Begin()
	if e != nil {
		return e
	}

	stmt, e := tx.Prepare(`insert into doms (id, domain) values (?, ?)`)
	if e != nil {
		return e
	}

	for i, d := range j.Domains {
		_, e = stmt.Exec(i+1, d.String())
		if e != nil {
			tx.Rollback()
			return e
		}
	}

	e = tx.Commit()
	if e != nil {
		return e
	}

	return nil
}

// Do performs the job.
func (j *Job) Do() error {
	log.Printf("[%s] job started", j.Name)
	defer log.Printf("[%s] job finished", j.Name)
	defer func() {
		if j.db != nil {
			j.db.Close()
		}
	}()

	e := j.createDB()
	if e != nil {
		j.errProg(e)
		return e
	}

	// TODO
	return nil
}
