#!/usr/bin/env bash

set -e
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "Patching CRDs..."

yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.1.0"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumes.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsredisinstances.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.3"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumes.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpredisinstances.yaml
