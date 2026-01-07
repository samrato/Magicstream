
---

# MagicStream ✨

**Backend API for a Movie Streaming Platform** built with **Go (gin-gonic) and MongoDB**.

This backend provides RESTful API endpoints for **user management, movie handling, genres, and AI-powered recommendations**. Designed for scalability, security, and easy integration with any frontend.

---

## Table of Contents

* [About](#about)

* [Features](#features)

* [Tech Stack](#tech-stack)

* [Installation](#installation)

* [Environment Variables](#environment-variables)

* [API Documentation](#api-documentation)

  * [Unprotected Routes](#unprotected-routes)
  * [Authenticated Routes](#authenticated-routes)
  * [Admin Routes](#admin-routes)

* [Folder Structure](#folder-structure)

* [License](#license)

---

## About

MagicStream backend simulates a modern movie streaming service API.

* Handles **user registration, login, and authentication**
* CRUD operations for **movies** and **genres**
* Admin review updates and **AI-powered recommendations**
* Secure JWT-based authentication and refresh tokens
* Built with **Go + gin-gonic** for high performance
* Uses **MongoDB** for scalable document storage

---

## Features

* **User Management**: Signup, login, logout, JWT auth, role-based access
* **Movie Management**: Add, fetch, and update movies
* **Recommendation System**: Fetch recommended movies based on AI logic
* **Secure Backend**: Input validation, hashed passwords, token refresh
* **CORS Support**: Configurable allowed origins

---

## Tech Stack

| Layer               | Technology        |
| ------------------- | ----------------- |
| Backend / API       | Go / gin-gonic    |
| Database            | MongoDB           |
| AI / Recommendation | OpenAI (optional) |

---

## Installation

1. Clone the repository:

```bash
git clone https://github.com/samrato/MagicStream.git
cd MagicStream/Server/MagicStreamServer
```

2. Install Go dependencies:

```bash
go mod tidy
```

3. Create a `.env` file in the root:

```env
MONGO_URI=<your-mongodb-uri>
JWT_SECRET=<your-secret-key>
JWT_REFRESH_SECRET=<your-refresh-secret>
ALLOWED_ORIGINS=http://localhost:5173
```

4. Run the server:

```bash
go run main.go
```

Server will start at: `http://localhost:8080`

---

## API Documentation

### Unprotected Routes (No Auth Required)

| Method | Endpoint               | Description                       |
| ------ | ---------------------- | --------------------------------- |
| POST   | `/users/register`      | Register a new user               |
| POST   | `/users/login`         | Login user and get JWT tokens     |
| POST   | `/users/refresh-token` | Refresh JWT token                 |
| GET    | `/movies`              | Fetch all movies                  |
| GET    | `/movies/:imdb_id`     | Fetch a specific movie by IMDb ID |
| GET    | `/movies/recommended`  | Fetch recommended movies          |
| GET    | `/genres`              | Fetch all genres                  |

---

### Authenticated Routes (JWT Required)

> Middleware: `AuthMiddleware()`

| Method | Endpoint                  | Description                           |
| ------ | ------------------------- | ------------------------------------- |
| GET    | `/users/profile`          | Get logged-in user profile            |
| PUT    | `/users/favourite-genres` | Update user's favourite genres        |
| POST   | `/users/logout`           | Logout user (invalidate token)        |
| POST   | `/movies`                 | Add a new movie (authenticated users) |

---

### Admin Routes (JWT + Admin Role)

> Middleware: `AuthMiddleware() + AdminOnly()`

| Method | Endpoint                        | Description                              |
| ------ | ------------------------------- | ---------------------------------------- |
| PUT    | `/admin/movies/:imdb_id/review` | Update admin review and ranking of movie |

---

## Folder Structure

```
MagicStreamServer/
├── controllers/     # API handlers (business logic)
├── database/        # MongoDB connection and setup
├── middleware/      # JWT auth middleware
├── models/          # Database schemas (User, Movie, Genre)
├── routes/          # Route definitions
├── utils/           # Helper functions (JWT, hashing, env)
└── main.go          # Entry point of the application
```

---

## Environment Variables

| Variable             | Description                                  |
| -------------------- | -------------------------------------------- |
| `MONGO_URI`          | MongoDB connection string                    |
| `JWT_SECRET`         | Secret key for JWT signing                   |
| `JWT_REFRESH_SECRET` | Secret key for refresh token                 |
| `ALLOWED_ORIGINS`    | Comma-separated list of allowed CORS origins |

---

## Example Usage

**Register a User:**

```bash
POST /users/register
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "password123",
  "favourite_genres": [{"genre_id":1,"genre_name":"Action"}]
}
```

**Login:**

```bash
POST /users/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**

```json
{
  "user_id": "12345",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "role": "USER",
  "token": "<jwt-token>",
  "refresh_token": "<refresh-token>",
  "favourite_genres": [{"genre_id":1,"genre_name":"Action"}]
}
```

---

## License

MIT License © 2026 MagicStream

---


