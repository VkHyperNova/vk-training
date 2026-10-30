# vk-training

A small command-line training tracker written in Go. Log what you did, see running totals per exercise at a glance, and keep a backup on a USB drive automatically.

```
VK TRAINING v1

	pushups  120  ████████████████████  +30
	run       15  ███
	squats     8  █

< log, undo >
```

The last logged entry is highlighted in yellow, so you can see what you just added.

## Features

- Log any exercise with a numeric value (whole numbers or decimals)
- Totals per exercise, sorted highest first, with bars scaled to the largest total
- Undo the last entry
- Data stored as plain JSON
- Atomic writes, so a crash mid-save can't leave a half-written file
- Best-effort backup to an external drive (a failed backup never blocks logging)

## Requirements

- Go 1.21 or newer
- Linux (screen clearing is supported on Linux and Windows; the backup drive mount is aimed at Linux)

## Installation

```sh
git clone https://github.com/VkHyperNova/vk-training.git
cd vk-training
go build -o training .
```

Or run it directly without building:

```sh
go run . <command> [arguments]
```

## Usage

```
training <command> [arguments]
```

| Command                      | Description                                       |
| ---------------------------- | ------------------------------------------------- |
| *(none)*                     | Show the summary                                  |
| `log`, `add`, `a <name> <value>` | Log an item (value can be a decimal, e.g. `12.5`) |
| `undo`                       | Remove the last logged item                       |
| `help`, `-h`, `--help`       | Show help                                         |

### Examples

```sh
training log pushups 30     # log 30 pushups
training add run 5.25       # log a 5.25 km run
training a squats 8         # short alias
training undo               # oops, remove the last entry
training                    # just show the summary
```

Names are lowercased, so `Pushups` and `pushups` are counted together. Exercises whose total adds up to zero are hidden from the summary.

## Data

Data is saved to `DATABASES/TRAINING/training.json`, **relative to the directory you run the program from**. Run it from the same directory each time (or adjust the paths in `main.go`).

The file looks like this:

```json
{
  "training": [
    {
      "name": "pushups",
      "value": 30,
      "date": "2026-10-30T08:00:00+02:00"
    },
    {
      "name": "run",
      "value": 5.25,
      "date": "2026-10-30T09:00:00+02:00"
    }
  ]
}
```

### Backup

After every save the program tries to mount the drive labelled `VK DATA` (using [vkmount](https://github.com/VkHyperNova/vkmount)) and copy the data file to `/media/veikko/VK DATA/DATABASES/TRAINING/training.json`.

The local save is the one that has to succeed. If the drive isn't plugged in or can't be mounted, you'll see a yellow `Backup failed: ...` notice and your data is still safely saved locally.

To use this on your own machine, change `driveLabel` and `driveMountPoint` in `main.go`.

## Project structure

```
├── main.go           # Entry point, constants, command dispatch
├── training.go       # Item/Training types, Log and Undo
├── storage.go        # Load, Save, backup, atomic file writes
├── report.go         # Summary rendering (totals, bars, colors)
├── storage_test.go   # Tests for loading, saving and backup
└── training_test.go  # Tests for Log and Undo
```

## Tests

```sh
go test ./...
```

The tests use a temporary directory and a stubbed drive mount, so they never touch your real data or backup drive.

## License

No license specified. All rights reserved by the author unless stated otherwise.