package config

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

type Config struct {
	Addr                 string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ParseConfig() (Config, error) {
	cfg := new(Config)
	flag.StringVar(&cfg.Addr, "a", ":8080", "Service run address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Postgres URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system address")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	return *cfg, err
}
