#!/bin/bash
# This is a simple wrapper script around the cockroach binary.
# It must be in the same directory as the cockroach binary, and invoked in one of two ways:
#
# To start the first node and initialize the cluster:
# ./launch init [load balancer address:port]
# To start a node:
# ./launch start [load balancer address:port]
set -ex

DATA_DIR="data"
LOG_DIR="logs"
STORES="ssd=${DATA_DIR}"
COMMON_FLAGS="--log-dir=${LOG_DIR} --logtostderr=false --stores=${STORES}"
START_FLAGS="--insecure"

action=$1
if [ "${action}" != "init" -a "${action}" != "start" ]; then
	echo "Usage: ${0} [init|start] [load balancer address:port]"
	exit 1
fi

gossip=$2
if [ -z "${gossip}" ]; then
	echo "Usage: ${0} [init|start] [gossip flag]"
	exit 1
fi

chmod 755 cockroach
mkdir -p ${DATA_DIR} ${LOG_DIR}

if [ "${action}" == "init" ]; then
	./cockroach init ${COMMON_FLAGS}
fi

cmd="./cockroach start ${COMMON_FLAGS} ${START_FLAGS} --gossip=${gossip}"
nohup ${cmd} > ${LOG_DIR}/cockroach.STDOUT 2> ${LOG_DIR}/cockroach.STDERR < /dev/null &
pid=$!
echo "Launched ${cmd}: pid=${pid}"
# Sleep a bit to let the process start before we terminate the ssh connection.
sleep 3
