# Friemon-go 🤖

Welcome to **Friemon-go**, a Discord bot inspired by Pokémon, built with Go! This bot allows users to catch, train, and battle characters within your Discord server.

## ✨ Features

* **Character Spawning & Claiming**: Characters randomly appear in channels and can be claimed by users.
* **Character Stats & Info**: View detailed information about your characters, including stats, IVs, and personality.
* **XP & Leveling**: Your selected character gains experience from messages sent in the server and levels up.
* **List & Select**: Manage your collection by listing all your characters and selecting your favorite.
* **Trivia Game**: Engage your community with a fun trivia game.
* **Database & Caching**: Uses PostgreSQL for persistent storage and Redis for caching and temporary data.

## 🛠️ Tech Stack

* **Language**: [Go](https://golang.org/)
* **Discord API Library**: [disgo](https://github.com/disgoorg/disgo)
* **Database**: [PostgreSQL](https://www.postgresql.org/)
* **Cache**: [Redis](https://redis.io/)
* **Containerization**: [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
* **Database Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)

---

## 🚀 Getting Started

Follow these instructions to set up the project for local development and start the bot.

### Prerequisites

* [Git](https://git-scm.com/)
* [Go](https://golang.org/dl/) (version 1.22 or higher)
* [Docker](https://www.docker.com/products/docker-desktop)
* [Docker Compose](https://docs.docker.com/compose/install/)
* [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Installation & Setup

1.  **Clone the Repository**
    ```bash
    git clone [https://github.com/TheOreoTM/friemon-go.git](https://github.com/TheOreoTM/friemon-go.git)
    cd friemon-go
    ```

2.  **Configuration**
    You can configure the application using a `.env` file or by editing `config.toml`. Environment variables take precedence.

    * **Create a `.env` file** by copying the example:
        ```bash
        cp .env.example .env
        ```
    * **Edit the `.env` file** with your settings. You **must** provide a Discord bot token.
        ```dotenv
        # .env
        # Discord Bot Token (Required)
        BOT_TOKEN=your_discord_bot_token_here

        # PostgreSQL Settings
        POSTGRES_USER=friemon
        POSTGRES_PASSWORD=friemonpass
        POSTGRES_DB=friemon

        # Redis Settings (optional, defaults are fine for local)
        # REDIS_PASSWORD=
        ```
    > **Note**: To get a `BOT_TOKEN`, you need to create a new application on the [Discord Developer Portal](https://discord.com/developers/applications). Ensure your bot has the `GUILDS`, `GUILD_MESSAGES`, and `MESSAGE_CONTENT` intents enabled.

3.  **Build and Run with Docker Compose**
    This command will build the Docker images and start all the services (`bot`, `postgres`, `redis`).
    ```bash
    docker-compose up --build -d
    ```

4.  **Run Database Migrations**
    After the `postgres` container is running and healthy, apply the database schema.
    ```bash
    migrate -database "postgres://friemon:friemonpass@localhost:5432/friemon?sslmode=disable" -path ./friemon/db/migrations up
    ```
    *You may need to wait a few seconds for the database to initialize before running this command.*

5.  **Check the Logs**
    You can view the bot's logs to ensure it started correctly.
    ```bash
    docker-compose logs -f bot
    ```

You should see a "friemon ready" message, and the bot will appear as online in your Discord server.

---

## 📂 Project Structure

A brief overview of the key directories in the project:
```
friemon-go/
├── friemon/
│   ├── commands/     # Slash command definitions and handlers
│   ├── components/   # Message component (e.g., button) handlers
│   ├── db/           # Database logic, models, queries, and migrations
│   ├── entities/     # Core data structures (Character, User, etc.)
│   ├── handlers/     # Event handlers (e.g., onMessage)
│   ├── memstore/     # In-memory cache implementation
│   ├── bot.go        # Main bot struct and setup logic
│   └── config.go     # Configuration loading and management
├── .github/          # GitHub-specific files (e.g., workflows)
├── main.go           # Main application entry point
├── Dockerfile        # Instructions to build the bot's Docker image
└── docker-compose.yml# Defines the services for running the application
```

---

## ✍️ Coding Style & Conventions

To maintain code quality and consistency, please adhere to the following guidelines.

* **Formatting**: All Go code should be formatted with `go fmt`. The `Makefile` includes a command for this:
    ```bash
    make fmt
    ```
* **Linting**: While not enforced in the current CI, we recommend using a linter like [golangci-lint](https://golangci-lint.run/) to catch common issues.
* **Naming Conventions**: Follow standard Go naming conventions.
    * `camelCase` for internal variables and functions.
    * `PascalCase` for exported identifiers.
    * Keep names short and descriptive.
* **Commit Messages**: Please use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for your commit messages. This helps in maintaining a clear and automated version history.
    * Example: `feat: add shiny character notifications`
    * Example: `fix: correct IV calculation for low-level characters`
    * Example: `docs: update README with setup instructions`

---

## 🤝 Contributing

We welcome contributions! Please follow these steps to contribute:

1.  **Fork the Repository**: Create your own fork of the project on GitHub.
2.  **Create a Branch**: Make a new branch for your feature or bug fix.
    ```bash
    git checkout -b feat/my-new-feature
    ```
3.  **Make Your Changes**: Write your code and any accompanying tests.
4.  **Test Your Changes**: Ensure your changes don't break existing functionality and that all tests pass.
    ```bash
    make test
    ```
5.  **Format and Tidy**: Run `make fmt` and `make tidy` to format your code and clean up the Go modules.
6.  **Submit a Pull Request**: Push your branch to your fork and open a pull request to the `main` branch of the original repository. Provide a clear description of your changes.

---

## 🐳 Docker Usage

Here are some common Docker Compose commands for managing the application:

* **Start all services**: `docker-compose up -d`
* **Stop all services**: `docker-compose down`
* **View logs**: `docker-compose logs -f <service_name>` (e.g., `bot`, `postgres`)
* **Restart a service**: `docker-compose restart <service_name>`