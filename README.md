# Computer Management API

A REST API for managing company-issued computers built with Go, GORM, and PostgreSQL/SQLite.

## Overview

This API allows system administrators to track company-issued computers, assign them to employees, and receive notifications when an employee has 3 or more computers assigned.

## Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        CLI[curl/HTTP Client]
        API_CALLS[API Calls]
    end

    subgraph "HTTP Layer"
        ROUTER[Gorilla Mux Router]
        MIDDLEWARE[Middleware<br/>- Logging<br/>- CORS]
        HANDLERS[Computer Handler<br/>- CreateComputer<br/>- GetAllComputers<br/>- GetComputerByID<br/>- UpdateComputer<br/>- DeleteComputer<br/>- GetComputersByEmployee]
    end

    subgraph "Business Logic Layer"
        SERVICE[Computer Service<br/>- Validation<br/>- 3+ Computer Check<br/>- Business Rules]
    end

    subgraph "Data Layer"
        REPO[Computer Repository<br/>- CRUD Operations<br/>- Count Operations]
        DB[(Database<br/>SQLite/PostgreSQL)]
    end

    subgraph "External Services"
        NOTIFY_CLIENT[Notification Client<br/>- Retry Logic<br/>- Exponential Backoff]
        MOCK_SERVER[Mock Notification Server<br/>:9090/api/notify]
        GREENBONE_SERVER[Greenbone Notification<br/>:8080/api/notify]
    end

    subgraph "Models"
        COMPUTER_MODEL[Computer Model<br/>- ID: uint<br/>- MACAddress: string*<br/>- ComputerName: string*<br/>- IPAddress: string*<br/>- EmployeeAbbreviation: *string<br/>- Description: string<br/>- CreatedAt/UpdatedAt]
    end

    CLI --> ROUTER
    API_CALLS --> ROUTER
    ROUTER --> MIDDLEWARE
    MIDDLEWARE --> HANDLERS
    HANDLERS --> SERVICE
    SERVICE --> REPO
    SERVICE --> NOTIFY_CLIENT
    REPO --> DB
    NOTIFY_CLIENT -.-> MOCK_SERVER
    NOTIFY_CLIENT -.-> GREENBONE_SERVER
    
    SERVICE --> COMPUTER_MODEL
    REPO --> COMPUTER_MODEL
    HANDLERS --> COMPUTER_MODEL

    classDef clientLayer fill:#e1f5fe
    classDef httpLayer fill:#f3e5f5
    classDef businessLayer fill:#e8f5e8
    classDef dataLayer fill:#fff3e0
    classDef externalLayer fill:#fce4ec
    classDef modelLayer fill:#f1f8e9

    class CLI,API_CALLS clientLayer
    class ROUTER,MIDDLEWARE,HANDLERS httpLayer
    class SERVICE businessLayer
    class REPO,DB dataLayer
    class NOTIFY_CLIENT,MOCK_SERVER,GREENBONE_SERVER externalLayer
    class COMPUTER_MODEL modelLayer
```

## API Flow Diagram

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Database
    participant NotificationClient
    participant NotificationServer

    Note over Client,NotificationServer: Create Computer Flow

    Client->>Handler: POST /api/computers
    Handler->>Service: CreateComputer(computer)
    
    Service->>Service: validateComputer()
    Service->>Repository: CountByEmployee(abbr)
    Repository->>Database: SELECT COUNT(*) WHERE employee_abbreviation = ?
    Database-->>Repository: currentCount
    Repository-->>Service: currentCount
    
    Service->>Repository: Create(computer)
    Repository->>Database: INSERT INTO computers
    Database-->>Repository: success
    Repository-->>Service: success
    
    alt currentCount + 1 >= 3
        Service->>NotificationClient: SendNotification() [async]
        NotificationClient->>NotificationServer: POST /api/notify
        Note over NotificationClient,NotificationServer: Retry with exponential backoff
        NotificationServer-->>NotificationClient: 200 OK
    end
    
    Service-->>Handler: success
    Handler-->>Client: 201 Created + computer data

    Note over Client,NotificationServer: Get Computers Flow

    Client->>Handler: GET /api/computers
    Handler->>Service: GetAllComputers()
    Service->>Repository: GetAll()
    Repository->>Database: SELECT * FROM computers
    Database-->>Repository: computers[]
    Repository-->>Service: computers[]
    Service-->>Handler: computers[]
    Handler-->>Client: 200 OK + computers data

    Note over Client,NotificationServer: Update Computer (Reassign) Flow

    Client->>Handler: PUT /api/computers/{id}
    Handler->>Service: UpdateComputer(computer)
    Service->>Repository: GetByID(id)
    Repository->>Database: SELECT * FROM computers WHERE id = ?
    Database-->>Repository: existingComputer
    Repository-->>Service: existingComputer
    
    Service->>Service: validateComputer()
    Service->>Repository: Update(computer)
    Repository->>Database: UPDATE computers SET ... WHERE id = ?
    Database-->>Repository: success
    Repository-->>Service: success
    
    alt employee changed AND newEmployee != ""
        Service->>Repository: CountByEmployee(newEmployee)
        Repository->>Database: SELECT COUNT(*) WHERE employee_abbreviation = ?
        Database-->>Repository: count
        Repository-->>Service: count
        
        alt count >= 3
            Service->>NotificationClient: SendNotification() [async]
            NotificationClient->>NotificationServer: POST /api/notify
            NotificationServer-->>NotificationClient: 200 OK
        end
    end
    
    Service-->>Handler: success
    Handler-->>Client: 200 OK + updated computer
```

## Quick Start

### Local Development
```bash
# Clone and setup
git clone <repository-url>
cd computer-management-api
go mod tidy

# Run with SQLite (default)
go run cmd/api/main.go

# Run mock notification server (separate terminal)
go run cmd/mock-notify/main.go
```

### Docker Setup
```bash
docker-compose up --build
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/computers` | Create a new computer |
| GET | `/api/computers` | Get all computers |
| GET | `/api/computers/{id}` | Get computer by ID |
| PUT | `/api/computers/{id}` | Update computer |
| DELETE | `/api/computers/{id}` | Delete computer |
| GET | `/api/employees/{abbr}/computers` | Get computers by employee |
| GET | `/api/health` | Health check |

## Usage Examples

### Create Computer
```bash
curl -X POST http://localhost:8080/api/computers \
  -H "Content-Type: application/json" \
  -d '{
    "mac_address": "00:11:22:33:44:55",
    "computer_name": "MacBook Pro 2023",
    "ip_address": "192.168.1.100",
    "employee_abbreviation": "mmu",
    "description": "Development laptop"
  }'
```

### Get All Computers
```bash
curl http://localhost:8080/api/computers
```

### Test 3-Computer Notification
```bash
# Create 3 computers for employee "mmu" to trigger notification
for i in {1..3}; do
  curl -X POST http://localhost:8080/api/computers \
    -H "Content-Type: application/json" \
    -d "{
      \"mac_address\": \"00:11:22:33:44:0$i\",
      \"computer_name\": \"Computer $i\",
      \"ip_address\": \"192.168.1.$i\",
      \"employee_abbreviation\": \"mmu\",
      \"description\": \"Test computer $i\"
    }"
done
```

## Data Model

**Computer:**
- MAC Address (required, 17 chars, unique)
- Computer Name (required, max 100 chars)
- IP Address (required, max 15 chars)
- Employee Abbreviation (optional, exactly 3 lowercase letters)
- Description (optional, max 500 chars)

## Notification System

When an employee is assigned 3+ computers, the system sends a notification:

```json
{
  "level": "warning",
  "employeeAbbreviation": "mmu",
  "message": "Employee mmu has been assigned 3 computers"
}
```

Features:
- Automatic triggering on 3+ computer assignment
- Retry logic with exponential backoff
- Mock notification server included for testing

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `DB_TYPE` | Database type (sqlite/postgres) | `sqlite` |
| `DATABASE_URL` | Database connection string | `computers.db` |
| `NOTIFICATION_URL` | Notification service URL | `http://localhost:9090` |
| `PORT` | Server port | `8080` |
(
➜  export DB_TYPE=sqlite
➜  export DATABASE_URL=computers.db
➜  export NOTIFICATION_URL=http://localhost:9090
➜  export PORT=8080
)

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Test specific package
go test ./pkg/services/
```

## Project Structure

```
├── cmd/
│   ├── api/main.go          # Main API server
│   └── mock-notify/main.go  # Mock notification server
├── pkg/
│   ├── handlers/            # HTTP handlers
│   ├── services/            # Business logic
│   ├── models/              # Data models & repository
│   └── notifications/       # Notification client
├── internal/db/             # Database setup
├── docker-compose.yml       # Docker services
└── Dockerfile              # Container build
```