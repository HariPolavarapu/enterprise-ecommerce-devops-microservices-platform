# Service Interaction Diagram

```mermaid
graph TD
    subgraph "User Layer"
        User[("User/Browser")]
    end

    subgraph "API Gateway Layer"
        GW[("API Gateway<br/>(Kong/Envoy)")]
        GW -->|"JWT Validation"| JWT[("JWT Validator")]
        GW -->|"Rate Limiting"| RL[("Rate Limiter")]
        GW -->|"Routing"| Router[("Router")]
    end

    subgraph "Auth Layer"
        KC[("Keycloak")]
        KC -->|"OIDC"| Token[("Token Service")]
    end

    subgraph "Frontend Layer"
        FE[("Frontend<br/>(Go HTTP)"]
    end

    subgraph "Core Services"
        PC[("Product Catalog<br/>(Go, gRPC)")]
        Cart[("Cart Service<br/>(C#, gRPC)")]
        Checkout[("Checkout Service<br/>(Go, gRPC)")]
        Payment[("Payment Service<br/>(Node.js)"]
        Ship[("Shipping Service<br/>(Go, gRPC)"]
        Currency[("Currency Service<br/>(Node.js)"]
        Email[("Email Service<br/>(Python)"]
        Rec[("Recommendation<br/>(Python)"]
        Ad[("Ad Service<br/>(Java)"]
    end

    subgraph "New Enterprise Services"
        US[("User Service<br/>(Go)"]
        OS[("Order Service<br/>(Go)"]
        IS[("Inventory Service<br/>(Go)"]
        AS[("Analytics Service<br/>(Go)"]
        NS[("Notification Service<br/>(Go)"]
    end

    subgraph "Data Layer"
        PG[("PostgreSQL")]
        Redis[("Redis")]
        Kafka[("Kafka")]
        OpenSearch[("OpenSearch")]
        S3[("S3/MinIO")]
    end

    %% Connections
    User -->|"HTTPS"| GW
    GW -->|"Route /"| FE
    GW -->|"/auth"| KC
    GW -->|"API /api/v1/*"| Router

    FE -->|"gRPC"| PC
    FE -->|"gRPC"| Cart
    FE -->|"gRPC"| Checkout
    FE -->|"gRPC"| Ship
    FE -->|"gRPC"| Currency
    FE -->|"gRPC"| Rec
    FE -->|"gRPC"| Ad
    FE -->|"HTTP"| US

    Checkout -->|"gRPC"| PC
    Checkout -->|"gRPC"| Cart
    Checkout -->|"gRPC"| Ship
    Checkout -->|"gRPC"| Payment
    Checkout -->|"gRPC"| Email
    Checkout -->|"gRPC"| Currency

    %% New connections
    PC -->|"SQL"| PG
    Cart -->|"SQL"| PG
    Checkout -->|"SQL"| PG
    Payment -->|"SQL"| PG
    Ship -->|"SQL"| PG
    US -->|"SQL"| PG
    OS -->|"SQL"| PG
    IS -->|"SQL"| PG

    %% Events
    Checkout -->|"order-created"| Kafka
    Payment -->|"payment-success"| Kafka
    Payment -->|"payment-failed"| Kafka
    IS -->|"inventory-updated"| Kafka
    Ship -->|"shipment-created"| Kafka
    Cart -->|"cart-updated"| Kafka

    %% Consume
    OS -->|"order-created"| Kafka
    IS -->|"order-created"| Kafka
    NS -->|"notification-requested"| Kafka
    AS -->|"All Events"| Kafka

    %% Search
    PC -->|"Index"| OpenSearch
    Rec -->|"KNN"| OpenSearch

    %% Storage
    PC -->|"Images"| S3
    Checkout -->|"Invoices"| S3

    %% Style
    classDef user fill:#e1f5fe,stroke:#01579b
    classDef gateway fill:#fff3e0,stroke:#e65100
    classDef auth fill:#f3e5f5,stroke:#7b1fa2
    classDef existing fill:#e8f5e9,stroke:#2e7d32
    classDef new fill:#fce4ec,stroke:#c62828
    classDef infra fill:#f5f5f5,stroke:#616161

    class User user
    class GW,JWT,RL,Router gateway
    class KC,Token auth
    class FE,PC,Cart,Checkout,Payment,Ship,Currency,Email,Rec,Ad existing
    class US,OS,IS,AS,NS new
    class PG,Redis,Kafka,OpenSearch,S3 infra
```