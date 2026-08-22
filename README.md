# Wallet-App

## Overview

This is a RESTful API for digital **Wallet Application** built with **Go** programming language, and the **gin-gonic/gin** framework. The application manages secure financial transactions, user wallet balances, deposits, withdrawals, and transfers. To guarantee **data consistency** and prevent **race conditions** during concurrent transactions, it utilizes GORM with PostgreSQL, applying and transaction operation using strict database-level **row locking** and **atomic transaction management**. The API is secured using JWT authentication and **role-based access control**.

## Features

Our app has the following features:

### Auth & Authorization Endpoints:
  - **POST `/signup`**: Registers in the system with unique username and password and authomatically generate a wallet for new user with default balance equal to zero.
  - **POST `/login`**: Login to the system with your username and password.

### Wallet Endpoints:
- `GET /wallet` : get the current user's wallet and view its balance.
- `POST /wallet/deposit` : creates a `deposit` transaction on the user's wallet.
- `POST /wallet/withdraw` : creates a `withdraw` transaction on the user's wallet.
- `POST /wallet/transfer` : transfers money to another user's wallet and creates both `transfer_out` and `transfer_in` transactions.

### Transaction Endpoints:
- `GET /transactions` : list current user's transactions.
- `GET /transactions?category=Food` : get user's transactions filter by category.
- `GET /transactions?from=&to=` : get user's transactions filter by date range.
- `GET /transactions/summary` : get transactions summary (totals grouped by category for the current month).

### Budget Endpoints:
- `POST /budgets` : Creates a new monthly spending limit for a category.
- `PUT /budgets/:category` : Updates the limit for an existing category budget.
- `GET /budgets/status` : get spending progress, limits, and over-budget status across all budget records.

### Error Handling: 
  Return appropriate HTTP status codes with clear messages to explain the error:
  - 200: Successful operations.
  - 201: User created (on signup).
  - 400: Invalid input or JSON.
  - 401: Unauthorized (missing or invalid JWT, invalid login).
  - 403: Forbidden (user lacks required permissions).
  - 404: Resource not found (e.g., non-existent wallet).
  - 500: Internal server/database error.

### Clean Code & Architecture
  This project is built following clean code principles and modular architecture to ensure high maintainability, readability and scalable growth.
  
#### Key Architectural Highlights:
  - **Separation of Concerns**: The application logic is decoupled into **distinct layers** (model, repository, service, handler, and database configurations) to keep business logic isolated and easy to test.
  -  **Interface Based Architecture**: Lower layers define clear **Go interfaces** instead of relying on fixed implementations. This decouples the layers, make it easy to replace database implementations without touching the other layers, and allows simple mocking for unit tests.
  -  **Layer-by-Layer Validation**: Data is validated at each layer before calling the next lower layer. This ensure that every layer only receives correct and safe data, detecting errors as early as possible.
  -  **Clear Error Handling**: Errors are handled at their origin, returning simple **JSON responses** with standard HTTP status codes for easy debugging.

### Authentication & Authorization:
  This project provides a secure **JWT authentication** and **Role-Based Access Control** (RBAC) to ensure strict data privacy and isolation.

  - **JWT Authentication**: Protects all wallet and transaction endpoints by requiring a **valid JSON Web Token** generated upon Login.
  - **Strict Data Isolation**: Restrict database queries to the logged-in user's ID, ensuring regular users can only view their own wallet balance and transactions.
  - **Least Privilege**: Automatically assigns all **new registrations** the standard **user role** by default, ensuring accounts start with the minimal permissions.
  - **Admin Privileges**: Grants administrative accounts **system-wide access** to inspect and view records across all users.
  
### Concurrency & Race Condition Safety:
To handle **high-concurrency** environments and prevent critical financial issues like double-spending , lost updates, or negative balances during **Transfer operations**, The application enforces **atomic operations**:
- **Row Locking** (`FOR UPDATE`): During balance transfers, the system locks both the sender and receiver wallet rows in database (`SELECT ... FOR UPDATE`). This ensures that no concurrent HTTP requests can read or modify the same wallet balance until the active transaction completes.
- **Atomic Database Transactions**: Transfers, deposits, and withdrawals run inside isolated, atomic transactions (`db.Transaction`). If any condition fails, all state changes are automatically rolled back.

### Budget Alerts Mechanism:
- **Non-blocking Warnings**: Exceeding a budget on `withdraw` or `transfer` operations does not block the transaction. The operations completes normally and returns a warning message in response.
- **Dynamic Tracking**: Monthly spending is calculated dynamically per category to evaluate budget limits in real time.

## Testing

### Unit Testing:
  The application includes unit tests covering core business logic, input validation, user management and error handling:
  - **User Authentication**: Tests user creation (signup) and ensure **automatic wallet creation** for new users, credential checking and login logic.
  - **Wallet logic**: Tests balance checks, insufficient funds handling, deposits, withdrawals and transfer rules.
  - **Database Error Masking**: Ensures raw database errors are intercepted at the service layer and replaced with safe, generic messages to protect internal system details. 

### Integration Testing
The application includes **end-to-end** integration tests to validate full API workflows, database transactions, edge cases, and concurrency safety:
- **Transfer Rules & Validation**:
    - **Non-existent Receiver**: Confirms transfers fail gracefully when the recipient account does not exist.
    - **Insufficient Balance**: Ensures transactions are rejected when requested funds exceed current balance.
    - **Self-Transfer**: Prevents users from initiating transfers to their own account.
- **Concurrent Withdrawals**: Simulates concurrent withdrawal requests on the same wallet to guarantee row locking, atomic operations, and double-spending prevention under high load.


To run all tests, execute the following command in terminal:

  ```
    go test -v ./...
  ```

## How to Run using Docker:
Start the application and database using docker command in terminal:

```
  docker-compose up
```

After this command, the API server will start listening on: http://localhost:8080.

To stop and remove all containers use this command:

```
  docker-compose down
```

### Practical Example:

1. **Signup**: Register a new user by sending a `POST` request to `http://localhost:8080/signup` with JSON body:
   
    ```
    {
        "username": "your_name",
        "password": "your_secret_password"
    }
    ```
2. **Login**: Obtain a JWT token by sending a `POST` request to `http://localhost:8080/login` with your credentials.
3. **Deposit Funds**: Send a `POST` request to `http://localhost:8080/wallet/deposit` with your JWT attached in `Authorization` header:
   
    ```
    {
        "amount": 150,
        "category": "work",
        "note": "salary"
    }
    ```
The system will execute the operation and increase your wallet balance then generate a transaction, store it in the system and return it to you in response:

  ```
    {
        "id": 1,
        "amount": 150,
        "wallet_id": "your_wallet_id",
        "type": "deposit",
        "category": "work",
        "note": "salary",
        "created_at": "2026-08-19T10:00:00Z"
    }
  ```

## API Documentation

To view an **interactive API documentation** generated using swagger. Once the server is running, access Swagger UI at this address:  

http://localhost:8080/swagger/index.html