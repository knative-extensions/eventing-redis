# Redis Stream Source

The Redis Stream Event Source for Knative reads messages from a Redis Stream and
sends them as CloudEvents to the referenced Sink, which can be a Kubernetes service
or a Knative Serving service, etc. It is configured to retry sending of CloudEvents
so that events are not lost.

The Redis Stream Source can work with a local version of Redis database instance or
a cloud based instance whose [`address`](config/300-redisstreamsource.yaml) will be
specified in the Source spec. Additionally, the specified [`stream`](config/300-redisstreamsource.yaml)
name and consumer [`group`](config/300-redisstreamsource.yaml) name will be
created by the receive adapter, if they don't already exist.

The number of consumers in the consumer group can also be configured via data in
[`config-redis`](config/config-redis.yaml). This makes it possible for each consumer to
consume different messages arriving in the stream. Each consumer has an unique consumer
name which is a string created by the receive adapter.

When a Redis Stream Source resource is deleted, all the consumers in the group
are gracefully shutdown/deleted, before the consumer group itself is destroyed.
Before a consumer is shut down, all its pending messages are sent as CloudEvents and acknowledged.


## Getting started

### Install

#### Prerequisite for a cloud based Redis instance:

If you are using a local Redis instance, you can skip this step. If you are
using a cloud instance of Redis (for example, Redis DB on IBM Cloud), a TLS
certificate will need to be configured, prior to installing the event source.

Edit the [`config-tls`](config/config-tls.yaml) Config Map to add the TLS Certicate
from your cloud instance of Redis to the `cert.pem` data key:

```
vi source/config/config-tls.yaml
```

Add your certificate to the file, and save the file. Will be applied in the next step.

#### Create the `RedisStreamSource` source definition, and all of its components:

You can also, configure the receive adapter with the number of consumers in a group,
prior to installing the event source.

Edit the [`config-redis`](config/config-redis.yaml) Config Map to edit the `numConsumers` data key:

```
vi source/config/config-redis.yaml
```

Then, apply [`source/config`](../source/config)

```sh
ko apply -f source/config
```


### Example

In this example, you create one Redis Stream event source listening for items added to
the `mystream` stream. The items are then sent to the event display service as
CloudEvent events.

1. Install a local Redis by running this command:

```sh
kubectl apply -f samples/redis
```

2. Create a namespace for this example source:

```sh
kubectl create ns redex
```

3. Install a Redis Stream Source example resource by running this command:

Note: In addition to configuring your TLS Certificate, if you are using a cloud
instance of Redis DB, you will need to set the appropriate address in
[`redisstreamsource`](../samples/source/redisstreamsource.yaml) source yaml.
Here's an example connection string:

```
address: "rediss://$USERNAME:$PASSWORD@7f41ece8-ccb3-43df-b35b-7716e27b222e.b2b5a92ee2df47d58bad0fa448c15585.databases.appdomain.cloud:32086"
```

Then, apply [`samples/source`](../samples/source) which creates an event-display service and a Redis Stream Source resource

```sh
kubectl apply -n redex -f samples/source
```

4. Verify the Redis Stream Source is ready:

```sh
kubectl get  -n redex redisstreamsources.sources.knative.dev mystream
NAME       SINK                                            AGE   READY   REASON
mystream   http://event-display.redex.svc.cluster.local/   38s   True
```

5. Add an item to `mystream`:

```sh
kubectl exec -n redis svc/redis redis-cli xadd mystream '*' fruit banana color yellow
```

6. Check the event display sink to see if the event was received:

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

7. To cleanup, delete the Redis Stream Source example, and redex namespace:

```sh
kubectl delete -f samples/source
kubectl delete ns redex
```

## Reference

### Prerequisites

* A Redis installation. (Instructions to deploy a local Redis are above)

* An understanding of Redis Stream basics: https://redis.io/topics/streams-intro,
and some of the commands specific to Streams: https://redis.io/commands#stream.

### Resource fields

`RedisStreamSource` sources are Kubernetes objects. In addition to the standard Kubernetes
`apiVersion`, `kind`, and `metadata`, they have the following `spec` fields:

| Field       | Value       |
| ----------- | ----------- |
| `address`   | The Redis TCP address
| `stream`    | Name of the Redis stream
| `group`     | Name of the consumer group associated to this source. When left empty, a group is automatically created for this source and deleted when this source is deleted. {optional}
| `sink`      | A reference to an `Addressable` Kubernetes object that will resolve to a uri to use as the sink

{optional} These attributes are optional.

The source will provide output information about readiness or errors via the
`status` field on the object once it has been created in the cluster.

### Debugging tips

* You can check the Redis Stream Source resource's `status.condition` values to diagnose any issues by running either of commands below:

```
kubectl get redisstreamsource -n redex
kubectl describe redisstreamsource mystream -n redex
```

* You can also read the logs to check for issues with the receive adapter's deployment:

```
kubectl logs redissource-mystream-1234-0 -n redex
```

* You can also read the logs to check for issues with the source controller's
deployment:

```
kubectl logs redis-controller-manager-0  -n knative-sources
```

* KO install issues?

Reference: https://github.com/google/ko/issues/106

Try re-installing KO and setting `export GOROOT=$(go env GOROOT)`

### Configuration options

The [`config-observability`](config/config-observability.yaml) and [`config-logging`](config/config-logging.yaml)
ConfigMaps may be used to manage the logging and metrics configuration.
