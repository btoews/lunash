package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"

	"github.com/mastahyeti/lunash"
	"github.com/pkg/errors"
)

var (
	pathArg = flag.String("path", "", "where to put the file on the HSM")
	nameArg = flag.String("name", "", "name of HSM to put file on")
	confArg = flag.String("config", "./lunash.json", "path to the config file")

	path     string
	name     string
	confPath string
)

func parseFlags() {
	flag.Parse()

	if pathArg != nil && len(*pathArg) > 0 {
		path = *pathArg
	} else {
		flag.Usage()
		os.Exit(1)
	}

	if nameArg != nil && len(*nameArg) > 0 {
		name = *nameArg
	} else {
		flag.Usage()
		os.Exit(1)
	}

	if confArg != nil && len(*confArg) > 0 {
		confPath = *confArg
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	parseFlags()

	config, err := lunash.LoadConfig(confPath, name)
	if err != nil {
		log.Fatal(err)
	}

	client := config.Client()
	if err = client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, os.Stdin); err != nil {
		log.Fatal(errors.Wrap(err, "Error reading file from stdin"))
	}

	if err := client.ScpPut(path, buf.Bytes()); err != nil {
		log.Fatal(err)
	}
}
