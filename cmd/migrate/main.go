package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	dbDriver   = "pgx"
	configFile = "config/pg.env"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "migrations", "directory with migration files")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbString := createDBString()

	err := flags.Parse(os.Args[1:])
	if err != nil {
		panic(fmt.Errorf("failed parsing flags: %w", err))
	}

	args := flags.Args()

	goose_command := args[1]

	db, err := goose.OpenDBWithDriver(dbDriver, dbString)
	if err != nil {
		panic(fmt.Errorf("failed open db: %w", err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			panic(fmt.Errorf("failed closing db: %w", err))
		}
	}()

	if err := goose.RunContext(ctx, goose_command, db, *dir, args[1:]...); err != nil {
		panic(fmt.Errorf("failed migrate "+goose_command+": %w", err))
	}
}

func createDBString() string {
	type DBConfig struct {
		DB       string `env:"POSTGRES_DB"`
		User     string `env:"POSTGRES_USER"`
		Password string `env:"POSTGRES_PASSWORD"`
	}

	var cfg DBConfig

	err := cleanenv.ReadConfig(configFile, &cfg)
	if err != nil {
		panic(fmt.Errorf("failed reading config: %w", err))
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		"localhost",
		cfg.User,
		cfg.Password,
		cfg.DB,
	)
}
