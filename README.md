# OTP-Based Authentication Service

A Go backend service that implements OTP-based login and registration with user management features.

## Features

- **OTP Authentication**: Phone number-based OTP login and registration
- **Rate Limiting**: Maximum 3 OTP requests per phone number within 10 minutes
- **User Management**: REST endpoints for user retrieval with pagination and search
- **JWT Tokens**: Standard JWT token generation for authenticated sessions
- **Database**: PostgreSQL for persistent user data storage
- **Caching**: Redis for OTP storage and rate limiting
- **API Documentation**: Swagger/OpenAPI documentation
- **Containerization**: Docker and docker-compose setup

## Architecture

The service follows a clean, layered architecture:

```
cmd/server/          # Application entry point
internal/
├── api/             # HTTP handlers and routing
├── auth/            # JWT token generation and validation
├── config/          # Configuration management
├── db/              # Database operations and models
├── otp/             # OTP generation, validation, and rate limiting
└── types/           # Data structures and types
migirations/         # Database schema migrations
docs/                # Swagger documentation
```

## Database Choice: PostgreSQL

**Why PostgreSQL?**
For this OTP authentication service, PostgreSQL provides the perfect balance of reliability, performance, and features needed for user management and authentication systems.

## Quick Start with Docker

1. **Clone the repository**:
   ```bash
   git clone https://github.com/MiladJlz/dekamond-task.git
   ```

2. **Run with Docker Compose**:
   ```bash
   docker-compose up --build
   ```

3. **Access the service**:
   - API Base: http://localhost:8080/v1
   - Swagger Documentation: http://localhost:8080/v1/swagger/


## Database Migrations

Migrations are applied automatically by PostgreSQL on container start via files in `migirations/`.

## API Endpoints

### Authentication

#### Request OTP
```http
POST /v1/request-otp
Content-Type: application/json

{
  "phone": "+1234567890"
}
```

**Response**:
```json
{
  "message": "OTP sent"
}
```

#### Verify OTP
```http
POST /v1/verify-otp
Content-Type: application/json

{
  "phone": "+1234567890",
  "code": "123456"
}
```

**Response**:
```json
{
  "message": "Login success",
  "token": "generated-jwt-token"
}
```

### User Management

#### Get Users List
```http
GET /v1/users?offset=0&limit=10&search=123
Authorization: Bearer <jwt-token>
```

**Response**:
```json
{
  "users": [
    {
      "id": 1,
      "phone": "+1234567890",
      "created_at": "2025-08-19T12:00:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 10
}
```

#### Get User by ID
```http
GET /v1/users/1
Authorization: Bearer <jwt-token>
```

**Response**:
```json
{
  "id": 1,
  "phone": "+1234567890",
  "created_at": "2025-08-19T12:00:00Z"
}
```

## OTP Flow

1. **User requests OTP** by sending phone number
2. **System generates** a random 6-digit OTP
3. **OTP is printed** to console (for development)
4. **OTP is stored** in Redis with 2-minute expiration
5. **Rate limiting** is applied (max 3 requests per 10 minutes)
6. **User submits** phone number + OTP
7. **System validates** OTP and creates/logs in user
8. **JWT token** is returned for authentication

## Rate Limiting

- **Maximum 3 OTP requests** per phone number within 10 minutes
- **429 Too Many Requests** response when limit exceeded
- **Redis-based** rate limiting for scalability

## Security Features

- **Standard JWT tokens** for session management
- **OTP expiration** after 2 minutes
- **Rate limiting** to prevent abuse
- **Input validation** for all endpoints
- **SQL injection protection** through parameterized queries

## JWT Authentication

The service uses JWT (JSON Web Tokens) for authentication:

- **Token Format**: `Bearer <jwt-token>`
- **Algorithm**: HS256
- **Expiration**: 24 hours
- **Claims**: Phone number, issuer, issued/expiration times

### Using JWT Tokens

1. **Get token** by completing OTP verification
2. **Include in requests** as Authorization header:
   ```
   Authorization: Bearer <jwt-token>
   ```
3. **Protected endpoints** require valid JWT token

 