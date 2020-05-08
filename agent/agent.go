package agent

import (
	"context"
	"encoding/json"
	"fmt"
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

// Label represents structured label
type Label struct {
	Service  string
	HostID   string
	Name     string
	EmitZero bool
}

func (l Label) IsService() bool {
	return l.Service != ""
}

func (l Label) String() string {
	var s string
	if l.Service != "" {
		s = fmt.Sprintf("service=%s:%s", l.Service, l.Name)
	} else if l.HostID != "" {
		s = fmt.Sprintf("host=%s:%s", l.HostID, l.Name)
	}
	if l.EmitZero {
		s = s + ";emit_zero"
	}
	return s
}

// parse parses a label string as service, hostID, metric name.
func parseLabel(label string) (Label, error) {
	l := strings.SplitN(label, ":", 2)
	if len(l) != 2 {
		return Label{}, errors.New("invalid label format")
	}
	s := strings.SplitN(l[0], "=", 2)
	if len(s) != 2 {
		return Label{}, errors.New("invalid label format")
	}
	t, id, name := s[0], s[1], l[1]
	if t == "" || id == "" || name == "" {
		return Label{}, errors.New("invalid label format")
	}
	var parsed Label
	if strings.Contains(name, ";") {
		nameWithOpts := strings.Split(l[1], ";")
		name = nameWithOpts[0]
		for _, o := range nameWithOpts[1:] {
			switch o {
			case "emit_zero":
				parsed.EmitZero = true
			default:
				log.Printf("[warn] unknown option %s in label %s", o, label)
			}
		}
	}
	parsed.Name = name
	switch t {
	case "service":
		parsed.Service = id
	case "host":
		parsed.HostID = id
	default:
		return Label{}, fmt.Errorf("unknown label type %s", t)
	}
	return parsed, nil
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
	log.Printf("[debug] query %#v", qs)
	results, err := fetchMetrics(ctx, opt, qs)
	if err != nil {
		return errors.Wrap(err, "failed to fetch metrics from CloudWatch")
	}
	log.Printf("[debug] results %#v", results)
	serviceMetrics, hostMetrics := buildMetrics(results)
	log.Printf("[debug] service metrics %#v", serviceMetrics)
	log.Printf("[debug] host metrics %#v", hostMetrics)

	postMetrics(ctx, opt, serviceMetrics, hostMetrics)
	return nil
}

func fetchMetrics(ctx context.Context, opt Option, qs []*cloudwatch.MetricDataQuery) (map[Label]*cloudwatch.MetricDataResult, error) {
	svc := cloudwatch.New(opt.Session)
	var nextToken *string
	results := make(map[Label]*cloudwatch.MetricDataResult)
	for {
		if nextToken != nil {
			log.Printf("[debug] GetMetricData nextToken:%s", *nextToken)
		}
		log.Printf("[debug] GetMetricData from %s to %s", opt.StartTime, opt.EndTime)
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
			return nil, errors.Wrap(err, "failed to GetMetricData")
		}

		for _, r := range res.MetricDataResults {
			result := r
			label, err := parseLabel(*result.Label)
			if err != nil {
				log.Printf("[warn] %s label:%s", err, label)
				continue
			}
			results[label] = result
		}
		if res.NextToken == nil {
			// no more metrics
			break
		}
		nextToken = res.NextToken
	}
	for _, query := range qs {
		label, err := parseLabel(*query.Label)
		if err != nil {
			log.Printf("[warn] %s label:%s", err, label)
			continue
		}
		var period time.Duration
		if query.MetricStat.Period != nil {
			period = time.Duration(*query.MetricStat.Period) * time.Second
		} else if query.Period != nil {
			period = time.Duration(*query.Period) * time.Second
		}
		if res, ok := results[label]; ok && label.EmitZero && period > 0 {
			fillResult(opt, label, period, res)
		}
	}
	return results, nil
}

func fillResult(opt Option, label Label, period time.Duration, res *cloudwatch.MetricDataResult) {
	log.Printf("[debug] filling missing data points for %s (period %s)", label, period)
TIMESTAMP:
	for t := opt.StartTime.Truncate(time.Minute); t.Before(opt.EndTime); t = t.Add(period) {
		ts := t
		for _, v := range res.Timestamps {
			if v.Equal(ts) {
				continue TIMESTAMP
			}
		}
		res.Timestamps = append(res.Timestamps, &ts)
		res.Values = append(res.Values, aws.Float64(0))
	}
}

func buildMetrics(results map[Label]*cloudwatch.MetricDataResult) (serviceMetrics map[string][]*mackerel.MetricValue, hostMetrics []*mackerel.HostMetricValue) {
	serviceMetrics = make(map[string][]*mackerel.MetricValue)
	hostMetrics = make([]*mackerel.HostMetricValue, 0)
	for label, r := range results {
		log.Printf("[debug] build metrics for %s", label)
		if serviceMetrics[label.Service] == nil {
			serviceMetrics[label.Service] = make([]*mackerel.MetricValue, 0)
		}
		for i, ts := range r.Timestamps {
			tsUnix, value := (*ts).Unix(), *(r.Values[i])
			mv := &mackerel.MetricValue{
				Name:  label.Name,
				Time:  tsUnix,
				Value: value,
			}
			if label.IsService() {
				serviceMetrics[label.Service] = append(serviceMetrics[label.Service], mv)
				log.Printf("[debug] service:%s metric:%v", label.Service, mv)
			} else {
				hostMetrics = append(hostMetrics, &mackerel.HostMetricValue{
					HostID:      label.HostID,
					MetricValue: mv,
				})
				log.Printf("[debug] host:%s metric:%v", label.HostID, mv)
			}
		}
	}
	return serviceMetrics, hostMetrics
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
