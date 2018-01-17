#!/bin/bash

set -ex

current_version=$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
HELM_MONITOR_VERSION=${HELM_MONITOR_VERSION:-$current_version}

dir=${HELM_PLUGIN_DIR:-"$(helm home)/plugins/helm-monitor"}
os=$(uname -s | tr '[:upper:]' '[:lower:]')
release_file="helm-monitor_${os}_${HELM_MONITOR_VERSION}.tgz"
url="https://github.com/ContainerSolutions/helm-monitor/releases/download/v${HELM_MONITOR_VERSION}/${release_file}"

mkdir -p $dir

if command -v wget
then
  wget -O ${dir}/${release_file} $url
elif command -v curl; then
  curl -o ${dir}/${release_file} $url
fi

tar xf ${dir}/${release_file} -C $dir

chmod +x ${dir}/helm-monitor

rm ${dir}/${release_file}
