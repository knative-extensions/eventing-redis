# Knative Redis Source

Redis Event Source for Knative

## Getting started

1. If you are using a local Redis instance, you can skip this step. If you are
   using a cloud instance of Redis (for example, Redis on IBM Cloud), a TLS
   certificate will need to be configured, prior to installing the event source.

Edit the [`config-tls`](config/config-tls.yaml) Config Map to add TLS Certicate
from your cloud instance of Redis to the `cert.pem` data key:

```
vi source/config/config-tls.yaml
```

2. Install the source:

```sh
ko apply -f source/config
```

Note: In addition to configuring your TLS Certificate, if you are using a cloud
instance of Redis, you will need to set the appropriate address in
[`redisstreamsource`](../samples/source/redisstreamsource.yaml) source yaml.
Here's an example connection string:

```
address: "rediss://$USERNAME:$PASSWORD@7f41ece8-ccb3-43df-b35b-7716e27b222e.b2b5a92ee2df47d58bad0fa448c15585.databases.appdomain.cloud:32086"
```

## Example

In this example, you create one Redis Stream source listening for items added to
the `mystream` stream. The items are sent to the event display service as
CloudEvent.

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
kubectl apply -n redex -f samples/source
```

Verify the source is ready:

```sh
kubectl get  -n redex redisstreamsources.sources.knative.dev mystream
NAME       SINK                                            AGE   READY   REASON
mystream   http://event-display.redex.svc.cluster.local/   38s   True
```

Add an item to `mystream`:

```sh
kubectl exec -n redis svc/redis redis-cli xadd mystream '*' fruit banana color yellow
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
    "fruit",
    "banana"
    "color",
    "yellow"
  ]
```

The data contains the list of field-value pairs added to the stream.

To cleanup, delete the redex namespace:

```sh
kubectl delete ns redex
```

## Release Notes

- Consume stream via a consumer group (internal change) (08/24/2020)
- RedisStreamSource can now be deployment in any namespace (08/21/2020)
- The redis address can now be specified in RedisStreamSource (08/21/2020)
