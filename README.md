# Banking Service REST API

A secure and efficient banking service REST API built with Go.

## Features

- User registration and authentication with JWT
- Bank account management
- Card operations (generation, viewing, payments)
- Money transfers between accounts
- Credit operations with payment schedules
- Financial analytics
- Integration with external services (Central Bank of Russia, SMTP)
- Secure data encryption and hashing

## Prerequisites

- Go 1.23+
- PostgreSQL 17
- PGP encryption tools

## Installation

1. Clone the repository:
```bash
git clone https://github.com/Abigotado/abi_banking.git
cd abi_banking
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Initialize the database:
```bash
psql -U your_user -d your_database -f migrations/init.sql
```

## Configuration

Create a `.env` file with the following variables:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=your_database
JWT_SECRET=your_jwt_secret
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_email
SMTP_PASSWORD=your_password
```

## API Endpoints

### Public Endpoints
- `POST /register` - User registration
- `POST /login` - User authentication

### Protected Endpoints
- `POST /accounts` - Create account
- `POST /cards` - Issue card
- `POST /transfer` - Transfer funds
- `GET /analytics` - Get analytics
- `GET /credits/{creditId}/schedule` - Get credit payment schedule
- `GET /accounts/{accountId}/predict` - Get balance prediction

## Security Features

- JWT-based authentication
- PGP encryption for card data
- HMAC for data integrity
- Bcrypt for password hashing
- Access control for accounts and cards

## Running the Application

```bash
go run cmd/main.go
```

## Testing

```bash
go test ./...
```

## License

MIT 