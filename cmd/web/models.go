package main

import (
	"html/template"
	"log/slog"
	"sync"
	
	"HomeIoT/internal/data"
	"HomeIoT/internal/mailer"
	"HomeIoT/internal/validator"
	
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

// config represents the configuration variables for the application.
type config struct {
	websocketURL string
	port         int64
	env          string
	broker       struct {
		host                string
		port                int64
		subscriptionChannel string
		qos                 byte
	}
	db struct {
		dsn string
	}
	smtp struct {
		host     string
		port     int64
		username string
		password string
		sender   string
	}
}

// application represents the application configuration.
type application struct {
	logger         *slog.Logger
	mailer         mailer.Mailer
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	Models         data.Models
	config         *config
	wg             *sync.WaitGroup
}

// templateData represents the data structure used in templates.
type templateData struct {
	WebsocketURL string
	Title        string
	CurrentYear  int
	Form         any
	Flash        string
	Nonce        string
	CSRFToken    string
	ResetToken   string
	
	Devices   []*data.Device
	Locations []*data.Location
	Device    *data.Device
	Location  *data.Location
	
	Error struct {
		Title   string
		Message string
	}
	FieldErrors    map[string]string
	NonFieldErrors []string
}

// envelope is a data type for JSON responses.
type envelope map[string]any

// userLoginForm represents the form used for user login.
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}
