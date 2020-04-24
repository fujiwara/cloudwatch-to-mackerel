package agent

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

const batchSize = 100

// Option represents agent option.
type Option struct {
	StartTime time.Time
	EndTime   time.Time
	APIKey    string
	Query     []byte
	Session   *session.Session
}

type ParsedLabel struct {
	Service string
	HostID  string
	Name    string
	Options map[string]struct{}
}

// parse parses a label string as service, hostID, metric name.
func parseLabel(label string) (*ParsedLabel, error) {
	l := strings.SplitN(label, ":", 2)
	if len(l) != 2 {
		return nil, errors.New("invalid label format")
	}
	s := strings.SplitN(l[0], "=", 2)
	if len(s) != 2 {
		return nil, errors.New("invalid label format")
	}
	t, id, name := s[0], s[1], l[1]
	if t == "" || id == "" || name == "" {
		return nil, errors.New("invalid label format")
	}

	options := make(map[string]struct{}, 0)
	if strings.Contains(name, ";") {
		nameWithOpts := strings.Split(l[1], ";")
		name = nameWithOpts[0]
		for _, o := range nameWithOpts[1:] {
			o := o
			options[o] = struct{}{}
		}
	}

	switch t {
	case "service":
		return &ParsedLabel{
			Service: id,
			Name:    name,
			Options: options,
		}, nil
	case "host":
		return &ParsedLabel{
			HostID:  id,
			Name:    name,
			Options: options,
		}, nil
	}
	return nil, errors.New("invalid label format")
}

func validateOption(opt *Option) (err error) {
	sess := opt.Session
	if sess == nil {
		opt.Session, err = session.NewSession(&aws.Config{})
		if err != nil {
			return errors.Wrap(err, "failed to new session")
		}
	}
	if opt.APIKey == "" {
		opt.APIKey = os.Getenv("MACKEREL_APIKEY")
	}
	if opt.APIKey == "" {
		return errors.New("Option.APIKey or MACKEREL_APIKEY envrionment variable is required")
	}
	now := time.Now()
	if opt.StartTime.IsZero() {
		opt.StartTime = now.Add(-3 * time.Minute)
	}
	if opt.EndTime.IsZero() {
		opt.EndTime = now
	}
	return nil
}

// Run fetches metrics from CloudWatch by MetricDataQuery and post these to Mackerel.
func Run(opt Option) error {
	return RunWithContext(context.Background(), opt)
}

// RunWithContext fetches metrics from CloudWatch by MetricDataQuery with context and post these to Mackerel.
func RunWithContext(ctx context.Context, opt Option) error {
	if err := validateOption(&opt); err != nil {
		return err
	}
	var qs []*cloudwatch.MetricDataQuery
	if err := json.Unmarshal(opt.Query, &qs); err != nil {
		return errors.Wrap(err, "failed to parse query as MetricDataQuery")
	}

	serviceMetrics, hostMetrics, err := fetchMetrics(ctx, opt, qs)
	if err != nil {
		return errors.Wrap(err, "failed to fetch metrics from CloudWatch")
	}

	postMetrics(ctx, opt, serviceMetrics, hostMetrics)
	return nil
}

func fetchMetrics(ctx context.Context, opt Option, qs []*cloudwatch.MetricDataQuery) (map[string][]*mackerel.MetricValue, []*mackerel.HostMetricValue, error) {
	svc := cloudwatch.New(opt.Session)

	serviceMetrics := make(map[string][]*mackerel.MetricValue)
	hostMetrics := []*mackerel.HostMetricValue{}
	var nextToken *string
	results := make(map[string]*cloudwatch.MetricDataResult)
	for {
		if nextToken != nil {
			log.Printf("[debug] GetMetricData nextToken:%s", *nextToken)
		}
		res, err := svc.GetMetricDataWithContext(
			ctx,
			&cloudwatch.GetMetricDataInput{
				StartTime:         aws.Time(opt.StartTime),
				EndTime:           aws.Time(opt.EndTime),
				MetricDataQueries: qs,
				NextToken:         nextToken,
			},
		)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to GetMetricData")
		}

		for _, r := range res.MetricDataResults {
			result := r
			label := *result.Label
			results[label] = result
		}
		if res.NextToken == nil {
			// no more metrics
			break
		}
		nextToken = res.NextToken
	}

	for _, query := range qs {
		label := *query.Label
		p, err := parseLabel(label)
		if err != nil {
			log.Printf("[warn] %s label:%s", err, label)
			continue
		}
		r, ok := results[label]
		if !ok {
			// data points not found
			if _, ok := p.Options["emit_zero"]; ok {
				r = &cloudwatch.MetricDataResult{
					Timestamps: []*time.Time{&opt.EndTime},
					Values:     []*float64{aws.Float64(0)},
				}
			}
		}
		if r == nil {
			continue
		}
		for i, ts := range r.Timestamps {
			tsUnix, value := (*ts).Unix(), *(r.Values[i])
			mv := &mackerel.MetricValue{
				Name:  p.Name,
				Time:  tsUnix,
				Value: value,
			}
			if p.Service != "" {
				serviceMetrics[p.Service] = append(serviceMetrics[p.Service], mv)
				log.Printf("[debug] service:%s metric:%v", p.Service, mv)
			} else {
				hostMetrics = append(hostMetrics, &mackerel.HostMetricValue{
					HostID:      p.HostID,
					MetricValue: mv,
				})
				log.Printf("[debug] host:%s metric:%v", p.HostID, mv)
			}
		}
	}
	return serviceMetrics, hostMetrics, nil
}

func postMetrics(ctx context.Context, opt Option, serviceMetrics map[string][]*mackerel.MetricValue, hostMetrics []*mackerel.HostMetricValue) {
	client := mackerel.NewClient(opt.APIKey)

	// post service metrics
	for service, values := range serviceMetrics {
		service, values := service, values
		size := len(values)
		for i := 0; i < size; i += batchSize {
			start, end := i, i+batchSize
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
	for i := 0; i < size; i += batchSize {
		start, end := i, i+batchSize
		if size < end {
			end = size
		}
		log.Printf("[info] PostHostMetricValues values[%d:%d]", start, end)
		err := client.PostHostMetricValues(hostMetrics[start:end])
		if err != nil {
			log.Println("[warn] failed to PostHostMetricValues", err)
		}
	}
}
