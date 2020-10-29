module knative.dev/eventing-redis

go 1.15

require (
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/gomodule/redigo v1.8.2
	github.com/google/go-cmp v0.5.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.16.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.18.1-0.20201029043934-af98d5579864
	knative.dev/hack v0.0.0-20201028205534-fe80f1c8af68
	knative.dev/pkg v0.0.0-20201029122234-6d905b3f84a6
	knative.dev/serving v0.18.1-0.20201029100934-fbc2b6640b56
)

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/apiserver => k8s.io/apiserver v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)
