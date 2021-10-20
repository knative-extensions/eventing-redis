module knative.dev/eventing-redis

go 1.16

require (
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/go-redis/redis/v8 v8.4.2
	github.com/gomodule/redigo v1.8.3
	github.com/google/go-cmp v0.5.6
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	k8s.io/api v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
	knative.dev/eventing v0.26.1-0.20211020051652-2db2eef82875
	knative.dev/hack v0.0.0-20211019034732-ced8ce706528
	knative.dev/pkg v0.0.0-20211019132235-ba2b2b1bf268
	knative.dev/serving v0.26.1-0.20211020110753-e4b3034b0df6
)
