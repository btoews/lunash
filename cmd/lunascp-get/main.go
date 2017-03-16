package main

import (
	"flag"
	"log"
	"os"

	"github.com/mastahyeti/lunash"
)

var (
	pathArg = flag.String("path", "", "path of file to get from HSM")
	nameArg = flag.String("name", "", "name of HSM to get file from")
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

	file, err := client.ScpGet(path)
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(file)
}
