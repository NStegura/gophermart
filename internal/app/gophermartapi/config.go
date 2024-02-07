package gophermartapi

import (
	"flag"
	"os"
)

const (
	defaultServerAddr  = ":8080"
	defaultDatabaseDSN = ""
	defaultLogLevel    = "debug"
	defaultAccrualAddr = "accrual-api:8082"
	defaultSecretKey   = "gljfsj;312sf;kdhrf;" // only for tests
)

type Config struct {
	RunAddress  string
	DatabaseDSN string
	SecretKey   string
	LogLevel    string
	AccrualAddr string
	TracerURL   string
}

func NewConfig() *Config {
	return &Config{
		RunAddress:  defaultServerAddr,
		DatabaseDSN: defaultDatabaseDSN,
		SecretKey:   defaultSecretKey,
		AccrualAddr: defaultAccrualAddr,
		LogLevel:    defaultLogLevel,
	}
}

func (c *Config) ParseFlags() (err error) {
	var (
		runAddress  = defaultServerAddr
		DatabaseDSN = defaultDatabaseDSN
		accrualAddr = defaultAccrualAddr
		secretKey   = defaultSecretKey
	)

	if envRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		runAddress = envRunAddr
	}

	if dbDsn, ok := os.LookupEnv("DATABASE_URI"); ok {
		DatabaseDSN = dbDsn
	}

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		c.LogLevel = envLogLevel
	}

	if acAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		accrualAddr = acAddr
	}

	if sk, ok := os.LookupEnv("SECRET_KEY"); ok {
		secretKey = sk
	}

	if tu, ok := os.LookupEnv("TRACER_URL"); ok {
		c.TracerURL = tu
	}

	flag.StringVar(&c.RunAddress, "a", runAddress, "address and port to run server")
	flag.StringVar(&c.DatabaseDSN, "d", DatabaseDSN, "database dsn")
	flag.StringVar(&c.AccrualAddr, "r", accrualAddr, "address and port accrual cli")
	flag.StringVar(&c.SecretKey, "s", secretKey, "secret key to hash auth")
	flag.Parse()
	return
}
