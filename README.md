# vk-training

A command-line training/workout tracker written in Go. Data is stored locally and the app runs as an interactive shell in your terminal.

## Tech Stack

- [Go](https://go.dev/) 1.25
- [peterh/liner](https://github.com/peterh/liner) — interactive line editing / readline-style shell
- [mattn/go-runewidth](https://github.com/mattn/go-runewidth) — terminal width handling

## Project Structure

```
├── DATABASES/TRAINING/   # Local data storage for training records
├── pkg/
│   ├── cmd/               # Command handling / interactive shell logic
│   ├── config/             # Configuration (e.g. local storage file path)
│   ├── db/                 # Training data model and load/save logic
│   └── util/                # Utility helpers (e.g. local storage init)
├── main.go                # Entry point
├── go.mod
└── go.sum
```

## How It Works

On startup, `vk-training`:

1. Initializes local storage (`util.InitLocalStorage`).
2. Loads existing training data from a local file (`db.Training.LoadFromFile`).
3. Starts an interactive command shell (`cmd.Start`) where you can enter commands to manage your training data.

## Getting Started

### Prerequisites

- Go 1.25 or later

### Installation

```bash
git clone https://github.com/VkHyperNova/vk-training.git
cd vk-training
go build -o vk-training .
```

### Usage

Run the built binary:

```bash
./vk-training
```

Or run directly with Go:

```bash
go run main.go
```

This launches an interactive shell where you can enter commands to log and manage your training data, which is persisted locally.

## License

No license specified — all rights reserved by the author unless stated otherwise.
