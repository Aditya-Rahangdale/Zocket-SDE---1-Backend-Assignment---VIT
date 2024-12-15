
# Product Management System

This is a Golang-based backend project for managing products. It features API endpoints for creating, retrieving, and listing products, along with PostgreSQL database integration.

## Features
- RESTful API Endpoints:
  - `POST /products`: Create a product.
  - `GET /products/{id}`: Retrieve a product by ID.
  - `GET /products`: List all products with optional filters.

## Setup Instructions
1. Clone the repository.
2. Create a PostgreSQL database and configure the `POSTGRES_CONN` environment variable.
3. Run the project:
   ```bash
   go run main.go
   ```

## Assumptions
- Environment variable `POSTGRES_CONN` should contain the connection string for the PostgreSQL database.
