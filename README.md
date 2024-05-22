# cloudwatch-to-mackerel

[![GoDoc](https://godoc.org/github.com/fujiwara/cloudwatch-to-mackerel?status.svg)][godoc]

[godoc]: https://godoc.org/github.com/fujiwara/cloudwatch-to-mackerel

Copy metrics from [Amazon CloudWatch](https://aws.amazon.com/cloudwatch/) to [Mackerel](https://mackerel.io).

cloudwatch-to-mackerel agent fetches metrics from Amazon CloudWatch by MetricDataQuery, and post these metrics to Mackerel as service/host metrics.

## Usage (cw2mkr command)

### Install

Homebrew or [binary packages](https://github.com/fujiwara/cloudwatch-to-mackerel/releases).

```console
$ brew install fujiwara/tap/cloudwatch-to-mackerel
```

### Usage

```console
$ cw2mkr metric-data-query.json
```

```
Usage of cw2mkr:
  -dump-metrics
        dump metrics to stdout instead of posting to mackerel. env:CW2MKR_DUMP_METRICS
  -end-time int
        end time(unix) if negative, relative time from now. env:CW2MKR_END_TIME
  -log-level string
        log level (debug, info, warn, error) env:CW2MKR_LOG_LEVEL (default "info")
  -query string
        query content. env:CW2MKR_QUERY
  -query-file string
        query file path. env:CW2MKR_QUERY_FILE
  -start-time int
        start time(unix) if negative, relative time from now. env:CW2MKR_START_TIME
  -time-offset duration
        time offset for relative time. adds the offset to mackerel's metrics env:CW2MKR_TIME_OFFSET
  -version
        show version
```

Environment variable `AWS_REGION` and `MACKEREL_APIKEY` are required both.

By the default, end-time is now, start-time is 3 min ago.

Some flags can be set by environment variables. The following is an example.

```console
$ cw2mkr -query-file query.json
$ CW2MKR_QUERY_FILE=query.json cw2mkr
```

## Usage (as library)

```go
import (
	"github.com/fujiwara/cloudwatch-to-mackerel/agent"
)

// MetricDataQuery JSON
query := []byte(`
[
  {
    "Id": "m1",
    "Label": "service=MyService:alb.my-alb.response-time.p99",
    "MetricStat": {
      "Metric": {
        "Namespace": "AWS/ApplicationELB",
        "MetricName": "TargetResponseTime",
        "Dimensions": [
          {
            "Name": "LoadBalancer",
            "Value": "app/my-alb/8e0641feccf3491c"
          }
        ]
      },
      "Period": 60,
      "Stat": "p99"
    }
  },
  {
    "Id": "m2",
    "Label": "service=MyService:alb.my-alb.response-time.p90",
    "MetricStat": {
      "Metric": {
        "Namespace": "AWS/ApplicationELB",
        "MetricName": "TargetResponseTime",
        "Dimensions": [
          {
            "Name": "LoadBalancer",
            "Value": "app/my-alb/8e0641feccf3491c"
          }
        ]
      },
      "Period": 60,
      "Stat": "p90"
    }
  }
]
`)

err := agent.Run(agent.Option{Query: query})
```

See [pkg.go.dev](https://pkg.go.dev/github.com/fujiwara/cloudwatch-to-mackerel/agent) in details.

## Usage (Lambda)

A `cw2mkr` binary can be used as an AWS Lambda function.

1. Rename `cw2mkr` binary to `bootstrap`.
2. Create a zip archive with `bootstrap` and `metric-data-query.json`.
3. Upload the zip archive to AWS Lambda.
4. Invoke the Lambda function.

You can set flags by environment variables (`CW2MKR_QUERY_FILE` etc.).

Supported runtime is `provided.al2` and `provided.al2023`.

If the binary name is "cw2mkr", cw2mkr runs as a CLI command whenever it is executed on a Lambda environment or not.

### Query format

Same as [MetricDataQuery](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDataQuery.html) JSON format.

See also `aws cloudwatch get-metric-data help`.

### Label syntax

cloudwatch-to-mackerel parses `Label` fields in MetricDataQuery as Mackerel's service/host which to post metrics.

```ebnf
(* service_name, host_id, metric_name are defined by Mackerel *)
service = "service=" , service_name ;
host    = "host=" , host_id ;
option  = "emit_zero"
label   = ( service | host ) , ":" , metric_name , { ";", option } ;
```

## License

Copyright 2019 FUJIWARA Shunichiro.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
