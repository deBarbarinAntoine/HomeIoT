package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	
	"HomeIoT/internal/data"
	"HomeIoT/ui"
)

// functions is a map of template functions available in all templates.
var functions = template.FuncMap{
	"humanDate":         humanDate,
	"bytesToString":     bytesToString,
	"increment":         increment,
	"decrement":         decrement,
	"transactionStatus": transactionStatus,
	"moduleName":        moduleName,
	"floatPrecision1":   floatPrecision1,
	"isActuator":        isActuator,
	"isFalse":           isFalse,
	"isTrue":            isTrue,
}

func isTrue(s string) bool {
	return strings.ToLower(s) == "true" || s == "1"
}

func isFalse(s string) bool {
	return strings.ToLower(s) == "false" || s == "0"
}

func isActuator(name string) bool {
	switch name {
	case data.LIGHT_CONTROLLER:
		return true
	case data.PRESENCE_DETECTOR,
		data.LUMINOSITY_SENSOR,
		data.CONSUMPTION_SENSOR,
		data.LIGHT_SENSOR,
		data.TEMPERATURE_SENSOR:
		return false
	default:
		return false
	}
}

func floatPrecision1(value string) string {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}
	return fmt.Sprintf("%.1f", floatValue)
}

func moduleName(name string) string {
	switch name {
	case data.LIGHT_CONTROLLER:
		return "Lighting (ON/OFF)"
	case data.PRESENCE_DETECTOR:
		return "Presence Detector (True/False)"
	case data.TEMPERATURE_SENSOR:
		return "Temperature (ºC)"
	case data.LUMINOSITY_SENSOR:
		return "Luminosity (Lux)"
	case data.CONSUMPTION_SENSOR:
		return "Consumption (W/H)"
	case data.LIGHT_SENSOR:
		return "Lighting (ON/OFF)"
	default:
		return name
	}
}

func transactionStatus(transactionStatus any, status string) string {
	trStatus := transactionStatus.(*string)
	if *trStatus == status {
		return "selected"
	}
	return ""
}

// humanDate formats a time.Time value to a human-readable string.
//
// Parameters:
//
//	t - The time.Time value to format
//
// Returns:
//
//	string - The formatted date and time
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// bytesToString converts a byte slice to a string.
//
// Parameters:
//
//	b - The byte slice to convert
//
// Returns:
//
//	string - The converted string, or an empty string if the input is nil
func bytesToString(b []byte) string {
	if b != nil {
		return string(b)
	}
	return ""
}

// increment increments an integer by 1.
//
// Parameters:
//
//	n - The integer to increment
//
// Returns:
//
//	int - The incremented value
func increment(n int) int {
	return n + 1
}

// decrement decrements an integer by 1.
//
// Parameters:
//
//	n - The integer to decrement
//
// Returns:
//
//	int - The decremented value
func decrement(n int) int {
	return n - 1
}

// newTemplateCache creates a template cache from the templates in the ui.Files file system.
//
// Returns:
//
//	map[string]*template.Template - A map of template names to their corresponding Template instances
//	error - If any error occurs during the process
func newTemplateCache() (map[string]*template.Template, error) {
	
	cache := map[string]*template.Template{}
	
	pages, err := fs.Glob(ui.Files, "templates/pages/*.tmpl")
	if err != nil {
		return nil, err
	}
	
	for _, page := range pages {
		name := filepath.Base(page)
		
		patterns := []string{
			"templates/base.tmpl",
			"templates/partials/*.tmpl",
			page,
		}
		
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		
		cache[name] = ts
	}
	
	return cache, nil
}
