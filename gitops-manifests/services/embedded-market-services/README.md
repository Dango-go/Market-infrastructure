# embedded-market-services

Helm chart for the 17 application deployments currently stored under `gitops-manifests/services`.

## Render

```sh
helm template embedded-market-services gitops-manifests/services/embedded-market-services --namespace application
```

## Install or upgrade

```sh
helm upgrade --install embedded-market-services gitops-manifests/services/embedded-market-services --namespace application --create-namespace
```

## Secrets

JWT secrets are intentionally not stored in `values.yaml`. By default, the chart references the existing `jwt-token` and `jwt-token-private` secrets. If you want Helm to create them, set `jwtSecrets[].create: true` and provide `stringData` from a secure values file that is not committed.

## Notes

The frontend deployment is included, but its Service remains disabled by default because the source `frontend.yaml` did not define a Service. Kafka topic names that are invalid as Kubernetes resource names, such as `cart.checked_out`, are rendered with a valid `metadata.name` and the original topic in `spec.topicName`.
