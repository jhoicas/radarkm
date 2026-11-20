package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"radarkm/api/internal/db"
	"radarkm/api/internal/migrate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate [up|down|reset]")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, db.DatabaseURL())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer pool.Close()

	switch os.Args[1] {
	case "up":
		err = migrate.Up(ctx, pool)
	case "down":
		err = migrate.Down(ctx, pool)
	case "reset":
		err = migrate.Reset(ctx, pool)
	default:
		fmt.Fprintln(os.Stderr, "usage: migrate [up|down|reset]")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("ok:", os.Args[1])
}
