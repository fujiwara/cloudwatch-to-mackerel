module github.com/fujiwara/cloudwatch-to-mackerel

go 1.21

toolchain go1.22.0

require (
	github.com/aws/aws-sdk-go-v2 v1.30.3
	github.com/aws/aws-sdk-go-v2/config v1.27.27
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.40.3
	github.com/fujiwara/lamblocal v0.0.4
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/logutils v1.0.0
	github.com/mackerelio/mackerel-client-go v0.31.0
	github.com/shogo82148/go-retry v1.2.0
)

require (
	github.com/aws/aws-lambda-go v1.45.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.27 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.3 // indirect
	github.com/aws/smithy-go v1.20.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)
