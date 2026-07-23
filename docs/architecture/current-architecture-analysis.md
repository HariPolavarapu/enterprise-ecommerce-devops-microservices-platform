# Current Architecture Analysis

## Overview

This document analyzes the existing Google Online Boutique (GoogleCloudPlatform/microservices-demo) implementation as the foundation for an enterprise-grade e-commerce platform. The application consists of 11 microservices communicating via gRPC, deployed on Kubernetes with optional Istio service mesh.

---

## Service Inventory

| Service | Language | Port | Protocol | Primary Responsibility |
|---------|----------|------|----------|------------------------|
| **Frontend** | Go | 8080 | HTTP/gRPC | Web server, session management, API gateway |
| **Cart Service** | C# (.NET) | 7070 | gRPC | Shopping cart storage in Redis |
| **Product Catalog Service** | Go | 3550 | gRPC | Product listing, search, detail retrieval |
| **Currency Service** | Node.js | 7000 | gRPC | Currency conversion (ECB rates) |
| **Payment Service** | Node.js | 50051 | gRPC | Mock credit card charging |
| **Shipping Service** | Go | 50051 | gRPC | Shipping quotes and order shipment |
| **Email Service** | Python | 5000 | gRPC | Order confirmation emails |
| **Checkout Service** | Go | 5050 | gRPC | Order orchestration (payment, shipping, email) |
| **Recommendation Service** | Python | 8080 | gRPC | Product recommendations |
| **Ad Service** | Java | 9555 | gRPC | Contextual advertisements |
| **Load Generator** | Python/Locust | N/A | HTTP | Synthetic traffic generation |
| **Shopping Assistant** | Python | 80 | HTTP | AI-powered product suggestions (optional) |
| **Redis** | - | 6379 | TCP | Cart persistence |

---

## Communication Patterns

### gRPC (Primary)
- All service-to-service communication uses gRPC with Protocol Buffers
- Unary RPCs only (no streaming)
- Health checking via `grpc.health.v1.Health` service
- OpenTelemetry instrumentation for trace propagation

### HTTP (Frontend Only)
- Frontend exposes HTTP endpoints for browser clients
- Server-side rendered HTML templates (Go html/template)
- REST-like routes: `/`, `/product/{id}`, `/cart`, `/cart/checkout`, `/setCurrency`, `/logout`

### Asynchronous Patterns
- **None currently implemented** - all communication is synchronous request-response
- No message queues, event streaming, or pub/sub

---

## API Definitions (from protos/demo.proto)

### CartService
```protobuf
service CartService {
    rpc AddItem(AddItemRequest) returns (Empty);
    rpc GetCart(GetCartRequest) returns (Cart);
    rpc EmptyCart(EmptyCartRequest) returns (Empty);
}
```

### ProductCatalogService
```protobuf
service ProductCatalogService {
    rpc ListProducts(Empty) returns (ListProductsResponse);
    rpc GetProduct(GetProductRequest) returns (Product);
    rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse);
}
```

### CurrencyService
```protobuf
service CurrencyService {
    rpc GetSupportedCurrencies(Empty) returns (GetSupportedCurrenciesResponse);
    rpc Convert(CurrencyConversionRequest) returns (Money);
}
```

### PaymentService
```protobuf
service PaymentService {
    rpc Charge(ChargeRequest) returns (ChargeResponse);
}
```

### ShippingService
```protobuf
service ShippingService {
    rpc GetQuote(GetQuoteRequest) returns (GetQuoteResponse);
    rpc ShipOrder(ShipOrderRequest) returns (ShipOrderResponse);
}
```

### EmailService
```protobuf
service EmailService {
    rpc SendOrderConfirmation(SendOrderConfirmationRequest) returns (Empty);
}
```

### CheckoutService
```protobuf
service CheckoutService {
    rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
}
```

### RecommendationService
```protobuf
service RecommendationService {
    rpc ListRecommendations(ListRecommendationsRequest) returns (ListRecommendationsResponse);
}
```

### AdService
```protobuf
service AdService {
    rpc GetAds(AdRequest) returns (AdResponse);
}
```

---

## Service Dependencies (Call Graph)

```
Frontend (HTTP) 
  ├─▶ ProductCatalogService (gRPC)
  ├─▶ CurrencyService (gRPC)
  ├─▶ CartService (gRPC)
  ├─▶ RecommendationService (gRPC)
  ├─▶ CheckoutService (gRPC)
  ├─▶ ShippingService (gRPC)
  ├─▶ AdService (gRPC)
  └─▶ ShoppingAssistantService (HTTP, optional)

CheckoutService (gRPC)
  ├─▶ CartService (GetCart, EmptyCart)
  ├─▶ ProductCatalogService (GetProduct)
  ├─▶ CurrencyService (Convert)
  ├─▶ ShippingService (GetQuote, ShipOrder)
  ├─▶ PaymentService (Charge)
  └─▶ EmailService (SendOrderConfirmation)

RecommendationService (gRPC)
  └─▶ ProductCatalogService (ListProducts)
```

---

## Deployment Architecture

### Kubernetes Manifests (kubernetes-manifests/)
- **Deployments**: Each service has its own Deployment with resource limits/requests
- **Services**: ClusterIP for internal gRPC, LoadBalancer for frontend-external
- **ServiceAccounts**: Per-service accounts for RBAC
- **Security Context**: Non-root user (1000), read-only root filesystem, dropped capabilities
- **Probes**: HTTP (frontend) or gRPC (others) for readiness/liveness

### Helm Chart (helm-chart/)
- **values.yaml**: Centralized configuration for all services
- **templates/**: Individual service templates + common.yaml (NetworkPolicy, AuthorizationPolicy)
- **Features**: Network policies, Istio sidecars, OpenTelemetry collector, Google Cloud Operations

### Kustomize (kustomize/)
- **base/**: Core manifests
- **components/**: Optional features (network policies, Istio, Spanner, Memorystore, AlloyDB, Cymbal branding, shopping assistant)
- **tests/**: Integration test configurations

### Istio (istio-manifests/)
- **frontend-gateway.yaml**: Ingress gateway for external access
- **allow-egress-googleapis.yaml**: Egress control for Google APIs

### CI/CD (.github/workflows/)
- **ci-main.yaml**: Main branch CI
- **ci-pr.yaml**: PR validation
- **helm-chart-ci.yaml**: Helm chart validation
- **kustomize-build-ci.yaml**: Kustomize build validation
- **terraform-validate-ci.yaml**: Terraform validation

---

## Configuration Management

### Environment Variables (per service)
Each service reads configuration from environment variables:
- `PORT`: Listen port
- `*_SERVICE_ADDR`: Downstream service addresses (DNS-based service discovery)
- `ENABLE_TRACING`: Enable OpenTelemetry tracing (1/0)
- `ENABLE_PROFILER`: Enable Google Cloud Profiler
- `COLLECTOR_SERVICE_ADDR`: OTLP trace collector address
- `ENV_PLATFORM`: Platform identifier (local, gcp, aws, azure, onprem, alibaba)
- `DISABLE_PROFILER`, `DISABLE_TRACING`, `DISABLE_STATS`: Feature toggles

### Secrets Management
- **No secrets management** currently implemented
- All configuration via environment variables (including service addresses)
- No TLS/mTLS between services (insecure gRPC)

---

## Observability

### Tracing
- OpenTelemetry (OTLP/gRPC) to collector
- Trace context propagation via W3C TraceContext and Baggage
- Go services: `otelgrpc` stats handlers
- Node.js: `@opentelemetry/instrumentation-grpc`
- Python: `GrpcInstrumentorServer`/`Client`
- Java: Not implemented (placeholder)

### Metrics
- **Not implemented** (TODO comments in all services)
- Prometheus metrics endpoint not exposed

### Logging
- Structured JSON logging (logrus for Go, pino for Node.js, custom for Python)
- Fields: timestamp, severity, message, service-specific context
- Request ID and session ID propagation in frontend

### Health Checks
- gRPC Health Checking Protocol (`grpc.health.v1.Health`)
- Kubernetes readiness/liveness probes use health check endpoint

---

## Security Posture

### Current State
- **No authentication/authorization** on any service
- **No mTLS** between services (insecure gRPC)
- **No API gateway** - frontend directly exposes HTTP
- **No rate limiting**
- **No JWT/OAuth** - session IDs are random UUIDs in cookies
- **No RBAC** - all users have same permissions
- Network policies available but disabled by default
- Istio authorization policies available but disabled

### Container Security
- Non-root user (UID 1000)
- Read-only root filesystem
- Dropped capabilities
- No privilege escalation

---

## Data Persistence

### Redis (Cart Service)
- In-memory key-value store
- Key: session/user ID
- Value: Serialized Cart protobuf
- No persistence configuration (data loss on restart)
- Single instance (no HA)

### Product Catalog
- Static JSON file loaded at startup (`data/products.json`)
- In-memory only
- Reloadable via SIGUSR1 signal

### Other Services
- **No persistent storage** - all stateless
- Payment, Shipping, Email: Mock implementations
- Currency: Static JSON file (ECB rates)

---

## Scaling Characteristics

### Horizontal Scaling
- All services designed as stateless (except Cart Service ↔ Redis)
- Deployments support replica scaling
- Resource limits defined per service

### Bottlenecks
- **Cart Service**: Single Redis instance (no clustering)
- **Product Catalog**: In-memory catalog, no sharding
- **Currency Service**: Highest QPS service, no caching layer
- **Frontend**: Session affinity via cookies (sticky sessions)

---

## Technology Stack Summary

| Layer | Technologies |
|-------|-------------|
| **Languages** | Go, C# (.NET 6), Node.js, Python, Java |
| **RPC** | gRPC + Protocol Buffers v3 |
| **HTTP** | Go net/http + gorilla/mux |
| **Templating** | Go html/template |
| **Service Discovery** | Kubernetes DNS |
| **Tracing** | OpenTelemetry (OTLP/gRPC) |
| **Logging** | logrus (Go), pino (Node), Python logging |
| **Profiling** | Google Cloud Profiler |
| **Container** | Docker multi-arch (amd64/arm64) |
| **Orchestration** | Kubernetes (Deployments, Services, ConfigMaps) |
| **Package Manager** | Helm 3, Kustomize |
| **Service Mesh** | Istio (optional) |
| **CI/CD** | GitHub Actions, Skaffold |
| **Infrastructure** | Terraform (GKE) |

---

## Key Architectural Decisions (Existing)

1. **Polyglot Microservices**: Each service written in language best suited for its domain
2. **gRPC for Internal Communication**: Strong typing, performance, code generation
3. **Frontend as API Gateway**: BFF pattern - aggregates backend calls
4. **Session-Based (No Auth)**: Cookie-based session IDs, no user identity
5. **In-Memory/Redis State**: Minimal persistence, optimized for demo
6. **Kubernetes-Native**: Deployments, Services, Health Checks
7. **Observability-First**: OpenTelemetry instrumentation built-in
8. **Security Hardened Containers**: Non-root, read-only FS, dropped caps
9. **Configuration via Env Vars**: 12-factor app principles
10. **Optional Service Mesh**: Istio integration available

---

## Gaps for Enterprise Production

| Area | Current State | Enterprise Requirement |
|------|---------------|------------------------|
| Authentication | None | OAuth2/OIDC, JWT, SSO |
| Authorization | None | RBAC, ABAC, fine-grained |
| API Gateway | Frontend only | Kong/Envoy, rate limiting, routing |
| Data Persistence | Redis (cart), JSON (catalog) | PostgreSQL (orders, users, inventory) |
| Event Streaming | None | Kafka for async communication |
| Search | In-memory linear scan | OpenSearch/Elasticsearch |
| Object Storage | None | S3-compatible (images, invoices) |
| Audit Logging | None | Immutable audit trail |
| Multi-tenancy | None | Namespace/isolation |
| Disaster Recovery | None | Backup/restore, multi-region |
| Compliance | None | PCI-DSS, GDPR, SOC2 |
| API Versioning | None | Versioned APIs, deprecation policy |
| Circuit Breaking | None | Resilience patterns |
| Distributed Transactions | None | Saga pattern, eventual consistency |

---

## Conclusion

The Google Online Boutique provides a solid **microservices foundation** with:
- Clean service boundaries
- gRPC contracts
- Kubernetes-native deployment
- Observability instrumentation
- Security-hardened containers

However, it lacks **all enterprise capabilities** required for production e-commerce:
- User identity and access management
- Persistent transactional data (orders, payments, inventory)
- Asynchronous event-driven architecture
- Advanced search and analytics
- API management and security
- Compliance and audit features

The transformation must **preserve the existing service boundaries and gRPC contracts** while adding enterprise layers around and between them.