# Enterprise E-Commerce Microservices Platform
## Architecture & Design Guide

**Based on:** Google Online Boutique (GoogleCloudPlatform/microservices-demo) v0  
**License:** Apache 2.0  
**Last Updated:** 2026-07-23  

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current Architecture Analysis](#2-current-architecture-analysis)
3. [Enterprise Architecture Overview](#3-enterprise-architecture-overview)
4. [Existing Service Modifications](#4-existing-service-modifications)
5. [New Enterprise Services](#5-new-enterprise-services)
6. [PostgreSQL Database Design](#6-postgresql-database-design)
7. [Kafka Event-Driven Architecture](#7-kafka-event-driven-architecture)
8. [API Gateway Design](#8-api-gateway-design)
9. [Authentication & Authorization (Keycloak)](#9-authentication--authorization)
10. [OpenSearch Search Design](#10-opensearch-search-design)
11. [Object Storage (S3) Design](#11-object-storage-s3-design)
12. [Service Interaction Diagrams](#12-service-interaction-diagrams)
13. [Sequence Diagrams](#13-sequence-diagrams)
14. [Deployment Considerations](#14-deployment-considerations)
15. [Design Decisions & Trade-offs](#15-design-decisions--trade-offs)

---

## 1. Executive Summary

This document describes the transformation of **Google Online Boutique** (a cloud-native microservices demo) into a **production-ready enterprise e-commerce platform**. The strategy is to **preserve the existing architecture** as the foundation and **incrementally add** the missing enterprise capabilities — authentication, persistent storage, event-driven messaging, search, and analytics — rather than rewriting from scratch.

### Key Design Principles

| Principle | Application |
|-----------|-------------|
| **Preserve existing** | Keep all 11 existing microservices; modify them minimally to integrate with new components. |
| **gRPC-first** | Existing services use gRPC for internal communication. New services expose gRPC for inter-service calls and REST/HTTP for external-facing APIs. |
| **Event-driven** | Use Kafka as the backbone for asynchronous communication between services (order lifecycle, notifications). |
| **Persistent storage** | Replace in-memory/in-file storage with PostgreSQL for stateful services. |
| **API-first** | Frontend becomes a client of the API Gateway; all external traffic goes through the gateway. |
| **Security** | JWT-based authentication via Keycloak; RBAC with 4 roles (Customer, Admin, Inventory Manager, Support). |
| **Search** | OpenSearch for product catalog search with autocomplete, filtering, and ranking. |
| **Observability** | OpenTelemetry for distributed tracing (already partially implemented); add metrics and logging. |

---

## 2. Current Architecture Analysis

### 2.1 Existing Services Overview

The current application has **11 microservices** + **1 Redis instance**, communicating over **gRPC** with **HTTP** for the frontend only.

| Service | Language | Database | Persistence | gRPC Server? | gRPC Client? | Current Data Source |
|---------|----------|----------|-------------|-------------|-------------|---------------------|
| **frontend** | Go | None | None | No | Yes (all downstream) | In-memory session |
| **cartservice** | C# | Redis | Yes | No | No (C# gRPC is not used) | Redis cache |
| **checkoutservice** | Go | None | None | Yes | Yes | Transient |
| **currencyservice** | Node.js | None | None | No | No | ECB API |
| **paymentservice** | Node.js | None | None | No | No | Mock |
| **shippingservice** | Go | None | None | Yes | No | Mock |
| **productcatalogservice** | Go | None (JSON file) | None | Yes | No | products.json |
| **recommendationservice** | Python | None | None | No | No | ML model |
| **adservice** | Java | None | None | No | No | In-memory |
| **emailservice** | Python | None | None | No | No | Mock SMTP |
| **loadgenerator** | Python | None | None | No | No | Locust |

### 2.2 Current Communication Flow

```
User → Browser → frontend (HTTP:8080)
                     │
                     ├─→ productcatalogservice (gRPC:3550)
                     ├─→ currencyservice (gRPC:7000)
                     ├─→ cartservice (→ Redis:6379)
                     ├─→ checkoutservice (gRPC:5050)
                     │     ├─→ productcatalogservice
                     │     ├─→ cartservice
                     │     ├─→ currencyservice
                     │     ├─→ shippingservice
                     │     ├─→ paymentservice
                     │     └─→ emailservice
                     ├─→ shippingservice
                     ├─→ recommendationservice
                     └─→ adservice
```

### 2.3 Key Architectural Gaps

| Gap | Description |
|-----|-------------|
| **No authentication** | All users are anonymous with session IDs. No login, no user accounts, no RBAC. |
| **No persistent storage** | Product catalog is a JSON file. Cart is in Redis (volatile). Orders are not persisted. |
| **No event-driven messaging** | All calls are synchronous gRPC. No async workflows. |
| **No search** | Product search is a basic `SearchProducts` gRPC call against in-memory data. |
| **No analytics** | No user behavior tracking, no order history, no business intelligence. |
| **No API Gateway** | Frontend directly calls downstream services. No rate limiting, no JWT validation, no routing. |
| **No object storage** | Product images are served from a `static/` directory. |
| **Mock payments** | `paymentservice` uses a mock that always succeeds. |
| **No audit logging** | No transaction history, no user activity trail. |

---

## 3. Enterprise Architecture Overview

### 3.1 High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              USER / BROWSER                                 │
└───────────────────────────┬─────────────────────────────────────────────────┘
                            │ HTTPS (TLS)
                            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          API GATEWAY (Kong / Envoy)                         │
│                    ┌──────────┬──────────┬─────────────────┐               │
│                    │ JWT Auth │ Rate-Lim │ Route / Load-Bal │               │
│                    │ 0.5K     │ 100 req/s│ Path-based       │               │
│                    └──────────┴──────────┴─────────────────┘               │
└──────┬─────────────────────────────────────────────────┬────────────────────┘
       │                                                  │
       ▼                                                  ▼
┌──────────────┐                                    ┌──────────────┐
│  KEYCLOAK    │                                    │   FRONTEND   │
│  (Auth)      │                                    │  (Go / HTTP) │
│              │                                    │              │
│  /auth       │◄─── OIDC ────│                     │  /login      │
│  /token      │               │                     │  /register   │
│  /userinfo    │               │                     │  /products   │
│  /logout      │               │                     │  /cart      │
└──────┬───────┘               │                     │  /checkout  │
       │                       │                     └──────┬───────┘
       │                       │                            │
       ▼                       ▼                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           gRPC / HTTP (Internal)                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Existing Microservices (unchanged core logic, modified integrations)    │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Product   │  │ Cart     │  │ Checkout │  │ Payment  │              │
│  │ Catalog   │  │ Service  │  │ Service  │  │ Service  │              │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘              │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Shipping  │  │ Currency  │  │ Email    │  │ Currency │              │
│  │ Service  │  │ Service  │  │ Service   │  │ Service  │              │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘              │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Ad       │  │ Recom-   │  │ User     │  │ Order    │              │
│  │ Service  │  │ mendation│  │ Service  │  │ Service  │              │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘              │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Inventory │  │ Notifi-  │  │Analytics │  │Shopping   │              │
│  │ Service   │  │ cation   │  │ Service  │  │Assistant │              │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘              │
│                                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                          INFRASTRUCTURE LAYER                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐           │
│  │PostgreSQL │   │  Redis   │   │  Kafka   │   │OpenSearch│           │
│  │ (Orders,  │   │ (Cart,   │   │ (Events) │   │ (Search) │           │
│  │  Users,   │   │ Session) │   │          │   │          │           │
│  │ Inventory)│   │          │   │          │   │          │           │
│  └──────────┘   └──────────┘   └──────────┘   └──────────┘           │
│                                                                          │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐                          │
│  │  S3      │   │  Kafka   │   │  Jaeger  │                          │
│  │ (Images, │   │  Connect │   │ (Tracing)│                          │
│  │ Invoices)│   │          │   │          │                          │
│  └──────────┘   └──────────┘   └──────────┘                          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Component Summary

| Layer | Component | Status | Notes |
|-------|-----------|--------|-------|
| **User** | Browser / Mobile App | N/A | Any HTTP client |
| **Edge** | API Gateway (Kong/Envoy) | **NEW** | JWT validation, rate limiting, routing |
| **Auth** | Keycloak | **NEW** | OIDC provider, RBAC, user federation |
| **Frontend** | Frontend (Go) | **MODIFIED** | Add login/register, JWT support |
| **BFF** | User Service | **NEW** | User management, profiles |
| **Product** | Product Catalog Service | **MODIFIED** | Add PostgreSQL, OpenSearch indexing |
| **Cart** | Cart Service | **MODIFIED** | Add Kafka producer for cart events |
| **Order** | Checkout Service | **MODIFIED** | Add Order -> Kafka; Add OrderService |
| **Payment** | Payment Service | **MODIFIED** | Add PostgreSQL for transactions |
| **Shipping** | Shipping Service | **MODIFIED** | Add Kafka consumer for shipment tracking |
| **Notification** | Email Service | **MODIFIED** | Convert to Kafka consumer |
| **Recommendation** | Recommendation Service | **MODIFIED** | Add OpenSearch-based recommendations |
| **Search** | OpenSearch | **NEW** | Product search, autocomplete, filtering |
| **Analytics** | Analytics Service | **NEW** | Kafka consumer for events |
| **Inventory** | Inventory Service | **NEW** | Stock management |
| **Order** | Order Service | **NEW** | Order lifecycle management |
| **Notification** | Notification Service | **NEW** | Push notifications, email |

---

## 4. Existing Service Modifications

### 4.1 Product Catalog Service

**Current state:** Go service, reads products from `products.json` file, in-memory catalog, no persistence.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add PostgreSQL for product storage | Code + DB | Replace `products.json` with PostgreSQL `products` table |
| 2 | Add gRPC `GetProductsByIDs` for batch fetching | Code | Add `rpc GetProductsByIDs(GetProductsByIDsRequest) returns (GetProductsByIDsResponse)` |
| 3 | Add `ManageProduct` CRUD APIs | Code | `CreateProduct`, `UpdateProduct`, `DeleteProduct` (admin only) |
| 4 | Add OpenSearch indexing for search | Code + Config | Index products to OpenSearch on create/update |
| 5 | Add Kafka producer for `product-updated` | Code + Config | Emit when product changes (price, stock, description) |
| 6 | Add image URL support | Config | Product schema: add `image_urls[]`, `thumbnail_url` |
| 7 | Add category/subcategory hierarchy | Config | Add `category_id`, `subcategory_id` to Product |
| 8 | Add product variants (size, color, SKU) | Code | `ProductVariant` message |

**PostgreSQL Schema for Product Catalog:**

```sql
CREATE TABLE products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description      TEXT,
    sku             VARCHAR(50) UNIQUE NOT NULL,
    price           NUMERIC(10,2) NOT NULL,
    currency_code   VARCHAR(3) DEFAULT 'USD',
    category_id     UUID REFERENCES categories(id),
    subcategory_id  UUID REFERENCES subcategories(id),
    image_urls     TEXT[],       -- Array of image URLs
    thumbnail_url   TEXT,
    attributes      JSONB,       -- Flexible product attributes
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ   -- Soft delete
);

CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id   UUID REFERENCES categories(id), -- hierarchy
    sort_order  INT DEFAULT 0
);

CREATE TABLE subcategories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    category_id UUID REFERENCES categories(id),
    parent_id   UUID REFERENCES subcategories(id),
    sort_order  INT DEFAULT 0
);

CREATE TABLE product_variants (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id        UUID REFERENCES products(id),
    name              VARCHAR(100),  -- e.g., "Large", "Red"
    sku               VARCHAR(50) UNIQUE NOT NULL,
    price_adjustment  NUMERIC(10,2) DEFAULT 0,
    stock             INT DEFAULT 0,
    is_active         BOOLEAN DEFAULT true
);
```

**Kafka Events:**

| Topic | Producer | Consumer | When |
|-------|----------|----------|------|
| `product-created` | Product Catalog | OpenSearch Indexer, Analytics Service | New product created |
| `product-updated` | Product Catalog | OpenSearch Indexer, Analytics Service | Product modified |
| `product-deleted` | Product Catalog | OpenSearch Indexer, Analytics Service | Product removed |

**OpenSearch Index Mapping:**

```json
{
  "index": "products",
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "name": { 
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "description": { "type": "text" },
      "categories": { "type": "keyword" },
      "price": { "type": "float" },
      "sku": { "type": "keyword" },
      "is_active": { "type": "boolean" },
      "created_at": { "type": "date" }
    }
  }
}
```

**API Changes:**

- `GET /api/v1/products` — List with pagination, filtering, search support
- `GET /api/v1/products/{id}` — Get single product
- `POST /api/v1/products` — Create (Admin only)
- `PUT /api/v1/products/{id}` — Update (Admin/Inventory Manager)
- `DELETE /api/v1/products/{id}` — Soft delete (Admin only)
- `GET /api/v1/products/search?q=...` — Search via OpenSearch
- `GET /api/v1/products/suggest?q=...` — Autocomplete

**New Dependencies:**

```go
// go.mod additions
github.com/jackc/pgx/v5           // PostgreSQL driver
github.com/opensearch-project/opensearch-go/v3  // OpenSearch client
github.com/segmentio/kafka-go      // Kafka client
```

**Configuration Changes:**

```yaml
# environment variables
POSTGRES_DSN: postgresql://user:pass@postgres:5432/productcatalog
OPENSEARCH_URL: https://opensearch:9200
KAFKA_BROKERS: kafka:9092
PRODUCT_IMAGES_BUCKET: s3://products
```

---

### 4.2 Cart Service

**Current state:** C# (.NET) service, stores cart in Redis, no persistence.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add PostgreSQL for persistent cart | Code + DB | Add `shopping_carts` table |
| 2 | Add Kafka producer for `cart-updated` | Code + Config | Emit on add/remove from cart |
| 3 | Add cart expiration TTL | Config | Configurable TTL (default: 30 days) |
| 4 | Add cart merge on login | Code | Merge anonymous cart → authenticated user cart |
| 5 | Support for guest/identified user | Code | Store `user_id` or `session_id` |

**PostgreSQL Schema:**

```sql
CREATE TABLE shopping_carts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id),  -- NULL for guest
    session_id  VARCHAR(255),               -- For anonymous users
    status      VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'abandoned', 'converted')),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    expires_at  TIMESTAMPTZ DEFAULT NOW() + INTERVAL '30 days'
);

CREATE TABLE cart_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id     UUID REFERENCES shopping_carts(id) ON DELETE CASCADE,
    product_id  UUID REFERENCES products(id),
    variant_id  UUID REFERENCES product_variants(id),
    quantity     INT NOT NULL CHECK (quantity > 0),
    price        NUMERIC(10,2) NOT NULL, -- snapshot at add time
    added_at    TIMESTAMPTZ DEFAULT NOW()
);
```

**Kafka Events:**

| Topic | Producer | Consumer | When |
|-------|----------|----------|------|
| `cart-updated` | Cart Service | Analytics, Recommendation | Item added/removed |

---

### 4.3 Checkout Service

**Current state:** Go service, orchestrates order placement, no persistence.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add Kafka producer for `order-created` | Code | After `PlaceOrder` succeeds |
| 2 | Add persistence for orders | Code + DB | Write to `orders` table |
| 3 | Add inventory validation before order | Code | Call Inventory Service |
| 4 | Add user validation | Code | Verify JWT, get user from User Service |
| 5 | Add idempotency key | Code | `X-Idempotency-Key` header |
| 6 | Add order status tracking | Code | Emit `order-status-changed` events |

**PostgreSQL Schema:**

```sql
-- In checkoutservice or orderservice
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    status          VARCHAR(30) DEFAULT 'pending' CHECK (status IN (
                        'pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded'
                    )),
    total_amount    NUMERIC(10,2),
    currency        VARCHAR(3) DEFAULT 'USD',
    shipping_address JSONB,  -- Full address object
    billing_address  JSONB,
    payment_method  VARCHAR(50),
    payment_id      VARCHAR(255),  -- External payment reference
    idempotency_key VARCHAR(255) UNIQUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    cancelled_at    TIMESTAMPTZ
);

CREATE TABLE order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id      UUID REFERENCES products(id),
    variant_id      UUID REFERENCES product_variants(id),
    product_name    VARCHAR(255), -- snapshot at order time
    product_sku     VARCHAR(50),
    quantity         INT NOT NULL CHECK (quantity > 0),
    unit_price       NUMERIC(10,2),
    total_price      NUMERIC(10,2),
    created_at       TIMESTAMPTZ DEFAULT NOW()
);
```

**Kafka Events:**

| Topic | Producer | Consumer | When |
|-------|----------|----------|------|
| `order-created` | Checkout Service | Order Service, Inventory, Notification, Analytics | Order placed |
| `order-cancelled` | Checkout/Order | Inventory (restock), Notification | Order cancelled |
| `order-status-changed` | Checkout/Order | All interested | Status transition |

---

### 4.4 Payment Service

**Current state:** Node.js service, mock charge that always succeeds.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add PostgreSQL for transaction records | Code + DB | `payments` table |
| 2 | Add real payment provider integration | Code | Stripe / Adyen / Braintree |
| 3 | Add payment status webhook handler | Code | Handle async payment status |
| 4 | Add Kafka producer for `payment-success` / `payment-failed` | Code | Emit after charge |
| 5 | Add refund API | Code | `POST /api/v1/payments/{id}/refund` |
| 6 | Add idempotency for charge | Code | Prevent double charges |

**PostgreSQL Schema:**

```sql
CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID REFERENCES orders(id),
    user_id         UUID REFERENCES users(id),
    amount          NUMERIC(10,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    status          VARCHAR(30) DEFAULT 'pending' CHECK (status IN (
                        'pending', 'processing', 'succeeded', 'failed', 'refunded', 'partially_refunded'
                    )),
    payment_method   VARCHAR(50),  -- 'card', 'paypal', 'stripe'
    payment_provider VARCHAR(50),   -- 'stripe', 'adyen'
    provider_payment_id VARCHAR(255), -- Stripe PaymentIntent ID
    provider_customer_id VARCHAR(255), -- Stripe Customer ID
    error_message    TEXT,
    idempotency_key  VARCHAR(255) UNIQUE,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    refunded_at      TIMESTAMPTZ
);

CREATE TABLE payment_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      UUID REFERENCES payments(id),
    action          VARCHAR(50) NOT NULL, -- 'created', 'succeeded', 'failed', 'refunded'
    old_status      VARCHAR(30),
    new_status      VARCHAR(30),
    metadata        JSONB,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

**Kafka Events:**

| Topic | Producer | Consumer | When |
|-------|----------|----------|------|
| `payment-success` | Payment Service | Order Service, Email, Analytics | Payment succeeded |
| `payment-failed` | Payment Service | Order Service, Notification | Payment failed |
| `payment-refunded` | Payment Service | Order Service, Notification | Refund processed |

**Kafka Payload:**

```json
{
  "event_id": "uuid",
  "payment_id": "uuid",
  "order_id": "uuid",
  "user_id": "uuid",
  "amount": 29.99,
  "currency": "USD",
  "status": "succeeded",
  "timestamp": "2026-07-23T12:00:00Z",
  "metadata": {
    "payment_method": "card",
    "provider": "stripe"
  }
}
```

---

### 4.5 Shipping Service

**Current state:** Go service, mock shipping quotes.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add Kafka consumer for `shipment-requested` | Code | Listen to `order-created` |
| 2 | Add PostgreSQL for shipping records | Code + DB | `shipments` table |
| 3 | Add real shipping carrier integration | Code | FedEx, UPS, USPS APIs |
| 4 | Add tracking number generation | Code | Real tracking IDs |
| 5 | Add Kafka producer for `shipment-created` | Code | On shipment dispatch |
| 6 | Add shipment status webhook | Code | Carrier webhook for status updates |

**PostgreSQL Schema:**

```sql
CREATE TABLE shipments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID REFERENCES orders(id),
    carrier          VARCHAR(50) NOT NULL, -- 'fedex', 'ups', 'usps'
    tracking_number  VARCHAR(100),
    status           VARCHAR(30) DEFAULT 'pending' CHECK (status IN (
                        'pending', 'picked_up', 'in_transit', 'delivered', 'exception'
                    )),
    origin_address   JSONB,
    destination_address JSONB,
    estimated_delivery TIMESTAMPTZ,
    actual_delivery   TIMESTAMPTZ,
    weight_kg        NUMERIC(10,2),
    cost             NUMERIC(10,2),
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);
```

**Kafka Events:**

| Topic | Producer | Consumer | When |
|-------|----------|----------|------|
| `shipment-created` | Shipping Service | Notification, Order Service | Shipment created |
| `shipment-delivered` | Shipping Service | Order Service | Package delivered |
| `shipment-exception` | Shipping Service | Notification | Delivery exception |

---

### 4.6 Email Service

**Current state:** Python service, sends order confirmation emails (mock).

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Convert from gRPC client to Kafka consumer | Code | Subscribe to `notification-requested` topic |
| 2 | Add real email provider (SendGrid, SES) | Code | Replace mock SMTP |
| 3 | Add email template system | Code | HTML templates with Go templates |
| 4 | Add Kafka consumer for `order-created` | Code | Send order confirmation |
| 5 | Add transactional email support | Code | Welcome emails, password reset, etc. |

**Kafka Events:**

| Topic | Consumer | When |
|-------|----------|------|
| `order-created` → `notification-requested` | Email Service | Order placed |
| `payment-success` → `notification-requested` | Email Service | Payment confirmed |

**New Dependencies:**

```python
# requirements.txt additions
sendgrid>=6.0.0
boto3>=1.26.0   # AWS SES
kafka-python>=2.0.2
```

---

### 4.7 Recommendation Service

**Current state:** Python service, basic collaborative filtering.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add OpenSearch for ML-based recommendations | Code | Use OpenSearch's KNN for similarity search |
| 2 | Add Kafka consumer for `cart-updated` | Code | Real-time recommendations based on cart |
| 3 | Add user-based recommendations | Code | `GET /api/v1/recommendations?user_id=...` |
| 4 | Add A/B testing framework | Config | Feature flags for recommendation strategies |

**OpenSearch KNN Mapping:**

```json
{
  "index": "product_embeddings",
  "mappings": {
    "properties": {
      "product_id": { "type": "keyword" },
      "embedding": { 
        "type": "knn_vector",
        "dimension": 768
      },
      "category": { "type": "keyword" }
    }
  }
}
```

---

### 4.8 Currency Service

**Current state:** Node.js service, fetches rates from ECB.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add caching (Redis) | Code | Cache rates to reduce API calls |
| 2 | Add more exchange rate sources | Config | ECB + OpenExchangeRates + fallback |
| 3 | No code changes to core logic | — | Keep as-is, add cache layer |

---

### 4.9 Ad Service

**Current state:** Java (Spring Boot), in-memory ads.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add PostgreSQL for ad targeting | Code + DB | `ads` table with targeting rules |
| 2 | Add Kafka producer for `ad-impression` / `ad-click` | Code | Analytics for ad performance |
| 3 | Add A/B ad placement | Code | Multiple ad strategies |

---

### 4.10 Frontend

**Current state:** Go HTTP server, serves HTML templates.

**Required changes:**

| # | Change | Type | Details |
|---|--------|------|---------|
| 1 | Add login/register UI | Templates | Login page, registration form |
| 2 | Add JWT token management | Code | Store in `httpOnly` cookie |
| 3 | Add RBAC UI elements | Templates | Admin dashboard, role-based views |
| 4 | Update to use API Gateway | Code | Route all API calls through gateway |
| 5 | Add session management via Redis | Code | Store sessions in Redis |
| 6 | Add logout functionality | Code | `POST /logout` — invalidate token |

---

### 4.11 Product Catalog Service (Detailed)

**Full transformation:**

- **Service:** `productcatalogservice` (Go, gRPC)
- **Current:** Reads `products.json` into memory on startup
- **Target:** PostgreSQL-backed, gRPC + REST, OpenSearch-indexed
- **Code changes:** ~500 new lines (data access, search integration, API handlers)
- **Database:** PostgreSQL `products`, `categories`, `product_variants` tables
- **Kafka:** Producer for `product-created`/`updated`/`deleted`
- **APIs:** `ListProducts`, `GetProduct`, `SearchProducts`, `CreateProduct`, `UpdateProduct`, `DeleteProduct`

---

## 5. New Enterprise Services

### 5.1 User Service

**Purpose:** User profile management, address book, preferences.

**Responsibilities:**
- CRUD for user profiles
- Address management (shipping/billing)
- Payment methods management
- User preferences (currency, language, notifications)

**REST APIs:**

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/users/me` | Get current user | JWT |
| PUT | `/api/v1/users/me` | Update profile | JWT |
| GET | `/api/v1/users/{id}` | Get user (Admin) | Admin |
| GET | `/api/v1/users/addresses` | List addresses | JWT |
| POST | `/api/v1/users/addresses` | Add address | JWT |
| PUT | `/api/v1/users/addresses/{id}` | Update address | JWT |
| DELETE | `/api/v1/users/addresses/{id}` | Delete address | JWT |
| GET | `/api/v1/users/payment-methods` | List saved payment methods | JWT |
| POST | `/api/v1/users/payment-methods` | Add payment method | JWT |

**gRPC Services:**

```protobuf
service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc CreateUser(CreateUserRequest) returns (User);
    rpc UpdateUser(UpdateUserRequest) returns (User);
    rpc DeleteUser(DeleteUserRequest) returns (Empty);
    rpc GetAddresses(GetAddressesRequest) returns (AddressList);
    rpc AddAddress(AddAddressRequest) returns (Address);
    rpc GetPaymentMethods(GetPaymentMethodsRequest) returns (PaymentMethodList);
}
```

**PostgreSQL Schema:**

```sql
-- Users table (in shared `users` schema)
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    keycloak_id     VARCHAR(255) UNIQUE, -- Keycloak user UUID
    email           VARCHAR(255) UNIQUE NOT NULL,
    username        VARCHAR(100) UNIQUE NOT NULL,
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    phone           VARCHAR(20),
    avatar_url      TEXT,
    role            VARCHAR(30) DEFAULT 'customer' CHECK (role IN (
                        'customer', 'admin', 'inventory_manager', 'support'
                    )),
    is_active       BOOLEAN DEFAULT true,
    email_verified  BOOLEAN DEFAULT false,
    preferences     JSONB DEFAULT '{}', -- language, currency, notifications
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

CREATE TABLE user_addresses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    label           VARCHAR(50), -- 'Home', 'Work'
    street_address  TEXT NOT NULL,
    city            VARCHAR(100) NOT NULL,
    state           VARCHAR(100),
    postal_code     VARCHAR(20),
    country         VARCHAR(100) NOT NULL,
    is_default       BOOLEAN DEFAULT false,
    is_billing      BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_payment_methods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    provider        VARCHAR(50) NOT NULL, -- 'stripe', 'paypal'
    provider_customer_id VARCHAR(255),
    type            VARCHAR(20) NOT NULL, -- 'card', 'paypal'
    last_four       VARCHAR(4),
    expiry_month     INT,
    expiry_year     INT,
    is_default       BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

**Kafka:**

- Consumer: `user-registered` — from Keycloak event listener
- Producer: `user-updated` — profile changes

---

### 5.2 Order Service

**Purpose:** Order lifecycle management, status tracking, history.

**Responsibilities:**
- Order CRUD
- Order status state machine
- Order history
- Order cancellation
- Order returns

**REST APIs:**

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/orders` | List user orders | JWT |
| GET | `/api/v1/orders/{id}` | Get order details | JWT |
| POST | `/api/v1/orders` | Create order | JWT |
| POST | `/api/v1/orders/{id}/cancel` | Cancel order | JWT |
| GET | `/api/v1/orders/admin` | List all orders (Admin) | Admin |
| PUT | `/api/v1/orders/{id}/status` | Update order status | Admin |

**gRPC:**

```protobuf
service OrderService {
    rpc CreateOrder(CreateOrderRequest) returns (Order);
    rpc GetOrder(GetOrderRequest) returns (Order);
    rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
    rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (Order);
    rpc CancelOrder(CancelOrderRequest) returns (Order);
}
```

**PostgreSQL Schema:**

```sql
-- See Orders/Order Items in PostgreSQL Design section
```

**Order Status State Machine:**

```
pending → confirmed → processing → shipped → delivered
                              ↘ cancelled
                              ↘ returned (from delivered)
```

**Kafka:**

- Consumer: `order-created`, `order-cancelled`
- Producer: `order-status-changed`

---

### 5.3 Inventory Service

**Purpose:** Real-time stock management, reservations, restocking.

**Responsibilities:**
- Stock levels
- Inventory reservations
- Low-stock alerts
- Restock requests
- Warehouse management

**REST APIs:**

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/inventory/{productId}` | Get stock level | Public |
| POST | `/api/v1/inventory/reserve` | Reserve inventory | Internal |
| POST | `/api/v1/inventory/release` | Release reservation | Internal |
| PUT | `/api/v1/inventory/{productId}/stock` | Update stock | Inventory Manager |
| GET | `/api/v1/inventory/low-stock` | List low-stock items | Inventory Manager |

**PostgreSQL Schema:**

```sql
CREATE TABLE inventory (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id        UUID UNIQUE REFERENCES products(id),
    variant_id        UUID REFERENCES product_variants(id),
    sku               VARCHAR(50) UNIQUE NOT NULL,
    quantity           INT NOT NULL DEFAULT 0,
    reserved           INT NOT NULL DEFAULT 0,
    low_stock_threshold INT DEFAULT 10,
    warehouse_id       UUID REFERENCES warehouses(id),
    location           VARCHAR(100), -- aisle, shelf, bin
    is_active          BOOLEAN DEFAULT true,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE inventory_reservations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id          UUID REFERENCES orders(id),
    product_id        UUID REFERENCES products(id),
    quantity           INT NOT NULL,
    status            VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'released', 'expired')),
    expires_at        TIMESTAMPTZ DEFAULT NOW() + INTERVAL '30 minutes',
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE warehouses (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(100) NOT NULL,
    code              VARCHAR(10) UNIQUE NOT NULL,
    address           TEXT,
    is_active         BOOLEAN DEFAULT true
);

CREATE TABLE restock_requests (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id        UUID REFERENCES products(id),
    quantity           INT NOT NULL,
    status            VARCHAR(20) DEFAULT 'pending',
    priority          VARCHAR(10) DEFAULT 'normal',
    created_by        UUID REFERENCES users(id),
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    fulfilled_at      TIMESTAMPTZ
);
```

**Available Quantity Calculation:**

```sql
-- View for available stock
CREATE VIEW available_stock AS
SELECT 
    i.product_id,
    i.quantity - i.reserved AS available
FROM inventory i;
```

**Kafka:**

| Topic | Producer | Consumer | When |
|-------|----------|----------|------|
| `inventory-reserved` | Inventory Service | Order Service | Inventory reserved |
| `inventory-released` | Inventory Service | Order Service | Reservation expired |
| `inventory-updated` | Inventory Service | Product Catalog, Analytics | Stock change |
| `low-stock-alert` | Inventory Service | Notification, Inventory Manager | Below threshold |

---

### 5.4 Notification Service

**Purpose:** Multi-channel notifications (email, SMS, push, in-app).

**Responsibilities:**
- Send emails (via SendGrid/SES)
- Send SMS (via Twilio)
- Send push notifications (via Firebase)
- Notification preferences
- Notification templates

**REST APIs:**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/notifications/send` | Send notification |
| GET | `/api/v1/notifications/types` | Get supported types |
| POST | `/api/v1/notifications/preferences` | Update user preferences |

**Kafka:**

- Consumer: `notification-requested`
- Producer: `notification-sent`, `notification-failed`

---

### 5.5 Analytics Service

**Purpose:** Business intelligence, user behavior, sales analytics.

**Responsibilities:**
- User behavior tracking
- Sales analytics
- Conversion funnel analysis
- Product performance
- Custom dashboards

**Kafka:**

- Consumer: All events (`order-created`, `payment-success`, `cart-updated`, `product-viewed`, etc.)
- Producer: `analytics-report-generated`

**Storage:** Time-series data in PostgreSQL or OpenSearch.

---

### 5.6 Shopping Assistant Service (Existing)

**Purpose:** AI-powered shopping assistant using Gemini.

**Integration:** Already exists as a separate service. Add Kafka consumer for user queries.

---

## 6. PostgreSQL Database Design

### 6.1 Entity-Relationship Diagram

```
┌───────────┐       ┌───────────────┐       ┌───────────┐
│   users   │───1:N──│  addresses    │       │ products   │
└───────────┘       └───────────────┘       └───────────┘
      │                                            │
      │                                            │
      1:N                                        1:N
      │                                            │
      ▼                                            ▼
┌───────────┐       ┌───────────────┐       ┌───────────────┐
│  orders   │───1:N──│ order_items   │───N:1──│  order_items  │
└───────────┘       └───────────────┘       └───────────────┘
      │                                            │
      │                                            │
      1:1                                         1:1
      ▼                                            │
┌───────────┐                                     ▼
│ payments  │                               ┌───────────────┐
└───────────┘                               │  inventory    │
                                            └───────────────┘

┌───────────┐       ┌───────────────┐       ┌───────────────┐
│   cart    │───1:N──│  cart_items   │       │  categories   │
└───────────┘       └───────────────┘       └───────────────┘

┌───────────┐       ┌───────────────┐       ┌───────────────┐
│ products   │───N:N──│  categories   │       │  shipments    │
└───────────┘       └───────────────┘       └───────────────┘

┌───────────┐       ┌───────────────┐
│  audit_log │       │  transactions  │
└───────────┘       └───────────────┘
```

### 6.2 Complete PostgreSQL Schema

```sql
-- ============================================================
-- SCHEMA: public (shared)
-- ============================================================

-- CATEGORIES
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id   UUID REFERENCES categories(id),
    sort_order INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_parent ON categories(parent_id);

-- SUB-CATEGORIES
CREATE TABLE subcategories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    category_id UUID REFERENCES categories(id),
    parent_id   UUID REFERENCES subcategories(id),
    sort_order  INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- PRODUCTS
CREATE TABLE products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description      TEXT,
    sku             VARCHAR(50) UNIQUE NOT NULL,
    brand           VARCHAR(100),
    price           NUMERIC(10,2) NOT NULL,
    compare_at_price NUMERIC(10,2), -- original price for discount display
    currency_code   VARCHAR(3) DEFAULT 'USD',
    category_id     UUID REFERENCES categories(id),
    subcategory_id  UUID REFERENCES subcategories(id),
    weight_kg       NUMERIC(10,3),
    dimensions      JSONB, -- { "length": 10, "width": 5, "height": 3, "unit": "cm" }
    attributes       JSONB, -- { "color": "red", "size": "M", "material": "cotton" }
    image_urls      TEXT[], -- S3 URLs
    thumbnail_url    TEXT,
    is_active       BOOLEAN DEFAULT true,
    is_featured     BOOLEAN DEFAULT false,
    metadata        JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_products_sku ON products(sku) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_active ON products(is_active) WHERE is_active = true;

-- PRODUCT VARIANTS
CREATE TABLE product_variants (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id        UUID REFERENCES products(id) ON DELETE CASCADE,
    name              VARCHAR(100) NOT NULL,
    sku               VARCHAR(50) UNIQUE NOT NULL,
    price_adjustment  NUMERIC(10,2) DEFAULT 0,
    weight_adjustment NUMERIC(10,3) DEFAULT 0,
    attributes        JSONB, -- e.g., { "color": "red", "size": "XL" }
    image_url         TEXT,
    is_active         BOOLEAN DEFAULT true,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

-- USERS
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    keycloak_id     VARCHAR(255) UNIQUE,
    email           VARCHAR(255) UNIQUE NOT NULL,
    username        VARCHAR(100) UNIQUE NOT NULL,
    password_hash   VARCHAR(255), -- only if not using Keycloak
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    phone           VARCHAR(20),
    avatar_url      TEXT,
    role            VARCHAR(30) DEFAULT 'customer' CHECK (role IN (
                        'customer', 'admin', 'inventory_manager', 'support'
                    )),
    is_active       BOOLEAN DEFAULT true,
    email_verified  BOOLEAN DEFAULT false,
    preferences     JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

-- AUDIT LOG (all changes)
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name      VARCHAR(100) NOT NULL,
    record_id       UUID NOT NULL,
    action          VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values      JSONB,
    new_values      JSONB,
    changed_by      UUID REFERENCES users(id),
    changed_at      TIMESTAMPTZ DEFAULT NOW()
);

-- SHOPPING CARTS
CREATE TABLE shopping_carts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    session_id      VARCHAR(255),
    status          VARCHAR(20) DEFAULT 'active',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

-- CART ITEMS
CREATE TABLE cart_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id         UUID REFERENCES shopping_carts(id) ON DELETE CASCADE,
    product_id      UUID REFERENCES products(id),
    variant_id      UUID REFERENCES product_variants(id),
    quantity         INT NOT NULL CHECK (quantity > 0),
    price_snapshot  NUMERIC(10,2),
    added_at        TIMESTAMPTZ DEFAULT NOW()
);

-- ORDERS
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    status          VARCHAR(30) DEFAULT 'pending' CHECK (status IN (
                        'pending', 'confirmed', 'processing', 'shipped', 
                        'delivered', 'cancelled', 'refunded', 'partially_shipped'
                    )),
    total_amount    NUMERIC(10,2),
    subtotal        NUMERIC(10,2),
    tax_amount      NUMERIC(10,2),
    shipping_cost   NUMERIC(10,2),
    currency         VARCHAR(3) DEFAULT 'USD',
    shipping_address JSONB,
    billing_address  JSONB,
    payment_method  VARCHAR(50),
    payment_id      VARCHAR(255),
    idempotency_key VARCHAR(255) UNIQUE,
    notes           TEXT,
    coupon_code     VARCHAR(50),
    discount_amount NUMERIC(10,2),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    cancelled_at    TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at);

-- ORDER ITEMS
CREATE TABLE order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id      UUID REFERENCES products(id),
    variant_id      UUID REFERENCES product_variants(id),
    product_name    VARCHAR(255),
    product_sku     VARCHAR(50),
    quantity         INT NOT NULL CHECK (quantity > 0),
    unit_price       NUMERIC(10,2),
    total_price      NUMERIC(10,2),
    tax_applied      NUMERIC(10,2) DEFAULT 0,
    discount_applied NUMERIC(10,2) DEFAULT 0,
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

-- PAYMENTS
CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID REFERENCES orders(id),
    user_id         UUID REFERENCES users(id),
    amount          NUMERIC(10,2) NOT NULL,
    currency         VARCHAR(3) DEFAULT 'USD',
    status          VARCHAR(30) DEFAULT 'pending' CHECK (status IN (
                        'pending', 'processing', 'succeeded', 'failed', 
                        'refunded', 'partially_refunded'
                    )),
    payment_method   VARCHAR(50),
    payment_provider VARCHAR(50),
    provider_payment_id VARCHAR(255),
    provider_customer_id VARCHAR(255),
    error_message    TEXT,
    idempotency_key  VARCHAR(255) UNIQUE,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    refunded_at      TIMESTAMPTZ
);

-- USER ADDRESSES
CREATE TABLE user_addresses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    label           VARCHAR(50),
    street_address  TEXT NOT NULL,
    city            VARCHAR(100) NOT NULL,
    state           VARCHAR(100),
    postal_code     VARCHAR(20),
    country         VARCHAR(100) NOT NULL,
    is_default       BOOLEAN DEFAULT false,
    is_billing      BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- INVENTORY
CREATE TABLE inventory (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id       UUID UNIQUE REFERENCES products(id),
    variant_id       UUID REFERENCES product_variants(id),
    sku              VARCHAR(50) UNIQUE NOT NULL,
    quantity         INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved         INT NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    low_stock_threshold INT DEFAULT 10,
    location         VARCHAR(100),
    warehouse_id     UUID REFERENCES warehouses(id),
    is_active        BOOLEAN DEFAULT true,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

-- AVAILABLE STOCK VIEW
CREATE VIEW available_stock AS
SELECT 
    i.product_id,
    i.variant_id,
    i.quantity - i.reserved AS available,
    i.quantity,
    i.reserved,
    i.low_stock_threshold
FROM inventory i;

-- WAREHOUSES
CREATE TABLE warehouses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(10) UNIQUE NOT NULL,
    address         TEXT,
    is_active       BOOLEAN DEFAULT true
);

-- SHIPMENTS
CREATE TABLE shipments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID REFERENCES orders(id),
    carrier          VARCHAR(50) NOT NULL,
    tracking_number  VARCHAR(100),
    status           VARCHAR(30) DEFAULT 'pending',
    origin_address   JSONB,
    destination_address JSONB,
    weight_kg        NUMERIC(10,2),
    cost             NUMERIC(10,2),
    estimated_delivery TIMESTAMPTZ,
    actual_delivery   TIMESTAMPTZ,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

-- TRANSACTIONS (for payment audit)
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id      UUID REFERENCES payments(id),
    type            VARCHAR(30) NOT NULL CHECK (type IN (
                        'charge', 'refund', 'authorization', 'capture', 'void'
                    )),
    amount          NUMERIC(10,2) NOT NULL,
    currency         VARCHAR(3) DEFAULT 'USD',
    status          VARCHAR(30) NOT NULL,
    provider_transaction_id VARCHAR(255),
    request         JSONB, -- original request
    response        JSONB, -- provider response
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 7. Kafka Event-Driven Architecture

### 7.1 Topic Definitions

| Topic Name | Partitions | Replication Factor | Retention | Cleanup Policy |
|------------|-----------|--------------------|-----------|----------------|
| `order-created` | 3 | 3 | 7 days | delete |
| `order-cancelled` | 3 | 3 | 7 days | delete |
| `order-status-changed` | 3 | 3 | 7 days | compact |
| `payment-success` | 3 | 3 | 30 days | compact |
| `payment-failed` | 3 | 3 | 30 days | compact |
| `inventory-updated` | 3 | 3 | 7 days | compact |
| `inventory-reserved` | 3 | 3 | 7 days | delete |
| `shipment-created` | 3 | 3 | 7 days | compact |
| `product-created` | 3 | 3 | 30 days | compact |
| `product-updated` | 3 | 3 | 30 days | compact |
| `product-deleted` | 3 | 3 | 7 days | delete |
| `notification-requested` | 3 | 3 | 7 days | delete |
| `cart-updated` | 3 | 3 | 1 day | delete |
| `user-registered` | 3 | 3 | 30 days | compact |
| `user-updated` | 3 | 3 | 7 days | compact |
| `analytics-event` | 3 | 3 | 90 days | delete |
| `low-stock-alert` | 3 | 3 | 7 days | delete |

### 7.2 Topic Detail: `order-created`

**Producer:** Checkout Service (after successful `PlaceOrder`)

**Consumers:**
1. Order Service — Create order record in PostgreSQL
2. Inventory Service — Reserve inventory
3. Notification Service — Send order confirmation
4. Analytics Service — Track order metric

**Payload:**

```json
{
  "event_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "event_type": "order.created",
  "version": "1.0",
  "timestamp": "2026-07-23T14:30:00Z",
  "order": {
    "order_id": "uuid-from-checkout",
    "user_id": "user-uuid",
    "user_email": "customer@example.com",
    "items": [
      {
        "product_id": "prod-uuid",
        "variant_id": "var-uuid",
        "sku": "PROD-SKU-001",
        "quantity": 2,
        "unit_price": 29.99,
        "total": 59.98
      }
    ],
    "total": 59.98,
    "currency": "USD",
    "shipping_address": {
      "street": "123 Main St",
      "city": "Springfield",
      "state": "IL",
      "zip": "62701",
      "country": "US"
    },
    "payment_method": "card"
  },
  "metadata": {
    "source": "checkout-service",
    "correlation_id": "trace-id-from-otel"
  }
}
```

**Retry Strategy:**

| Attempt | Delay | Action |
|---------|-------|--------|
| 1 | 1s | Immediate retry |
| 2 | 5s | Retry |
| 3 | 30s | Retry |
| 4 | 2m | Retry |
| 5 | 10m | Send to DLQ |

**Dead Letter Queue:** `order-created-dlq` — Dead letter topic. Manual inspection/replay.

### 7.3 Topic Detail: `payment-success`

**Producer:** Payment Service

**Consumers:**
1. Order Service — Update order status to `paid`
2. Notification Service — Send payment confirmation email
3. Analytics Service — Record successful payment

**Payload:**

```json
{
  "event_id": "uuid",
  "event_type": "payment.success",
  "version": "1.0",
  "timestamp": "2026-07-23T14:30:05Z",
  "payment": {
    "payment_id": "uuid",
    "order_id": "uuid",
    "amount": 59.98,
    "currency": "USD",
    "status": "succeeded",
    "payment_method": "card",
    "provider": "stripe",
    "provider_payment_intent": "pi_3N...",
    "transaction_id": "txn_..."
  },
  "metadata": {
    "idempotency_key": "idem-key"
  }
}
```

### 7.4 Topic Detail: `inventory-updated`

**Producer:** Inventory Service

**Consumers:**
1. Product Catalog Service — Update product stock display
2. Analytics Service — Track inventory changes
3. Notification Service — Low-stock alerts

**Payload:**

```json
{
  "event_id": "uuid",
  "event_type": "inventory.updated",
  "version": "1.0",
  "timestamp": "2026-07-23T14:30:10Z",
  "inventory": {
    "product_id": "prod-uuid",
    "variant_id": "var-uuid",
    "sku": "PROD-SKU-001",
    "old_quantity": 100,
    "new_quantity": 98,
    "reserved": 2,
    "available": 96,
    "warehouse": "WH-NORTH",
    "change_type": "sale"
  }
}
```

### 7.5 Topic Detail: `notification-requested`

**Producer:** Any service needing to send a notification

**Consumers:**
1. Email Service (Python) — Send transactional email
2. Notification Service — Send push/SMS

**Payload:**

```json
{
  "event_id": "uuid",
  "event_type": "notification.requested",
  "version": "1.0",
  "timestamp": "2026-07-23T14:30:00Z",
  "notification": {
    "type": "email",
    "to": "customer@example.com",
    "template": "order-confirmation",
    "data": {
      "order_id": "ORDER-123",
      "customer_name": "John Doe",
      "total": "$59.98",
      "items": [
        { "name": "Product 1", "qty": 2 }
      ]
    }
  }
}
```

### 7.6 Topic Detail: `order-cancelled`

**Producer:** Order/Checkout Service

**Consumers:**
1. Inventory — Release reserved stock
2. Payment — Process refund if paid
3. Notification — Send cancellation email
4. Analytics — Track cancellation

---

## 8. API Gateway Design

### 8.1 Architecture

Use **Kong Gateway** (OSS) or **Envoy Proxy** with custom configurations.

**Why Kong:**
- Proven in production (many enterprises)
- Plugin ecosystem (JWT, Rate Limiting, ACL)
- Built-in gRPC support (Kong 3.x+)
- Admin API for dynamic configuration

**Why Envoy (alternative):**
- Istio-native (if using service mesh)
- L4/L7 proxy capabilities
- Lower latency

### 8.2 Gateway Configuration

```yaml
# kong.yml - Declarative config
_format_version: "3.0"
services:
  - name: frontend-service
    host: frontend
    port: 8080
    protocol: http
    routes:
      - name: frontend-route
        paths:
          - /
        strip_path: false
        plugins:
          - name: rate-limiting
            config:
              minute: 100
              hour: 1000
  
  - name: auth-service
    host: keycloak
    port: 8080
    protocol: http
    routes:
      - name: auth-login
        paths:
          - /auth/login
        plugins:
          - name: rate-limiting
            config:
              minute: 20  # Stricter for auth
  
  - name: api-service
    host: api-gateway-internal
    port: 8080
    protocol: http
    routes:
      - name: api-v1
        paths:
          - /api/v1
        strip_path: false
        plugins:
          - name: jwt
            config:
              claims_to_verify: ["iss", "exp"]
              maximum_expiration: 3600
              key_claim_name: "kid"
          - name: cors
          - name: rate-limiting
            config:
              minute: 60
  
  - name: product-catalog
    host: productcatalogservice
    port: 3550
    protocol: grpc
    routes:
      - name: product-grpc
        paths:
          - /hipstershop.ProductCatalogService
        strip_path: true
        plugins:
          - name: jwt
          - name: acl
            config:
              allow: ["customer", "admin", "inventory_manager"]
```

### 8.3 JWT Validation Flow

```
1. Client → POST /api/v1/auth/login
2. Keycloak → Returns JWT + Refresh Token
3. Client stores JWT in httpOnly cookie (or localStorage)
4. Every API request includes: Authorization: Bearer <jwt>
5. API Gateway validates JWT:
   - Checks signature (RS256, using Keycloak public key)
   - Checks expiry
   - Checks audience
   - Extracts roles from `realm_access.roles`
6. If valid → Forward to backend with `X-User-ID` and `X-User-Roles` headers
7. If invalid → Return 401
```

### 8.4 Rate Limiting

| Route | Limit | Window | Strategy |
|-------|-------|--------|---------|
| `/api/v1/products` | 100 | 60s | Per-IP + Per-User |
| `/api/v1/orders` | 30 | 60s | Per-User |
| `/api/v1/cart` | 100 | 60s | Per-User |
| `/auth/login` | 20 | 60s | Per-IP |
| `/auth/register` | 5 | 60s | Per-IP |
| `/api/v1/admin/*` | 10 | 60s | Per-Admin |

### 8.5 Load Balancing

- Round-robin for stateless services
- Least-connections for stateful services (cart, shipping)
- Consistent hashing for session affinity (Redis)
- Health checks every 10s (HTTP 200 on `/health`)

---

## 9. Authentication & Authorization

### 9.1 Keycloak Configuration

**Realm:** `online-boutique`

**Clients:**
- `frontend` (public) — for browser-based login
- `backend-service` (confidential) — for service-to-service

**Roles:**
- `customer` — Default for all registered users
- `admin` — Full access
- `inventory_manager` — Manage stock, products
- `support` — View orders, manage returns

### 9.2 Flow: Login

```
1. User → POST /api/v1/auth/login (email + password)
2. Frontend → POST /auth/realms/online-boutique/protocol/openid-connect/token
3. Keycloak validates credentials
4. Returns:
   {
     "access_token": "jwt...",
     "refresh_token": "...",
     "expires_in": 3600,
     "refresh_expires_in": 86400,
     "token_type": "Bearer",
     "not-before-policy": 0,
     "session_state": "uuid"
   }
5. Frontend stores:
   - access_token: httpOnly cookie (secure, sameSite=strict)
   - refresh_token: httpOnly cookie (separate)
   - session_state: LocalStorage (for SPA detection)
6. Frontend redirects to /dashboard
```

### 9.3 Flow: Registration

```
1. User → POST /api/v1/auth/register
2. Frontend → POST /auth/admin/realms/online-boutique/users (admin token)
   OR → POST /auth/realms/online-boutique/protocol/openid-connect/registrations
3. Keycloak creates user
4. User Service creates profile in PostgreSQL
5. Returns 201 Created
```

### 9.4 Flow: Token Refresh

```
1. Frontend detects token expiry (via 401 response)
2. Frontend → POST /auth/realms/online-boutique/protocol/openid-connect/token
   Body: grant_type=refresh_token&refresh_token=...
3. Keycloak validates refresh token
4. Returns new access_token + refresh_token
```

### 9.5 RBAC Implementation

**gRPC Interceptor (Go):**

```go
// auth_interceptor.go
func AuthUnaryClientInterceptor(ctx context.Context, method string, 
    req interface{}, reply interface{}, cc *grpc.ClientConn, 
    invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    
    // Extract JWT from context metadata
    md, _ := metadata.FromIncomingContext(ctx)
    jwt := md.Get("authorization")
    
    // Validate JWT with Keycloak
    token, err := validateToken(jwt[0])
    if err != nil {
        return status.Errorf(codes.Unauthenticated, "invalid token")
    }
    
    // Extract roles
    roles := extractRoles(token)
    
    // Check permission for this method
    if !hasPermission(method, roles) {
        return status.Errorf(codes.PermissionDenied, "insufficient permissions")
    }
    
    return invoker(ctx, method, req, reply, cc, opts...)
}
```

**Permission Matrix:**

| Resource | Method | Customer | Admin | Inventory Manager | Support |
|----------|--------|----------|-------|-------------------|---------|
| Products | GET | ✓ | ✓ | ✓ | ✓ |
| Products | POST | ✗ | ✓ | ✓ | ✗ |
| Products | PUT | ✗ | ✓ | ✓ | ✗ |
| Products | DELETE | ✗ | ✓ | ✗ | ✗ |
| Orders | GET (own) | ✓ | ✓ | ✗ | ✓ |
| Orders | GET (all) | ✗ | ✓ | ✗ | ✓ |
| Orders | PUT | ✗ | ✓ | ✗ | ✓ |
| Inventory | GET | ✗ | ✓ | ✓ | ✗ |
| Inventory | PUT | ✗ | ✓ | ✓ | ✗ |
| Users | GET | ✗ | ✓ | ✗ | ✗ |
| Users | PUT | ✗ | ✓ | ✗ | ✗ |

### 9.6 User ↔ Keycloak Mapping

```go
// KeycloakUser struct
type KeycloakUser struct {
    ID               string   `json:"id"`
    CreatedTimestamp int64    `json:"createdTimestamp"`
    Username         string   `json:"username"`
    Enabled          bool     `json:"enabled"`
    Email            string   `json:"email"`
    EmailVerified     bool     `json:"emailVerified"`
    FirstName        string   `json:"firstName"`
    LastName         string   `json:"lastName"`
    Attributes        map[string][]string `json:"attributes"`
    RealmRoles       []string `json:"realmRoles"`
}

// Sync: When user logs in through Keycloak, 
// User Service checks if user exists in PostgreSQL.
// If not, creates new user record.
// Updates `last_login_at` on each login.
```

---

## 10. OpenSearch Search Design

### 10.1 Index Architecture

**Product Index:**

```json
PUT /products
{
  "settings": {
    "analysis": {
      "analyzer": {
        "autocomplete": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "edge_ngram"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "name": {
        "type": "text",
        "analyzer": "autocomplete",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "description": { "type": "text" },
      "category": { "type": "keyword" },
      "price": { "type": "float" },
      "rating": { "type": "float" },
      "in_stock": { "type": "boolean" },
      "created_at": { "type": "date" },
      "tags": { "type": "keyword" },
      "attributes": { "type": "nested" }
    }
  }
}
```

### 10.2 Search Queries

**Full-text search:**

```json
POST /products/_search
{
  "query": {
    "bool": {
      "must": [
        { "multi_match": {
          "query": "blue running shoes",
          "fields": ["name^3", "description", "category"],
          "fuzziness": "AUTO"
        }}
      ],
      "filter": [
        { "term": { "category": "footwear" }},
        { "range": { "price": { "gte": 10, "lte": 100 }}}
      ]
    }
  },
  "sort": [
    { "_score": "desc" },
    { "price": "asc" }
  ]
}
```

### 10.3 Autocomplete

```json
POST /products/_search
{
  "suggest": {
    "product-suggest": {
      "prefix": "run",
      "completion": {
        "field": "suggest",
        "fuzzy": {
          "fuzziness": 2
        }
      }
    }
  }
}
```

### 10.4 Filtering

- **Category:** `{ "term": { "category": "clothing" }}`
- **Price range:** `{ "range": { "price": { "gte": 10, "lte": 50 }}}`
- **In stock:** `{ "term": { "in_stock": true }}`
- **Rating:** `{ "range": { "rating": { "gte": 4 }}}`

### 10.5 Ranking

- **Boost by:** Popularity score, rating, recency
- **Function score query:**

```json
{
  "function_score": {
    "query": { ... },
    "functions": [
      { "field_value_factor": {
        "field": "popularity",
        "factor": 1.5
      }},
      { "gauss": {
        "date": {
          "origin": "now",
          "scale": "30d",
          "decay": 0.5
        }
      }}
    ]
  }
}
```

---

## 11. Object Storage (S3) Design

### 11.1 Bucket Structure

```
s3://products/
  ├── images/
  │   ├── {product-id}/
  │   │   ├── original.jpg       (2048x2048)
  │   │   ├── large.jpg          (1024x1024)
  │   │   ├── medium.jpg         (512x512)
  │   │   ├── small.jpg          (256x256)
  │   │   └── thumbnail.jpg       (128x128)
  │   └── ...
  ├── documents/
  │   ├── spec-sheets/
  │   └── manuals/
  └── invoices/
      └── {year}/{month}/
          ├── {order-id}.pdf
          └── ...

s3://static/
  ├── css/
  ├── js/
  └── fonts/
```

### 11.2 Presigned URLs

- Generate presigned URLs for uploads (5 min expiry)
- Serve images directly via CDN (CloudFront)
- Protected: only authenticated users can view invoices

---

## 12. Service Interaction Diagrams

### 12.1 Service Map

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INTERNET                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                          ┌─────────────────┐
                          │   API Gateway    │
                          │  (Kong/Envoy)    │
                          └────────┬────────┘
                                   │
                                   ▼
                          ┌─────────────────┐
                          │    Frontend      │
                          │  (Go, HTTP:8080) │
                          └────────┬────────┘
                                   │
                                   ├──────────────────────┐
                                   │ HTTP / gRPC          │
                                   ▼                      ▼
                          ┌──────────────────┐   ┌──────────────────┐
                          │   Keycloak       │   │  User Service    │
                          │  (Auth, OIDC)    │   │  (PostgreSQL)    │
                          └──────────────────┘   └──────────────────┘
                                   │
                                   │ gRPC
                                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│ Product       │   │  Cart         │   │  Checkout    │   │  Payment     │
│ Catalog       │◄──│  Service      │◄──│  Service     │──►│  Service     │
│ (PostgreSQL)   │   │  (Redis)     │   │  (PostgreSQL) │   │  (PostgreSQL)│
└──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘
       │                   │                   │                   │
       └─────────┬─────────┘                    │                   │
                 │                              │                   │
                 ▼                              ▼                   ▼
          ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
          │  OpenSearch  │   │   Kafka      │   │  Order       │   │  Inventory   │
          │  (Search)    │   │  (Events)    │   │  Service     │   │  Service     │
          └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘
                                │       │               │               │
                                │       │               │               │
                                ▼       ▼               ▼               ▼
                          ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
                          │  Email       │   │  Notification │   │  Analytics    │
                          │  Service      │   │  Service      │   │  Service      │
                          └──────────────┘   └──────────────┘   └──────────────┘
```

### 12.2 Data Flow

```
HTTP (External):     User → Browser → API Gateway → Frontend
gRPC (Internal):     Frontend → Product Catalog, Cart, Checkout
                     Checkout → Payment, Shipping, Email, Cart
                     Product Catalog → OpenSearch (search index)
                     Cart Service → Redis (cache)
Kafka (Async):        Checkout → Kafka (order-created)
                     Payment → Kafka (payment-success)
                     Inventory → Kafka (inventory-updated)
                     Shipping → Kafka (shipment-created)
                     Email ← Kafka (notification-requested)
```

---

## 13. Sequence Diagrams

### 13.1 Order Placement Sequence

```
┌─────┐   ┌──────────────┐   ┌────────────┐   ┌────────────┐   ┌──────────────┐   ┌────────┐   ┌────────────┐
│User │   │   Frontend   │   │  Checkout   │   │  Cart      │   │  Product     │   │Payment  │   │  Shipping   │
│     │   │              │   │  Service    │   │  Service   │   │  Catalog     │   │Service  │   │  Service    │
└──┬──┘   └──────┬───────┘   └─────┬──────┘   └─────┬──────┘   └──────┬───────┘   └───┬────┘   └──────┬───────┘
   │            │                 │            │              │              │            │
   │POST /cart/ │                 │            │              │              │            │
   │checkout    │                 │            │              │              │            │
   ├────────────┼─────────────────┤            │              │              │            │
   │            │ PlaceOrder()    │            │              │              │            │
   │            ├─────────────────┼────────────┤              │              │            │
   │            │                 │ GetCart()   │              │              │            │
   │            │                 ├────────────┼─────► Cart  │              │            │
   │            │                 │            │  Redis      │              │            │
   │            │                 │◄───────────┴───── Cart   │              │            │
   │            │                 │            │  Items      │              │            │
   │            │                 ├────────────┼─────►       │              │            │
   │            │                 │ GetProduct │             │              │            │
   │            │                 │───────────►│  Catalog    │Product Data  │            │
   │            │                 │◄───────────┤             │              │            │
   │            │                 │            │             │              │            │
   │            │                 ├────────────┼────────────►──────────────┤            │
   │            │                 │            │             │   Convert   │            │
   │            │                 │            │             │   Currency  │            │
   │            │                 │            │             │  Service    │            │
   │            │                 │◄───────────┴────────────┴──────────────┘            │
   │            │                 │            │             │              │            │
   │            │                 ├────────────┼────────────►──────────────┤            │
   │            │                 │            │             │  Get Quote  │            │
   │            │                 │            │  Shipping   │──────────────┤            │
   │            │                 │◄───────────┤             │              │            │
   │            │                 │            │             │              │            │
   │            │                 ├────────────┼────────────►──────────────┤            │
   │            │                 │            │             │   Charge    │            │
   │            │                 │            │  Payment    │──────────────┤            │
   │            │                 │◄─────────┬─┴───────────┴───────────────┤            │
   │            │                 │          │  Transaction ID            │            │
   │            │                 │          │                            │            │
   │            │                 ├────────────┼───────────────────────────┤            │
   │            │                 │            │             │   Ship      │            │
   │            │                 │            │  Shipping   │──────────────┤            │
   │            │                 │◄─────────┬─┴────────────┴───────────────┤            │
   │            │                 │          │  Tracking ID                │            │
   │            │                 │          │                            │            │
   │            │                 ├────────────┼───────────────────────────┤            │
   │            │                 │            │             │   Email     │            │
   │            │                 │  EmptyCart │             │             │            │
   │            │                 ├────────────┤             │             │            │
   │            │                 │◄───────────┤  Order      │             │            │
   │            │                 │            │  Result     │             │            │
   │            │                 │            │             │             │            │
   │            │  Order          │            │             │             │            │
   │            │  Confirmation   │            │             │             │            │
   │◄───────────┴─────────────────┴────────────┴────────────┴─────────────┴─────────────┘
```

### 13.2 Payment Sequence

```
┌─────────┐   ┌──────────────┐   ┌────────────┐   ┌──────────────┐   ┌────────────┐
│  User   │   │  Frontend    │   │  Checkout   │   │  Payment     │   │   Kafka    │
│         │   │              │   │  Service    │   │  Service    │   │            │
└────┬────┘   └──────┬───────┘   └──────┬─────┘   └──────┬───────┘   └──────┬──────┘
     │                │                │                │                │
     │ POST /checkout  │                │                │                │
     ├─────────────────┤                │                │                │
     │                 │ PlaceOrder()   │                │                │
     │                 ├────────────────┼────────────────┤                │
     │                 │                │                │                │
     │                 │                │                │                │
     │                 │                │                │                │
     │                 │                │                │                │
     │                 │   Charge()    │                │                │
     │                 ├───────────────┤                │                │
     │                 │               │                │                │
     │                 │               │  Stripe API     │                │
     │                 │               │  (external)    │                │
     │                 │               ├───────────────►│                │
     │                 │               │                │                │
     │                 │               │  pi_3N...      │                │
     │                 │               │◄───────────────┤                │
     │                 │               │                │                │
     │                 │               │  Record in     │                │
     │                 │               │  PostgreSQL    │                │
     │                 │               │  (payments)    │                │
     │                 │               ├────────────────┤                │
     │                 │               │                │                │
     │                 │               │                │                │
     │                 │               ├───────────────┤                │
     │                 │               │  payment-success │              │
     │                 │               │──────────────────┼───────────────┤
     │                 │               │                │                │
     │                 │               │                │                │
     │                 │               │                │                │
     │                 │               │  payment-success │              │
     │                 │               │──────────────────┤                │
     │                 │               │                │                │
     │                 │  Transaction   │                │                │
     │                 │  ID          │                │                │
     │◄────────────────┤               │                │                │
     │                 │               │                │                │
```

### 13.3 Inventory Update Sequence

```
┌─────────┐   ┌──────────────┐   ┌────────────┐   ┌────────────┐   ┌────────────┐
│  Checkout │   │  Inventory   │   │   PostgreSQL │   │   OpenSearch│   │   Kafka    │
│  Service  │   │  Service     │   │             │   │            │   │            │
└─────┬────┘   └──────┬───────┘   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
      │                │                │                │                │
      │                │                │                │                │
      │                │                │                │                │
      │  reserve()     │                │                │                │
      ├────────────────┤                │                │                │
      │                │                │                │                │
      │                │  UPDATE        │                │                │
      │                │  (reserved++)  │                │                │
      │                ├────────────────►│                │                │
      │                │                │                │                │
      │                │  inventory-    │                │                │
      │                │  updated      │                │                │
      │                ├─────────────────────────────────────┤           │
      │                │                │                │                │
      │                │                │                │                │
      │                │  Confirm       │                │                │
      │                │  OK           │                │                │
      │◄───────────────┤                │                │                │
      │                │                │                │                │
      │                │                │                │                │
      │                │                │                │                │
      │                │                │                │                │
      │                │  Ref          │                │                │
      │                │  Order        │                │                │
      │                │  (30 min)     │                │                │
      │                │                │                │                │
      │                │  If expired:   │                │                │
      │                │  inventory-    │                │                │
      │                │  released     │                │                │
      │                ├─────────────────────────────────────┤           │
      │                │                │                │                │
```

---

## 14. Deployment Considerations

### 14.1 Kubernetes Namespace Strategy

```yaml
# Namespaces
apiVersion: v1
kind: Namespace
metadata:
  name: online-boutique
---
apiVersion: v1
kind: Namespace
metadata:
  name: online-boutique-infra  # PostgreSQL, Kafka, Redis, OpenSearch
```

### 14.2 Service Mesh

- Use **Istio** (already integrated in kustomize)
- Enable mTLS between all services
- Circuit breakers: max 5 retries, 50ms timeout

### 14.3 Helm Chart Additions

Add these new services to the existing `helm-chart/values.yaml`:

```yaml
# New services for enterprise deployment
userService:
  create: true
  name: userservice
  database: postgresql
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi

orderService:
  create: true
  name: orderservice
  database: postgresql
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi

inventoryService:
  create: true
  name: inventoryservice
  database: postgresql
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi

notificationService:
  create: true
  name: notificationservice
  resources:
    requests:
      cpu: 100m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 128Mi

analyticsService:
  create: true
  name: analyticsservice
  database: postgresql
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 400m
      memory: 512Mi

apiGateway:
  create: true
  name: apigateway
  type: kong
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi

keycloak:
  create: true
  name: keycloak
  database: postgresql
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1
      memory: 2Gi

opensearch:
  create: true
  name: opensearch
  resources:
    requests:
      cpu: 500m
      memory: 2Gi
    limits:
      cpu: 1
      memory: 4Gi

kafka:
  create: true
  name: kafka
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1
      memory: 2Gi

postgresql:
  create: true
  name: postgresql
  resources:
    requests:
      cpu: 500m
      memory: 2Gi
    limits:
      cpu: 1
      memory: 4Gi

objectStorage:
  create: true
  name: minio  # or S3-compatible
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 200m
      memory: 512Mi
```

### 14.4 Health Checks & Readiness

- gRPC health probes (already implemented via `grpc_health_v1`)
- Database connection pool: max 25 connections per service
- Kafka consumer: max 5 concurrent consumers per partition

### 14.5 Resource Sizing

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit | Replicas | 
|---------|------------|----------|---------------|-------------|---------|
| Frontend | 100m | 200m | 64Mi | 128Mi | 2-3 |
| Product Catalog | 100m | 200m | 64Mi | 128Mi | 2-3 |
| Cart | 200m | 300m | 128Mi | 256Mi | 2-3 |
| Checkout | 100m | 200m | 64Mi | 128Mi | 2-3 |
| Payment | 100m | 200m | 128Mi | 256Mi | 2-3 |
| Shipping | 100m | 200m | 64Mi | 128Mi | 2-3 |
| Email | 100m | 200m | 64Mi | 128Mi | 1-2 |
| Recommendation | 100m | 200m | 220Mi | 450Mi | 2-3 |
| Ad | 200m | 300m | 180Mi | 300Mi | 1-2 |
| User | 100m | 200m | 128Mi | 256Mi | 2-3 |
| Order | 100m | 200m | 128Mi | 256Mi | 2-3 |
| Inventory | 100m | 200m | 128Mi | 256Mi | 2-3 |
| Notification | 100m | 200m | 64Mi | 128Mi | 1-2 |
| Analytics | 200m | 400m | 256Mi | 512Mi | 1-2 |
| API Gateway | 200m | 500m | 256Mi | 512Mi | 2-3 |
| Keycloak | 500m | 1 | 1Gi | 2Gi | 2-3 |
| PostgreSQL | 500m | 1 | 2Gi | 4Gi | 2-3 |
| Kafka | 500m | 1 | 1Gi | 2Gi | 3 |
| OpenSearch | 500m | 1 | 2Gi | 4Gi | 2-3 |
| Redis | 200m | 300m | 128Mi | 256Mi | 1-2 |

---

## 15. Design Decisions & Trade-offs

### 15.1 Keep vs Replace

| Decision | Rationale | Trade-off |
|----------|-----------|-----------|
| **Keep gRPC** | Existing services use gRPC; migration would be costly | gRPC requires HTTP/2, service mesh for external |
| **Keep existing services** | Minimal changes required, focus on new capabilities | May miss optimization opportunities |
| **Add PostgreSQL** | Needed for persistent state, audit, analytics | Adds complexity (migrations, connections) |
| **Add Kafka** | Asynchronous order lifecycle, notification | Latency for synchronous operations |
| **Add Keycloak** | Battle-tested, OIDC compliant, RBAC support | Additional infrastructure, learning curve |
| **Add OpenSearch** | Full-text search, autocomplete, ranking | Indexing latency, storage overhead |
| **S3 over local storage** | Scalable, CDN-ready, persistent | Network dependency, added cost |
| **Kong over custom** | Plugin ecosystem, dynamic config | Extra deployment (Kong DB, migrations) |
| **gRPC health probes** | Already implemented | Not as rich as HTTP probes |

### 15.2 Mono-repo vs Poly-repo

**Decision:** Keep mono-repo (current structure).

**Rationale:** Simpler for development, consistent CI/CD, shared protos and libraries.

**Trade-off:** Build times increase with more services. Use `skaffold` for selective builds.

### 15.3 Synchronous vs Async

**Decision:** Hybrid.
- **Synchronous:** gRPC for critical paths (cart retrieval, product lookup)
- **Asynchronous:** Kafka for order lifecycle, notifications, analytics

**Rationale:** User-facing operations need immediate feedback. Backend operations (notifications, analytics) can be async.

**Trade-off:** Eventual consistency for some operations (e.g., inventory may lag by seconds).

### 15.4 Database per Service

**Decision:** Database per service (microservices pattern).

**Rationale:** Loose coupling, independent scaling, separate failure domains.

**Trade-off:** Join queries require service calls. Use Kafka for data synchronization.

### 15.5 Authentication Strategy

**Decision:** Keycloak for auth → User Service for profile (PostgreSQL).

**Rationale:** Separation of concerns. Keycloak handles auth, User Service handles business data.

**Trade-off:** Two user stores to keep in sync. Use Keycloak event listeners to sync.

### 15.6 API Gateway Strategy

**Decision:** Kong Gateway (OSS).

**Rationale:**
- Built-in JWT, rate limiting, CORS plugins
- Declarative config (YAML)
- gRPC support (Kong 3.x+)
- Well-documented for enterprises

**Trade-off:** Extra hop for all traffic. Additional infrastructure to manage.

---

## Appendix A: Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)
- Deploy PostgreSQL with schema migrations
- Deploy Keycloak with realm configuration
- Create User Service (Go, gRPC + PostgreSQL)
- Add JWT validation to Frontend
- Update Frontend with login/register

### Phase 2: Event-Driven (Weeks 3-4)
- Deploy Kafka
- Add Kafka producers/consumers to Checkout, Payment
- Create Order Service
- Create Notification Service (Kafka consumer)
- Create Inventory Service

### Phase 3: Search & Analytics (Weeks 5-6)
- Deploy OpenSearch
- Index products
- Add search endpoint to Product Catalog
- Create Analytics Service
- Add OpenSearch KNN to Recommendations

### Phase 4: Polish (Weeks 7-8)
- API Gateway (Kong)
- S3 object storage for images
- RBAC implementation
- Rate limiting
- Admin dashboard
- Documentation

## Appendix B: Technology Stack

| Component | Technology | Version | Notes |
|-----------|------------|---------|-------|
| **Runtime** | Go | 1.21+ | Existing services |
| | Node.js | 18+ | Currency, Payment |
| | Python | 3.11+ | Email, Recommendation |
| | Java | 17+ | Ad Service |
| | C# | .NET 8 | Cart Service |
| **Database** | PostgreSQL | 15+ | Primary database |
| | Redis | 7+ | Cart cache, session |
| **Messaging** | Kafka | 3.5+ | Event bus |
| **Search** | OpenSearch | 2.11+ | Product search |
| **Auth** | Keycloak | 22+ | OIDC provider |
| **Gateway** | Kong | 3.x+ | API Gateway |
| **Storage** | MinIO | Latest | S3-compatible |
| **Observability** | OpenTelemetry | Latest | Distributed tracing |

## Appendix C: Protobuf Extensions

```protobuf
// extensions to demo.proto for enterprise features
syntax = "proto3";
package hipstershop;

// New service definitions
service AuthService {
    rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse);
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
    rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
}

service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc CreateUser(CreateUserRequest) returns (User);
    rpc UpdateUser(UpdateUserRequest) returns (User);
    rpc GetAddresses(GetAddressesRequest) returns (AddressList);
    rpc AddAddress(AddAddressRequest) returns (Address);
}

service InventoryService {
    rpc CheckStock(CheckStockRequest) returns (StockResponse);
    rpc ReserveInventory(ReserveInventoryRequest) returns (ReserveResponse);
    rpc ReleaseInventory(ReleaseInventoryRequest) returns (Empty);
    rpc GetStockLevel(GetStockLevelRequest) returns (StockLevelResponse);
}

service OrderService {
    rpc CreateOrder(CreateOrderRequest) returns (Order);
    rpc GetOrder(GetOrderRequest) returns (Order);
    rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
    rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (Order);
}

message User {
    string id = 1;
    string keycloak_id = 2;
    string email = 3;
    string username = 4;
    string first_name = 5;
    string last_name = 6;
    string role = 7;
    bool is_active = 8;
}

message Address {
    string id = 1;
    string user_id = 2;
    string label = 3;
    string street_address = 4;
    string city = 5;
    string state = 6;
    string postal_code = 7;
    string country = 8;
    bool is_default = 9;
}

message StockLevel {
    string product_id = 1;
    string variant_id = 2;
    int32 quantity = 3;
    int32 reserved = 4;
    int32 available = 5;
    string sku = 6;
}

message Order {
    string id = 1;
    string user_id = 2;
    string status = 3;
    repeated OrderItem items = 4;
    Address shipping_address = 5;
    Money total = 6;
    string currency = 7;
    string created_at = 8;
    string updated_at = 9;
}
```