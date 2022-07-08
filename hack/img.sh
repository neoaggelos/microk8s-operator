#!/bin/bash

export IMG=`cat deploy/deploy.yaml | grep "image:" | cut -f2,3 -d:`

make docker-build
