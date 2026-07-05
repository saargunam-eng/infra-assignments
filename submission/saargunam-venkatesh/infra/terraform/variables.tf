variable "cluster_name" {
  description = "Name of the Kind cluster — must match the name used in 'make cluster'"
  type        = string
  default     = "config-service"
}

variable "namespace" {
  description = "Kubernetes namespace to deploy into"
  type        = string
  default     = "config-service"
}

variable "db_name" {
  description = "Postgres database name"
  type        = string
  default     = "configs"
}

variable "db_user" {
  description = "Postgres username"
  type        = string
  default     = "config_user"
}

variable "db_password" {
  description = "Postgres password — supply via TF_VAR_db_password or terraform.tfvars (gitignored)"
  type        = string
  sensitive   = true
}
