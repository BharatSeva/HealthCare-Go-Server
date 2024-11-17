# Bharat Seva+ Healthcare Service API 🚀  

Welcome to the **Bharat Seva+ Healthcare Service API**!  
This API is a high-performance backend solution designed to streamline healthcare management. Built using **Golang**, it handles complex workflows like managing patient records, medical history, appointments, and notifications with ease. Designed for **scalability**, **security**, and **efficiency**, this API ensures fast, reliable service in high-concurrency environments.  

---

## Table of Contents
- [Key Features](#key-features)  
- [Tech Requirements](#tech-requirements)  
- [Setup & Installation](#setup--installation)  
- [API Endpoints](#api-endpoints)  
- [License](#license)  

---

## Key Features  

- **📂 Patient Record Management:**  
  Effortless creation, retrieval, updating, and deletion (CRUD) of patient data, with robust error handling and optimized query performance.  

- **🗂️ Medical History Access:**  
  A secure and structured repository for accessing patients’ healthcare histories, ensuring seamless integrations with other systems.  

- **🔒 JWT-Based Security:**  
  End-to-end protection of API endpoints using **JSON Web Tokens (JWT)**, ensuring secure authentication and data privacy.  

- **💾 Multi-Database Integration:**  
  Harnesses the power of both **PostgreSQL** for relational data and **MongoDB** for NoSQL needs, ensuring data flexibility and resilience.  

- **⚡ Redis Caching:**  
  Implements real-time caching and advanced rate limiting, reducing response times while maintaining server health.  

- **📩 RabbitMQ for Async Tasks:**  
  Processes background tasks like notifications and logs asynchronously, ensuring smooth user experiences.  

- **🚀 High-Performance & Concurrent:**  
  Optimized for environments with high request rates, delivering rapid responses with minimal latency.  

- **🐳 Containerized with Docker:**  
  Streamlined deployments using **Docker**, enabling reliable and platform-agnostic setups.  

---

## Tech Requirements  

To run this API, you’ll need:  
- **Go** v1.22+  
- **PostgreSQL** and **MongoDB** for persistent storage  
- **Docker** for containerized deployments  
- **RabbitMQ** for task queuing  
- **Redis** for caching and rate limiting  

---

## Setup & Installation  

### 1. Clone the Repository
```bash
git clone https://github.com/BharatSeva/Healthcare-Server.git
cd Healthcare-Server
```

### 2. Configure Environment Variables  
Set up a `.env` file with the following variables for smooth deployment:  
```bash
PORT=:3002
MONGOURL=mongodb://rootuser:rootuser@mongodb:27017 
POSTGRES=postgres://rootuser:rootuser@postgres:5432/postgres?sslmode=disable
RABBITMQ=amqp://rootuser:rootuser@rabbitmq:5672/
REDIS=redis:6379
KEY=VAIBHAVYADAV
```

### 3. Install Dependencies
```bash
go mod download
```
### 4. Launch the Server Locally
```bash
go run main.go
```
Alternatively, Deploy Using Docker
```bash
docker run -d -p 3002:3002 --name healthcare --env-file .env healthcare
```
### API Endpoints
Explore the full range of available endpoints and their usage with our Postman collection.
Find it here: [Healthcare API Postman Collection](./Healthcare.postman_collection.json).

## License
This project is licensed under the AGPL-3.0 License. For more details, check the [LICENSE](./LICENSE) file.
