# ─── Variables ────────────────────────────────────────────────────────────────
CLUSTER_NAME  := config-service
IMAGE_NAME    := config-service
IMAGE_TAG     := local
TF_DIR        := infra/terraform
HELM_DIR      := helm/config-service
NAMESPACE     := config-service
APP_PORT      := 8080

.PHONY: help up down \
        cluster cluster-down \
        build load \
        deploy tf-down \
        wait validate \
        test lint tf-validate helm-lint

# ─── Default target ───────────────────────────────────────────────────────────
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ─── Primary entrypoints ──────────────────────────────────────────────────────
up: cluster build load deploy wait ## Full setup — cluster, image, deploy, wait
	@echo "✅ Stack is up. Run 'make validate' to smoke test."

down: tf-down cluster-down ## Tear down everything
	@echo "✅ Stack torn down."

# ─── Cluster ──────────────────────────────────────────────────────────────────
cluster: ## Create Kind cluster (idempotent)
	@kind get clusters | grep -q "^$(CLUSTER_NAME)$$" \
	  && echo "⏭  Cluster '$(CLUSTER_NAME)' already exists, skipping." \
	  || kind create cluster --name $(CLUSTER_NAME)

cluster-down: ## Delete Kind cluster
	@kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true

# ─── Image ────────────────────────────────────────────────────────────────────
build: ## Build Docker image
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

load: ## Load image into Kind cluster
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name $(CLUSTER_NAME)

# ─── Infra / Deploy ───────────────────────────────────────────────────────────
deploy: ## Provision Postgres + deploy app via Terraform
	cd $(TF_DIR) && terraform init -input=false
	cd $(TF_DIR) && terraform apply -auto-approve -input=false

tf-down: ## Destroy Terraform-managed resources
	cd $(TF_DIR) && terraform destroy -auto-approve -input=false 2>/dev/null || true

# ─── Rollout ──────────────────────────────────────────────────────────────────
wait: ## Wait for app and Postgres to be ready
	kubectl rollout status deployment/config-service -n $(NAMESPACE) --timeout=120s
	@echo "✅ Deployment ready."

# ─── Smoke test ───────────────────────────────────────────────────────────────
validate: ## Run end-to-end smoke test (port-forward + curl)
	@bash scripts/smoke-test.sh

# ─── Development ──────────────────────────────────────────────────────────────
test: ## Run Go unit tests with race detector
	go test ./... -race -count=1

lint: ## Run gofmt and go vet
	@echo "→ gofmt"
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	@echo "→ go vet"
	go vet ./...
	@echo "✅ Lint passed."

tf-validate: ## Run terraform fmt check and validate
	cd $(TF_DIR) && terraform fmt -check -recursive
	cd $(TF_DIR) && terraform init -backend=false -input=false
	cd $(TF_DIR) && terraform validate

helm-lint: ## Run helm lint on the app chart
	helm lint $(HELM_DIR)
