package main

import (
    "log"
    "asa/shell/internal/shell"
)

func main() {
    sh := shell.New()
    if err := sh.Start(); err != nil {
        log.Fatalf("Shell error: %v", err)
    }
}