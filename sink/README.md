# Knative Redis Sink

Redis Event Sink for Knative

## Prerequisites

- Knative Serving

## Getting started

Install the sink:

```sh
ko apply -f sink/config
```

## Example

In this example, you create one Redis Stream sink for adding items into
the `mystream` stream.

Install Redis by running this command:

```sh
kubectl apply -f samples/redis
```

Create a namespace:

```sh
kubectl create ns redex
```

Install the example by running this command:

```sh
kubectl apply -n redex -f samples/sink
```

Verify the sink is ready:

```sh
kubectl get -n redex redisstreamsink.sinks.knative.dev mystream
NAME       URL                                                     AGE   READY   REASON
mystream   http://redistreamsinkmystream.redex.svc.cluster.local   35s   True
```

Send an event to the sink:

```sh
curl $(kubectl get ksvc redistreamsinkmystream -ojsonpath='{.status.url}') \
 -H "ce-specversion: 1.0" \
 -H "ce-type: dev.knative.sources.redisstream" \
 -H "ce-source: cli" \
 -H "ce-id: 1" \
 -H "datacontenttype: application/json" \
 -d '["fruit", "orange"]'
```

Check a new message has been added to redis:

```sh
kubectl exec -n redis svc/redis redis-cli xinfo stream mystream
...
last-entry
1598652372717-0
fruit
orange
```

To cleanup, delete the redex namespace:

```sh
kubectl delete ns redex
```

