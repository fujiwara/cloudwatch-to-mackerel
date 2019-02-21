package agent

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

const MaxMetricValues = 100

type Option struct {
	StartTime time.Time
	EndTime   time.Time
}

// parse parses a label string as service, hostID, metric name.
func parseLabel(label string) (string, string, string, error) {
	l := strings.SplitN(label, ":", 2)
	if len(l) != 2 {
		return "", "", "", errors.New("invalid label format")
	}
	s := strings.SplitN(l[0], "=", 2)
	if len(s) != 2 {
		return "", "", "", errors.New("invalid label format")
	}
	t, id, name := s[0], s[1], l[1]
	if t == "" || id == "" || name == "" {
		return "", "", "", errors.New("invalid label format")
	}
	switch t {
	case "service":
		return id, "", name, nil
	case "host":
		return "", id, name, nil
	}
	return "", "", "", errors.New("invalid label format")
}

// Run fetches metrics from CloudWatch by MetricDataQuery and post these to Mackerel.
func Run(ctx context.Context, sess *session.Session, query []byte, apiKey string, opt Option) error {
	var qs []*cloudwatch.MetricDataQuery
	if err := json.Unmarshal(query, &qs); err != nil {
		return errors.Wrap(err, "failed to parse query as MetricDataQuery")
	}

	svc := cloudwatch.New(sess)
	now := time.Now()
	if opt.StartTime.IsZero() {
		opt.StartTime = now.Add(-3 * time.Minute)
	}
	if opt.EndTime.IsZero() {
		opt.EndTime = now
	}

	res, err := svc.GetMetricData(&cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(opt.StartTime),
		EndTime:           aws.Time(opt.EndTime),
		MetricDataQueries: qs,
	})
	if err != nil {
		return errors.Wrap(err, "failed to GetMetricData")
	}

	serviceMetrics := make(map[string][]*mackerel.MetricValue)
	hostMetrics := make([]*mackerel.HostMetricValue, 0)

	for _, r := range res.MetricDataResults {
		for i, ts := range r.Timestamps {
			tsUnix, value := (*ts).Unix(), *(r.Values[i])
			service, hostID, name, err := parseLabel(*r.Label)
			if err != nil {
				log.Printf("[warn] %s label:%s", err, *r.Label)
				continue
			}
			mv := &mackerel.MetricValue{
				Name:  name,
				Time:  tsUnix,
				Value: value,
			}
			if service != "" {
				serviceMetrics[service] = append(serviceMetrics[service], mv)
			} else {
				hostMetrics = append(hostMetrics, &mackerel.HostMetricValue{
					HostID:      hostID,
					MetricValue: mv,
				})
			}
		}
	}

	client := mackerel.NewClient(apiKey)

	// post service metrics
	for service, values := range serviceMetrics {
		service, values := service, values
		size := len(values)
		for i := 0; i < size; i += MaxMetricValues {
			start, end := i, i+MaxMetricValues
			if size < end {
				end = size
			}
			log.Printf("[info] PostServiceMetricValues %s values[%d:%d]", service, start, end)
			err := client.PostServiceMetricValues(service, values[start:end])
			if err != nil {
				log.Printf("[warn] failed to PostServiceMetricValues service:%s %s", service, err)
			}
		}
	}

	// post host metrics
	size := len(hostMetrics)
	for i := 0; i < size; i += MaxMetricValues {
		start, end := i, i+MaxMetricValues
		if size < end {
			end = size
		}
		log.Printf("[info] PostHostMetricValues values[%d:%d]", start, end)
		err := client.PostHostMetricValues(hostMetrics[start:end])
		if err != nil {
			log.Println("[warn] failed to PostHostMetricValues", err)
		}
	}

	return nil
}
