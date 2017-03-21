package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mastahyeti/lunash"
	"github.com/mastahyeti/lunash/scp"
)

var (
	loginArg = flag.Bool("login", false, "run 'hsm login' before the commands")
	cmdArg   = flag.String("command", "", "a semicolon delimited list of commands to run")
	namesArg = flag.String("names", "", "comma separated list of HSMs to send command to")
	allArg   = flag.Bool("all", false, "send commands to all HSMs in the config file")
	confArg  = flag.String("config", "./lunash.json", "path to the config file")
	debugArg = flag.Bool("debug", false, "whether to output debugging information")

	login    bool
	confPath string
	all      bool
	names    []string
	commands []string
)

func parseFlags() {
	flag.Parse()

	if loginArg != nil && *loginArg {
		login = true
	}

	if allArg != nil && *allArg {
		all = true
	} else if namesArg != nil && len(*namesArg) > 0 {
		names = strings.Split(*namesArg, ",")
	} else {
		flag.Usage()
		os.Exit(1)
	}

	if cmdArg != nil && len(*cmdArg) > 0 {
		for _, cmd := range strings.Split(*cmdArg, ";") {
			commands = append(commands, strings.TrimSpace(cmd))
		}
	} else {
		flag.Usage()
		os.Exit(1)
	}

	if len(commands) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if confArg != nil && len(*confArg) > 0 {
		confPath = *confArg
	} else {
		flag.Usage()
		os.Exit(1)
	}

	if debugArg != nil && *debugArg {
		scp.Debug = true
	}
}

func main() {
	parseFlags()

	var configs []*lunash.Config
	var err error

	if all {
		configs, err = lunash.LoadAllConfigs(confPath)
	} else {
		configs, err = lunash.LoadConfigs(confPath, names)
	}
	if err != nil {
		log.Fatal(err)
	}

	for _, config := range configs {
		client := config.Client()
		if err := client.Connect(); err != nil {
			log.Fatalf("host=%s error='%s'", config.Hostname, err.Error())
		}

		outputs, err := client.Run(commands, login)
		for i := range outputs {
			log.Printf("host=%s cmd=%s\n%s\n",
				config.Hostname,
				strconv.QuoteToASCII(commands[i]),
				outputs[i],
			)
		}

		if err != nil {
			log.Fatalf("host=%s error='%s'", config.Hostname, err.Error())
		}

		if err := client.Close(); err != nil {
			log.Fatalf("host=%s error='%s'", config.Hostname, err.Error())
		}
	}
}
