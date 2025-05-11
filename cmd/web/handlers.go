package main

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	
	"HomeIoT/internal/data"
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
	for _, device := range devices {
		slices.SortFunc(device.Modules, func(a, b data.Module) int {
			return cmp.Compare(a.ID, b.ID)
		})
	}
	tmplData.Devices = devices
	
	// DEBUG
	for device := range slices.Values(devices) {
		app.logger.Debug(fmt.Sprintf("devices: %+v", *device))
	}
	
	app.render(w, r, http.StatusOK, "home.tmpl", tmplData)
}

func (app *application) updateLocation(w http.ResponseWriter, r *http.Request) {
	var jsonData envelope
	
	// Decode form values
	form := newLocationUpdateForm()
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	
	// DEBUG
	app.logger.Debug(fmt.Sprintf("updateLocationForm: %+v", *form))
	
	if form.LocationName == "" {
		app.ajaxResponse(w, http.StatusBadRequest, envelope{"LocationName": "cannot be empty"}, fmt.Errorf("location name is empty"))
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
	app.logger.Debug(fmt.Sprintf("form: %+v", form))
	
	// Retrieve the location by ID
	location := app.Models.Location.GetByID(uint(locationID))
	if location == nil {
		app.ajaxResponse(w, http.StatusNotFound, envelope{"message": "location not found"}, fmt.Errorf("location with ID %d not found", locationID))
		return
	}
	
	// Update the name
	location.Name = form.LocationName
	err = app.Models.Location.UpdateName(location)
	if err != nil {
		app.ajaxResponse(w, http.StatusInternalServerError, jsonData, fmt.Errorf("update location name: %w", err))
		return
	}
	
	// Success response
	jsonData = envelope{"message": fmt.Sprintf("location name updated: %s", form.LocationName)}
	app.ajaxResponse(w, http.StatusOK, jsonData, nil)
}

// CommandDevice handler - allows sending a command to a specific IoT device
func (app *application) commandDevice(w http.ResponseWriter, r *http.Request) {
	var jsonData envelope
	
	deviceID, err := getPathString(r, "deviceID")
	if err != nil {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, err)
		return
	}
	
	StrModuleID, err := getPathString(r, "moduleID")
	if err != nil {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, err)
		return
	}
	
	moduleID, err := strconv.Atoi(StrModuleID)
	if err != nil {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, err)
		return
	}
	
	form := newCommandValueForm()
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.ajaxResponse(w, http.StatusInternalServerError, jsonData, err)
		return
	}
	
	// Check the form Value
	if form.Value == "" {
		app.ajaxResponse(w, http.StatusBadRequest, jsonData, err)
		return
	}
	
	// Get the module from the database
	module, err := app.Models.Module.GetByID(uint(moduleID))
	if err != nil {
		app.ajaxResponse(w, http.StatusNotFound, envelope{"message": "module not found"}, fmt.Errorf("module with ID %d not found", moduleID))
		return
	}
	
	// Send command via MQTT
	err = app.Models.ModuleModels.Set(module, form.Value)
	if err != nil {
		app.ajaxResponse(w, http.StatusInternalServerError, jsonData, err)
		return
	}
	
	// DEBUG
	app.logger.Debug(fmt.Sprintf("Sending value '%s' to moduleID '%d' on device '%s'", form.Value, moduleID, deviceID))
	
	// Send response
	jsonData = envelope{"message": fmt.Sprintf("module %d updated successfully with value %s", moduleID, form.Value)}
	app.ajaxResponse(w, http.StatusOK, jsonData, nil)
}
