package noblankimport

import (
	"fmt"
	_ "net/http/pprof" // want `blank import of "net/http/pprof" outside package main`
)

func Use() { fmt.Println() }
