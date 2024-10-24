# Healthcare Service API - Golang

This repository contains the **Healthcare Service API**, a scalable and high-performance service built with **Golang** for managing patient healthcare records, appointments, and notifications. The API is optimized for **low latency** and **high concurrency**, leveraging Golang.

## Features

- **Patient Management:** Handle patient information efficiently with CRUD operations (Create, Retrieve, Update, Delete).
- **Appointment Scheduling:** Seamless management of appointment booking, rescheduling, and cancellations.
- **Healthcare Records:** Store and retrieve patient medical history and records securely.
- **Notifications:** Send email/SMS notifications for appointment reminders and updates.
- **JWT-based Authentication:** Secure access to all API endpoints with JSON Web Token (JWT) for authentication.
- **Database Integration:** Seamless data management with PostgreSQL and MongoDB.
- **Redis Caching:** Utilize Redis for caching and rate limiting to enhance performance. (coming Soon)
- **RabbitMQ:** Process Request Asynchronously (coming Soon)
- **High Performance:** Optimized for handling large requests concurrently with low response time.
- **Docker Support:** Deploy easily with Docker for a consistent and reliable environment.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Requirements

- Go 1.20+
- PostgreSQL and MongoDB
- Docker (For containerization)

## Environment Variables

Make sure to set up the following environment variables:

```bash
MONGOURL=mongodb://rootuser:rootuser@mongodb:27017 
POSTGRES=postgres://rootuser:rootuser@postgres:5432/postgres?sslmode=disable
PORT=:3000
KEY=VAIBHAVYADAV
```


## Installation

1. Clone the repository:
    ```bash
    https://github.com/BharatSeva/HealthCare-Go-Server.git
    cd HealthCare-Go-Server
    ```

2. Install Go modules:
    ```bash
    go mod download
    ```

3. Run Docker to set up PostgreSQL, MongoDB, and Redis:
    ```bash
    docker-compose up -d
    ```

4. Start the server:
    ```bash
    go run main.go
    ```

5. Alternatively, you can start the docker container for same (make sure you've set .env file before this else it will be rejected)
    ```bash
    docker-compose up -d
    ```




## API Endpoints  
Please find Postman API Collection [here](./Golang_HealthCare_BharatSeva.postman_collection.json)  


## LICENSE  
This project is licensed under the Apache-2.0 license - see the [LICENSE](./LICENSE) file for details.   


## Contributing
Please find [CONTRIBUTING](./CONTRIBUTING.md) file to know how to get started with contributing.  
