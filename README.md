<div align="center">

# 🎮 Next Turn

**Play classic board games directly in Telegram with your friends!**

[![Telegram Bot](https://img.shields.io/badge/Telegram-@NextTurnBot-blue?style=for-the-badge&logo=telegram)](https://t.me/NextTurnBot)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Database](https://img.shields.io/badge/Database-Neon_Postgres-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://neon.tech/)

[**Try it Now**](https://t.me/NextTurnBot) • [Report Bug](https://github.com/The0xArchitect/next-turn/issues) • [Request Feature](https://github.com/The0xArchitect/next-turn/issues)

</div>

---

## ✨ Features

🎯 **6 Classic Games** - Enjoy a variety of board games with different difficulty levels  
💬 **Inline Mode** - Start games directly in any chat without leaving the conversation  
👥 **Multiplayer** - Challenge friends and track game states in real-time  
🔄 **Turn-Based** - Asynchronous gameplay that fits your schedule  
💾 **Persistent State** - All games are saved to the database automatically  
🚀 **Fast & Reliable** - Built with Go for high performance and low latency  

---

## 🎲 Available Games

<table>
<tr>
<td width="50%">

### ⭕ **Tic Tac Toe**
Classic 3×3 grid game. Get three in a row to win!

**Board:** 3×3 | **Win Condition:** 3 in a row

</td>
<td width="50%">

### 🎯 **Four XO**
Extended tic-tac-toe on a larger board.

**Board:** 5×5 | **Win Condition:** 4 in a row

</td>
</tr>
<tr>
<td width="50%">

### 🐘 **Elephant XO**
Even bigger board for longer strategic games.

**Board:** 7×7 | **Win Condition:** 5 in a row

</td>
<td width="50%">

### 🔴🔵 **Connect 4**
Drop discs and connect four vertically, horizontally, or diagonally!

**Board:** 7×6 | **Win Condition:** 4 in a row

</td>
</tr>
<tr>
<td width="50%">

### ⚫⚪ **Checkers**
Classic checkers with traditional rules.

**Board:** 8×8 | **Win Condition:** Capture all opponent pieces

</td>
<td width="50%">

### 🎱 **Pool Checkers**
Checkers variant with flying kings for advanced players.

**Board:** 8×8 | **Win Condition:** Capture all opponent pieces

</td>
</tr>
</table>

---

## 🚀 Quick Start

### For Players

1. **Open Telegram** and search for [@NextTurnBot](https://t.me/NextTurnBot)
2. **Start the bot** with `/start` command
3. **Choose a game** from the menu
4. **Share with a friend** and start playing!

### Using Inline Mode

Type `@NextTurnBot` in any chat followed by a game code:

```
@NextTurnBot ttt    # Tic Tac Toe
@NextTurnBot 4xo    # Four XO
@NextTurnBot exo    # Elephant XO
@NextTurnBot c4     # Connect 4
@NextTurnBot ck     # Checkers
@NextTurnBot pc     # Pool Checkers
```

---

## 🛠️ Tech Stack

<div align="center">

| Component | Technology |
|-----------|-----------|
| **Language** | ![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white) Go 1.25.4 |
| **Database** | ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=flat&logo=postgresql&logoColor=white) Neon Postgres |
| **Bot API** | ![Telegram](https://img.shields.io/badge/Telegram_Bot_API-26A5E4?style=flat&logo=telegram&logoColor=white) go-telegram-bot-api/v5 |
| **Database Driver** | ![pgx](https://img.shields.io/badge/pgx-v5-blue?style=flat) jackc/pgx/v5 |
| **Deployment** | ![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white) Containerized |

</div>

---

## 📦 Installation & Development

### Prerequisites

- Go 1.25.4 or higher
- PostgreSQL database (or Neon account)
- Telegram Bot Token ([Get one from @BotFather](https://t.me/BotFather))

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/The0xArchitect/next-turn.git
   cd next-turn
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   
   Create a `.env` file in the project root:
   ```env
   BOT_TOKEN=your_telegram_bot_token
   DATABASE_URL=postgres://user:password@host/database
   PORT=8080
   ```

4. **Run the bot**
   ```bash
   go run cmd/main.go
   ```

### Docker Deployment

Build and run with Docker:

```bash
# Build the image
docker build -t next-turn .

# Run the container
docker run -d \
  -p 8080:8080 \
  -e BOT_TOKEN=your_token \
  -e DATABASE_URL=your_database_url \
  next-turn
```

---

## 🏗️ Project Structure

```
next-turn/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── core/                # Game engine registry
│   ├── db/                  # Database connection & queries
│   ├── games/               # Game implementations
│   │   ├── tictactoe/       # Tic Tac Toe engine
│   │   ├── fourxo/          # Four XO engine
│   │   ├── elephantxo/      # Elephant XO engine
│   │   ├── connect4/        # Connect 4 engine
│   │   ├── checkers/        # Checkers engine
│   │   └── poolcheckers/    # Pool Checkers engine
│   ├── handlers/            # Telegram update handlers
│   └── models/              # Data models
├── Dockerfile               # Container configuration
├── go.mod                   # Go module definition
└── README.md                # This file
```

---

## 🎯 How It Works

```mermaid
graph LR
    A[User] -->|Inline Query| B[Telegram]
    B -->|Webhook| C[Next Turn Bot]
    C -->|Game Logic| D[Game Engine]
    C -->|Save State| E[(Neon DB)]
    D -->|Board State| C
    E -->|Load State| C
    C -->|Response| B
    B -->|Game Board| A
```

1. **User initiates** a game via inline query
2. **Bot receives** the request through webhook
3. **Game engine** processes the move and validates rules
4. **Database** persists the game state
5. **Bot sends** updated board back to Telegram
6. **Players continue** taking turns asynchronously

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/AmazingFeature`)
3. **Commit** your changes (`git commit -m 'Add some AmazingFeature'`)
4. **Push** to the branch (`git push origin feature/AmazingFeature`)
5. **Open** a Pull Request

### Ideas for Contributions

- 🎮 Add new games (Chess, Othello, etc.)
- 🌐 Internationalization support
- 📊 Player statistics and leaderboards
- 🎨 Custom board themes
- 🤖 AI opponents
- 📱 Better mobile UI/UX

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - Telegram Bot API wrapper for Go
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit for Go
- [Neon](https://neon.tech/) - Serverless Postgres platform
- All contributors and players who make this project possible!

---

## 📧 Contact

**Developer:** [@The0xArchitect](https://github.com/The0xArchitect)

**Bot:** [@NextTurnBot](https://t.me/NextTurnBot)

**Issues:** [GitHub Issues](https://github.com/The0xArchitect/next-turn/issues)

---

<div align="center">

**Made with ❤️ for the Telegram bots**

⭐ Star this repo if you enjoy playing!

[🎮 Start Playing](https://t.me/NextTurnBot)

</div>
