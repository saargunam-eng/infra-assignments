output "port_forward_cmd" {
  description = "Command to expose the app locally after deployment"
  value       = "kubectl port-forward -n ${var.namespace} svc/config-service 8080:8080"
}

output "ping_url" {
  description = "Health check URL (after port-forward)"
  value       = "curl http://localhost:8080/ping"
}
