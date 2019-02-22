package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fujiwara/cloudwatch-to-mackerel/agent"
	"github.com/hashicorp/logutils"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	var startTime, endTime int64
	var logLevel string
	flag.Int64Var(&startTime, "start-time", 0, "start time(unix)")
	flag.Int64Var(&endTime, "end-time", 0, "end time(unix)")
	flag.StringVar(&logLevel, "log-level", "warn", "log level (debug, info, warn, error)")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	opt := agent.Option{}
	if startTime != 0 {
		opt.StartTime = time.Unix(startTime, 0)
	}
	if endTime != 0 {
		opt.EndTime = time.Unix(endTime, 0)
	}

	if len(flag.Args()) != 1 {
		return errors.New("an argument MetricDataQuery json file path is requried")
	}
	f, err := os.Open(flag.Args()[0])
	if err != nil {
		return err
	}
	opt.Query, err = ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return agent.Run(opt)
}
