#!/usr/bin/env bash

# Copyright 2021 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script runs the end-to-end tests against eventing-contrib built
# from source.

# If you already have the *_OVERRIDE environment variables set, call
# this script with the --run-tests arguments and it will use the cluster
# and run the tests.
# Note that local clusters often do not have the resources to run 12 parallel
# tests (the default) as the tests each tend to create their own namespaces and
# dispatchers.  For example, a local Docker cluster with 4 CPUs and 8 GB RAM will
# probably be able to handle 6 at maximum.  Be sure to adequately set the
# MAX_PARALLEL_TESTS variable before running this script, with the caveat that
# lowering it too much might make the tests run over the timeout that
# the go_test_e2e commands are using below.

# This script includes common functions for testing setup and teardown.

TEST_PARALLEL=${MAX_PARALLEL_TESTS:-12}

source $(dirname $0)/../vendor/knative.dev/hack/e2e-tests.sh

# If gcloud is not available make it a no-op, not an error.
which gcloud &> /dev/null || gcloud() { echo "[ignore-gcloud $*]" 1>&2; }

# Use GNU Tools on MacOS (Requires the 'grep' and 'gnu-sed' Homebrew formulae)
if [ "$(uname)" == "Darwin" ]; then
  sed=gsed
  grep=ggrep
fi

# Eventing main config path from HEAD.
readonly EVENTING_CONFIG="./config/"
readonly EVENTING_MT_CHANNEL_BROKER_CONFIG="./config/brokers/mt-channel-broker"
readonly EVENTING_IN_MEMORY_CHANNEL_CONFIG="./config/channels/in-memory-channel"

# Vendored Eventing Test Images.
readonly VENDOR_EVENTING_TEST_IMAGES="vendor/knative.dev/eventing/test/test_images/"
# HEAD eventing test images.
readonly HEAD_EVENTING_TEST_IMAGES="${GOPATH}/src/knative.dev/eventing/test/test_images/"

# Config tracing config.
readonly CONFIG_TRACING_CONFIG="test/config/config-tracing.yaml"

readonly REDEX_NAMESPACE="redex" # Installation Namespace
readonly REDIS_NAMESPACE="redis" # Local Redis instance Namespace
readonly REDIS_ADDRESS="rediss://redis.redis.svc.cluster.local:6379"

readonly REDEX_SOURCE_INSTALLATION_CONFIG_TEMPLATE="samples/source"
readonly REDEX_SOURCE_INSTALLATION_CONFIG="$(mktemp -d)"

readonly REDEX_SINK_INSTALLATION_CONFIG_TEMPLATE="samples/sink"
readonly REDEX_SINK_INSTALLATION_CONFIG="$(mktemp -d)"

readonly REDIS_INSTALLATION_CONFIG_TEMPLATE="samples/redis"
readonly REDIS_INSTALLATION_CONFIG="$(mktemp -d)"

# Real Redis Stream Source CRD config, generated from the template directory and modified template file.
readonly REDISSTREAM_SOURCE_CRD_CONFIG_DIR="$(mktemp -d)"
# Real Redis Stream Sink CRD config, generated from the template directory and modified template file.
readonly REDISSTREAM_SINK_CRD_CONFIG_DIR="$(mktemp -d)"

# Remove the temporary directories on exit (avoiding "rm -rf" to prevent disaster if something is wrong with the variables)
trap "{ for dirrm in \"${REDISSTREAM_SOURCE_CRD_CONFIG_DIR}\" \"${REDISSTREAM_SINK_CRD_CONFIG_DIR}\"; do rm \"\${dirrm}\"/*; rmdir \"\${dirrm}\"; done }" EXIT

# Redis Stream Source and Sink CRD config template directory
readonly REDISSTREAM_SOURCE_TEMPLATE_DIR="source/config"
readonly REDISSTREAM_SINK_TEMPLATE_DIR="sink/config"

# Namespaces where we install Eventing components
# This is the namespace of knative-eventing itself
export EVENTING_NAMESPACE="knative-eventing"

# Namespace where we install eventing-redis components (may be different than EVENTING_NAMESPACE)
readonly SYSTEM_NAMESPACE="knative-eventing"
export SOURCES_NAMESPACE="knative-sources"
export SYSTEM_NAMESPACE

# Zipkin setup
readonly KNATIVE_EVENTING_MONITORING_YAML="test/config/monitoring.yaml"

#
# TODO - Consider adding this function to the test-infra library.sh utilities ?
#
# Add The kn-eventing-test-pull-secret To Specified ServiceAccount & Restart Pods
#
# If the default namespace contains a Secret named 'kn-eventing-test-pull-secret',
# then copy it into the specified Namespace and add it to the specified ServiceAccount,
# and restart the specified Pods.
#
# This utility function exists to support local cluster testing with a private Docker
# repository, and is based on the CopySecret() functionality in eventing/pkg/utils.
#
function add_kn_eventing_test_pull_secret() {

  # Local Constants
  local secret="kn-eventing-test-pull-secret"

  # Get The Function Arguments
  local namespace="$1"
  local account="$2"
  local deployment="$3"

  # If The Eventing Test Pull Secret Is Present & The Namespace Was Specified
  if [[ $(kubectl get secret $secret -n default --ignore-not-found --no-headers=true | wc -l) -eq 1 && -n "$namespace" ]]; then

      # If The Secret Is Not Already In The Specified Namespace Then Copy It In
      if [[ $(kubectl get secret $secret -n "$namespace" --ignore-not-found --no-headers=true | wc -l) -lt 1 ]]; then
        kubectl get secret $secret -n default -o yaml | sed "s/namespace: default/namespace: $namespace/" | kubectl create -f -
      fi

      # If Specified Then Patch The ServiceAccount To Include The Image Pull Secret
      if [[ -n "$account" ]]; then
        kubectl patch serviceaccount -n "$namespace" "$account" -p "{\"imagePullSecrets\": [{\"name\": \"$secret\"}]}"
      fi

      # If Specified Then Restart The Pods Of The Deployment
      if [[ -n "$deployment" ]]; then
        kubectl rollout restart -n "$namespace" deployment "$deployment"
      fi
  fi
}

function knative_setup() {
  if is_release_branch; then
    echo ">> Install Knative Eventing from ${KNATIVE_EVENTING_RELEASE}"
    kubectl apply -f ${KNATIVE_EVENTING_RELEASE}
  else
    echo ">> Install Knative Eventing from HEAD"
    pushd .
    cd ${GOPATH} && mkdir -p src/knative.dev && cd src/knative.dev
    git clone https://github.com/knative/eventing
    cd eventing
    ko apply -f "${EVENTING_CONFIG}"
    # Install MT Channel Based Broker
    ko apply -f "${EVENTING_MT_CHANNEL_BROKER_CONFIG}"
    # Install IMC
    ko apply -f "${EVENTING_IN_MEMORY_CHANNEL_CONFIG}"
    popd
  fi
  wait_until_pods_running "${EVENTING_NAMESPACE}" || fail_test "Knative Eventing did not come up"

  install_zipkin
}

# Setup zipkin
function install_zipkin() {
  echo "Installing Zipkin..."
  sed "s/\${SYSTEM_NAMESPACE}/${SYSTEM_NAMESPACE}/g" < "${KNATIVE_EVENTING_MONITORING_YAML}" | kubectl apply -f -
  wait_until_pods_running "${SYSTEM_NAMESPACE}" || fail_test "Zipkin inside eventing did not come up"
  # Setup config tracing for tracing tests
  sed "s/\${SYSTEM_NAMESPACE}/${SYSTEM_NAMESPACE}/g" <  "${CONFIG_TRACING_CONFIG}" | kubectl apply -f -
}

# Remove zipkin
function uninstall_zipkin() {
  echo "Uninstalling Zipkin..."
  sed "s/\${SYSTEM_NAMESPACE}/${SYSTEM_NAMESPACE}/g" <  "${KNATIVE_EVENTING_MONITORING_YAML}" | kubectl delete -f -
  wait_until_object_does_not_exist deployment zipkin "${SYSTEM_NAMESPACE}" || fail_test "Zipkin deployment was unable to be deleted"
  kubectl delete -n "${SYSTEM_NAMESPACE}" configmap config-tracing
}

function knative_teardown() {
  echo ">> Stopping Knative Eventing"
  if is_release_branch; then
    echo ">> Uninstalling Knative Eventing from ${KNATIVE_EVENTING_RELEASE}"
    kubectl delete -f "${KNATIVE_EVENTING_RELEASE}"
  else
    echo ">> Uninstalling Knative Eventing from HEAD"
    pushd .
    cd ${GOPATH}/src/knative.dev/eventing
    # Remove IMC
    ko delete -f "${EVENTING_IN_MEMORY_CHANNEL_CONFIG}"
    # Remove MT Channel Based Broker
    ko delete -f "${EVENTING_MT_CHANNEL_BROKER_CONFIG}"
    # Remove eventing
    ko delete -f "${EVENTING_CONFIG}"
    popd
  fi
  wait_until_object_does_not_exist namespaces "${EVENTING_NAMESPACE}"
}

# Add function call to trap
# Parameters: $1 - Function to call
#             $2...$n - Signals for trap
function add_trap() {
  local cmd=$1
  shift
  for trap_signal in $@; do
    local current_trap="$(trap -p $trap_signal | cut -d\' -f2)"
    local new_cmd="($cmd)"
    [[ -n "${current_trap}" ]] && new_cmd="${current_trap};${new_cmd}"
    trap -- "${new_cmd}" $trap_signal
  done
}

function test_setup() {
  install_sources_crds || return 1
  install_sinks_crds || return 1

  # Install kail if needed.
  if ! which kail > /dev/null; then
    bash <( curl -sfL https://raw.githubusercontent.com/boz/kail/master/godownloader.sh) -b "$GOPATH/bin"
  fi

  # Capture all logs.
  kail > "${ARTIFACTS}/k8s.log.txt" &
  local kail_pid=$!
  # Clean up kail so it doesn't interfere with job shutting down
  add_trap "kill $kail_pid || true" EXIT

  # Publish test images.
  echo ">> Publishing test images from eventing"
  # We vendor test image code from eventing, in order to use ko to resolve them into Docker images, the
  # path has to be a GOPATH.  The two slashes at the beginning are to anchor the match so that running the test
  # twice doesn't re-parse the yaml and cause errors.
  sed -i '' 's@//knative.dev/eventing/test/test_images@//knative.dev/eventing-redis/vendor/knative.dev/eventing/test/test_images@g' "${VENDOR_EVENTING_TEST_IMAGES}"*/*.yaml
  $(dirname $0)/upload-test-images.sh "${VENDOR_EVENTING_TEST_IMAGES}" e2e || fail_test "Error uploading test images"
  $(dirname $0)/upload-test-images.sh "test/test_images" e2e || fail_test "Error uploading test images"
}

function test_teardown() {
  uninstall_sources_crds
  uninstall_sinks_crds
}

function install_sources_crds() {
  echo "Installing Redis Stream Source CRD"
  rm "${REDISSTREAM_SOURCE_CRD_CONFIG_DIR}/"*yaml
  cp "${REDISSTREAM_SOURCE_TEMPLATE_DIR}/"*yaml "${REDISSTREAM_SOURCE_CRD_CONFIG_DIR}"
  sed -i '' "s/namespace: knative-eventing/namespace: ${SOURCES_NAMESPACE}/g" "${REDISSTREAM_SOURCE_CRD_CONFIG_DIR}/"*yaml
  ko apply -f "${REDISSTREAM_SOURCE_CRD_CONFIG_DIR}" || return 1
  wait_until_pods_running "${SOURCES_NAMESPACE}" || fail_test "Failed to install the Redis Stream Source CRD"
}

function uninstall_sources_crds() {
  echo "Uninstalling Redis Stream Source CRD"
  ko delete --ignore-not-found=true --now --timeout 180s -f "${REDISSTREAM_SOURCE_CRD_CONFIG_DIR}"
}

function install_sinks_crds() {
  echo "Installing Redis Stream Sink CRD"
  rm "${REDISSTREAM_SINK_CRD_CONFIG_DIR}/"*yaml
  cp "${REDISSTREAM_SINK_TEMPLATE_DIR}/"*yaml "${REDISSTREAM_SINK_CRD_CONFIG_DIR}"
  sed -i '' "s/namespace: knative-eventing/namespace: ${SYSTEM_NAMESPACE}/g" "${REDISSTREAM_SINK_CRD_CONFIG_DIR}/"*yaml
  ko apply -f "${REDISSTREAM_SINK_CRD_CONFIG_DIR}" || return 1
  wait_until_pods_running "${EVENTING_NAMESPACE}" || fail_test "Failed to install the Redis Stream Sink CRD"
}

function uninstall_sinks_crds() {
  echo "Uninstalling Redis Stream Sink CRD"
  ko delete --ignore-not-found=true --now --timeout 180s -f "${REDISSTREAM_SINK_CRD_CONFIG_DIR}"
}

function redisstreamsource_setup() {
  # Create The Namespace Where Redis Stream Source Will Be Installed
  echo "Installing Redis Stream Source example and local Redis instance"
  kubectl create namespace ${REDEX_NAMESPACE}

  cp "${REDIS_INSTALLATION_CONFIG_TEMPLATE}/"*yaml "${REDIS_INSTALLATION_CONFIG}"
  sed -i '' "s/namespace: knative-eventing/namespace: ${SYSTEM_NAMESPACE}/g" "${REDIS_INSTALLATION_CONFIG}/"*yaml

  cp "${REDEX_SOURCE_INSTALLATION_CONFIG_TEMPLATE}/"*yaml "${REDEX_SOURCE_INSTALLATION_CONFIG}"
  sed -i '' "s/namespace: knative-eventing/namespace: ${SYSTEM_NAMESPACE}/g" "${REDEX_SOURCE_INSTALLATION_CONFIG}/"*yaml

  kubectl apply -f "${REDIS_INSTALLATION_CONFIG}" -n "${REDIS_NAMESPACE}"
  kubectl apply -f "${REDEX_SOURCE_INSTALLATION_CONFIG}" -n "${REDEX_NAMESPACE}"

  # Delay Pod Running Check Until All Pods Are Created To Prevent Race Condition
  local iterations=0
  local progress="Waiting for Redis Stream Source Pods to be created..."
  while [[ $(kubectl get pods --no-headers=true -n ${REDEX_NAMESPACE} | wc -l) -lt 2 && $iterations -lt 60 ]]
  do
    echo -ne "${progress}\r"
    progress="${progress}."
    iterations=$((iterations + 1))
    sleep 3
  done
  echo "${progress}"

  # Wait For The Pods To Be Ready (Forcing Delay To Ensure CRDs Are Installed To Prevent Race Condition)
  wait_until_pods_running "${REDEX_NAMESPACE}" || fail_test "Failed to start up a Redis Instance"
}

function redisstreamsource_teardown() {
  echo "Uninstalling Redis Stream Source"
  kubectl delete -f ${REDEX_SOURCE_INSTALLATION_CONFIG} -n "${REDEX_NAMESPACE}" --ignore-not-found
  kubectl delete -f "${REDIS_INSTALLATION_CONFIG}" -n "${REDIS_NAMESPACE}" --ignore-not-found
  kubectl delete namespace "${REDEX_NAMESPACE}" --ignore-not-found
}

function redisstreamsink_setup() {
  # Create The Namespace Where Redis Stream Source Will Be Installed
  echo "Installing Redis Stream Sink example and local Redis instance"
  kubectl create namespace ${REDEX_NAMESPACE}

  cp "${REDIS_INSTALLATION_CONFIG_TEMPLATE}/"*yaml "${REDIS_INSTALLATION_CONFIG}"
  sed -i '' "s/namespace: knative-eventing/namespace: ${SYSTEM_NAMESPACE}/g" "${REDIS_INSTALLATION_CONFIG}/"*yaml

  cp "${REDEX_SINK_INSTALLATION_CONFIG_TEMPLATE}/"*yaml "${REDEX_SINK_INSTALLATION_CONFIG}"
  sed -i '' "s/namespace: knative-eventing/namespace: ${SYSTEM_NAMESPACE}/g" "${REDEX_SINK_INSTALLATION_CONFIG}/"*yaml

  kubectl apply -f "${REDIS_INSTALLATION_CONFIG}" -n "${REDIS_NAMESPACE}"
  kubectl apply -f "${REDEX_SINK_INSTALLATION_CONFIG}" -n "${REDEX_NAMESPACE}"

  # Delay Pod Running Check Until All Pods Are Created To Prevent Race Condition
  local iterations=0
  local progress="Waiting for Redis Stream Sink Pods to be created..."
  while [[ $(kubectl get pods --no-headers=true -n ${REDEX_NAMESPACE} | wc -l) -lt 1 && $iterations -lt 60 ]]
  do
    echo -ne "${progress}\r"
    progress="${progress}."
    iterations=$((iterations + 1))
    sleep 3
  done
  echo "${progress}"

  # Wait For The Pods To Be Ready (Forcing Delay To Ensure CRDs Are Installed To Prevent Race Condition)
  wait_until_pods_running "${REDEX_NAMESPACE}" || fail_test "Failed to start up a Redis Instance"
}

function redisstreamsink_teardown() {
  echo "Uninstalling Redis Stream Sink"
  kubectl delete -f ${REDEX_SINK_INSTALLATION_CONFIG} -n "${REDEX_NAMESPACE}" --ignore-not-found
  kubectl delete -f "${REDIS_INSTALLATION_CONFIG}" -n "${REDIS_NAMESPACE}" --ignore-not-found
  kubectl delete namespace "${REDEX_NAMESPACE}" --ignore-not-found
}

# Installs the resources necessary to test the redis stream source, runs those tests, and then cleans up those resources
function test_redisstream_source() {
  echo "Testing the redis stream source"
  redisstreamsource_setup || return 1

  go_test_e2e -tags=e2e,source -timeout=40m -test.parallel=${TEST_PARALLEL} ./test/e2e -sources=sources.knative.dev/v1alpha1:RedisStreamSource  || fail_test
  go_test_e2e -tags=e2e,source -timeout=5m -test.parallel=${TEST_PARALLEL} ./test/conformance -sources=sources.knative.dev/v1alpha1:RedisStreamSource || fail_test

  redisstreamsource_teardown || return 1
}

function test_redisstream_sink() {
  echo "Testing the redis stream sink"
  redisstreamsink_setup || return 1

  # Add sink tests here, if any

  redisstreamsink_teardown || return 1
}

function test_redisstream_integration() {
  echo "Testing the redis sink + source integration test"
  redisstreamsink_setup || return 1
  redisstreamsource_setup || return 1

  sleep 3m
  echo "Sending a CE to sink"
  curl $(kubectl get ksvc redistreamsinkmystream -ojsonpath='{.status.url}' -n ${REDEX_NAMESPACE}) \
    -H "ce-specversion: 1.0" \
    -H "ce-type: dev.knative.sources.redisstream" \
    -H "ce-source: cli" \
    -H "ce-id: 1" \
    -H "datacontenttype: application/json" \
    -d '["fruit", "orange"]'

  sleep 3m

  echo "Confirm that Redis DB has the event (Sink works!)"
  local iterations=0
  local progress="Waiting for Redis CLI Xinfo info..."
  while [[ $(kubectl exec svc/redis -n ${REDIS_NAMESPACE}  redis-cli xinfo stream mystream | wc -l) -lt 18 && $iterations -lt 60 ]]
  do
    echo -ne "${progress}\r"
    progress="${progress}."
    iterations=$((iterations + 1))
    sleep 3
  done
  echo "${progress}"
  kubectl exec svc/redis -n "${REDIS_NAMESPACE}" redis-cli xinfo stream mystream 

  echo "Confirm the sink for the Redis Stream Source has received the event (Source works!)"
  local iterations=0
  local progress="Waiting for Event display log..."
  while [[ $(kubectl logs svc/event-display -n ${REDEX_NAMESPACE} | wc -l) -lt 14 && $iterations -lt 60 ]]
  do
    echo -ne "${progress}\r"
    progress="${progress}."
    iterations=$((iterations + 1))
    sleep 3
  done
  echo "${progress}"
  kubectl logs svc/event-display -n "${REDEX_NAMESPACE}"

  # TODO: Compare Redis-CLI text output with event-display logs to confirm data is same

  redisstreamsource_teardown || return 1
  redisstreamsink_teardown || return 1
}

function parse_flags() {
  # This function will be called repeatedly by initialize() with one fewer
  # argument each time and expects a return value of "the number of arguments to skip"
  # so we can just check the first argument and return 1 (to have it redirected to the
  # test container) or 0 (to have initialize() parse it normally).
  case $1 in
    --source)
      TEST_REDISSTREAMSOURCE=1
      return 1
      ;;
    --sink)
      TEST_REDISSTREAMSINK=1
      return 1
      ;;
    --integration)
      TEST_REDISSTREAMINTEGRATION=1
      return 1
      ;;
  esac
  return 0
}
