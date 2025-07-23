# Sabor da Rondonia - Product Microservice

This repository contains the source code for the Product Management Microservice for the "Sabor da Rondonia" project. It is a backend service written in Go, responsible for handling all CRUD (Create, Read, Update, Delete) operations for products, including inventory management and image uploads.

This service is designed to be consumed by other backend services (like the main Node.js API) and is secured via an API key.

## Features

-   **RESTful API** for product management.
-   Built with **Go** and the **Fiber** web framework for high performance.
-   Uses **GORM** for database interaction with a **PostgreSQL** database.
-   **Secure Endpoints** protected by an API key middleware.
-   **Image Uploads** handled by **Cloudinary** for scalable and persistent storage.
-   **Automated API Documentation** with Swagger.
-   **Unit and Integration Tests** with coverage reports.

## Prerequisites

Before you begin, ensure you have the following installed on your local machine:

-   [Go](https://golang.org/doc/install) (version 1.18 or higher)
-   A running [PostgreSQL](https://www.postgresql.org/) instance.
-   [Make](https://www.gnu.org/software/make/) (optional, for running test commands easily). For Windows, you can install it via [Chocolatey](https://chocolatey.org/install) (`choco install make`).

## Getting Started

Follow these steps to get the project up and running on your local machine.

### 1. Clone the Repository

```bash
git clone <your-repository-url>
cd products
```

### 2. Set Up Environment Variables

The application requires a `.env` file in the root directory to store sensitive information like database credentials and API keys.

Create a file named `.env` and copy the contents of `.env.example` into it.

**.env.example**
```env
# PostgreSQL Database Connection URL
DATABASE_URL="postgres://user:password@localhost:5432/sabordarondoniadb?sslmode=disable"

# Secret key for inter-service communication
API_SECRET_KEY="your-long-and-random-secret-key"

# Cloudinary Environment Variable URL for image uploads
CLOUDINARY_URL="cloudinary://API-Key:API-Secret@cloud_name"
```

Fill in the `.env` file with your actual credentials for PostgreSQL and Cloudinary.

### 3. Install Dependencies

The project uses Go Modules to manage dependencies. Run the following command to download and install all the required packages:

```bash
go mod tidy
```
### 4. Run the Application

To start the server, run:

```bash
go run main.go
```

The server will start on `http://localhost:3000` by default.

## API Endpoints

All endpoints are prefixed with `/api`. Access to the product endpoints requires an `X-API-KEY` header with the value defined in your `.env` file.

-   `POST /products`: Create a new product.
-   `GET /products`: Get a list of all products.
-   `GET /products/:id`: Get a single product by its ID.
-   `PATCH /products/:id`: Partially update a product's details.
-   `DELETE /products/:id`: Delete a product.
-   `POST /products/:id/upload`: Upload an image for a product.
-   `POST /products/batch`: Get multiple products by a list of IDs.
-   `POST /products/:id/stock`: Update a product's stock.
### API Documentation

This project uses Swagger for API documentation. Once the server is running, you can access the interactive documentation at:

[http://localhost:3000/swagger/index.html](http://localhost:3000/swagger/index.html)

## Running Tests

This project includes a suite of unit and integration tests.

### Run All Tests

To run all tests, execute the following command from the root directory:

```bash
go test ./...
```

### Run Tests with Coverage Report

To run the tests and generate a visual HTML coverage report (excluding generated files), you can use the `Makefile`:

```bash
make test-coverage
```

This will run the tests and automatically open the coverage report in your browser.

If you don't have `make` installed, you can run the commands manually:

```bash
# Generate the coverage file
go test -coverpkg="./handlers,./middleware,./models" -coverprofile=coverage.out ./...

# View the HTML report
go tool cover -html coverage.out
```