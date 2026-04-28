# Order & Payment Microservice System

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=flat)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat&logo=docker)

Backend microservice yang memisahkan domain **Order** dan **Payment** ke dalam service independen yang saling terintegrasi via internal REST callback. Dirancang untuk merepresentasikan pola transactional system pada e-commerce.

---

## Arsitektur

```
Client
  │
  │  REST API
  ▼
Order Service (:8080) ◄──── Internal Callback ──── Payment Service (:8081)
  │                                                        │
  ▼                                                        ▼
order_db (PostgreSQL)                             payment_db (PostgreSQL)
```

Setiap service memiliki database sendiri — tidak ada shared database antar service.

---

## Tech Stack

| Komponen         | Teknologi               |
| ---------------- | ----------------------- |
| Language         | Go 1.21+                |
| HTTP Framework   | Fiber v2                |
| ORM              | GORM                    |
| Database         | PostgreSQL              |
| Containerization | Docker & Docker Compose |
| API Docs         | OpenAPI 3.0             |

---

## Cara Menjalankan

### Prasyarat

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- Atau Go 1.21+ dan PostgreSQL (untuk local tanpa Docker)

### 1. Clone repository

```bash
git clone https://github.com/DhahikaR/microservice-order-payment.git
cd microservice-order-payment
```

### 2. Setup environment variables

```bash
# Order Service
cp order-service/.env.example order-service/.env

# Payment Service
cp payment-service/.env.example payment-service/.env
```

Edit masing-masing file `.env` sesuai konfigurasi lokal kamu (lihat section [Environment Variables](#environment-variables)).

### 3. Jalankan dengan Docker (Direkomendasikan)

```bash
docker-compose up --build
```

Setelah berhasil, kedua service akan berjalan di:

- **Order Service** → http://localhost:8080
- **Payment Service** → http://localhost:8081

### 4. Jalankan secara manual (tanpa Docker)

Pastikan PostgreSQL sudah berjalan dan database sudah dibuat.

```bash
# Jalankan Order Service
cd order-service
go mod tidy
go run main.go

# Jalankan Payment Service (terminal baru)
cd payment-service
go mod tidy
go run main.go
```

---

## Environment Variables

### Order Service (`order-service/.env`)

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=order_db
```

### Payment Service (`payment-service/.env`)

```env
APP_PORT=8081

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=payment_db

ORDER_SERVICE_URL=http://localhost:8080
```

---

## API Reference

Format response sukses selalu konsisten:

```json
{
  "code": 200,
  "status": "OK",
  "data": {}
}
```

---

### Order Service — `http://localhost:8080`

#### Buat Order Baru

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "item_name": "Sepatu Lari Nike",
    "quantity": 2,
    "price": 350000
  }'
```

**Response:**

```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "item_name": "Sepatu Lari Nike",
    "quantity": 2,
    "price": 350000,
    "total_amount": 700000,
    "status": "pending",
    "payment_id": null,
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:00:00Z"
  }
}
```

---

#### Ambil Semua Order

```bash
curl http://localhost:8080/orders
```

---

#### Ambil Order berdasarkan ID

```bash
curl http://localhost:8080/orders/550e8400-e29b-41d4-a716-446655440000
```

---

#### Update Order

```bash
curl -X PUT http://localhost:8080/orders/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "item_name": "Sepatu Lari Adidas",
    "quantity": 1,
    "price": 450000
  }'
```

---

#### Hapus Order

```bash
curl -X DELETE http://localhost:8080/orders/550e8400-e29b-41d4-a716-446655440000
```

**Response:**

```json
{
  "message": "order deleted",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### Payment Service — `http://localhost:8081`

#### Buat Payment (Auto-Success Mock)

> Payment yang dibuat akan otomatis diproses dan mengirim callback ke Order Service untuk mengubah status order menjadi `paid`.

```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 700000,
    "provider": "bank_transfer"
  }'
```

**Response:**

```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "id": "7f3b9c20-1234-5678-abcd-ef0123456789",
    "order_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 700000,
    "status": "paid",
    "provider": "bank_transfer",
    "paid_at": "2025-01-15T10:05:00Z",
    "created_at": "2025-01-15T10:05:00Z",
    "updated_at": "2025-01-15T10:05:00Z"
  }
}
```

---

#### Ambil Payment berdasarkan ID

```bash
curl http://localhost:8081/payments/7f3b9c20-1234-5678-abcd-ef0123456789
```

---

#### Tandai Payment Berhasil (Manual)

```bash
curl -X PUT http://localhost:8081/payments/success/7f3b9c20-1234-5678-abcd-ef0123456789
```

---

#### Tandai Payment Gagal (Manual)

```bash
curl -X PUT http://localhost:8081/payments/failed/7f3b9c20-1234-5678-abcd-ef0123456789
```

---

## Alur Payment (Business Logic)

```
1. Client  →  POST /orders             →  Order dibuat, status: "pending"
                                                   │
2. Client  →  POST /payments           →  Payment dibuat & diproses
                                                   │
3. Payment Service  →  Internal Callback  →  Order Service
                                                   │
4. Order Service memperbarui status:
   - Payment sukses  →  status order: "paid"
   - Payment gagal   →  status order: "failed"
```

Setiap domain tetap menjadi **single source of truth** untuk datanya masing-masing. Order Service tidak tahu cara memproses payment — ia hanya menerima notifikasi hasilnya.

---

## Testing

Jalankan test dengan coverage report:

```bash
# Order Service
cd order-service
go test ./... -v -coverpkg=./...

# Payment Service
cd payment-service
go test ./... -v -coverpkg=./...
```

Coverage mencakup: Controller · Service · Repository · Middleware · Helper · Exception handling

---

## API Documentation (OpenAPI)

Seluruh endpoint didokumentasikan di `openapi.yaml`. Cara membukanya:

1. Buka [Swagger Editor](https://editor.swagger.io)
2. Import file `openapi.yaml` dari root project
3. Eksplorasi dan test endpoint secara interaktif
   Atau import ke **Postman**: File → Import → pilih `openapi.yaml`

---

## Author

**Dhahika Rahmadani**
Backend Developer · Go Enthusiast
📧 dhahikardani@gmail.com
