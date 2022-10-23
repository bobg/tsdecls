package main

import (
	"flag"
	"log"
	"os"

	"github.com/bobg/tsdecls"
)

func main() {
	var dir, typename string
	flag.StringVar(&dir, "dir", "", "directory containing Go type")
	flag.StringVar(&typename, "type", "App", "type name")
	flag.Parse()

	err := tsdecls.Write(os.Stdout, dir, typename)
	if err != nil {
		log.Fatal(err)
	}
}
