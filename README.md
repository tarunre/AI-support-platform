# AI-Powered Customer Support Platform

SupportIQ is a production-oriented, multi-tenant customer support platform built with **Go and Gin** on the backend and **React + Vite** on the frontend.

It combines traditional helpdesk functionality with AI-powered ticket analysis, automated reply generation, SLA management, background job processing, real-time WebSocket updates, email integration, analytics, knowledge-base retrieval, and enterprise integrations.

---

## Features

### Ticket Management

* Create, update, assign, and manage support tickets
* Ticket status and priority management
* Search and filtering
* Pagination
* Ticket ownership and assignment
* Internal notes and public comments
* Complete activity timeline

### AI-Powered Support

* Automatic ticket categorization
* Priority classification
* Sentiment analysis
* AI-generated ticket analysis
* AI-generated customer replies
* Human approval workflow for AI replies
* Retry failed AI analysis
* Google Gemini and Groq provider support

### Knowledge Base & RAG

* Create and manage knowledge-base articles
* PostgreSQL-based knowledge retrieval
* Context-aware AI response generation
* Retrieval-augmented AI replies

### SLA Management

* Configurable SLA policies
* First-response deadlines
* Resolution deadlines
* Automatic SLA status tracking
* At-risk detection
* SLA escalation
* SLA breach detection
* Real-time SLA updates

### Email Integration

* IMAP inbound email processing
* SMTP outbound email
* Email threading
* Attachment support
* Email account management
* Email health monitoring

### Analytics & Reporting

* Ticket analytics
* Agent performance metrics
* AI performance metrics
* Queue monitoring
* Email analytics
* Trend analysis
* PDF/CSV report generation
* Scheduled analytics processing

### Background Processing

* Redis-backed job queue
* Dedicated worker process
* Automatic retries
* Exponential backoff
* Dead-letter queue
* AI processing jobs
* Email processing jobs
* Integration synchronization jobs

### Enterprise Integrations

SupportIQ provides integration support for:

* Jira
* Linear
* GitHub Issues
* Slack
* Microsoft Teams
* Discord
* Salesforce
* HubSpot
* Google Calendar
* Webhooks

### Real-Time Updates

WebSocket-based real-time events for:

* Ticket updates
* AI analysis completion
* AI reply generation
* SLA changes
* Background job completion
* Analytics refreshes

### Multi-Tenancy

* Tenant-isolated data
* JWT-based authentication
* Role-based access control
* Admin users
* Support agents
* SuperAdmin platform management

---

## Technology Stack

### Backend

* **Go 1.21+**
* **Gin**
* **GORM**
* **PostgreSQL**
* **Redis**
* **JWT**
* **WebSocket**
* **Google Gemini API**
* **Groq API**

### Frontend

* **React 18**
* **Vite**
* **Tailwind CSS**
* **React Router**
* **Axios**
* **Recharts**

### Infrastructure

* Docker
* GitHub Actions
* PostgreSQL
* Redis

---

## Project Structure

```text
supportiq-ai-support-platform/
│
├── backend/
│   ├── cmd/
│   │   ├── main.go
│   │   └── worker/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── ai/
│   │   ├── analytics/
│   │   ├── config/
│   │   ├── database/
│   │   ├── dto/
│   │   ├── email/
│   │   ├── events/
│   │   ├── handlers/
│   │   ├── integrations/
│   │   ├── jwt/
│   │   ├── knowledge/
│   │   ├── middleware/
│   │   ├── models/
│   │   ├── queue/
│   │   ├── repositories/
│   │   ├── routes/
│   │   ├── services/
│   │   ├── utils/
│   │   └── websocket/
│   │
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── contexts/
│   │   ├── layouts/
│   │   ├── pages/
│   │   ├── routes/
│   │   └── services/
│   │
│   ├── package.json
│   └── vite.config.js
│
├── docs/
├── .github/
│   └── workflows/
│
└── README.md
```

---

# ⚙️ Prerequisites

Install the following before running the project:

| Requirement | Version |
| ----------- | ------- |
| Go          | 1.21+   |
| Node.js     | 18+     |
| PostgreSQL  | 14+     |
| Redis       | 7+      |

Redis is required for background processing and real-time job workflows.

---

# Running Locally

## 1. Clone the Repository

```bash
git clone https://github.com/YOUR_USERNAME/supportiq-ai-support-platform.git
cd supportiq-ai-support-platform
```

---

## 2. Configure the Backend

Navigate to the backend:

```bash
cd backend
```

Copy the example environment file:

```bash
cp .env.example .env
```

On Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

Configure the required environment variables.

Example:

```env
PORT=8080
APP_ENV=development

DATABASE_URL=postgres://postgres:password@localhost:5432/supportiq?sslmode=disable

JWT_ACCESS_SECRET=your_access_secret
JWT_REFRESH_SECRET=your_refresh_secret

GEMINI_API_KEY=your_gemini_api_key
GEMINI_MODEL=gemini-2.0-flash

REDIS_URL=redis://localhost:6379
QUEUE_NAME=ai_jobs

WORKER_COUNT=3
MAX_RETRIES=3
RETRY_DELAY=5

WEBSOCKET_ORIGIN=http://localhost:5173

VITE_API_URL=http://localhost:8080
```

**Never commit `.env` or API keys to GitHub.**

---

## 3. Create the PostgreSQL Database

Create a PostgreSQL database named:

```text
supportiq
```

For example:

```bash
createdb supportiq
```

Or create it using pgAdmin.

The application uses GORM migrations to create the required database tables.

---

## 4. Install Backend Dependencies

From the `backend` directory:

```bash
go mod tidy
```

Start the backend:

```bash
go run ./cmd
```

The API will run on:

```text
http://localhost:8080
```

Health check:

```text
http://localhost:8080/api/v1/health
```

---

# Running the Frontend

Open another terminal:

```bash
cd frontend
```

Install dependencies:

```bash
npm install
```

Create the frontend environment file:

```bash
cp .env.example .env
```

On Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

Set:

```env
VITE_API_URL=http://localhost:8080
```

Start the frontend:

```bash
npm run dev
```

The frontend will normally be available at:

```text
http://localhost:5173
```

---

# Running the Background Worker

The worker processes asynchronous jobs such as AI analysis, email processing, and integration events.

Open another terminal:

```bash
cd backend
```

Run:

```bash
go run ./cmd/worker
```

The worker requires Redis to be configured.

---

# Running with Docker

SupportIQ also includes Docker configuration for containerized deployment.

Build the backend image:

```bash
cd backend
docker build -t supportiq-backend .
```

Run the container:

```bash
docker run -p 8080:8080 supportiq-backend
```

For production deployments, configure PostgreSQL, Redis, environment variables, and the frontend separately or through a container orchestration solution.

---

# Hosting / Deployment

SupportIQ can be deployed using platforms that support Go, React, PostgreSQL, and Redis.

A typical production architecture is:

```text
                    ┌──────────────────┐
                    │    Frontend      │
                    │   React + Vite    │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │   Go / Gin API   │
                    │     :8080        │
                    └──────┬─────┬─────┘
                           │     │
                 ┌─────────┘     └──────────┐
                 ▼                          ▼
        ┌────────────────┐          ┌────────────────┐
        │  PostgreSQL    │          │     Redis      │
        │   Database     │          │ Queue / PubSub │
        └────────────────┘          └───────┬────────┘
                                           │
                                           ▼
                                  ┌─────────────────┐
                                  │ Background      │
                                  │ Worker          │
                                  └─────────────────┘
```

## Backend Hosting

Deploy the Go backend to a cloud platform that supports Go applications.

Set the start command to:

```bash
go run ./cmd
```

For a production build:

```bash
go build -o supportiq ./cmd
```

Then run:

```bash
./supportiq
```

Configure the following environment variables on the hosting platform:

```text
DATABASE_URL
JWT_ACCESS_SECRET
JWT_REFRESH_SECRET
GEMINI_API_KEY
GROQ_API_KEY
REDIS_URL
WEBSOCKET_ORIGIN
APP_URL
```

---

## Frontend Hosting

Build the React application:

```bash
cd frontend
npm install
npm run build
```

The production files will be generated in:

```text
frontend/dist/
```

These files can be hosted using a static hosting provider.

Set:

```env
VITE_API_URL=https://your-backend-domain.com
```

Build again after changing the environment variable:

```bash
npm run build
```

---

# Security

Before deploying SupportIQ publicly:

* Use strong JWT secrets.
* Never commit `.env` files.
* Never expose Gemini/Groq API keys.
* Use HTTPS in production.
* Configure CORS for trusted frontend domains.
* Use secure database credentials.
* Restrict PostgreSQL access.
* Restrict Redis access.
* Configure secure WebSocket origins.
* Rotate compromised credentials immediately.
* Store production secrets in the hosting provider's secret manager.

---

# User Roles

| Role          | Description                                |
| ------------- | ------------------------------------------ |
| Admin         | Full access within a tenant                |
| Support Agent | Manage tickets and customer interactions   |
| SuperAdmin    | Manage tenants and platform-wide resources |

---

# Core Workflow

```text
Customer
   │
   ▼
Create Support Ticket
   │
   ▼
AI Ticket Analysis
   │
   ├── Category
   ├── Priority
   └── Sentiment
   │
   ▼
Assign Support Agent
   │
   ▼
SLA Monitoring
   │
   ▼
Knowledge Base Retrieval
   │
   ▼
AI Reply Generation
   │
   ▼
Human Approval
   │
   ▼
Send Customer Response
   │
   ▼
Resolve Ticket
```

---

# 🔌 API

The API uses RESTful endpoints under:

```text
/api/v1/
```

Example endpoints:

```text
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/auth/me

POST   /api/v1/tickets
GET    /api/v1/tickets
GET    /api/v1/tickets/:id
PUT    /api/v1/tickets/:id
PATCH  /api/v1/tickets/:id/status

GET    /api/v1/analytics/overview
GET    /api/v1/analytics/tickets
GET    /api/v1/analytics/agents

GET    /api/v1/knowledge
POST   /api/v1/knowledge

GET    /api/v1/sla-policies
POST   /api/v1/sla-policies

GET    /api/v1/integrations
POST   /api/v1/integrations

GET    /api/v1/health
GET    /api/v1/ws
```

Authenticated endpoints require:

```text
Authorization: Bearer <access_token>
```

---

# Testing

Run backend tests:

```bash
cd backend
go test ./...
```

Build the backend:

```bash
go build -o supportiq ./cmd
```

Build the frontend:

```bash
cd frontend
npm run build
```

