#!/bin/bash

set -e

current_version=$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
HELM_MONITOR_VERSION=${HELM_MONITOR_VERSION:-$current_version}

file="${HELM_PLUGIN_DIR:-"$(helm home)/plugins/helm-monitor"}/helm-monitor"
os=$(uname -s | tr '[:upper:]' '[:lower:]')
url="https://github.com/ContainerSolutions/helm-monitor/releases/download/v${HELM_MONITOR_VERSION}/helm-monitor_${os}_amd64"

if command -v wget
then
  wget -O "${file}" "${url}"
elif command -v curl; then
  curl -o "${file}" "${url}"
fi

chmod +x "${file}"
