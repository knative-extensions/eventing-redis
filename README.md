# Knative Redis Source

Redis Event Source for Knative

## Getting started

Install the source:

```sh
ko apply -f source/config
```

The source controller assumes Redis is installed in the `redis` namespace.
Install Redis by running this command:

```sh
kubectl apply -f samples/redis
```

## Example

In this example, you create one Redis Stream source listening for items added to
the `mystream` stream. The items are sent to the event display service as
CloudEvent.

Create a namespace:

```sh
kubectl create ns redex
```

Install the example by running this command:

```sh
kubectl apply -n redex -f samples/
```

Verify the source is ready:

```sh
kubectl get  -n redex redisstreamsources.sources.knative.dev mystream
NAME       SINK                                            AGE   READY   REASON
mystream   http://event-display.redex.svc.cluster.local/   38s   True
```

Add an item to `mystream`:

```sh
kubectl exec -n redis svc/redis redis-cli xadd mystream '*' fruit banana
```

Check the event display received the event:

```sh
kubectl logs -n redex svc/event-display
☁️  cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 1.0
  type: dev.knative.sources.redisstream
  source: /mystream
  id: 1597775814718-0
  time: 2020-08-18T18:36:54.719802342Z
  datacontenttype: application/json
Data,
  [
    "ZnJ1aXQ=",
    "YmFuYW5h"
  ]
```

The data contains the raw binary stream item.

To cleanup, delete the redex namespace:

```sh
kubectl delete ns redex
```

## Release Notes

- Consume stream via a consumer group (internal change) (08/24/2020)
- RedisStreamSource can now be deployment in any namespace (08/21/2020)
- The redis address can now be specified in RedisStreamSource (08/21/2020)
