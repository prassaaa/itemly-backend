# Itemly - Enterprise Inventory Management System (Backend)

Itemly adalah solusi manajemen inventaris tingkat lanjut yang dirancang untuk skala enterprise. Backend ini dibangun menggunakan **Go (Golang)** dengan fokus pada performa tinggi, konkurensi, dan arsitektur yang bersih (*Clean Architecture*).

Sistem ini mendukung pengelolaan multi-gudang, pelacakan stok real-time, hingga fitur mutasi barang yang kompleks untuk kebutuhan rantai pasok modern.

---

## 🛠 Tech Stack

- **Language:** Go (Golang) 1.2x
- **Framework/Router:** Gin Gonic / Fiber
- **Database:** PostgreSQL / MySQL
- **ORM:** GORM / SQLX
- **Caching & Pub/Sub:** Redis
- **Authentication:** JWT (JSON Web Token)
- **Real-time:** Gorilla WebSocket
- **Documentation:** Swagger (go-swagger)

---

## 🚀 Key Features

### 🔐 Enterprise Core

- **Authentication & RBAC:** Pengamanan API menggunakan JWT dengan Middleware untuk Role-Based Access Control (Admin, Warehouse Manager, Staff).
- **Multi-Warehouse Support:** Kelola stok di berbagai lokasi gudang secara independen dalam satu sistem.
- **Master Data Management:** Pengelolaan data Produk, Kategori, dan Supplier yang terintegrasi.
- **Inventory Core Engine:** Logika bisnis untuk Stock In, Stock Out, Adjustment, dan Mutasi antar gudang yang atomik.
- **Procurement System:** Alur kerja profesional mulai dari Purchase Order (PO) hingga Goods Received (Penerimaan Barang).
- **Audit Trail:** Pencatatan setiap perubahan stok (Stock Logs) untuk transparansi data dan *debugging* histori.

### 🌟 Killer Features

- **Real-time Stock Alert:** Notifikasi instan via WebSockets ketika stok mencapai ambang batas minimum (*safety stock*).
- **QR Code Generator:** Otomatisasi pembuatan QR Code untuk setiap unit barang dan dukungan Label Printing untuk efisiensi operasional.

---

## 📁 Project Structure

Proyek ini mengikuti standar **Standard Go Project Layout**:

```
.
├── cmd/                # Entry point aplikasi
├── internal/           # Private application & business logic
│   ├── delivery/       # HTTP Handlers / Controllers
│   ├── usecase/        # Business Logic / Services
│   ├── repository/     # Data Access Layer (Database)
│   └── model/          # Structs & Entities
├── pkg/                # Helper/Utility yang bisa digunakan kembali
├── config/             # Konfigurasi aplikasi & Environment
├── database/           # Migrasi & Seeder
└── docs/               # Dokumentasi API (Swagger)
```

---

## ⚡ Getting Started

### Prerequisites

- Go 1.2x installed
- PostgreSQL / MySQL
- Redis Server (untuk Real-time Alert)

### Installation

1. Clone repositori:

```
git clone https://github.com/prassaaa/itemly-backend.git
cd itemly-backend
```

2. Install dependencies:

```
go mod tidy
```

3. Konfigurasi Environment:

   Salin file `.env.example` menjadi `.env` dan sesuaikan kredensial database Anda.

4. Jalankan Migrasi:

```
go run main.go migrate
```

5. Jalankan Aplikasi:

```
go run main.go
```

---

## 📑 API Documentation

Dokumentasi API tersedia via Swagger. Setelah menjalankan aplikasi, akses di:

```
http://localhost:8080/swagger/index.html
```

---

## 🛡 License

Distribusi di bawah Lisensi MIT. Lihat `LICENSE` untuk informasi lebih lanjut.

---

**Itemly** - *Optimizing Supply Chain with Speed and Precision.*
