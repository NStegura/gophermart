package gophermartapi

import (
	"flag"
	"os"
)

const (
	defaultServerAddr  = ":8080"
	defaultDatabaseURI = ""
	defaultLogLevel    = "debug"
	defaultAccrualAddr = "accrual-api:8082"
	defaultSecretKey   = "gljfsj;312sf;kdhrf;" // ?????? very bad
)

type Config struct {
	RunAddress  string
	DatabaseURI string
	SecretKey   string
	LogLevel    string
	AccrualAddr string
}

func NewConfig() *Config {
	return &Config{
		RunAddress:  defaultServerAddr,
		DatabaseURI: defaultDatabaseURI,
		SecretKey:   defaultSecretKey,
		AccrualAddr: defaultAccrualAddr,
		LogLevel:    defaultLogLevel,
	}
}

func (c *Config) ParseFlags() (err error) {
	var (
		runAddress  = defaultServerAddr
		databaseURI = defaultDatabaseURI
		accrualAddr = defaultAccrualAddr
		secretKey   = defaultSecretKey
	)

	if envRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		runAddress = envRunAddr
	}

	if dbDsn, ok := os.LookupEnv("DATABASE_URI"); ok {
		databaseURI = dbDsn
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

	flag.StringVar(&c.RunAddress, "a", runAddress, "address and port to run server")
	flag.StringVar(&c.DatabaseURI, "d", databaseURI, "database dsn")
	flag.StringVar(&c.AccrualAddr, "r", accrualAddr, "address and port accrual cli")
	flag.StringVar(&c.SecretKey, "s", secretKey, "secret key to hash auth")
	flag.Parse()
	return
}
