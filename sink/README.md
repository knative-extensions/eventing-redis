# Redis Stream Sink

The Redis Stream Event Sink for Knative receives CloudEvents and adds them to
the specified [`stream`](config/300-redisstreamsink.yaml) of the Redis instance.

The Redis Stream Sink can work with a local version of Redis database instance
or a cloud based instance whose [`address`](config/300-redisstreamsink.yaml)
will be specified in the Sink spec. Additionally, the specified
[`stream`](config/300-redisstreamsink.yaml) name will be created by the
receiver, if they don't already exist.

## Getting started

### Install

#### Prerequisites

- Knative Serving (Install instructions here:
  https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-serving-component)

- If you are using a local Redis instance, you can skip this step. If you are
  using a cloud instance of Redis (for example, Redis DB on IBM Cloud), a TLS
  certificate will need to be configured, prior to installing the event sink.

  Edit the [`config-tls`](config/config-tls.yaml) Config Map to add the TLS Certicate
  from your cloud instance of Redis to the `cert.pem` data key:

  ```
  vi sink/config/config-tls.yaml
  ```

  Add your certificate to the file, and save the file. Will be applied in the next step.

#### Create the `RedisStreamSink` sink definition, and all of its components:

Apply [`sink/config`](../sink/config)

```sh
ko apply -f sink/config
```

### Example

In this example, you create one Redis Stream event sink that will receive
CloudEvent events and then add those items into the `mystream` stream.

1. Install a local Redis by running this command:

```sh
kubectl apply -f samples/redis
```

2. Create a namespace for this example sink:

```sh
kubectl create ns redex
```

3. Install a Redis Stream Sink example by running this command:

Note: In addition to configuring your TLS Certificate, if you are using a cloud
instance of Redis DB, you will need to set the appropriate address in
[`redisstreamsink`](../samples/sink/redisstreamsink.yaml) sink yaml. Here's an
example connection string:

```
address: "rediss://$USERNAME:$PASSWORD@7f41ece8-ccb3-43df-b35b-7716e27b222e.b2b5a92ee2df47d58bad0fa448c15585.databases.appdomain.cloud:32086"
```

Then, apply [`samples/sink`](../samples/sink) which creates a Redis Stream Sink
resource

```sh
kubectl apply -n redex -f samples/sink
```

4. Verify the Redis Stream Sink is ready:

```sh
kubectl get -n redex redisstreamsink.sinks.knative.dev mystream
NAME       URL                                                     AGE   READY   REASON
mystream   http://redistreamsinkmystream.redex.svc.cluster.local   35s   True
```

5. Send an event to the sink:

```sh
curl $(kubectl get ksvc redistreamsinkmystream -ojsonpath='{.status.url}' -n redex) \
 -H "ce-specversion: 1.0" \
 -H "ce-type: dev.knative.sources.redisstream" \
 -H "ce-source: cli" \
 -H "ce-id: 1" \
 -H "datacontenttype: application/json" \
 -d '["fruit", "orange"]'
```

6. Check a new message has been added to redis:

```sh
kubectl exec -n redis svc/redis redis-cli xinfo stream mystream
...
last-entry
1598652372717-0
fruit
orange
```

7. To cleanup, delete the Redis Stream Sink example, and redex namespace:

```sh
kubectl delete -f samples/sink
kubectl delete ns redex
```

## Reference

### Prerequisites

- A Redis installation. Instructions to deploy a local Redis are above.

- An understanding of Redis Stream basics:
  https://redis.io/topics/streams-intro, and some of the commands specific to
  Streams: https://redis.io/commands#stream

### Resource fields

`RedisStreamSink` resources are Kubernetes objects. In addition to the standard
Kubernetes `apiVersion`, `kind`, and `metadata`, they have the following `spec`
fields:

| Field     | Value                    |
| --------- | ------------------------ |
| `address` | The Redis TCP address    |
| `stream`  | Name of the Redis stream |

The sink will provide output information about readiness or errors via the
`status` field on the object once it has been created in the cluster.

### Debugging tips

- You can check the Redis Stream Sink resource's `status.condition` values to
  diagnose any issues by running either of commands below:

```
kubectl get redisstreamsinks -n redex
kubectl describe redisstreamsinks mystream -n redex
```

- You can also read the logs to check for issues with the receiver's deployment:

```
kubectl get pods -n redex
kubectl logs {podname} -n redex
```

- You can also read the logs to check for issues with the sink controller's
  deployment:

```
kubectl logs redis-controller-manager-0  -n knative-sinks
```

- KO install issues?

Reference: https://github.com/google/ko/issues/106

Try re-installing KO and setting `export GOROOT=$(go env GOROOT)`

### Configuration options

The [`config-observability`](config/config-observability.yaml) and
[`config-logging`](config/config-logging.yaml) ConfigMaps may be used to manage
the logging and metrics configuration.
