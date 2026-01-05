# DingTalk Approval Dashboard

A dashboard to view and manage DingTalk approval form data with scheduled syncing.

## Tech Stack

- **Backend**: Go (Fiber framework)
- **Frontend**: React (Vite)
- **Database**: PostgreSQL
- **Authentication**: External API (api-incoming.ws-allure.com)

## Setup

### Prerequisites

- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL

### 1. Start Database

```bash
docker compose up -d postgres
```

### 2. Configure Environment

Copy and edit the environment file:

```bash
cd backend
cp .env.example .env
# Edit .env with your DingTalk and JWT credentials
```

Required environment variables:
- `DINGTALK_APP_KEY` - Your DingTalk App Key
- `DINGTALK_APP_SECRET` - Your DingTalk App Secret
- `APPROVAL_PROCESS_CODE` - The approval form process code
- `JWT_SECRET` - Same JWT secret as your auth system

### 3. Run Backend

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

The backend runs on `http://localhost:8081`

### 4. Run Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on `http://localhost:5173`

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/approvals` | List approvals (with pagination/filters) |
| GET | `/api/v1/approvals/:id` | Get approval details |
| GET | `/api/v1/approvals/stats` | Dashboard statistics |
| GET | `/api/v1/sync/logs` | Sync history |
| POST | `/api/v1/sync/trigger` | Trigger manual sync |

## Scheduler

Data syncs automatically at: **8:00, 11:00, 13:00, 16:00, 18:00** (Jakarta time)

## Project Structure

```
DingTalk Dashboard/
├── backend/
│   ├── cmd/server/main.go       # Entry point
│   ├── internal/
│   │   ├── config/              # Configuration
│   │   ├── database/            # DB connection & migrations
│   │   ├── dingtalk/            # DingTalk API client
│   │   ├── domain/approval/     # Models, repository, service
│   │   ├── handler/             # HTTP handlers
│   │   ├── middleware/          # Auth & CORS
│   │   └── scheduler/           # Cron jobs
│   ├── go.mod
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── components/          # React components
│   │   ├── contexts/            # Auth context
│   │   ├── pages/               # Login, Dashboard
│   │   └── services/            # API services
│   ├── package.json
│   └── vite.config.js
└── docker-compose.yml
```
