# Art Print Backend - System Summary

A comprehensive backend system for managing art print orders, print shop services, payments, and order fulfillment. Built with Go, Firebase, and designed for seamless computation and delivery.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Core Modules](#core-modules)
- [Print Shop Console Management](#print-shop-console-management)
- [Admin Management](#admin-management)
- [Service Architecture](#service-architecture)
- [API Endpoints](#api-endpoints)
- [Installation](#installation)
- [Configuration](#configuration)
- [Development](#development)

## Architecture Overview

The backend is built using a modular service-oriented architecture that separates concerns and enables seamless computation and delivery:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handlers Layer                      │
│  (REST API Endpoints, Authentication, Request Validation)   │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    Service Layer                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Pricing    │  │   Matching   │  │   Payment   │      │
│  │   Service    │  │   Service    │  │   Service   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Order      │  │   Delivery   │  │    Admin    │      │
│  │   Service    │  │   Service    │  │   Service   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                  Repository Layer                            │
│  (Data Access Abstraction, Firestore Operations)             │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    Data Layer                                │
│              (Firebase Firestore)                            │
└─────────────────────────────────────────────────────────────┘
```

## Core Modules

### 1. User Management
- **Buyers**: Purchase art prints
- **Artists**: Upload and manage artworks
- **Print Shops**: Offer printing services
- **Admins**: Platform management

### 2. Print Shop Console Management
Comprehensive management system for print shop owners to configure their services and pricing.

### 3. Order Management
- Order creation and tracking
- Shop assignment via matching algorithms
- Status management

### 4. Payment System
- Half payment (50% deposit) on order confirmation
- Full payment support
- Payment verification and refunds
- Simulated provider for testing
- Ready for Stripe/M-Pesa integration

### 5. Matching Service
Intelligent order-to-shop matching with three modes:
- **Auto**: Cheapest available shop
- **Smart**: Multi-factor scoring (price, rating, delivery time, technology)
- **Manual**: Admin review and selection

## Print Shop Console Management

The Print Shop Console allows shop owners to manage their entire service catalog and pricing structure.

### Features

#### 1. Shop Profile Management
- Create and update shop profile
- Manage contact information and location
- Set shop status (active/inactive)

**Endpoints:**
- `GET /printshop/profile` - Get shop profile
- `POST /printshop/profile/create` - Create shop profile
- `PUT /printshop/profile/update` - Update shop profile

#### 2. Service Configuration
Print shops can define multiple print services (e.g., Canvas Printing, Metal Prints, Giclée):

- **Service Details**: Name, description, technology type
- **Base Pricing**: Base price for each service
- **Dynamic Pricing Matrix**: Configurable pricing based on:
  - Size modifiers (A4, A3, A2, etc.)
  - Material markups (matte, glossy, canvas, metal)
  - Medium markups (paper, metal, premium)
  - Frame prices (classic, modern, premium)
  - Quantity discount tiers
  - Rush order fees

**Endpoints:**
- `GET /printshop/services` - List all services
- `POST /printshop/services/create` - Create new service
- `PUT /printshop/services/update/{id}` - Update service
- `DELETE /printshop/services/delete/{id}` - Delete service

#### 3. Configuration Management
Shops can configure their available options:

**Frames:**
- `GET /printshop/frames` - List frames
- `POST /printshop/frames/create` - Add frame
- `PUT /printshop/frames/update/{id}` - Update frame
- `DELETE /printshop/frames/delete/{id}` - Remove frame

**Sizes:**
- `GET /printshop/sizes` - List sizes
- `POST /printshop/sizes/create` - Add size
- `PUT /printshop/sizes/update/{id}` - Update size
- `DELETE /printshop/sizes/delete/{id}` - Remove size

**Materials:**
- `GET /printshop/materials` - List materials
- `POST /printshop/materials/create` - Add material
- `PUT /printshop/materials/update/{id}` - Update material
- `DELETE /printshop/materials/delete/{id}` - Remove material

#### 4. Pricing Configuration
- `GET /printshop/services/pricing/{serviceId}` - Get price matrix
- `PUT /printshop/services/pricing/update/{serviceId}` - Update price matrix
- `POST /printshop/services/calculate/{serviceId}` - Calculate price for options

### Authentication
All print shop endpoints require:
1. Standard authentication (session cookie)
2. `printShop` role in user profile
3. Shop ownership verification

## Admin Management

The Admin Dashboard provides comprehensive platform oversight and management capabilities.

### Features

#### 1. Platform Statistics
- Total users (buyers, artists, print shops)
- Total orders and revenue
- Active shops and pending orders
- Recent signups and activity

#### 2. Order Management
- View all orders across the platform
- Filter by status, shop, buyer
- Update order status
- View order details and history

#### 3. User Management
- View all users (buyers, artists, print shops)
- Filter by role and status
- Update user status
- View user profiles and activity

#### 4. Print Shop Oversight
- View all print shops
- Approve/reject shop registrations
- Monitor shop activity
- Review shop ratings and performance

#### 5. Catalog Change Supervision
Print shops can request catalog changes (new services, price updates, etc.):
- View pending catalog change requests
- Approve or reject changes
- Track change history
- Maintain catalog integrity

#### 6. Payment Management
- View all payments
- Payment statistics and analytics
- Process refunds
- Monitor payment status

#### 7. Review Management
- View all user/buyer reviews
- Approve or reject reviews
- Moderate review content
- Track review ratings

### Authentication
Admin endpoints require:
1. Standard authentication (session cookie)
2. `admin` role in user profile

## Service Architecture

### How Services Work Together

The backend uses a service-oriented architecture where specialized services handle specific domains and communicate through well-defined interfaces.

#### 1. Pricing Service
**Purpose**: Calculate prices for print orders using shop-specific pricing matrices.

**Key Features:**
- Real-time price calculation using Go's computational efficiency
- Support for shop-specific `PriceMatrix`
- Backward compatibility with hardcoded catalog
- Detailed price breakdowns

**How It Works:**
```go
// Shop-specific pricing
price = service.BasePrice
price *= sizeModifier[options.Size]
price *= materialMarkup[options.Material]
price *= mediumMarkup[options.Medium]
price += framePrice[options.Frame]
price *= (1.0 - quantityDiscount)
price += rushOrderFee (if applicable)
price *= quantity
```

**Integration:**
- Used by `MatchingService` to calculate prices for shop matching
- Used by `PublicPrintShopHandler` for price calculations
- Used by `CheckoutHandler` for order pricing

#### 2. Matching Service
**Purpose**: Match customer orders with suitable print shops.

**Components:**
- **ServiceDiscovery**: Finds shops that can fulfill order requirements
- **AutoMatcher**: Assigns to cheapest available shop
- **SmartMatcher**: Uses multi-factor scoring algorithm
- **ManualMatcher**: Returns all matches for admin review

**How It Works:**
1. **Service Discovery**:
   - Queries all active print shops
   - For each shop, checks services for compatibility
   - Validates service supports requested options (size, material, medium, frame)
   - Calculates price using `PricingService`

2. **Auto Matching**:
   - Finds all matching shops
   - Sorts by price (cheapest first)
   - Assigns to cheapest shop

3. **Smart Matching**:
   - Finds all matching shops
   - Calculates match score:
     - Price competitiveness: 40%
     - Shop rating: 30%
     - Delivery time: 20%
     - Technology quality: 10%
   - Assigns to highest-scoring shop

4. **Manual Matching**:
   - Returns all matches
   - Admin selects appropriate shop

**Integration:**
- Uses `PricingService` for price calculations
- Uses `PrintShopRepository` for shop data
- Called by `OrderService` during checkout

#### 3. Payment Service
**Purpose**: Handle payment processing and verification.

**Components:**
- **PaymentProvider Interface**: Abstraction for payment providers
- **SimulatedProvider**: Testing provider
- **PaymentRepository**: Payment data access

**How It Works:**
1. **Payment Creation**:
   - Calculates amount based on type (deposit 50%, full 100%, remaining 50%)
   - Creates payment with provider
   - Stores payment record
   - Returns transaction ID

2. **Payment Verification**:
   - Verifies with provider
   - Updates payment status
   - Updates order status if completed

3. **Payment Status Calculation**:
   - Aggregates all payments for order
   - Calculates overall status: unpaid/partial/paid

**Integration:**
- Updates `Order` model with payment status
- Triggers delivery creation on payment completion
- Used by `CheckoutHandler` for payment processing

#### 4. Order Service
**Purpose**: Orchestrate order lifecycle and shop assignment.

**How It Works:**
1. **Order Creation**:
   - Creates order from cart
   - Extracts print options
   - Calls `MatchingService` to assign shop
   - Saves order

2. **Shop Assignment**:
   - Determines fulfillment mode (auto/smart/manual)
   - Routes to appropriate matcher
   - Updates order with shop assignment

**Integration:**
- Uses `MatchingService` for shop assignment
- Uses `ConfigService` for fulfillment mode
- Called by `CheckoutHandler`

### Service Communication Flow

```
Order Creation Flow:
┌─────────────┐
│  Checkout   │
│  Handler    │
└──────┬──────┘
       │
       ▼
┌─────────────┐      ┌─────────────┐
│   Order     │─────▶│  Matching   │
│   Service   │      │   Service   │
└──────┬──────┘      └──────┬──────┘
       │                    │
       │                    ▼
       │            ┌─────────────┐
       │            │  Pricing    │
       │            │   Service   │
       │            └─────────────┘
       │
       ▼
┌─────────────┐
│   Payment   │
│   Service   │
└─────────────┘

Price Calculation Flow:
┌─────────────┐
│   Request   │
│  (Options)  │
└──────┬──────┘
       │
       ▼
┌─────────────┐      ┌─────────────┐
│  Pricing    │─────▶│   Shop      │
│   Service   │      │  Repository │
└──────┬──────┘      └─────────────┘
       │
       ▼
┌─────────────┐
│   Price     │
│  Breakdown  │
└─────────────┘
```

## Seamless Computation and Delivery

### Computational Efficiency

The backend leverages Go's strengths for high-performance computations:

1. **Real-Time Price Calculations**:
   - Efficient map lookups for price matrices
   - Fast arithmetic operations
   - No external API calls for pricing
   - Sub-millisecond response times

2. **Matching Algorithm**:
   - Concurrent shop discovery
   - Efficient filtering and sorting
   - In-memory scoring calculations
   - Fast shop assignment

3. **Payment Processing**:
   - Asynchronous webhook processing
   - Efficient status updates
   - Fast payment verification

### Delivery Architecture

The system ensures seamless order delivery through:

1. **Order Lifecycle Management**:
   ```
   pending → confirmed (payment) → processing → ready → delivered
   ```

2. **Status Tracking**:
   - Order status
   - Payment status
   - Delivery status
   - All tracked independently and updated in real-time

3. **Automatic Workflows**:
   - Payment completion triggers delivery creation
   - Shop updates trigger status changes
   - Buyer notifications on status changes

## API Endpoints

### Public Endpoints
- `POST /signup` - User registration
- `POST /sessionLogin` - User login
- `POST /sessionLogout` - User logout
- `GET /artworks` - List artworks
- `GET /artists` - List artists
- `GET /print-options` - Get print options
- `GET /printshops` - List active shops
- `GET /printshops/details` - Get shop details
- `POST /printshops/match` - Match shops for order
- `POST /printshops/calculate` - Calculate price

### Authenticated Endpoints
- `GET /getprofile` - Get user profile
- `PUT /updateprofile` - Update profile
- `POST /artworks/upload` - Upload artwork
- `POST /cart/add` - Add to cart
- `DELETE /cart/remove` - Remove from cart
- `GET /cart` - Get cart
- `POST /checkout` - Checkout
- `GET /orders` - Get orders
- `POST /calculate-price` - Calculate price
- `POST /payments/create` - Create payment
- `GET /payments/verify` - Verify payment
- `GET /payments` - Get payments
- `POST /orders/matches` - Get order matches
- `POST /orders/assign` - Assign shop to order

### Print Shop Console Endpoints
- `GET /printshop/profile` - Get shop profile
- `POST /printshop/profile/create` - Create shop
- `PUT /printshop/profile/update` - Update shop
- `GET /printshop/services` - List services
- `POST /printshop/services/create` - Create service
- `PUT /printshop/services/update/{id}` - Update service
- `DELETE /printshop/services/delete/{id}` - Delete service
- `GET /printshop/frames` - List frames
- `POST /printshop/frames/create` - Create frame
- `PUT /printshop/frames/update/{id}` - Update frame
- `DELETE /printshop/frames/delete/{id}` - Delete frame
- `GET /printshop/sizes` - List sizes
- `POST /printshop/sizes/create` - Create size
- `PUT /printshop/sizes/update/{id}` - Update size
- `DELETE /printshop/sizes/delete/{id}` - Delete size
- `GET /printshop/materials` - List materials
- `POST /printshop/materials/create` - Create material
- `PUT /printshop/materials/update/{id}` - Update material
- `DELETE /printshop/materials/delete/{id}` - Delete material
- `GET /printshop/services/pricing/{id}` - Get pricing
- `PUT /printshop/services/pricing/update/{id}` - Update pricing
- `POST /printshop/services/calculate/{id}` - Calculate price

## Installation

### Prerequisites
- Go 1.21+
- Firebase project with Firestore enabled
- Firebase service account JSON file
- Cloudinary account (for image storage)

### Setup

1. **Clone the repository**:
```bash
git clone <repository-url>
cd art-print-backend
```

2. **Install dependencies**:
```bash
go mod download
```

3. **Configure Firebase**:
   - Place `firebase-service-account.json` in project root
   - Set `FIREBASE_PROJECT_ID` environment variable

4. **Configure Cloudinary**:
   - Set `CLOUDINARY_CLOUD_NAME`
   - Set `CLOUDINARY_API_KEY`
   - Set `CLOUDINARY_API_SECRET`

5. **Run the server**:
```bash
go run cmd/server/main.go
```

Or using Docker:
```bash
./run-server.sh
```

## Configuration

### Environment Variables

- `FIREBASE_PROJECT_ID` - Firebase project ID
- `PORT` - Server port (default: 3001)
- `CLOUDINARY_CLOUD_NAME` - Cloudinary cloud name
- `CLOUDINARY_API_KEY` - Cloudinary API key
- `CLOUDINARY_API_SECRET` - Cloudinary API secret

### Firebase Configuration

The system uses Firebase for:
- Authentication (Firebase Auth)
- Database (Firestore)
- User management

### Cloudinary Configuration

Cloudinary is used for:
- Artwork image storage
- User avatar storage
- Image transformations

## Development

### Project Structure

```
art-print-backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── handlers/            # HTTP handlers
│   ├── models/              # Data models
│   ├── repositories/        # Data access layer
│   ├── services/           # Business logic services
│   │   ├── payment/        # Payment service
│   │   ├── matching/       # Matching service
│   │   ├── pricing/        # Pricing service
│   │   └── orders/         # Order service
│   ├── middleware/         # HTTP middleware
│   └── firebase/           # Firebase client
├── docker-compose.yml      # Docker Compose config
└── Dockerfile.server       # Server Dockerfile
```

### Key Design Patterns

1. **Repository Pattern**: Abstracts data access
2. **Service Layer**: Business logic separation
3. **Provider Pattern**: Payment provider abstraction
4. **Middleware Chain**: Authentication and logging

### Testing

Run tests:
```bash
go test ./...
```

### Building

Build the server:
```bash
go build -o server cmd/server/main.go
```

## License

[Your License Here]

## Contributing

[Contributing Guidelines]

