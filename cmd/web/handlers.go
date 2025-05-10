package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
)

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {

	// retrieving basic template data
	tmplData := app.newTemplateData(r)
	tmplData.Title = "Home IoT - Not Found"

	// rendering the template
	app.render(w, r, http.StatusNotFound, "error.tmpl", tmplData)
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {

	// retrieving basic template data
	tmplData := app.newTemplateData(r)
	tmplData.Title = "Home IoT - Oooops"

	// setting the error title and message
	tmplData.Error.Title = "Error 405"
	tmplData.Error.Message = "Something went wrong!"

	// rendering the template
	app.render(w, r, http.StatusMethodNotAllowed, "error.tmpl", tmplData)
}

func (app *application) index(w http.ResponseWriter, r *http.Request) {

	// retrieving basic template data
	tmplData := app.newTemplateData(r)
	tmplData.Title = "Home IoT - Home"

	// rendering the template
	app.render(w, r, http.StatusOK, "home.tmpl", tmplData)
}

// Dashboard handler - renders the IoT dashboard page
func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	devices, err := app.Models.Device.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	tmplData := app.newTemplateData(r)
	tmplData.Devices = devices

	// DEBUG
	for device := range slices.Values(devices) {
		app.logger.Debug(fmt.Sprintf("devices: %+v", *device))
	}

	app.render(w, r, http.StatusOK, "home.tmpl", tmplData)
}

func (app *application) updateLocation(w http.ResponseWriter, r *http.Request) {
	var jsonData envelope

	// Decode JSON from request body
	type LocationUpdateRequest struct {
		LocationName string `json:"locationName"`
	}

	var input LocationUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if input.LocationName == "" {
		app.ajaxResponse(w, http.StatusBadRequest, envelope{"locationName": "cannot be empty"}, fmt.Errorf("location name is empty"))
		return
	}

	// Get the location ID from URL path
	locationID, err := getPathID(r)
	if err != nil {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, err)
		return
	}

	// DEBUG: Log the incoming values
	app.logger.Debug(fmt.Sprintf("locationID: %d", locationID))
	app.logger.Debug(fmt.Sprintf("input: %+v", input))

	// Retrieve the location by ID
	location := app.Models.Location.GetByID(uint(locationID))
	if location == nil {
		app.ajaxResponse(w, http.StatusNotFound, envelope{"message": "location not found"}, fmt.Errorf("location with ID %d not found", locationID))
		return
	}

	// Update the name
	location.Name = input.LocationName
	err = app.Models.Location.UpdateName(location)
	if err != nil {
		app.ajaxResponse(w, http.StatusInternalServerError, jsonData, fmt.Errorf("update location name: %w", err))
		return
	}

	// Success response
	jsonData = envelope{"message": fmt.Sprintf("location name updated: %s", input.LocationName)}
	app.ajaxResponse(w, http.StatusOK, jsonData, nil)
}

// CommandDevice handler - allows sending a command to a specific IoT device
func (app *application) commandDevice(w http.ResponseWriter, r *http.Request) {
	// Ensure only POST requests are allowed
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w, r)
		return
	}

	// Parse device command from request body
	type CommandRequest struct {
		DeviceID string `json:"device_id"`
		Command  string `json:"command"`
	}

	var cmdReq CommandRequest
	err := json.NewDecoder(r.Body).Decode(&cmdReq)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Simulate sending the command to the device (e.g., via MQTT, API call, etc.)
	// For now, we just log it
	app.logger.Debug(fmt.Sprintf("Sending command '%s' to device '%s'", cmdReq.Command, cmdReq.DeviceID))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Command sent"})
	if err != nil {
		return
	}
}

// GetDeviceInfo handler - retrieves device details
func (app *application) getDeviceInfo(w http.ResponseWriter, r *http.Request) {
	// Ensure only GET requests are allowed
	if r.Method != http.MethodGet {
		app.methodNotAllowed(w, r)
		return
	}

	// Get device ID from query parameters
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "Missing device_id parameter", http.StatusBadRequest)
		return
	}

	// Simulated device info
	deviceInfo := map[string]interface{}{
		"device_id": deviceID,
		"name":      "Smart Light",
		"status":    "Online",
		"battery":   "85%",
	}

	// Send device info as JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(deviceInfo)
	if err != nil {
		return
	}
}
