#!/usr/bin/env bash
set -euo pipefail

cd /home/bodya/Projects/embedded-market-infrastructure/gitops-manifests/services/backend 

# kubectl create ns application
kubectl apply -f /home/bodya/Projects/embedded-market-infrastructure/gitops-manifests/services/backend/secrets/jwt-pub.yaml
kubectl apply -f /home/bodya/Projects/embedded-market-infrastructure/gitops-manifests/services/backend/secrets/jwt.yaml


ROOT_DIR="$(pwd)"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required but not installed" >&2
  exit 1
fi

for manifest in "$ROOT_DIR"/*.yaml; do
  [ -e "$manifest" ] || continue
  echo "Applying $(basename "$manifest")"
  kubectl apply -f "$manifest"
done

# bash /home/bodya/Projects/embedded-market-infrastructure/scripts/apply-backend.sh