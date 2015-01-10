package digo

import (
	"log"
)

func ne(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
