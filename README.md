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

Install the example by running this command:

```sh
kubectl apply -n knative-sources -f samples/
```

Note: the example needs to be installed in the `knative-sources` namespace.

Verify the source is ready:

```sh
kubectl get  -n knative-sources redisstreamsources.sources.knative.dev mystream
NAME       AGE
mystream   13m
```

NOTE: there is no ready status yet

Add an item to `mystream`:

```sh
kubectl exec -n redis svc/redis redis-cli xadd mystream '*' fruit banana
```

Check the event display received the event:

```sh
kubectl logs -n knative-sources svc/event-display
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
