package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"
	
	"HomeIoT/internal/data"
	"HomeIoT/internal/mailer"
	
	"github.com/alexedwards/scs/gormstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// main is the entry point of the application.
func main() {
	
	// setting the configuration variables
	var cfg config
	var err error
	
	// Web Server config
	cfg.port, err = strconv.ParseInt(os.Getenv("PORT"), 10, 64)
	if err != nil {
		fmt.Println("port is not a number")
		os.Exit(1)
	}
	cfg.env = os.Getenv("ENVIRONMENT")
	
	// WebSocket config
	cfg.websocketURL = os.Getenv("WEBSOCKET_URL")
	
	// Database config
	cfg.db.dsn = os.Getenv("DATABASE_DSN")
	
	// MQTT config
	cfg.broker.host = os.Getenv("BROKER_HOST")
	cfg.broker.subscriptionChannel = os.Getenv("BROKER_SUBSCRIPTION_CHANNEL")
	cfg.broker.port, err = strconv.ParseInt(os.Getenv("BROKER_PORT"), 10, 64)
	if err != nil {
		fmt.Println("MQTT Broker port is not a number")
		os.Exit(1)
	}
	intQos, err := strconv.ParseInt(os.Getenv("BROKER_QOS"), 10, 8)
	if err != nil {
		fmt.Println("MQTT Broker QoS is not a number")
		os.Exit(1)
	}
	cfg.broker.qos = byte(intQos)
	
	// SMTP config
	cfg.smtp.sender = os.Getenv("SMTP_SENDER")
	cfg.smtp.username = os.Getenv("SMTP_USERNAME")
	cfg.smtp.password = os.Getenv("SMTP_PASSWORD")
	cfg.smtp.host = os.Getenv("SMTP_HOST")
	cfg.smtp.port, err = strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 64)
	if err != nil {
		fmt.Println("SMTP port is not a number")
		os.Exit(1)
	}
	
	// setting the logging level according to the environment
	var opts *slog.HandlerOptions
	
	if cfg.env == "development" {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	} else {
		opts = &slog.HandlerOptions{Level: slog.LevelInfo}
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	
	// checking the Websocket URL info
	if cfg.websocketURL == "" {
		fmt.Println("WebSocket URL required")
		os.Exit(1)
	}
	
	// checking the SMTP info
	if cfg.smtp.username == "" || cfg.smtp.password == "" || cfg.smtp.host == "" {
		fmt.Println("SMTP credentials are required")
		os.Exit(1)
	}
	
	// checking the MQTT Broker info
	if cfg.broker.host == "" || cfg.broker.port == 0 || cfg.broker.subscriptionChannel == "" || cfg.broker.qos > 2 {
		fmt.Println("Valid MQTT Broker configuration is required")
		os.Exit(1)
	}
	
	// checking the dsn info
	if cfg.db.dsn == "" {
		logger.Error("DSN is required")
		os.Exit(1)
	}
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.db.dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database")
		os.Exit(1)
	}
	
	// Execute migrations if necessary
	err = db.AutoMigrate(&data.Data{}, &data.Module{})
	if err != nil {
		logger.Error(fmt.Sprintf("failed to auto migrate: %s", err.Error()))
		os.Exit(1)
	}
	
	// caching the templates
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	
	// initializing the application components
	formDecoder := form.NewDecoder()
	
	sessionManager := scs.New()
	sessionManager.Store, err = gormstore.New(db)
	if err != nil {
		logger.Error("Failed to set the session store")
		os.Exit(1)
	}
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Secure = true
	
	// connecting to the broker
	broker := data.NewBroker(cfg.broker.host, cfg.broker.port, cfg.broker.qos)
	
	app := &application{
		logger:         logger,
		mailer:         mailer.New(cfg.smtp.host, int(cfg.smtp.port), cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		sessionManager: sessionManager,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		config:         &cfg,
		Models:         data.NewModels(db, broker, logger),
		wg:             new(sync.WaitGroup),
	}
	
	// Subscription to MQTT channels
	app.Models.Data.Sub(app.config.broker.subscriptionChannel)
	
	// Running the server
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	
	os.Exit(1)
}
