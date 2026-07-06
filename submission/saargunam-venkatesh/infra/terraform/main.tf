locals {
  # Constructed here so the sensitive value is not repeated across set{} blocks.
  db_url = "postgres://${var.db_user}:${var.db_password}@postgres-postgresql.${var.namespace}.svc.cluster.local:5432/${var.db_name}?sslmode=disable"
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
  set_sensitive {
    name  = "auth.password"
    value = var.db_password
  }

  # Pin to a locally-available image tag.
  # Chart v16.3.0 pins 17.2.0-debian-12-r2 which was removed from Docker Hub.
  # Pre-load with: kind load docker-image bitnami/postgresql:latest --name config-service
  set {
    name  = "image.tag"
    value = "latest"
  }
  set {
    name  = "image.pullPolicy"
    value = "IfNotPresent"
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

  set_sensitive {
    name  = "db.url"
    value = local.db_url
  }

  depends_on = [helm_release.postgres]
}
