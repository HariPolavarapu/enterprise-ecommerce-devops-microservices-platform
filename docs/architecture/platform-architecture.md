# Architecture Overview

The platform uses AWS Organizations with dedicated accounts for management, security, log archive, networking, shared services, development, QA, pre-production and production. Terraform provisions the shared AWS foundation, Ansible bootstraps shared services, and ArgoCD applies Kubernetes manifests for the workload platform.

## Components
- AWS: VPC, IAM, EC2, EKS, ECR, ALB, API Gateway, Route53, CloudFront, WAF, Shield, ACM, RDS, ElastiCache, S3, OpenSearch, CloudWatch, Lambda, EventBridge, SNS and SQS
- CI/CD: Jenkins, Maven, SonarQube, Trivy, Nexus, Docker, Helm, ArgoCD
- Observability: Prometheus, Grafana, Fluent Bit, OpenSearch, Kibana, OpenTelemetry, Jaeger
