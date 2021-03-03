#!/bin/bash

set -e

DELETED_NAMESPACE="scheduled-scaler-operator-system"
kubectl get namespace $DELETED_NAMESPACE -o json > $DELETED_NAMESPACE.json
sed -i -e 's/"kubernetes"//' $DELETED_NAMESPACE.json
kubectl replace --raw "/api/v1/namespaces/$DELETED_NAMESPACE/finalize" -f ./$DELETED_NAMESPACE.json
rm ./$DELETED_NAMESPACE.json

exit 0