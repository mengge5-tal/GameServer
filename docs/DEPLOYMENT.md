# GameServer Deployment Guide

## Overview
This document provides instructions for deploying the GameServer with improved security and monitoring.

## Prerequisites
- Go 1.24 or later
- MySQL database
- Environment variables configured

## Security Improvements Implemented

### 1. Database Security
- ✅ **Hardcoded credentials removed**: Database credentials now use environment variables
- ✅ **Connection pooling**: Optimized database connection management
- ✅ **Configuration validation**: Proper validation of database parameters

### 2. WebSocket Security
- ✅ **CORS protection**: Origin validation for WebSocket connections
- ✅ **Configurable origins**: Allowed origins managed via environment variables

### 3. Password Security
- ✅ **bcrypt hashing**: Already implemented in the original code
- ✅ **Configurable cost**: bcrypt cost can be adjusted via environment variables

### 4. Field Naming
- ✅ **Database schema**: Fixed "bloodenergy" → "blood_energy" spelling

## Environment Configuration

### 1. Copy Environment Template
```bash
cp .env.example .env
```

### 2. Configure Environment Variables
Edit the `.env` file with your specific values:

```bash
# Database Configuration
DB_HOST=your-database-host.com
DB_PORT=3306
DB_NAME=your-database-name
DB_USER=your-username
DB_PASSWORD=your-secure-password

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# WebSocket Configuration
WS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Security Configuration
BCRYPT_COST=12

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

## Database Migration

### 1. Run Schema Migration
Execute the migration script to update the database schema:

```sql
-- Run migrate_bloodenergy.sql
mysql -h your-host -u your-user -p your-database < migrate_bloodenergy.sql
```

### 2. Verify Schema Changes
Check that the `blood_energy` field has been renamed correctly:

```sql
DESCRIBE playerinfo;
```

## Deployment Steps

### 1. Build Application
```bash
go mod tidy
go build -o gameserver .
```

### 2. Set Environment Variables
```bash
export DB_HOST=your-database-host
export DB_PASSWORD=your-secure-password
# ... other variables
```

### 3. Run Application
```bash
./gameserver
```

## Monitoring and Health Checks

### Health Check Endpoint
- **URL**: `GET /health`
- **Response**: JSON with server status and timestamp

### Metrics Endpoint
- **URL**: `GET /metrics`
- **Response**: JSON with connection counts, message statistics, and performance metrics

### Example Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "version": "1.0.0"
}
```

### Example Metrics Check
```bash
curl http://localhost:8080/metrics
```

Expected response:
```json
{
  "connection_count": 5,
  "total_connections": 127,
  "messages_processed": 1543,
  "error_count": 12,
  "database_queries": 856,
  "request_durations": {
    "login": [45, 32, 67, 23],
    "getPlayerInfo": [12, 15, 18, 11]
  },
  "last_updated": "2023-12-31T23:59:59Z"
}
```

## Architecture Improvements

### 1. Modular Structure
The codebase has been refactored into organized packages:
- `config/`: Configuration management
- `models/`: Data structures and message definitions
- `logger/`: Structured logging system
- `metrics/`: Performance monitoring and metrics collection

### 2. Configuration Management
- Centralized configuration with validation
- Environment variable support with fallbacks
- Type-safe configuration loading

### 3. Logging System
- Structured JSON logging
- Configurable log levels (debug, info, warn, error)
- Request/response tracking capabilities

### 4. Monitoring System
- Real-time connection tracking
- Message processing statistics
- Database query monitoring
- Request duration tracking

## Security Best Practices

### 1. Environment Variables
- Never commit `.env` files to version control
- Use strong, unique passwords
- Rotate credentials regularly

### 2. CORS Configuration
- Only allow trusted origins in `WS_ALLOWED_ORIGINS`
- Review and update allowed origins regularly
- Use HTTPS in production

### 3. Database Security
- Use strong database passwords
- Limit database user permissions
- Enable database SSL/TLS in production
- Regular database backups

### 4. Network Security
- Use firewalls to restrict access
- Consider using a reverse proxy (nginx, Apache)
- Enable rate limiting if needed

## Production Considerations

### 1. Process Management
Consider using a process manager like systemd or PM2:

```bash
# systemd service example
sudo systemctl enable gameserver
sudo systemctl start gameserver
```

### 2. Reverse Proxy
Use nginx or Apache as a reverse proxy for additional security:

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### 3. SSL/TLS
Always use HTTPS in production with proper SSL certificates.

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check environment variables
   - Verify database server is running
   - Confirm credentials and network connectivity

2. **WebSocket Connection Rejected**
   - Check `WS_ALLOWED_ORIGINS` configuration
   - Verify client origin matches allowed origins

3. **Permission Denied**
   - Check file permissions on binary
   - Verify port is available (not used by another process)

### Logs
Check application logs for detailed error information:
```bash
tail -f gameserver.log
```

## Support
For issues and questions, refer to the API documentation and test cases provided in the repository.