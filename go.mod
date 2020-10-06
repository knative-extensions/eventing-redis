module knative.dev/eventing-redis

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/gomodule/redigo v1.7.0
	github.com/google/go-cmp v0.5.1
	github.com/kelseyhightower/envconfig v1.4.0
	go.uber.org/zap v1.15.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.16.1-0.20200806200629-e4bc346017b6
	knative.dev/pkg v0.0.0-20200916171541-6e0430fd94db
	knative.dev/reconciler-test v0.0.0-20201001063329-8a22ebf8dbfc
	knative.dev/serving v0.17.1
	knative.dev/test-infra v0.0.0-20200911201000-3f90e7c8f2fa
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6

replace k8s.io/apiserver => k8s.io/apiserver v0.17.6
