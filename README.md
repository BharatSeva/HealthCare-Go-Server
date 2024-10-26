![Go-Logo_Blue (1)](https://github.com/user-attachments/assets/369e83fe-82c3-463e-85fa-1fb229f5e89f)          

# Bharat Seva+ Healthcare Service API 🚀

This **Healthcare Service API**, crafted in **Golang**, powers high-concurrency environments with low-latency responses. Optimized for scalability, the API manages healthcare records, appointments, and patient notifications efficiently and securely.


## Table of Contents
- [Key Features](#key-features)
- [Tech Requirements](#tech-requirements)
- [Setup & Installation](#setup--installation)
- [API Endpoints](#api-endpoints)
- [License](#license)


## Key Features
- **Patient Record Management:** Full CRUD operations for handling patient data with robust error handling and optimized retrieval.
- **Appointments System:** End-to-end support for appointment creation, scheduling, rescheduling, and cancellations.
- **Medical History Access:** A secure repository for patients' healthcare history, accessible through structured queries.
- **Notification Services:** Automated email/SMS reminders for appointments and other key updates.
- **JWT-Based Security:** JSON Web Tokens (JWT) for secure access to API endpoints.
- **Multi-Database Integration:** Optimized data flow across **PostgreSQL**  and **MongoDB**.
- **Redis Caching** : Real-time caching and rate-limiting for optimal response times and resource efficiency.
- **RabbitMQ for Async Tasks** : Enhances processing of background tasks for smoother user experiences.
- **High-Performance & Concurrent:** Built for high-request environments with minimal response times.
- **Containerized with Docker:** Simple deployment through Docker for a reliable, cross-platform experience.


## Tech Requirements
- **Go** v1.22+
- **PostgreSQL** and **MongoDB** for data persistence
- **Docker** for containerized environments
- **RabbitMQ** for asynchronous tasks
- **Redis** for caching and advance rate limiting

## Setup & Installation
Set up the following environment variables for smooth deployment:
```bash
PORT=:3000
MONGOURL=mongodb://rootuser:rootuser@mongodb:27017 
POSTGRES=postgres://rootuser:rootuser@postgres:5432/postgres?sslmode=disable
RABBITMQ=amqp://rootuser:rootuser@rabbitmq:5672/
REDIS=redis:6379
KEY=VAIBHAVYADAV
```

1. Clone the Repository:

```bash
git clone https://github.com/BharatSeva/HealthCare-Go-Server.git
cd HealthCare-Go-Server
```

2. Install Dependencies:

```bash
go mod download
```

3. Start Database Services:

```bash
docker-compose up -d
```

4. Launch the Server:

```bash
go run main.go
```
### Alternatively, deploy using Docker:

```bash
docker-compose up -d
```

## API Endpoints
Full documentation of the endpoints and their usage is available in our Postman collection [here](./Golang_HealthCare_BharatSeva.postman_collection.json).

## License
Licensed under the AGPL-3.0 license. See the [LICENSE](./LICENSE) file for full details.

