# Strato User Management Dashboard

Strato is a full-stack user management application with a Go backend and React frontend, containerized using Docker. The application provides a comprehensive dashboard for monitoring user accounts, focusing on security metrics like password age and last access dates.

## Features

- **User Management Dashboard**
  - View all users in a sortable, filterable table
  - Add new users to the system
  - Monitor security metrics in real-time
  - Visual indicators for security concerns (color-coded rows)

- **Security Monitoring**
  - Track days since last password change
  - Monitor account access patterns
  - Filter accounts by MFA status
  - Highlight accounts requiring attention:
    - Yellow: Password older than 90 days
    - Red: No account activity for over 365 days

- **Advanced Filtering**
  - Search by username
  - Filter by date ranges (create date, password change, last access)
  - Filter by MFA status (enabled/disabled)

## Technology Stack

- **Backend**
  - Go 1.22
  - PostgreSQL database
  - RESTful API endpoints
  - Docker containerization
  - Environment-based configuration

- **Frontend**
  - React 19
  - Tailwind CSS for styling
  - Solar Icons component library
  - Responsive design

## Design & Architecture

### Performance Optimizations

- **Database Interaction Strategy**: Data is loaded from the database only when the server starts or when a new user is added, not on every request. This in-memory caching approach significantly reduces database load and improves response times.

- **Stateless API Design**: The backend implements a stateless API design pattern, allowing for horizontal scaling without session synchronization concerns.

### Frontend Architecture

- **Tailwind CSS**: Used for rapid UI development with utility-first classes, eliminating the need for custom CSS in most cases and ensuring design consistency.

- **Component-Based Design**: UI is built with reusable components to maximize code reuse and maintainability.

### Containerization

- **Multi-Stage Builds**: The backend Dockerfile uses multi-stage builds to create a minimal production image without build tools and intermediate artifacts.

- **Service Dependency Management**: Docker Compose health checks ensure services start in the correct order, preventing connection failures during startup.

### Tradeoffs

- **In-Memory Data Storage**: While providing performance benefits, the current caching approach means changes made directly to the database won't be reflected without a server restart.

- **Monolithic API**: The current design favors simplicity with a monolithic API structure rather than microservices, trading some scalability for development speed and operational simplicity.

## Project Structure

```
strato
├── backend
│   ├── Dockerfile          # Multi-stage Docker build for Go
│   ├── .env                # Environment configuration 
│   ├── go.mod              # Go dependencies
│   ├── main.go             # API server implementation
│   ├── main_test.go        # Unit tests
│   └── users.json          # Initial user data
│
├── frontend
│   ├── Dockerfile          # Node/React Docker build
│   ├── public              # Static assets
│   │   └── index.html      # Main HTML template
│   ├── src
│   │   ├── App.js          # Root React component
│   │   ├── UserTable.js    # Main dashboard component
│   │   ├── index.js        # React entry point
│   │   └── *.css           # Styling files
│   ├── package.json        # Frontend dependencies
│   └── tailwind.config.js  # Tailwind configuration
│
├── docker-compose.yml      # Multi-container orchestration
└── README.md               # Project documentation
```

## Getting Started

### Prerequisites

- Docker and Docker Compose installed on your machine

### Running the Project

1. Clone the repository to your local machine
2. Navigate to the project directory
3. Start all services with Docker Compose:

   ```
   docker-compose up
   ```

4. Access the dashboard at `http://localhost:3000`
5. The API endpoints are available at `http://localhost:8080/api/users`

### Environment Variables

The backend service uses the following environment variables:
- `DB_CONN_STR`: PostgreSQL connection string
- `GO_ENV`: Runtime environment (development/production)

These are configured in the `docker-compose.yml` file for Docker deployment.

## API Documentation

### Endpoints

- `GET /api/users`: Retrieve all users with calculated metrics
- `POST /api/users`: Add a new user
  - Required fields: `humanUser`, `createDate`, `passwordChangedDate`, `lastAccessDate`, `mfaEnabled`

## Development

To develop locally:

1. Start the database with Docker:
   ```
   docker-compose up postgres
   ```

2. Update the `.env` file in the backend directory to use localhost for the DB connection:
   ```
   DB_CONN_STR=postgres://postgres:mysecretpassword@localhost:5432/strato?sslmode=disable
   ```

3. Run the backend:
   ```
   cd backend
   go run main.go
   ```

4. Run the frontend:
   ```
   cd frontend
   npm install
   npm start
   ```

## Testing

The backend includes unit tests. Run them with:

```
cd backend
go test
```

## License

This project is licensed under the MIT License.