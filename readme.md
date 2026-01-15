# Order & Payment Microservice System (Goalng)

## Overview

Project ini merupakan implementasi backend microservice yang memisahkan domain Order dan Payment ke dalam service yang berdiri sendiri namun saling terintegrasi. Sistem dirancang untuk merepresentasikan pola umum pada transactional system (misalnya e-commerce), dengan fokus pada API contract, alur pembayaran, dan separation of concerns.

Integrasi antar service dilakukan melalui internal REST callback, tanpa shared database.

---

## System Architecture

```
Client
  |
  | REST API
  v
Order Service  <---- Internal Callback ----  Payment Service
```

---

## Architectural Rationale

- Order Service tidak bergantung langsung pada implementasi Payment
- Payment Service tidak mengubah data Order secara langsung
- Komunikasi antar service dilakukan melalui endpoint internal yang terdefinisi jelas
- Setiap service memiliki database dan lifecycle sendiri
- Pendekatan ini menghindari tight coupling dan mencerminkan praktik microservice di dunia industri.

---

## Technology Stack

- Go (Golang)
- Fiber v2 – HTTP framework
- GORM – ORM
- PostgreSQL
- Docker & Docker Compose
- OpenAPI 3.0 (Swagger)
- SQLite (in-memory)

---

## API Documentation (OpenAPI)

Seluruh API didokumentasikan menggunakan OpenAPI 3.0 dan merepresentasikan kontrak aktual implementasi, bukan dokumentasi konseptual.

Dokumentasi tersedia pada:

- openapi.yaml (gabungan Order & Payment Service)

Swagger digunakan sebagai:

- sumber kebenaran kontrak API
- alat eksplorasi dan pengujian
- referensi integrasi antar service

---

## Project Structure

```
Microservice-Order-Payment/
│
├── openapi.yaml
├── docker-compose.yml
├── README.md
│
├── order-service/
│ ├── config/
│ ├── controller/
│ ├── exception/
│ ├── helper/
│ ├── models/
│ ├── repository/
│ ├── routes/
│ ├── service/
│ ├── test/
│ ├── .env
│ ├── Dockerfile
│ ├── main.go
│ ├── go.mod
│ └── go.sum
│
├── payment-service/
│ ├── config/
│ ├── controller/
│ ├── exception/
│ ├── helper/
│ ├── models/
│ ├── repository/
│ ├── routes/
│ ├── service/
│ ├── test/
│ ├── .env
│ ├── Dockerfile
│ ├── main.go
│ ├── go.mod
│ └── go.sum
```

---

## Structure Notes

- Struktur kedua service dibuat simetris untuk konsistensi dan maintainability
- Layering mengikuti pola:

```
Controller → Service → Repository → Database
```

- DTO (request/response) dipisahkan dari domain model

- Tidak ada shared database antar service

---

## Service Responsibilities

### Order Service

Bertanggung jawab atas:

- pembuatan dan manajemen order
- perhitungan total amount
- perubahan status order berdasarkan hasil pembayaran

### Public Endpoints

- POST /orders
- GET /orders
- GET /orders/{orderId}
- PUT /orders/{orderId}
- DELETE /orders/{orderId}

### Internal Endpoint

- POST /internal/payment-callback

---

## Payment Service

Bertanggung jawab atas:

- pembuatan payment
- manajemen status pembayaran
- mengirim callback ke order-service

### Endpoints

- POST /payments
- GET /payments/{paymentId}
- PUT /payments/success/{paymentId}
- PUT /payments/failed/{paymentId}

---

## Payment Flow (Business Logic)

1. Client membuat order → status awal pending
2. Client membuat payment untuk order tersebut
3. Payment Service memproses payment
4. Payment Service mengirim callback ke Order Service
5. Order Service memperbarui status order:

- pending → paid jika payment sukses
- pending → failed jika payment gagal

Setiap domain tetap menjadi single source of truth untuk datanya masing-masing.

---

## Instalasi & Setup

1. Clone Repo

```bash
git clone https://github.com/DahaikaR/microservice-order-payment.git
cd microservice-order-payment
```

2. Install Dependencies

```bash
go mod tidy
```

---

## Response Standardization

Seluruh response sukses menggunakan format konsisten:

```bash
{
"code": 200,
"status": "OK",
"data": {}
}
```

---

## Testing

Jalankan order-service test:

```bash
cd order-service
go test ./... -v -coverpkg=./...
```

Jalankan payment-service test:

```bash
cd payment-service
go test ./... -v -coverpkg=./...
```

Testing mencakup:

- Controller
- Service
- Middleware
- Repository
- Helper
- Exception handling

---

## Running the Project (Local Development)

Jalankan seluruh service menggunakan Docker Compose:

```bash
docker-compose up --build
```

Default ports:

- Order Service → http://localhost:8080
- Payment Service → http://localhost:8081

---

## Author

**Dhahika Rahmadani**  
Backend Developer • Go Enthusiast  
📧 [dhahikardani@gmail.com](mailto:dhahikardani@gmail.com)
