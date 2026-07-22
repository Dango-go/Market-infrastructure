# Embedded Market Infrastructure

Infrastructure and GitOps repository for **Embedded Market**, a Kubernetes-based marketplace platform composed of Go microservices, a frontend application, managed platform services, and observability tooling.

The repository keeps application runtime manifests, platform add-ons, database definitions, and cloud infrastructure code in one place so the cluster can be provisioned with Terraform and continuously reconciled through Argo CD.

## Overview

Embedded Market is organized as a microservice application running on Kubernetes. The application layer includes services for authentication, catalog, cart, inventory, pricing, orders, payment, shipping, wishlist, reviews, search, analytics, notifications, recommendations, users, and an API gateway.

The operational model is GitOps-first:

- **Terraform** provisions cloud infrastructure such as the VPC, GKE cluster, and IAM roles.
- **Argo CD** owns continuous delivery from this repository into the cluster.
- **Helm and Kubernetes manifests** define application services, platform dependencies, databases, ingress, certificates, Kafka, and monitoring.
- **Prometheus stack** provides metrics collection, alerting primitives, and Grafana integration.

## Repository Layout

```text
.
├── app/                    # Application source code, Dockerfiles, frontend, backend services
├── gitops-manifests/       # Argo CD, platform, database, monitoring, and service manifests
├── infrastructure-live/    # Terraform live configuration and reusable modules
├── scripts/                # Helper scripts for applying and deleting deployments
├── docker-compose.yml      # Local development support
└── README.md
```

## Infrastructure

Terraform configuration lives in `infrastructure-live/` and is split into small modules:

- `modules/vpc` creates the network and subnet used by the cluster.
- `modules/gke` provisions the GKE Kubernetes cluster.
- `modules/iam-roles` manages IAM integration for GitHub-based workflows.

This layer is responsible for the cloud foundation. Kubernetes workloads and platform components are managed separately through Argo CD.

## GitOps

Argo CD application definitions are stored in `gitops-manifests/apps/`. They connect cluster namespaces to the desired state in this repository:

- `argo.yaml` bootstraps Argo-related resources.
- `application.yaml` deploys the application service manifests.
- `databases.yaml` deploys database resources.
- `ingress.yaml` deploys ingress infrastructure.
- `monitoring.yaml` deploys observability components.

The rest of `gitops-manifests/` is grouped by operational concern:

```text
gitops-manifests/
├── argo/          # Argo CD values, ingress, repository secret, and project resources
├── apps/          # Argo CD Application manifests
├── database/      # PostgreSQL and database-related manifests
├── monitoring/    # Prometheus stack, Loki, and long-term monitoring components
├── platform/      # Kafka, ingress, cert-manager, Karpenter, Falco, and other platform add-ons
└── services/      # Application services, topics, secrets, and Helm chart templates
```

## Platform Components

### PostgreSQL

PostgreSQL manifests are located under `gitops-manifests/database/postgres/`. They define database clusters, users, databases, connection poolers, service accounts, and required secrets for application services.

### Kafka

Kafka platform resources live in `gitops-manifests/platform/kafka/`. The repository includes Helm values, Kafka cluster resources, Kafka users, and topic definitions used by the event-driven parts of the system.

Application services also include outbox repositories and event relay code, making Kafka part of the service integration model rather than a standalone infrastructure detail.

### Helm

The main application chart is stored in `gitops-manifests/services/embedded-market-services/`. It contains templates for deployments, services, secrets, service monitors, and Kafka topics.

Helm values are also used for platform components such as Argo CD, ingress-nginx, Kafka, and the Prometheus stack.

### Ingress

Ingress resources are managed in `gitops-manifests/platform/ingress/`. This area contains ingress-nginx Helm values, application ingress, Grafana ingress, monitoring rules, and service monitor resources.

### cert-manager

Certificate management lives in `gitops-manifests/platform/cert-manager/`. It contains the cluster issuer and certificates for the main application endpoint, Argo CD, and Grafana.

### Prometheus Stack

Monitoring is configured under `gitops-manifests/monitoring/prometheus-stack/` and through service monitors in the service and ingress manifests.

The stack is intended to provide:

- Kubernetes and workload metrics collection.
- Service-level scraping through `ServiceMonitor` resources.
- Grafana access through managed ingress.
- Alerting and rule integration for platform components.

### Loki

Loki is planned under `gitops-manifests/monitoring/loki/` and is intended to provide centralized log aggregation for application and platform workloads.

### Karpenter

Karpenter is planned under `gitops-manifests/platform/karpenter/` and is intended to manage dynamic Kubernetes node provisioning and improve workload scheduling efficiency.

### Falco

Falco is planned under `gitops-manifests/platform/falco/` and is intended to provide runtime security detection for suspicious container and host-level behavior.

## Application Services

Backend services are written in Go and live under `app/backend/services/`. Each service follows a similar structure:

- `cmd/` for the executable entrypoint.
- `internal/config` for configuration loading.
- `internal/domain` for domain models and ports.
- `internal/application` for use cases.
- `internal/repository/postgres` for PostgreSQL persistence.
- `internal/transport/http` for HTTP routing and handlers.
- `internal/infrastructure` for system wiring and event relays.
- `db/` for SQL migrations, SQLC queries, and embedded database assets.

The frontend application lives in `app/frontend/` and is built with Vite.

## Operational Flow

1. Provision cloud infrastructure with Terraform from `infrastructure-live/`.
2. Install or bootstrap Argo CD in the target cluster.
3. Apply the Argo CD applications from `gitops-manifests/apps/`.
4. Let Argo CD reconcile platform components, databases, monitoring, and application services.
5. Use Prometheus, Grafana, and planned Loki/Falco integrations for observability and runtime visibility.

## Helper Scripts

The `scripts/` directory contains small operational helpers:

- `apply-backend.sh` applies backend manifests.
- `delete-deployments.sh` removes deployments when a cleanup cycle is needed.

Component-specific `Makefile`s are also present in several GitOps directories, including Kafka, PostgreSQL, ingress, and kubeseal.

## Status

This repository currently contains the core infrastructure and GitOps layout for the Embedded Market platform. Karpenter, Falco, and Loki directories are already reserved in the structure and are expected to be completed and integrated as the platform matures.
