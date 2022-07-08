#!/bin/bash

make manifests
./bin/kustomize build config/default > deploy/deploy.yaml
