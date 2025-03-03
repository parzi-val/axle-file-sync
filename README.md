# Axle

Axle is a real-time file synchronization system that tracks changes in a directory and efficiently computes differences using an XOR-based mechanism.

## Installation

Make sure you have Go installed. Then, clone the repository and run:

```
git clone https://github.com/parzi-val/axle-file-sync
cd axle
go mod tidy
```

## Usage

1. Build the Project

```
go build -o axle.exe
```

2. Run the Application

```
./axle.exe
```
