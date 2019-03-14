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
  -end-time int
    	end time(unix)
  -log-level string
    	log level (debug, info, warn, error) (default "warn")
  -start-time int
    	start time(unix)
```

Environment variable `AWS_REGION` and `MACKEREL_APIKEY` are required both.

By the default, end-time is now, start-time is 3 minuts ago.

## Usage (as library)

```go
import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

See [godoc](https://godoc.org/github.com/fujiwara/cloudwatch-to-mackerel/agent) in details.

### Query format

Same as [MetricDataQuery](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDataQuery.html) JSON format.

See also `aws cloudwatch get-metric-data help`.

### Label syntax

cloudwatch-to-mackerel parses `Label` fields in MetricDataQuery as Mackerel's service/host which to post metrics.

```ebnf
(* service name, host id, metric name are defined by Mackerel *)
service = "service=" , service name ;
host    = "host=" , host id ;
label   = ( service | host ) , ":" , metric name ;
```

## License

Copyright 2019 FUJIWARA Shunichiro.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
