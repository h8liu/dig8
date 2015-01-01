package dig8

import (
	"database/sql"
	"log"
	"net/rpc"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // sqlite3 db support

	"lonnie.io/dig8/dns8"
)

type job struct {
	name     string
	archive  string
	domains  []*dns8.Domain
	callback string
	crawled  int
	respond  *Respond

	db     *sql.DB
	client *rpc.Client
	quotas chan int

	resChan   chan *task
	writeDone chan bool
}

func newJob(name string, doms []*dns8.Domain, cb string) *job {
	ret := new(job)
	ret.name = name
	ret.domains = doms
	ret.callback = cb

	ret.respond = new(Respond)
	ret.respond.Name = name
	ret.respond.Total = len(doms)

	return ret
}

func (j *job) connect() error {
	var e error
	j.client, e = rpc.DialHTTP("tcp", j.callback)
	return e
}

func (j *job) call() error {
	var s string
	return j.client.Call("Respond", j.respond, &s)
}

func (j *job) cb() {
	if j.callback == "" {
		return
	}

	// in callback, we log error
	var e error
	if j.client == nil {
		e = j.connect()
		if e != nil {
			log.Print(j.name, e)
			return
		}
	}

	e = j.call()
	if e == rpc.ErrShutdown {
		j.client.Close()
		e = j.connect()
		if e != nil {
			log.Print(j.name, e)
			return
		}
		e = j.call()
	}

	if e != nil {
		log.Print(j.name, e)
	}
}

func (j *job) cleanup() {
	if j.db != nil {
		e := j.db.Close()
		if e != nil {
			log.Print(j.db, e)
		}
	}

	if j.client != nil {
		e := j.client.Close()
		if e != nil {
			log.Print(j.name, e)
		}
	}
}

func (j *job) fail(e error) {
	j.respond.Error = e.Error()
	j.cb()
}

func (j *job) failOn(e error) bool {
	if e != nil {
		j.fail(e)
		return true
	}
	return false
}

func (j *job) run() {
	log.Printf("job %s started", j.name)
	defer log.Printf("job %s done", j.name)
	defer j.cleanup()

	db, err := sql.Open("sqlite3", j.name+".db")
	if j.failOn(err) {
		return
	}

	j.db = db

	q := func(sql string) bool {
		_, e := db.Exec(sql)
		if e != nil {
			log.Printf("sql fail: %s\n", sql)
			log.Print(e)
			j.fail(e)
			return false
		}

		return true
	}

	if !q(`create table jobs (
			id int not null primary key,
			domain text not null,
			output text not null,
			result text not null,
			err text not null,
			log text not null)`) {
		return
	}

	if !q(`create table doms (
			id int not null primary key,
			domain text not null
		)`) {
		return
	}

	tx, err := j.db.Begin()
	if j.failOn(err) {
		return
	}
	stmt, err := tx.Prepare(`insert into doms (id, domain) 
			values (?, ?)`)
	if j.failOn(err) {
		return
	}

	for i, d := range j.domains {
		_, err = stmt.Exec(i+1, d.String())
		if j.failOn(err) {
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if j.failOn(err) {
		return
	}

	log.Printf("job %s starts crawling", j.name)
	j.crawl()
}

const quota = 300

func (j *job) makeQuotas() chan int {
	nquota := quota
	ret := make(chan int, nquota)
	for i := 0; i < nquota; i++ {
		ret <- i
	}
	return ret
}

func (j *job) crawl() {
	c, e := dns8.NewClient()
	if j.failOn(e) {
		return
	}

	j.quotas = j.makeQuotas()
	j.resChan = make(chan *task, quota)
	defer close(j.resChan)

	// launch the jobs
	go func() {
		for i, d := range j.domains {
			quota := <-j.quotas
			task := &task{
				domain:   d,
				client:   c,
				quota:    quota,
				id:       i + 1,
				quotaRet: j.quotas,
				taskDone: j.resChan,
			}
			go task.run()
		}
	}()

	j.writeOut()
}

func (j *job) writeOut() {
	n := 0
	total := j.respond.Total

	chkerr := func(e error) bool {
		if e != nil {
			j.respond.Error = e.Error()
			j.cb()
			return true
		}
		return false
	}

	const insertStmt = `insert into jobs
		(domain, output, result, err, log, id) values
		(?, ?, ?, ?, ?, ?)`

	tx, err := j.db.Begin()
	if chkerr(err) {
		return
	}
	stmt, err := tx.Prepare(insertStmt)
	if chkerr(err) {
		return
	}

	ticker := time.Tick(time.Second * 3)
	bufed := 0

	for n < total {
		t := <-j.resChan

		_, err = stmt.Exec(t.domain.String(),
			t.out, t.res, t.err, t.log, t.id,
		)
		if chkerr(err) {
			return
		}

		n++
		bufed++
		j.respond.Crawled = n

		if bufed > 5000 {
			err = tx.Commit()
			if chkerr(err) {
				return
			}

			tx, err = j.db.Begin()
			if chkerr(err) {
				return
			}
			stmt, err = tx.Prepare(insertStmt)
			if chkerr(err) {
				return
			}

			bufed = 0
		}

		if len(ticker) > 0 {
			// report progress
			j.cb()
		}
	}

	err = tx.Commit()
	if chkerr(err) {
		return
	}

	j.cb()

	fout, err := os.Create(j.archive + j.name)
	if chkerr(err) {
		return
	}

	rows, err := j.db.Query(`select result from jobs order by id`)
	if chkerr(err) {
		return
	}

	var line string
	for rows.Next() {
		err = rows.Scan(&line)
		if chkerr(err) {
			fout.Close()
			return
		}
		_, err = fout.Write([]byte(line))
		if chkerr(err) {
			fout.Close()
			return
		}
	}

	err = fout.Close()
	if chkerr(err) {
		return
	}

	j.respond.Done = true
	j.cb()
}
