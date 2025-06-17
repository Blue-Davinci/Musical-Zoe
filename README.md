<div align="center">
  <img src="https://i.ibb.co/svfMTWLw/musical-zoe-high-resolution-logo-modified.png" alt="Musical Zoe Logo" width="400"/>
</div>

# Musical-Zoe Backend 🎵

A robust, modular Go backend service providing three core music-related APIs with comprehensive authentication, error handling, and developer-friendly automation.

## Table of Contents

- [About](#about)
- [Features](#features)
- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Environment Setup](#environment-setup)
- [Development](#development)
- [Error Handling](#error-handling)
- [Performance](#performance)
- [Project Structure](#project-structure)
- [Contributing](#contributing)

## About <a name="about"></a>

Musical-Zoe is a production-ready Go backend that aggregates music data from multiple external APIs to provide a unified interface for music news, trending tracks/artists, and song lyrics. Built with clean architecture principles, comprehensive error handling, and optimized performance for real-world applications.

The system integrates with:
- **NewsAPI.org** - Music news with intelligent filtering
- **Last.fm** - Music trends and charts  
- **Lyrics.ovh** - Song lyrics with text processing and analytics

## Features <a name="features"></a>

### 🎵 **Core Music Services**
- **Music News**: Real-time music news with genre filtering and music-specific content curation
- **Music Trends**: Trending tracks and artists from Last.fm with flexible time periods
- **Song Lyrics**: Lyrics retrieval with text cleaning, analytics, and processing

### 🔐 **Security & Authentication**
- Bearer token authentication for all protected endpoints
- CORS configuration for frontend integration
- Rate limiting and timeout protection

### ⚡ **Performance & Reliability**
- Retryable HTTP client with smart backoff strategies
- Optimized timeouts (5-11 seconds max response time)
- Comprehensive error handling with proper HTTP status codes
- Database connection pooling with health checks

### 🛠 **Developer Experience**
- One-command development setup with Docker Compose
- Automated database migrations using Goose
- Structured JSON logging with request tracing
- Comprehensive Makefile for all operations
- Environment variable management with .env support

## Quick Start <a name="quick-start"></a>

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make (optional but recommended)

### 1. Clone & Setup
```bash
git clone <repository-url>
cd Musical-Zoe
```

### 2. Environment Configuration
```bash
# Copy and edit environment variables
cp cmd/api/.env.example cmd/api/.env
```

Required environment variables:
```env
# API Keys
MUSICALZOE_NEWS_API_KEY=your_newsapi_key
MUSICALZOE_LASTFM_API_KEY=your_lastfm_key

# Database
MUSICALZOE_DB_DSN=postgres://MUSICALZOE:test@localhost/MUSICALZOE?sslmode=disable

# Optional: SMTP for email features
MUSICALZOE_SMTP_HOST=your_smtp_host
MUSICALZOE_SMTP_USERNAME=your_smtp_user
MUSICALZOE_SMTP_PASSWORD=your_smtp_password
```

### 3. Start Development Environment
```bash
# One command to rule them all
make dev-start

# Or step by step
make db-up        # Start PostgreSQL
make db-migrate   # Run migrations  
make run/api      # Start API server
```

### 4. Verify Setup
```bash
# Health check
curl http://localhost:4000/v1/health

# Test endpoint (requires authentication)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:4000/v1/musical/trends?limit=5"
```

## API Endpoints <a name="api-endpoints"></a>

All endpoints require Bearer token authentication except health check.

### Authentication
```bash
# Get API token (requires user registration first)
curl -X POST http://localhost:4000/v1/api/authentication \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'
```

### Core Endpoints

#### 🎵 Music News
```bash
GET /v1/musical/news
```
**Parameters:**
- `limit` (optional): Number of articles (1-100, default: 20)
- `country` (optional): Country code (default: "us")
- `type` (optional): "headlines" or "everything" (default: "everything")  
- `genre` (optional): Music genre filter

**Example:**
```bash
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:4000/v1/musical/news?genre=rock&limit=10"
```

#### 📈 Music Trends  
```bash
GET /v1/musical/trends
```
**Parameters:**
- `limit` (optional): Number of items (1-200, default: 50)
- `period` (optional): 7day, 1month, 3month, 6month, 12month, overall
- `type` (optional): "tracks" or "artists" (default: "tracks")

**Example:**
```bash
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:4000/v1/musical/trends?type=artists&period=1month&limit=20"
```

#### 🎤 Song Lyrics
```bash  
GET /v1/musical/lyrics
```
**Parameters:**
- `artist` (required): Artist name
- `title` or `song` (required): Song title
- `format` (optional): "processed" or "raw" (default: "processed")

**Example:**
```bash
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:4000/v1/musical/lyrics?artist=Coldplay&title=Yellow"
```

**Processed Response Includes:**
- Raw lyrics text
- Cleaned lyrics lines array
- Word count and line count
- Verse count estimation
- Chorus detection
- Source attribution

#### 🔍 Health Check
```bash
GET /v1/health
```
No authentication required.

## Environment Setup <a name="environment-setup"></a>

### API Keys Required

1. **NewsAPI.org**
   - Sign up at https://newsapi.org/
   - Add key to `MUSICALZOE_NEWS_API_KEY`

2. **Last.fm**  
   - Create application at https://www.last.fm/api/account/create
   - Add key to `MUSICALZOE_LASTFM_API_KEY`

3. **Lyrics.ovh**
   - No API key required (public API)

### Database Configuration

PostgreSQL is managed via Docker Compose with automatic health checks and migrations.

**Connection Details:**
- Host: localhost:5432
- Database: MUSICALZOE
- User: MUSICALZOE  
- Password: test

## Development <a name="development"></a>

### Makefile Commands

```bash
# Database Operations
make db-up          # Start PostgreSQL container
make db-down        # Stop PostgreSQL container  
make db-migrate     # Run database migrations
make db-reset       # Reset database (drop + recreate + migrate)
make db-connect     # Connect to database shell

# Development
make dev-start      # Start full development environment
make dev-stop       # Stop all services
make run/api        # Run API server with live reload

# Utilities  
make logs           # View container logs
make clean          # Clean up containers and volumes
```

### Development Scripts

```bash
# Full development setup script
./scripts/dev.sh start   # Start everything
./scripts/dev.sh stop    # Stop everything  
./scripts/dev.sh reset   # Reset and restart
./scripts/dev.sh logs    # View logs
```

### Project Structure <a name="project-structure"></a>

```
Musical-Zoe/
├── cmd/api/                    # Application entry point
│   ├── main.go                # Server setup and configuration
│   ├── routes.go              # Route definitions and middleware
│   ├── musicnews.go           # News API service and handlers
│   ├── musictrends.go         # Last.fm trends service and handlers
│   ├── musiclyrics.go         # Lyrics service and handlers
│   ├── http_clients.go        # Generic HTTP client with retries
│   ├── helpers.go             # Utility functions
│   ├── errors.go              # Error handling and responses
│   └── .env                   # Environment variables
├── internal/                   # Internal packages
│   ├── data/                  # Data models and database operations
│   ├── database/              # Database connection and queries
│   ├── logger/                # Structured logging setup
│   ├── mailer/                # Email functionality
│   └── vcs/                   # Version control information
├── scripts/                   # Development and deployment scripts
│   └── dev.sh                 # Development automation script
├── docker-compose.yml         # PostgreSQL container definition
├── makefile                   # Build and development commands
└── README.md                  # This file
```

## Error Handling <a name="error-handling"></a>

### HTTP Status Codes

The API uses standard HTTP status codes with consistent error response format:

```json
{
  "error": "descriptive error message"
}
```

**Status Code Guide:**
- `200` - Success
- `400` - Bad Request (validation errors, timeouts, rate limits)
- `401` - Unauthorized (invalid/missing token)
- `404` - Not Found (lyrics not found, invalid endpoints)
- `423` - Locked (inactive user account)
- `500` - Internal Server Error (server-side issues)

### Error Response Examples

```bash
# Missing Parameter
{"error": "artist parameter is required"}

# Invalid Parameter  
{"error": "invalid period. Valid periods: 7day, 1month, 3month, 6month, 12month, overall"}

# Service Timeout
{"error": "lyrics service is taking too long to respond. Please try again later"}

# Resource Not Found
{"error": "the requested resource could not be found"}

# Authentication Required
{"error": "invalid or missing authentication token"}
```

## Performance <a name="performance"></a>

### Response Times
- **Valid Requests**: 1-3 seconds
- **Timeout Errors**: Maximum 11 seconds (lyrics), 16 seconds (news/trends)
- **Validation Errors**: < 100ms

### Timeout Configuration
- **Lyrics Service**: 5s timeout, 1 retry (~10s max total)
- **News Service**: 8s timeout, 1 retry (~16s max total)  
- **Trends Service**: 8s timeout, 1 retry (~16s max total)

### HTTP Client Features
- Automatic retries with smart backoff (1-second intervals)
- Connection pooling and reuse
- Request/response logging for debugging
- Generic implementation supporting any JSON API

### Database Optimization
- Connection pooling (25 max open, 25 max idle)
- 15-minute idle timeout
- Health checks and automatic reconnection
- Prepared statements via SQLC

## Contributing <a name="contributing"></a>

### Development Workflow

1. **Setup Environment**
   ```bash
   make dev-start
   ```

2. **Make Changes**
   - Follow Go best practices
   - Add tests for new features
   - Update documentation

3. **Test Changes**
   ```bash
   # Test individual endpoints
   curl -H "Authorization: Bearer TOKEN" "http://localhost:4000/v1/musical/lyrics?artist=Test&title=Song"
   
   # Check logs
   make logs
   ```

4. **Database Migrations**
   ```bash
   # Create new migration
   goose -dir internal/sql/schema create migration_name sql
   
   # Apply migrations  
   make db-migrate
   ```

### Code Style
- Use `gofmt` for formatting
- Follow Go naming conventions
- Add comments for exported functions
- Use structured logging with zap
- Handle errors explicitly

### Testing External APIs
- Use the provided Bearer token for testing: `U2JNYCEIPF6FEBOE4AO3R44EE4`
- Test timeout scenarios with non-existent data
- Verify error responses match expected formats

---

**Musical-Zoe Backend - Built with ❤️ and Go** 🎵
