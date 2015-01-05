package main

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

// pumps domains out of the CESR redis domain queue
func pump() {
	c, e := redis.Dial("tcp", "ccied50.sysnet.ucsd.edu:6379")
	ne(e)

	n, e := redis.Uint64(c.Do("llen", "domains"))
	ne(e)
	fmt.Printf("%d domains in total\n", n)

	doms, e := redis.Strings(c.Do("lrange", "domains", 0, 10))
	ne(e)

	for _, d := range doms {
		fmt.Println(d)
	}

	ne(c.Close())
}
