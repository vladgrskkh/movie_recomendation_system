# 🎬 Movie Recommendation System



![Work in Progress](https://img.shields.io/badge/status-WIP-yellow)

[Swagger Docs](https://movierecomendation.duckdns.org/v1/swagger/)
A modern movie recommendation platform built with **Go** and **Python**, designed to provide personalized film suggestions through a clean REST API.

## 🚀 Features

- **RESTful API (Go)**: Fast and clean HTTP API for client interaction  
- **gRPC Communication**: High-performance connection between Go and Python services  
- **Machine Learning (Python)**: Personalized recommendation engine based on user preferences (work in progress)  
- **User Authentication**: Secure JWT-based authorization system  
- **Database Persistence**: PostgreSQL backend with optimized indexing  
- **Monitoring**: Prometheus metrics and Grafana dashboards  
- **API Documentation**: Swagger UI integration for all endpoints  
- **CI/CD**: Automated build and deployment using GitHub Actions  
- **Containerization**: Docker & Docker Compose for consistent environment setup
- **Reverse Proxy**: Automatic HTTPS and reverse proxy with Caddy   

---

## 🏗️ Architecture

The system is composed of two core services connected via gRPC:

Client → Caddy → REST API (Go) → ML Service (Python) → Recommendations
↓
PostgreSQL

- **Go Backend**: Handles REST API requests, user authentication, and communicates with the ML service.  
- **Python ML Service**: Generates personalized recommendations using machine learning models.  
- **gRPC**: Internal service-to-service communication.  

---

## 🛠️ Tech Stack

- **Language**: Go 1.25.3
- **Router**: Chi
- **Database**: PostgreSQL 18
- **Authentication**: JWT tokens
- **Logging**: Structured logging with slog
- **Monitoring**: Prometheus metrics, Grafana visualization
- **Containerization**: Docker & Docker Compose
- **Documentation**: Swag
- **Reverse Proxy**: Caddy

---

## 📡 API Endpoints

### Health
- `Get /v1/healthcheck` — Health check 

### Movies
- `GET /v1/movie` — List all movies  
- `GET /v1/movie/{id}` — Get movie details  
- `POST /v1/movie` — Add a movie
- `POST /v1/movie/predict` — Get a movie recommendation
- `DELETE /v1/movie/{id}` — Delete a movie  
- `PATCH /v1/movie/{id}` — Update a movie (supports partial update)

### Users
- `POST /v1/users` — Resister a new user
- `PUT /v1/users/activate` — Activate a user
- `PUT /v1/users/password` — Update user password

### Authentication
- `POST /v1/tokens/authentication` — Login user
- `POST /v1/tokens/password-reset` — Create token for password reset
- `PUT /v1/tokens/refresh` — Create refresh token

### System
- `GET /metrics` — Prometheus metrics  
- `GET /swagger/` — API documentation  

---

## 🧩 Database Schema

Core tables include:

- **users** — User accounts
- **movies** — Movie catalog and metadata  
- **tokens** — Tokens for activation and password reset

---

## 🛠️ Development

Configure .env files to run docker compose(see docker-compose.yml)

- **Docker Compose**: Run all services with `docker compose up -d --build`  
- **Makefile**: Contains useful rules for develop, audit, generate docs and proto.

All services communicate internally via gRPC, and metrics are exposed via Prometheus for monitoring.

---

## 👨‍💻 Author

**Vladislav Gorskikh**  

---

## 📝 License

This project is created for educational purposes and can be freely used, modified, and extended.
