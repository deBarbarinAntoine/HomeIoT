package main

import (
	"io/fs"
	"net/http"
	
	"HomeIoT/ui"
	
	"github.com/alexedwards/flow"
)

// routes sets up the HTTP routes for the application.
//
// Parameters:
//
//	app - The application instance
//
// Returns:
//
//	http.Handler - A new HTTP handler with all routes set up
func (app *application) routes() http.Handler {
	
	// setting the files to put in the static handler
	staticFs, err := fs.Sub(ui.StaticFiles, "assets")
	if err != nil {
		panic(err)
	}
	
	router := flow.New()
	
	router.NotFound = http.HandlerFunc(app.notFound)                 // error 404 page
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed) // error 405 page
	
	router.Handle("/static/...", http.StripPrefix("/static/", http.FileServerFS(staticFs)), http.MethodGet) // static files
	
	router.Use(app.recoverPanic, app.logRequest, commonHeaders, app.sessionManager.LoadAndSave, noSurf)
	
	// ###########################################################
	// #						COMMON						 	 #
	// ###########################################################
	
	router.HandleFunc("/", app.dashboard, http.MethodGet) // dashboard page
	
	// ###########################################################
	// #					   COMMANDS						 	 #
	// ###########################################################
	
	router.HandleFunc("/:location/:locationID/:device/:deviceID/:information", app.commandDevice, http.MethodPost) // command relay route
	
	// ###########################################################
	// #						AJAX							 #
	// ###########################################################
	
	router.HandleFunc("/:location/:locationID/:device/:deviceID", app.getDeviceInfo, http.MethodGet) // device info route
	
	router.HandleFunc("/locations/:id", app.updateLocation, http.MethodPost)
	
	router.HandleFunc("/ws", app.handleWebSocket, http.MethodGet)
	
	return router
}
