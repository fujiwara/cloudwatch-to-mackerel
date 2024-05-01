package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fujiwara/cloudwatch-to-mackerel/agent"
	"github.com/fujiwara/lamblocal"
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
	var showVersion bool
	var queryFile string
	var query string
	var dumpMetrics bool
	var timeOffset time.Duration
	flag.Int64Var(&startTime, "start-time", 0, "start time(unix) if negative, relative time from now. env:CW2MKR_START_TIME")
	flag.Int64Var(&endTime, "end-time", 0, "end time(unix) if negative, relative time from now. env:CW2MKR_END_TIME")
	flag.StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error) env:CW2MKR_LOG_LEVEL")
	flag.StringVar(&queryFile, "query-file", "", "query file path. env:CW2MKR_QUERY_FILE")
	flag.StringVar(&query, "query", "", "query content. env:CW2MKR_QUERY")
	flag.BoolVar(&dumpMetrics, "dump-metrics", false, "dump metrics to stdout instead of posting to mackerel. env:CW2MKR_DUMP_METRICS")
	flag.DurationVar(&timeOffset, "time-offset", 0, "time offset for relative time. env:CW2MKR_TIME_OFFSET")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.VisitAll(envToFlag)
	flag.Parse()

	if showVersion {
		fmt.Printf("cloudwatch-to-mackerel(cw2mkr) %s\n", agent.Version)
		return nil
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	opt := agent.Option{
		TimeOffset: timeOffset,
		DumpMetrics: dumpMetrics,
	}
	if startTime > 0 {
		opt.StartTime = time.Unix(startTime, 0)
	} else if startTime < 0 {
		opt.StartTime = time.Now().Add(time.Duration(startTime) * time.Second)
	}

	if endTime > 0 {
		opt.EndTime = time.Unix(endTime, 0)
	} else if endTime < 0 {
		opt.StartTime = time.Now().Add(time.Duration(endTime) * time.Second)
	}

	if query != "" && queryFile != "" {
		return errors.New("both query and query-file are specified")
	}
	if query != "" {
		opt.Query = []byte(query)
	} else if queryFile != "" {
		q, err := os.ReadFile(queryFile)
		if err != nil {
			return err
		}
		opt.Query = q
	} else {
		if len(flag.Args()) != 1 {
			return errors.New("an argument MetricDataQuery json file path is required")
		}
		q, err := os.ReadFile(flag.Args()[0])
		if err != nil {
			return err
		}
		opt.Query = q
	}

	lamblocal.CLISrc = strings.NewReader("{}") // empty payload
	return lamblocal.RunWithError(context.TODO(), func(ctx context.Context, _ struct{}) (struct{}, error) {
		return struct{}{}, agent.RunWithContext(ctx, opt)
	})
}

func envToFlag(f *flag.Flag) {
	names := []string{
		strings.ToUpper(strings.Replace(f.Name, "-", "_", -1)),
		strings.ToLower(strings.Replace(f.Name, "-", "_", -1)),
	}
	for _, name := range names {
		if s := os.Getenv("CW2MKR_" + name); s != "" {
			f.Value.Set(s)
			break
		}
		if s := os.Getenv(name); s != "" {
			f.Value.Set(s)
			break
		}
	}
}
