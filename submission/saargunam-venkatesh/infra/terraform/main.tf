terraform {
  required_version = ">= 1.8"

  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.27"
    }
  }
}

# The Kind cluster is created by 'make cluster' before Terraform runs.
# Terraform's scope: namespace + Bitnami Postgres + app Helm release.
provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-${var.cluster_name}"
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = "kind-${var.cluster_name}"
  }
}

# ─── Namespace ────────────────────────────────────────────────────────────────
resource "kubernetes_namespace" "this" {
  metadata {
    name = var.namespace
  }
}

# ─── Postgres (Bitnami) ───────────────────────────────────────────────────────
resource "helm_release" "postgres" {
  name       = "postgres"
  repository = "oci://registry-1.docker.io/bitnamicharts"
  chart      = "postgresql"
  version    = "16.3.0"
  namespace  = var.namespace

  wait    = true
  timeout = 300

  set {
    name  = "auth.database"
    value = var.db_name
  }
  set {
    name  = "auth.username"
    value = var.db_user
  }
  set {
    name  = "auth.password"
    value = var.db_password
  }
  # Minimal footprint for local dev
  set {
    name  = "primary.resources.requests.cpu"
    value = "100m"
  }
  set {
    name  = "primary.resources.requests.memory"
    value = "128Mi"
  }
  set {
    name  = "primary.resources.limits.memory"
    value = "256Mi"
  }

  depends_on = [kubernetes_namespace.this]
}

# ─── App ──────────────────────────────────────────────────────────────────────
resource "helm_release" "app" {
  name      = "config-service"
  chart     = "${path.module}/../../helm/config-service"
  namespace = var.namespace

  wait    = true
  timeout = 120

  set {
    name  = "db.url"
    value = "postgres://${var.db_user}:${var.db_password}@postgres-postgresql.${var.namespace}.svc.cluster.local:5432/${var.db_name}?sslmode=disable"
  }

  depends_on = [helm_release.postgres]
}
