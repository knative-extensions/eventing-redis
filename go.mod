module knative.dev/eventing-redis

go 1.16

require (
	github.com/cloudevents/sdk-go/v2 v2.7.0
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
	knative.dev/eventing v0.28.1-0.20220105073332-4348c292392a
	knative.dev/hack v0.0.0-20211222071919-abd085fc43de
	knative.dev/pkg v0.0.0-20220104185830-52e42b760b54
	knative.dev/serving v0.28.1-0.20220104122631-278af32f24ce
)
