# Service Analysis

Detailed analysis of each existing service in the Google Online Boutique.

---

## 1. Frontend Service

### Current Responsibility
- HTTP web server serving HTML templates (server-side rendering)
- Session management via cookies (shop_session-id, shop_currency)
- BFF (Backend for Frontend) - aggregates calls to backend gRPC services
- Currency conversion coordination
- Product browsing, cart management, checkout flow
- Static asset serving (/static/)
- Health endpoint (/healthz)

### Current Communication Pattern
- **Inbound**: HTTP/1.1 (browser clients)
- **Outbound**: gRPC to all backend services
- **Service Discovery**: Kubernetes DNS (env vars for service addresses)
- **Protocol**: gRPC with Protocol Buffers (generated from protos/demo.proto)

### Current APIs (HTTP Endpoints)
| Method | Path | Description |
|--------|------|-------------|
| GET | / | Home page with product listing |
| GET | /product/{id} | Product detail page |
| GET | /cart | View shopping cart |
| POST | /cart | Add item to cart |
| POST | /cart/empty | Empty cart |
| POST | /setCurrency | Change display currency |
| GET | /logout | Clear session cookies |
| POST | /cart/checkout | Place order (calls CheckoutService) |
| GET | /assistant | AI shopping assistant page |
| POST | /bot | Chat bot proxy to ShoppingAssistant |
| GET | /product-meta/{ids} | Product metadata (JSON) |
| GET | /_healthz | Health check |

### Current Dependencies
- ProductCatalogService (ListProducts, GetProduct, SearchProducts)
- CurrencyService (GetSupportedCurrencies, Convert)
- CartService (GetCart, AddItem, EmptyCart)
- RecommendationService (ListRecommendations)
- CheckoutService (PlaceOrder)
- ShippingService (GetQuote)
- AdService (GetAds)
- ShoppingAssistantService (HTTP POST /bot, optional)

### Current Deployment
- **Image**: `frontend` (Go, multi-arch amd64/arm64)
- **Port**: 8080 (HTTP)
- **Replicas**: 1 (default, scalable)
- **Resources**: 100m CPU / 64Mi request, 200m CPU / 128Mi limit
- **Security**: Non-root (UID 1000), read-only FS, dropped caps
- **Probes**: HTTP /_healthz with session cookie
- **Service**: ClusterIP (port 80) + LoadBalancer (frontend-external)

### Current Configuration (Environment Variables)
```bash
PORT=8080
PRODUCT_CATALOG_SERVICE_ADDR=productcatalogservice:3550
CURRENCY_SERVICE_ADDR=currencyservice:7000
CART_SERVICE_ADDR=cartservice:7070
RECOMMENDATION_SERVICE_ADDR=recommendationservice:8080
CHECKOUT_SERVICE_ADDR=checkoutservice:5050
SHIPPING_SERVICE_ADDR=shippingservice:50051
AD_SERVICE_ADDR=adservice:9555
SHOPPING_ASSISTANT_SERVICE_ADDR=shoppingassistantservice:80
ENABLE_PROFILER=0
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
ENV_PLATFORM=local  # or gcp, aws, azure, onprem, alibaba
CYMBAL_BRANDING=false
ENABLE_ASSISTANT=false
FRONTEND_MESSAGE=""
BASE_URL=""
```

---

## 2. Cart Service

### Current Responsibility
- Shopping cart persistence using Redis
- Add items, retrieve cart, empty cart
- Cart stored per user/session ID

### Current Communication Pattern
- **Inbound**: gRPC (CartService)
- **Outbound**: Redis (TCP 6379)
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| AddItem | AddItemRequest (user_id, CartItem) | Empty | Add/update item in cart |
| GetCart | GetCartRequest (user_id) | Cart (user_id, repeated CartItem) | Retrieve user's cart |
| EmptyCart | EmptyCartRequest (user_id) | Empty | Clear user's cart |

### Current Dependencies
- Redis (cartstorage) - single instance, no clustering

### Current Deployment
- **Image**: `cartservice` (C# .NET 6)
- **Port**: 7070 (gRPC)
- **Resources**: 200m CPU / 128Mi request, 300m CPU / 256Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 7070)
- **Redis**: Separate Deployment + Service (redis-cart:6379)

### Current Configuration
```bash
PORT=7070
REDIS_ADDR=redis-cart:6379  # via ASP.NET Core config
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
```

### Redis Cart Store Implementation
- Key: user_id (string)
- Value: Serialized `hipstershop.Cart` protobuf (byte array)
- Uses `IDistributedCache` abstraction (Microsoft.Extensions.Caching.Distributed)
- No TTL/expiration on cart entries
- No connection pooling configuration visible

---

## 3. Product Catalog Service

### Current Responsibility
- Product data management (list, get, search)
- Loads catalog from JSON file at startup
- In-memory storage with optional hot-reload via signal

### Current Communication Pattern
- **Inbound**: gRPC (ProductCatalogService)
- **Outbound**: None (stateless, file-based)
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| ListProducts | Empty | ListProductsResponse (repeated Product) | All products |
| GetProduct | GetProductRequest (id) | Product | Single product by ID |
| SearchProducts | SearchProductsRequest (query) | SearchProductsResponse (repeated Product) | Text search |

### Product Schema
```protobuf
message Product {
    string id = 1;
    string name = 2;
    string description = 3;
    string picture = 4;  # relative path to static image
    Money price_usd = 5;
    repeated string categories = 6;
}
```

### Current Dependencies
- Local JSON file (`data/products.json`)
- No external dependencies

### Current Deployment
- **Image**: `productcatalogservice` (Go)
- **Port**: 3550 (gRPC)
- **Resources**: 100m CPU / 64Mi request, 200m CPU / 128Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 3550)

### Current Configuration
```bash
PORT=3550
EXTRA_LATENCY=0s  # injected latency for testing
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
DISABLE_PROFILER=  # empty enables profiler
```

### Catalog Reload Mechanism
- SIGUSR1: Enable hot-reload (reload on each request)
- SIGUSR2: Disable hot-reload
- Not suitable for production (no cache invalidation strategy)

---

## 4. Currency Service

### Current Responsibility
- Currency conversion between supported currencies
- Uses European Central Bank (ECB) rates from static JSON
- Highest QPS service in the system

### Current Communication Pattern
- **Inbound**: gRPC (CurrencyService)
- **Outbound**: None (static data)
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| GetSupportedCurrencies | Empty | GetSupportedCurrenciesResponse (currency_codes) | List of ISO 4217 codes |
| Convert | CurrencyConversionRequest (from: Money, to_code) | Money | Convert amount |

### Money Representation
```protobuf
message Money {
    string currency_code = 1;  # ISO 4217
    int64 units = 2;           # whole units
    int32 nanos = 3;           # fractional (10^-9)
}
```

### Current Dependencies
- Static JSON: `data/currency_conversion.json` (ECB rates, EUR base)

### Current Deployment
- **Image**: `currencyservice` (Node.js)
- **Port**: 7000 (gRPC)
- **Resources**: 100m CPU / 128Mi request, 200m CPU / 256Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 7000)

### Current Configuration
```bash
PORT=7000
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
DISABLE_PROFILER=  # empty enables profiler
```

---

## 5. Payment Service

### Current Responsibility
- Mock credit card charging
- Returns fake transaction ID
- No actual payment processing

### Current Communication Pattern
- **Inbound**: gRPC (PaymentService)
- **Outbound**: None
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| Charge | ChargeRequest (amount: Money, credit_card: CreditCardInfo) | ChargeResponse (transaction_id) | Process payment |

### Credit Card Info
```protobuf
message CreditCardInfo {
    string credit_card_number = 1;
    int32 credit_card_cvv = 2;
    int32 credit_card_expiration_year = 3;
    int32 credit_card_expiration_month = 4;
}
```

### Current Dependencies
- None (mock implementation)

### Current Deployment
- **Image**: `paymentservice` (Node.js)
- **Port**: 50051 (gRPC)
- **Resources**: 100m CPU / 128Mi request, 200m CPU / 256Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 50051)

### Current Configuration
```bash
PORT=50051
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
DISABLE_PROFILER=  # empty enables profiler
```

---

## 6. Shipping Service

### Current Responsibility
- Shipping cost quotes based on item count
- Mock order shipment with tracking ID generation

### Current Communication Pattern
- **Inbound**: gRPC (ShippingService)
- **Outbound**: None
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| GetQuote | GetQuoteRequest (Address, repeated CartItem) | GetQuoteResponse (cost_usd: Money) | Shipping quote |
| ShipOrder | ShipOrderRequest (Address, repeated CartItem) | ShipOrderResponse (tracking_id) | Create shipment |

### Current Dependencies
- None (mock implementation)

### Current Deployment
- **Image**: `shippingservice` (Go)
- **Port**: 50051 (gRPC)
- **Resources**: 100m CPU / 64Mi request, 200m CPU / 128Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 50051)

### Current Configuration
```bash
PORT=50051
DISABLE_TRACING=  # empty enables tracing (currently unavailable)
DISABLE_PROFILER=  # empty enables profiler
```

---

## 7. Email Service

### Current Responsibility
- Send order confirmation emails
- Two implementations: Dummy (logs only) and Real (Google Cloud Mail API)

### Current Communication Pattern
- **Inbound**: gRPC (EmailService)
- **Outbound**: Google Cloud Mail API (real mode) or none (dummy)
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| SendOrderConfirmation | SendOrderConfirmationRequest (email, OrderResult) | Empty | Send confirmation |

### Current Dependencies
- Google Cloud Mail API (not implemented in current code)
- Jinja2 template (`templates/confirmation.html`)

### Current Deployment
- **Image**: `emailservice` (Python)
- **Port**: 5000 (gRPC)
- **Resources**: 100m CPU / 64Mi request, 200m CPU / 128Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 5000)
- **Mode**: Dummy only (real mode raises exception)

### Current Configuration
```bash
PORT=5000
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
DISABLE_PROFILER=  # empty enables profiler
```

---

## 8. Checkout Service

### Current Responsibility
- Order orchestration: coordinates cart, payment, shipping, email
- Currency conversion for order totals
- Generates order ID and order result

### Current Communication Pattern
- **Inbound**: gRPC (CheckoutService)
- **Outbound**: gRPC to 6 downstream services
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| PlaceOrder | PlaceOrderRequest (user_id, user_currency, address, email, credit_card) | PlaceOrderResponse (OrderResult) | Full checkout flow |

### Order Flow (PlaceOrder)
1. Get user cart from CartService
2. Get product details from ProductCatalogService
3. Convert prices to user currency via CurrencyService
4. Get shipping quote from ShippingService
5. Charge credit card via PaymentService
6. Ship order via ShippingService
7. Empty user cart via CartService
8. Send confirmation email via EmailService
9. Return OrderResult with tracking ID, items, costs

### Current Dependencies
- CartService (GetCart, EmptyCart)
- ProductCatalogService (GetProduct)
- CurrencyService (Convert)
- ShippingService (GetQuote, ShipOrder)
- PaymentService (Charge)
- EmailService (SendOrderConfirmation)

### Current Deployment
- **Image**: `checkoutservice` (Go)
- **Port**: 5050 (gRPC)
- **Resources**: 100m CPU / 64Mi request, 200m CPU / 128Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 5050)

### Current Configuration
```bash
PORT=5050
PRODUCT_CATALOG_SERVICE_ADDR=productcatalogservice:3550
SHIPPING_SERVICE_ADDR=shippingservice:50051
PAYMENT_SERVICE_ADDR=paymentservice:50051
EMAIL_SERVICE_ADDR=emailservice:5000
CURRENCY_SERVICE_ADDR=currencyservice:7000
CART_SERVICE_ADDR=cartservice:7070
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
DISABLE_PROFILER=  # empty enables profiler
```

---

## 9. Recommendation Service

### Current Responsibility
- Product recommendations based on cart contents
- Calls ProductCatalogService to get all products
- Returns random products not in user's cart

### Current Communication Pattern
- **Inbound**: gRPC (RecommendationService)
- **Outbound**: gRPC to ProductCatalogService
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| ListRecommendations | ListRecommendationsRequest (user_id, repeated product_ids) | ListRecommendationsResponse (repeated product_ids) | Get recommendations |

### Current Dependencies
- ProductCatalogService (ListProducts)

### Current Deployment
- **Image**: `recommendationservice` (Python)
- **Port**: 8080 (gRPC)
- **Resources**: 100m CPU / 220Mi request, 200m CPU / 450Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 8080)

### Current Configuration
```bash
PORT=8080
PRODUCT_CATALOG_SERVICE_ADDR=productcatalogservice:3550
ENABLE_TRACING=1
COLLECTOR_SERVICE_ADDR=otel-collector:4317
DISABLE_PROFILER=  # empty enables profiler
```

---

## 10. Ad Service

### Current Responsibility
- Contextual advertisement serving
- Static ad catalog mapped to categories
- Returns ads matching context keywords or random

### Current Communication Pattern
- **Inbound**: gRPC (AdService)
- **Outbound**: None (in-memory ad map)
- **Protocol**: gRPC with Protocol Buffers

### Current APIs (gRPC)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| GetAds | AdRequest (repeated context_keys) | AdResponse (repeated Ad) | Get ads for context |

### Ad Schema
```protobuf
message Ad {
    string redirect_url = 1;
    string text = 2;
}
```

### Current Dependencies
- None (static in-memory map)

### Current Deployment
- **Image**: `adservice` (Java)
- **Port**: 9555 (gRPC)
- **Resources**: 200m CPU / 180Mi request, 300m CPU / 300Mi limit
- **Security**: Non-root, read-only FS, dropped caps
- **Probes**: gRPC health check
- **Service**: ClusterIP (port 9555)

### Current Configuration
```bash
PORT=9555
DISABLE_TRACING=  # empty enables (currently unavailable)
DISABLE_STATS=  # empty enables (currently unavailable)
DISABLE_PROFILER=  # empty enables profiler
```

---

## 11. Redis (Cart Storage)

### Current Responsibility
- Persistent key-value store for cart data
- Single instance, no replication/clustering

### Current Deployment
- **Image**: `redis:alpine` (public Docker Hub)
- **Port**: 6379
- **Service**: ClusterIP (redis-cart:6379)
- **No persistence** configured (data lost on restart)
- **No HA** (single replica)

---

## 12. Load Generator

### Current Responsibility
- Synthetic traffic generation using Locust
- Simulates realistic user shopping flows

### Current Deployment
- **Image**: `loadgenerator` (Python/Locust)
- **Separate Skaffold config**
- **Resources**: 300m CPU / 256Mi request, 500m CPU / 512Mi limit

---

## 13. Shopping Assistant Service (Optional)

### Current Responsibility
- AI-powered product suggestions via Gemini
- HTTP endpoint for chat interactions

### Current Deployment
- **Image**: `shoppingassistantservice` (Python)
- **Port**: 80 (HTTP)
- **Disabled by default** in Helm chart

---

## Summary: Service Characteristics Matrix

| Service | Language | State | Persistence | Sync Calls | Async Calls | Mock/Real |
|---------|----------|-------|-------------|------------|-------------|-----------|
| Frontend | Go | Stateless | Cookies | 8 gRPC | 0 | Real |
| Cart | C# | Stateful | Redis | 0 | 0 | Real |
| ProductCatalog | Go | Stateless | JSON file | 0 | 0 | Real |
| Currency | Node.js | Stateless | JSON file | 0 | 0 | Real |
| Payment | Node.js | Stateless | None | 0 | 0 | Mock |
| Shipping | Go | Stateless | None | 0 | 0 | Mock |
| Email | Python | Stateless | None | 0 | 0 | Mock |
| Checkout | Go | Stateless | None | 6 gRPC | 0 | Real |
| Recommendation | Python | Stateless | None | 1 gRPC | 0 | Real |
| Ad | Java | Stateless | In-memory | 0 | 0 | Real |
| Redis | C | Stateful | Memory/RDB | N/A | N/A | Real |

---

## Key Observations for Enterprise Transformation

1. **No Database Layer**: Only Redis for cart; all other data is in-memory or mock
2. **Synchronous Only**: No async messaging, event streaming, or eventual consistency
3. **No User Identity**: Session IDs only, no authentication/authorization
4. **No Order Persistence**: Orders exist only in-memory during checkout flow
5. **No Inventory Management**: Product catalog is read-only
6. **Mock Critical Services**: Payment, Shipping, Email are not production-ready
6. **Single Point of Failure**: Redis, ProductCatalog (single instance)
7. **No API Gateway**: Frontend acts as BFF but no edge routing, auth, rate limiting
8. **No Observability Standardization**: Metrics not implemented, tracing optional