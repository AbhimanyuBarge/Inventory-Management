# 🛠️ SparePro - Spare Parts Inventory & POS Management System

A modern, mobile-friendly **Spare Parts Inventory & Point of Sale (POS) Billing System** built with **Go (Golang)**, **Vue.js 3**, and **PostgreSQL**, fully dockerized with a single `docker-compose.yml`.

Designed with an **Android Material Design 3 UI** that delivers an intuitive experience across desktop monitors and mobile devices.

---

## 📱 Features

- 👤 **Role-Based Access Control (RBAC)**:
  - **Admin**: Complete system control, user & seller management, GST rate and shop profile configuration.
  - **Seller (Sub-user)**: Fast billing, customer lookup & management, stock overview.
- 📦 **Spare Part & Inventory Management**:
  - SKU tracking, selling & cost prices, GST applicability, stock quantities, and low-stock warning thresholds.
  - Quick stock replenishment and adjustment dialogues (+ / - / =).
- 🗄️ **Storage Rack Compartments & Label Printing**:
  - Rack organization by Name & Number (e.g., `Rack A` - `A1`).
  - One-click **PDF Rack Label Generator** (A6 printable warehouse tags with parts directory).
- 👥 **Customer Directory & Fast Auto-Complete**:
  - Auto-fills customer profile by searching name or mobile number during billing.
  - Automatic sequence tracker per customer (`Bill #1`, `Bill #2`, etc.).
- 🧾 **Point of Sale (POS) Billing System**:
  - Rapid item search with live stock levels and rack location chips.
  - Instant stock deduction upon checkout.
  - **"Add GST" Checkbox Toggle**: Applies tax dynamically based on admin settings.
  - Built-in **Printable PDF Tax Invoice** generation with shop headers, GSTIN, itemized breakdown, and terms.
- 📊 **Dashboard & Real-time Analytics**:
  - KPI cards for stock valuation, storage racks, critical low stock alerts, today's sales, and month's revenue.

---

## 🚀 Quick Start with Docker Compose

### Prerequisites
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/) installed.

### 1. Launch All Services
Run the following command from the project root:
```bash
docker compose up -d --build
```

### 2. Access the Applications
| Service | URL | Description |
|---|---|---|
| **Frontend UI** | [http://localhost:8080](http://localhost:8080) | Vue 3 Android-style Web App |
| **Backend REST API** | [http://localhost:8081](http://localhost:8081) | Go Gin Engine & Swagger-ready API |
| **PostgreSQL Database** | `localhost:5432` | Persistent Database |

---

## 🔑 Default Login Credentials

The system automatically auto-migrates and seeds initial demo data and accounts on first launch:

| Role | Username | Password | Permissions |
|---|---|---|---|
| **Administrator** | `admin` | `admin123` | Full access to all modules, GST settings & user management |
| **Seller** | `seller` | `seller123` | POS Billing, customer management, inventory viewing |

*(Passwords can be changed at any time in the Settings panel).*

---

## 💻 Local Development Setup (Without Docker)

### 1. Backend (Go)
```bash
cd backend
# Optional: Set DB_TYPE=sqlite in .env for instant local testing without postgres
go run ./cmd/server
```
*Backend runs on port `8081`.*

### 2. Frontend (Vue 3 + Vite)
```bash
cd frontend
npm install
npm run dev
```
*Frontend runs on port `8080` (with automatic reverse proxy to `:8081`).*

---

## 🏗️ Project Structure

```
Inventory-Management/
├── backend/
│   ├── cmd/server/main.go            # Go HTTP server entrypoint & router
│   ├── internal/
│   │   ├── database/database.go      # PostgreSQL / SQLite connection, auto-migrations & seeders
│   │   ├── handlers/                 # REST API handlers (auth, parts, racks, bills, pdf, stats)
│   │   ├── middleware/auth.go        # JWT authentication & RBAC middleware
│   │   ├── models/models.go          # GORM Data models
│   │   └── pdf/pdf.go                # PDF Tax Invoice & Rack Label generator
│   ├── Dockerfile
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── api/client.js             # Axios client with JWT interceptors
│   │   ├── components/AppLayout.vue  # Desktop Drawer & Android Mobile Bottom Navigation
│   │   ├── stores/auth.js            # Pinia Authentication state store
│   │   ├── views/                    # Responsive Views (Dashboard, Billing, Parts, Racks, etc.)
│   │   ├── style.css                 # Android Material 3 Design System
│   │   ├── App.vue
│   │   └── main.js
│   ├── nginx.conf                    # Nginx SPA & API reverse proxy configuration
│   ├── Dockerfile
│   └── package.json
│
└── docker-compose.yml                # Unified orchestrator (Vue :8080, Go :8081, Postgres :5432)
```

---

## 📄 License
MIT License
