package dcrl

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

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

const nquota = 300

func makeQuotas() chan int {
	ret := make(chan int, nquota)
	for i := 0; i < nquota; i++ {
		ret <- i
	}
	return ret
}

func (j *Job) launch(c *dns8.Client, finished chan *task) {
	quotas := makeQuotas()

	for i, d := range j.Domains {
		quota := <-quotas
		t := &task{
			domain: d,
			client: c,
			id:     i,
		}

		go func(t *task, q int) {
			t.run()
			finished <- t
			quotas <- q
		}(t, quota)
	}
}

func (j *Job) crawl() error {
	c, e := dns8.NewClient()
	if e != nil {
		return e
	}

	finished := make(chan *task, 10)

	ins, err := newTaskInserter(j.db)
	if err != nil {
		return err
	}

	go j.launch(c, finished)

	ticker := time.Tick(time.Second * 3)
	n := 0
	for n < len(j.Domains) {
		select {
		case <-ticker:
			j.prog(n)
		case t := <-finished:
			err = ins.Insert(t)
			if err != nil {
				close(finished)
				return err
			}
			n++
		}
	}

	j.prog(n)

	return nil
}

func (j *Job) writeOut() error {
	if j.Archive != "" {
		e := os.MkdirAll(j.Archive, 0770)
		if e != nil {
			return e
		}
	}

	name := filepath.Join(j.Archive, j.Name)
	fout, err := os.Create(name)
	if err != nil {
		return err
	}

	defer fout.Close()

	rows, err := j.db.Query("select result from jobs order by id")
	if err != nil {
		return err
	}

	var line string
	for rows.Next() {
		err = rows.Scan(&line)
		if err != nil {
			return err
		}
		_, err = fout.Write([]byte(line))
		if err != nil {
			return err
		}
	}

	e := fout.Close()
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

	log.Printf("[%s] start crawling", j.Name)
	e = j.crawl()
	if e != nil {
		j.errProg(e)
		return e
	}

	log.Printf("[%s] generating output", j.Name)
	e = j.writeOut()
	if e != nil {
		j.errProg(e)
		return e
	}

	j.doneProg()
	return nil
}
