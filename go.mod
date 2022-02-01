module knative.dev/eventing-redis

go 1.16

require (
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/emicklei/go-restful v2.15.0+incompatible // indirect
	github.com/go-redis/redis/v8 v8.11.4
	github.com/gomodule/redigo v1.8.3
	github.com/google/go-cmp v0.5.6
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
	knative.dev/eventing v0.29.1-0.20220128133147-7a44b07f299d
	knative.dev/hack v0.0.0-20220201013531-82bfca153560
	knative.dev/pkg v0.0.0-20220131144930-f4b57aef0006
	knative.dev/serving v0.29.1-0.20220131194631-770b91180d43
)
