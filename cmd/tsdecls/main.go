package main

import (
	"flag"
	"log"
	"os"

	"github.com/bobg/tsdecls"
)

func main() {
	var dir, typename, prefix string
	flag.StringVar(&dir, "dir", "", "directory containing Go type")
	flag.StringVar(&typename, "type", "App", "type name")
	flag.StringVar(&prefix, "prefix", "", "prefix of endpoint paths")
	flag.Parse()

	err := tsdecls.Write(os.Stdout, dir, typename, prefix)
	if err != nil {
		log.Fatal(err)
	}
}
