# Friemon Bot 🚀

A modern Discord bot inspired by Pokémon, built with Go! Catch, train, and battle characters from the Frieren universe within your Discord server.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-2088FF?style=flat-square&logo=github-actions)](https://github.com/features/actions)

## ✨ Features

- 🎯 **Character Spawning & Claiming**: Characters randomly appear in channels and can be claimed by users
- 📊 **Character Stats & Info**: View detailed information about your characters, including stats, IVs, and personality
- ⬆️ **XP & Leveling**: Your selected character gains experience from messages and levels up
- 📋 **Collection Management**: List, select, and organize your character collection
- 🎮 **Interactive Commands**: Slash commands with autocomplete and button interactions
- 🗄️ **Persistent Storage**: PostgreSQL database with Redis caching for optimal performance
- 🔄 **Task Scheduling**: Automated cleanup and maintenance tasks
- 📝 **Structured Logging**: Comprehensive logging with Zap for monitoring and debugging
- 🐳 **Containerized**: Full Docker support with docker-compose for easy deployment
- 🚀 **CI/CD Ready**: GitHub Actions workflow for automated testing and deployment

## 🛠️ Tech Stack

- **Language**: [Go 1.22+](https://golang.org/)
- **Discord Library**: [disgo](https://github.com/disgoorg/disgo)
- **Database**: [PostgreSQL 16](https://www.postgresql.org/)
- **Cache**: [Redis 7](https://redis.io/)
- **Logging**: [Zap](https://github.com/uber-go/zap)
- **Task Queue**: [Asynq](https://github.com/hibiken/asynq)
- **Database Queries**: [sqlc](https://sqlc.dev/)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Containerization**: [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)

## 🚀 Quick Start

### Prerequisites

- [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/)
- A Discord Application with Bot Token ([Discord Developer Portal](https://discord.com/developers/applications))

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/TheOreoTM/friemon-go.git
cd friemon-go

# Copy environment template
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` file with your settings:

```bash
# Required - Get from Discord Developer Portal
BOT_TOKEN=your_discord_bot_token_here

# Required - Set secure passwords
DB_PASSWORD=your_secure_database_password
POSTGRES_PASSWORD=your_secure_database_password
PGADMIN_PASSWORD=your_secure_pgadmin_password

# Optional - Development settings
DEV_MODE=false
LOG_LEVEL=info
```

### 3. Run with Docker

```bash
# Start all services
docker-compose up -d --build

# Check if everything is running
docker-compose ps

# View logs
docker-compose logs -f bot
```

### 4. Setup Database

```bash
# Wait for PostgreSQL to be ready, then run migrations
sleep 10
migrate -database "postgres://friemon:your_password@localhost:5433/friemon?sslmode=disable" \
        -path ./internal/infrastructure/db/migrations up
```

### 5. Invite Bot to Server

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Select your application → OAuth2 → URL Generator
3. Select scopes: `bot`, `applications.commands`
4. Select permissions: `Send Messages`, `Use Slash Commands`, `Attach Files`
5. Copy and visit the generated URL to invite your bot

🎉 **Your bot is now ready!** Try `/character` in your Discord server.

## 🔧 Development Setup

### Local Development (without Docker)

```bash
# Install Go dependencies
go mod download

# Install required tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate database code
sqlc generate

# Set up local PostgreSQL and Redis
# Update .env with local database settings:
# DB_HOST=localhost
# DB_PORT=5433
# REDIS_ADDR=localhost:6379

# Run the bot
go run ./cmd/friemon/main.go
```

### Development with Docker Services

```bash
# Run only database services
docker-compose up postgres redis -d

# Run bot locally
go run ./cmd/friemon/main.go
```

### Useful Commands

```bash
# Run tests
go test -v ./...

# Run linting
golangci-lint run ./...

# Format code
go fmt ./...

# Update dependencies
go mod tidy

# Generate database code
sqlc generate

# View database logs
docker-compose logs -f postgres

# Access database directly
docker-compose exec postgres psql -U friemon -d friemon
```

## 📋 Available Commands

| Command      | Description                  | Usage                 |
| ------------ | ---------------------------- | --------------------- |
| `/character` | Generate a random character  | `/character`          |
| `/info`      | View character information   | `/info [character]`   |
| `/list`      | List your characters         | `/list [page]`        |
| `/select`    | Select your active character | `/select <character>` |
| `/version`   | Show bot version             | `/version`            |

## 🎮 How to Play

1. **Wait for Spawns**: Characters automatically spawn in active channels
2. **Claim Characters**: Click the "Claim" button when a character appears
3. **Check Your Collection**: Use `/list` to see all your characters
4. **Select Active Character**: Use `/select` to choose your active character
5. **Gain XP**: Your selected character gains XP from your messages
6. **View Stats**: Use `/info` to see detailed character information

## 🔧 Configuration

All configuration is done via environment variables. See `.env.example` for all available options:

### Core Settings
- `BOT_TOKEN` - Discord bot token (required)
- `DEV_MODE` - Enable development mode (default: false)
- `SYNC_COMMANDS` - Sync slash commands on startup (default: true)

### Database
- `DB_HOST` - Database host (default: postgres)
- `DB_PASSWORD` - Database password (required)
- `DB_NAME` - Database name (default: friemon)

### Logging
- `LOG_LEVEL` - Log level: debug/info/warn/error (default: info)
- `LOG_FORMAT` - Log format: console/json (default: console)
- `LOG_OUTPUT_PATH` - Log output: stdout or file path

### Advanced
- `DEV_GUILDS` - Comma-separated guild IDs for command testing
- `REDIS_ADDR` - Redis server address
- `TZ` - Timezone (default: UTC)

## 🚀 Deployment

### Production Deployment

1. **Set up your VPS** with Docker and Docker Compose
2. **Configure GitHub Secrets**:
   - `DOCKER_USERNAME` - Docker Hub username
   - `DOCKER_PASSWORD` - Docker Hub password
   - `VPS_SSH_HOST` - Your server IP
   - `VPS_SSH_USER` - SSH username
   - `VPS_SSH_KEY` - SSH private key
   - `BOT_TOKEN` - Discord bot token
   - `POSTGRES_PASSWORD` - Database password
   - `PGADMIN_PASSWORD` - PgAdmin password

3. **Push to main branch** - CI/CD will automatically deploy

### Manual Deployment

```bash
# On your server
git clone https://github.com/TheOreoTM/friemon-go.git
cd friemon-go

# Configure environment
cp .env.example .env
# Edit .env with your values

# Deploy
./deploy.sh
```

## 📊 Monitoring

### Database Management
- **PgAdmin**: Access at `http://your-server:5050`
  - Email: `admin@example.com`
  - Password: Your `PGADMIN_PASSWORD`

### Redis Management  
- **Redis Commander**: Access at `http://your-server:8081`

### Logs
```bash
# View real-time logs
docker-compose logs -f bot

# View specific service logs
docker-compose logs postgres
docker-compose logs redis

# Search logs for user actions
docker-compose logs bot | grep "discord_user_id.*123456789"

# Monitor errors
docker-compose logs bot | grep '"level":"ERROR"'
```

## 📁 Project Structure

```
friemon-go/
├── cmd/friemon/           # Application entry point
├── internal/
│   ├── application/       # Application layer
│   │   ├── bot/          # Bot configuration and setup
│   │   ├── commands/     # Slash commands
│   │   ├── components/   # Button/interaction handlers
│   │   └── handlers/     # Event handlers
│   ├── core/entities/    # Domain entities and business logic
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── db/          # Database queries, models, migrations
│   │   ├── memstore/    # Caching implementation
│   │   └── scheduler/   # Task scheduling
│   ├── pkg/logger/      # Logging utilities
│   └── types/           # Shared types
├── assets/               # Character images and resources
├── .github/workflows/    # CI/CD workflows
├── docker-compose.yml    # Docker services configuration
├── Dockerfile           # Bot container definition
└── .env.example         # Environment configuration template
```

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork the Repository**
2. **Create a Feature Branch**
   ```bash
   git checkout -b feat/amazing-feature
   ```
3. **Make Your Changes**
   - Follow Go conventions
   - Add tests for new features
   - Update documentation
4. **Test Your Changes**
   ```bash
   go test -v ./...
   go fmt ./...
   golangci-lint run ./...
   ```
5. **Commit with Conventional Commits**
   ```bash
   git commit -m "feat: add character trading system"
   ```
6. **Push and Create Pull Request**

### Coding Standards
- Use `go fmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Write tests for new features
- Use structured logging with appropriate context
- Document public APIs

## 🐛 Troubleshooting

### Common Issues

**Bot not responding to commands:**
```bash
# Check if bot is running
docker-compose ps

# Check bot logs
docker-compose logs bot

# Verify token and permissions
```

**Database connection errors:**
```bash
# Check PostgreSQL status
docker-compose logs postgres

# Verify database credentials in .env
# Ensure migrations are applied
```

**Characters not spawning:**
```bash
# Check Redis connection
docker-compose logs redis

# Verify cache settings
# Check interaction count in logs
```

**Build failures:**
```bash
# Clean Docker cache
docker system prune -a

# Rebuild from scratch
docker-compose build --no-cache
```

### Getting Help

- 📚 Check the [Issues](https://github.com/TheOreoTM/friemon-go/issues) page
- 💬 Join our [Discord Server](https://discord.gg/your-server) for support
- 📧 Contact the maintainers

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Frieren: Beyond Journey's End](https://frieren-anime.jp/) for character inspiration
- [disgo](https://github.com/disgoorg/disgo) for the excellent Discord library
- [sqlc](https://sqlc.dev/) for type-safe SQL generation
- All contributors who help improve this project

---

<div align="center">

**⭐ Star this repo if you find it useful!**

[Report Bug](https://github.com/TheOreoTM/friemon-go/issues) · [Request Feature](https://github.com/TheOreoTM/friemon-go/issues) · [Documentation](https://github.com/TheOreoTM/friemon-go/wiki)

</div>