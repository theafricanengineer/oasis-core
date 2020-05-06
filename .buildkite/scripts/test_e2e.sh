#!/bin/bash

############################################################
# This script tests the Oasis Core project.
#
# Usage:
# test_e2e.sh
############################################################

# Helpful tips on writing build scripts:
# https://buildkite.com/docs/pipelines/writing-build-scripts
set -euxo pipefail

# Working directory.
WORKDIR=$PWD

#################
# Run test suite.
#################
# Determine correct runtime to use for SGX.
runtime_target="default"
if [[ "${OASIS_TEE_HARDWARE:-""}" == "intel-sgx" ]]; then
    runtime_target="sgx/x86_64-fortanix-unknown-sgx"
fi

# We need a directory in the workdir so that Buildkite can fetch artifacts.
if [[ "${BUILDKITE:-""}" != "" ]]; then
    mkdir -p ${TEST_BASE_DIR:-$PWD}/e2e
fi

node_binary="${WORKDIR}/go/oasis-node/oasis-node"
test_runner_binary="${WORKDIR}/go/oasis-test-runner/oasis-test-runner"

# Use e2e-coverage-wrapper.sh as node binary if we need to compute E2E
# tests' coverage.
if [[ ${OASIS_E2E_COVERAGE:-""} != "" ]]; then
    test_runner_binary="${WORKDIR}/scripts/e2e-coverage-wrapper-arg.sh ${test_runner_binary}.test"

    # Use -env version of the wrapper as we can't pass additional arguments there.
    export E2E_COVERAGE_BINARY=${node_binary}.test
    node_binary="${WORKDIR}/scripts/e2e-coverage-wrapper-env.sh"
fi

# TODO:
# --e2e/runtime.ias.mock=false
# --e2e/runtime.ias.spid XXX
# --e2e/runtime.ias.cert_file "/code/ias-keys/ias.cert"
# --e2e/runtime.ias.key_file "/code/ias-keys/ias.key"
# Run Oasis test runner.
${test_runner_binary} \
    ${BUILDKITE:+--basedir ${TEST_BASE_DIR:-$PWD}/e2e} \
    --basedir.no_cleanup \
    --e2e.node.binary ${node_binary} \
    --e2e/runtime.client.binary_dir ${WORKDIR}/target/default/debug \
    --e2e/runtime.runtime.binary_dir ${WORKDIR}/target/${runtime_target}/debug \
    --e2e/runtime.runtime.loader ${WORKDIR}/target/default/debug/oasis-core-runtime-loader \
    --e2e/runtime.tee_hardware ${OASIS_TEE_HARDWARE:-""} \
    --remote-signer.binary ${WORKDIR}/go/oasis-remote-signer/oasis-remote-signer \
    --log.level info \
    ${BUILDKITE_PARALLEL_JOB_COUNT:+--parallel.job_count ${BUILDKITE_PARALLEL_JOB_COUNT}} \
    ${BUILDKITE_PARALLEL_JOB:+--parallel.job_index ${BUILDKITE_PARALLEL_JOB}} \
    "$@"

# Gather the coverage output.
if [[ "${BUILDKITE:-""}" != "" ]]; then
    if [[ ${OASIS_E2E_COVERAGE:-""} != "" ]]; then
        merged_file="coverage-merged-e2e-$BUILDKITE_PARALLEL_JOB.txt"
        if [[ "${OASIS_TEE_HARDWARE:-""}" == "intel-sgx" ]]; then
            merged_file="coverage-merged-e2e-sgx-$BUILDKITE_PARALLEL_JOB.txt"
        fi
        gocovmerge coverage-e2e-*.txt >"$merged_file"
    fi
fi
