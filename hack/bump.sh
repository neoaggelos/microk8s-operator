#!/bin/bash

IMG=`cat ./config/manager/kustomization.yaml | grep newTag | cut -f2 -d:`

# 0.0.1-dev12 --> BASE=0.0.1, NUMBER=12
BASE="${IMG%-dev*}"
NUMBER="${IMG#*-dev}"

NEW="${BASE}-dev$(( $NUMBER + 1 ))"

echo "Bump version to $NEW"

sed -i "s,$IMG,$NEW," ./config/manager/kustomization.yaml

./hack/update.sh
