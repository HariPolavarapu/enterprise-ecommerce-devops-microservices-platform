# Enterprise Platform Repository Reorganization Script
$ErrorActionPreference = "Stop"
$Root = "C:\Users\phari\OneDrive\Desktop\Devops\Projects\enterprise-ecommerce-devops-microservices-platform"
Set-Location $Root

function Ensure-Dir($path) {
    if (-not (Test-Path $path)) { New-Item -ItemType Directory -Path $path -Force | Out-Null }
}

function Move-ItemSafe($src, $dest) {
    if (Test-Path $src) {
        Ensure-Dir (Split-Path $dest -Parent)
        if (Test-Path $dest) { Remove-Item -Recurse -Force $dest }
        git mv $src $dest 2>$null
        if (-not $?) { Move-Item -Force $src $dest }
    }
}

Write-Host "=== Creating target directory structure ==="

$dirs = @(
    "infrastructure/terraform/backend",
    "infrastructure/terraform/modules/organizations",
    "infrastructure/terraform/modules/networking/vpc",
    "infrastructure/terraform/modules/networking/subnets",
    "infrastructure/terraform/modules/iam",
    "infrastructure/terraform/modules/eks",
    "infrastructure/terraform/modules/ec2",
    "infrastructure/terraform/modules/ecr",
    "infrastructure/terraform/modules/rds",
    "infrastructure/terraform/modules/redis",
    "infrastructure/terraform/modules/msk",
    "infrastructure/terraform/modules/opensearch",
    "infrastructure/terraform/modules/s3",
    "infrastructure/terraform/modules/api-gateway",
    "infrastructure/terraform/modules/alb",
    "infrastructure/terraform/modules/route53",
    "infrastructure/terraform/modules/cloudfront",
    "infrastructure/terraform/modules/waf",
    "infrastructure/terraform/modules/shield",
    "infrastructure/terraform/modules/lambda",
    "infrastructure/terraform/modules/eventbridge",
    "infrastructure/terraform/modules/sns",
    "infrastructure/terraform/modules/sqs",
    "infrastructure/terraform/modules/cloudwatch",
    "infrastructure/terraform/modules/backup",
    "infrastructure/terraform/accounts/management",
    "infrastructure/terraform/accounts/security",
    "infrastructure/terraform/accounts/log-archive",
    "infrastructure/terraform/accounts/networking",
    "infrastructure/terraform/accounts/shared-services",
    "infrastructure/terraform/accounts/development",
    "infrastructure/terraform/accounts/qa",
    "infrastructure/terraform/accounts/production",
    "infrastructure/ansible/inventory",
    "infrastructure/ansible/shared-services/jenkins",
    "infrastructure/ansible/shared-services/sonarqube",
    "infrastructure/ansible/shared-services/nexus",
    "infrastructure/ansible/shared-services/vault",
    "infrastructure/ansible/shared-services/ansible-controller",
    "infrastructure/ansible/shared-services/bastion",
    "infrastructure/ansible/playbooks",
    "kubernetes/platform-cluster/argocd",
    "kubernetes/platform-cluster/prometheus",
    "kubernetes/platform-cluster/grafana",
    "kubernetes/platform-cluster/elasticsearch",
    "kubernetes/platform-cluster/kibana",
    "kubernetes/platform-cluster/jaeger",
    "kubernetes/platform-cluster/opentelemetry-collector",
    "kubernetes/platform-cluster/fluent-bit",
    "kubernetes/platform-cluster/cert-manager",
    "kubernetes/platform-cluster/external-secrets",
    "kubernetes/platform-cluster/metrics-server",
    "kubernetes/platform-cluster/aws-load-balancer-controller",
    "kubernetes/platform-cluster/external-dns",
    "kubernetes/platform-cluster/istio",
    "kubernetes/application-clusters/base",
    "kubernetes/application-clusters/development",
    "kubernetes/application-clusters/qa",
    "kubernetes/application-clusters/production",
    "microservices/shared/enterprise",
    "gitops/platform/argocd",
    "gitops/applications/base",
    "gitops/applications/overlays/development",
    "gitops/applications/overlays/qa",
    "gitops/applications/overlays/production",
    "gitops/helm",
    "cicd/jenkins/agents",
    "cicd/github-actions",
    "security/vault",
    "security/external-secrets-operator",
    "security/cert-manager",
    "security/iam",
    "security/rbac",
    "security/network-policies",
    "observability/prometheus",
    "observability/grafana",
    "observability/elasticsearch",
    "observability/kibana",
    "observability/jaeger",
    "observability/opentelemetry",
    "observability/fluent-bit",
    "observability/cloudwatch",
    "automation/python",
    "automation/bash",
    "automation/lambda",
    "local-development/docker-compose",
    "docs/architecture",
    "docs/runbooks",
    "docs/releasing",
    "tools/load-generator",
    "legacy/gcp/terraform",
    "legacy/gcp/deploystack",
    "legacy/gcp/ci-terraform",
    "legacy/gcp/components"
)

foreach ($d in $dirs) { Ensure-Dir $d }

Write-Host "=== Moving Terraform infrastructure ==="
Move-ItemSafe "terraform/backend" "infrastructure/terraform/backend"
Move-ItemSafe "terraform/modules/organizations" "infrastructure/terraform/modules/organizations"
Move-ItemSafe "terraform/modules/vpc" "infrastructure/terraform/modules/networking/vpc"
Move-ItemSafe "terraform/modules/subnets" "infrastructure/terraform/modules/networking/subnets"
Move-ItemSafe "terraform/modules/iam" "infrastructure/terraform/modules/iam"
Move-ItemSafe "terraform/modules/eks" "infrastructure/terraform/modules/eks"
Move-ItemSafe "terraform/modules/rds" "infrastructure/terraform/modules/rds"
Move-ItemSafe "terraform/modules/rds-postgresql" "infrastructure/terraform/modules/rds-postgresql"
Move-ItemSafe "terraform/modules/redis" "infrastructure/terraform/modules/redis"
Move-ItemSafe "terraform/environments/management" "infrastructure/terraform/accounts/management"
Move-ItemSafe "terraform/environments/networking" "infrastructure/terraform/accounts/networking"
Move-ItemSafe "terraform/environments/shared-services" "infrastructure/terraform/accounts/shared-services"
Move-ItemSafe "terraform/environments/development" "infrastructure/terraform/accounts/development"
Move-ItemSafe "terraform/environments/qa" "infrastructure/terraform/accounts/qa"
Move-ItemSafe "terraform/environments/production" "infrastructure/terraform/accounts/production"
Move-ItemSafe "terraform/providers.tf" "infrastructure/terraform/providers.tf"
Move-ItemSafe "terraform/versions.tf" "infrastructure/terraform/versions.tf"
Move-ItemSafe "terraform/variables.tf" "infrastructure/terraform/variables.tf"
Move-ItemSafe "terraform/outputs.tf" "infrastructure/terraform/outputs.tf"
Move-ItemSafe "terraform/terraform.tfvars.example" "infrastructure/terraform/terraform.tfvars.example"

# Merge platform/terraform modules (non-duplicate)
if (Test-Path "platform/terraform/modules/networking") {
    Copy-Item -Recurse -Force "platform/terraform/modules/networking/*" "infrastructure/terraform/modules/networking/" -ErrorAction SilentlyContinue
}
if (Test-Path "platform/terraform/modules/observability") {
    Move-ItemSafe "platform/terraform/modules/observability" "infrastructure/terraform/modules/cloudwatch"
}

# Legacy GCP terraform
Move-ItemSafe "terraform/main.tf" "legacy/gcp/terraform/main.tf"
Move-ItemSafe "terraform/memorystore.tf" "legacy/gcp/terraform/memorystore.tf"
Move-ItemSafe "terraform/README.md" "legacy/gcp/terraform/README.md"
if (Test-Path "terraform/output.tf") { Remove-Item -Force "terraform/output.tf" }

Write-Host "=== Moving microservices ==="
$svcMap = @{
    "src/frontend" = "microservices/frontend"
    "src/productcatalogservice" = "microservices/product-service"
    "src/cartservice" = "microservices/cart-service"
    "src/checkoutservice" = "microservices/checkout-service"
    "src/paymentservice" = "microservices/payment-service"
    "src/shippingservice" = "microservices/shipping-service"
    "src/currencyservice" = "microservices/currency-service"
    "src/emailservice" = "microservices/email-service"
    "src/recommendationservice" = "microservices/recommendation-service"
    "src/enterprise" = "microservices/shared/enterprise"
    "src/loadgenerator" = "tools/load-generator"
    "src/shoppingassistantservice" = "legacy/gcp/shopping-assistant"
    "src/adservice" = "legacy/gcp/adservice"
}

foreach ($entry in $svcMap.GetEnumerator()) {
    Move-ItemSafe $entry.Key $entry.Value
}

Write-Host "=== Moving GitOps and Kubernetes ==="
Move-ItemSafe "kustomize/base" "gitops/applications/base"
Move-ItemSafe "kustomize/components" "gitops/applications/components"
Move-ItemSafe "kustomize/tests" "gitops/applications/tests"
Move-ItemSafe "helm-chart" "gitops/helm"
Move-ItemSafe "platform/argocd" "gitops/platform/argocd"
Move-ItemSafe "istio-manifests" "kubernetes/platform-cluster/istio/manifests"
Move-ItemSafe "platform/kubernetes/base/prometheus.yaml" "kubernetes/platform-cluster/prometheus/deployment.yaml"
Move-ItemSafe "platform/kubernetes/base/grafana.yaml" "kubernetes/platform-cluster/grafana/deployment.yaml"
Move-ItemSafe "platform/kubernetes/base/otel-collector.yaml" "kubernetes/platform-cluster/opentelemetry-collector/deployment.yaml"
Move-ItemSafe "platform/kubernetes/base/namespace.yaml" "kubernetes/platform-cluster/namespace.yaml"
Move-ItemSafe "platform/kubernetes/base/rbac.yaml" "kubernetes/platform-cluster/rbac.yaml"
Move-ItemSafe "platform/kubernetes/base/networkpolicy.yaml" "security/network-policies/platform-networkpolicy.yaml"

# Remove duplicate kubernetes-manifests and release
if (Test-Path "kubernetes-manifests") { Remove-Item -Recurse -Force "kubernetes-manifests" }
if (Test-Path "release") { Remove-Item -Recurse -Force "release" }
if (Test-Path "kustomize") { Remove-Item -Recurse -Force "kustomize" }

Write-Host "=== Moving CI/CD, Security, Observability, Automation ==="
Move-ItemSafe "platform/jenkins" "cicd/jenkins"
Move-ItemSafe ".github/workflows" "cicd/github-actions/workflows"
Move-ItemSafe "platform/iam" "security/iam"
Move-ItemSafe "platform/monitoring/prometheus.yml" "observability/prometheus/prometheus.yml"
Move-ItemSafe "platform/monitoring/alerts.yml" "observability/prometheus/alerts.yml"
Move-ItemSafe "platform/scripts/backup.sh" "automation/bash/backup.sh"
Move-ItemSafe "platform/scripts/restore.sh" "automation/bash/restore.sh"
Move-ItemSafe "platform/scripts/cleanup_lambda.py" "automation/lambda/cleanup_lambda.py"
Move-ItemSafe "platform/docker/docker-compose.yml" "local-development/docker-compose/shared-services.yml"
Move-ItemSafe "platform/ansible" "infrastructure/ansible"
Move-ItemSafe "platform/docs/runbook.md" "docs/runbooks/platform-runbook.md"
Move-ItemSafe "platform/docs/dr.md" "docs/runbooks/disaster-recovery.md"
Move-ItemSafe "platform/docs/architecture.md" "docs/architecture/platform-architecture.md"
Move-ItemSafe "docs/releasing" "docs/releasing"
Move-ItemSafe "docs/architecture" "docs/architecture/legacy"
Move-ItemSafe "platform/gitops" "gitops/README-legacy"
Move-ItemSafe "platform/charts" "gitops/helm/app-chart"

# Legacy GCP
Move-ItemSafe ".deploystack" "legacy/gcp/deploystack"
Move-ItemSafe ".github/terraform" "legacy/gcp/ci-terraform"
Move-ItemSafe "cloudbuild.yaml" "legacy/gcp/cloudbuild.yaml"
Move-ItemSafe "skaffold.yaml" "legacy/gcp/skaffold.yaml"

# Clean up empty/old directories
@("terraform", "platform", "src", ".github") | ForEach-Object {
    if (Test-Path $_) { Remove-Item -Recurse -Force $_ -ErrorAction SilentlyContinue }
}

Write-Host "=== Reorganization complete ==="
