package main

import (
	"log"
)

func ne(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func le(e error) {
	if e != nil {
		log.Print(e)
	}
}
