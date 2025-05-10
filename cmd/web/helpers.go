package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
	
	"HomeIoT/internal/validator"
	
	"github.com/alexedwards/flow"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// logout clears the session and renews the token.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	error - If any error occurs during the process
func (app *application) logout(r *http.Request) error {
	
	err := app.sessionManager.Clear(r.Context())
	if err != nil {
		return err
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}
	
	return nil
}

// newNonce generates a new nonce.
//
// Returns:
//
//	string - The generated nonce
//	error - If any error occurs during the process
func newNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(nonceBytes), nil
}

// getNonce retrieves the nonce from the request context.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	string - The nonce value, or an empty string if not found
func (app *application) getNonce(r *http.Request) string {
	nonce, ok := r.Context().Value(nonceContextKey).(string)
	if !ok {
		app.logger.Error("no nonce in request context")
		return ""
	}
	return nonce
}

// decodePostForm decodes the POST form data into a struct.
//
// Parameters:
//
//	r - The HTTP request
//	dst - The destination struct to decode the form data into
//
// Returns:
//
//	error - If any error occurs during the decoding process
func (app *application) decodePostForm(r *http.Request, dst any) error {
	
	err := r.ParseForm()
	if err != nil {
		return err
	}
	
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		
		return err
	}
	
	return nil
}

// clientError handles client-side errors.
//
// Parameters:
//
//	w - The HTTP response writer
//	r - The HTTP request
//	status - The HTTP status code
func (app *application) clientError(w http.ResponseWriter, r *http.Request, status int) {
	
	// setting the templateData
	tmplData := app.newTemplateData(r)
	
	// setting the error title and message
	tmplData.Error.Title = fmt.Sprintf("Error %d", status)
	
	if status == http.StatusNotFound {
		tmplData.Error.Message = "We didn't find what you were looking for :("
	} else {
		tmplData.Error.Message = "Something went wrong!"
	}
	
	// rendering the error page
	app.render(w, r, status, "error.tmpl", tmplData)
}

// failedValidationError handles validation errors.
//
// Parameters:
//
//	w - The HTTP response writer
//	r - The HTTP request
//	form - The form data
//	v - The validator instance
//	page - The template page to render
func (app *application) failedValidationError(w http.ResponseWriter, r *http.Request, form any, v *validator.Validator, page string) {
	
	// DEBUG
	app.logger.Debug(fmt.Sprintf("generic errors: %+v", v.NonFieldErrors))
	app.logger.Debug(fmt.Sprintf("field errors: %+v", v.FieldErrors))
	
	// retrieving basic template data
	tmplData := app.newTemplateData(r)
	
	tmplData.Form = form
	
	// render the template
	app.render(w, r, http.StatusUnprocessableEntity, page, tmplData)
}

// serverError handles server-side errors.
//
// Parameters:
//
//	w - The HTTP response writer
//	r - The HTTP request
//	err - The error that occurred
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		status = http.StatusInternalServerError
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)
	
	// logging the error
	app.logger.Error(err.Error(), slog.String("method", method), slog.String("URI", uri), slog.String("trace", trace))
	
	// setting the templateData
	tmplData := app.newTemplateData(r)
	
	// setting the error title and message
	tmplData.Error.Title = fmt.Sprintf("Error %d", status)
	tmplData.Error.Message = "Something went wrong!"
	
	// rendering the error page
	app.render(w, r, status, "error.tmpl", tmplData)
}

// ajaxResponse sends an AJAX response.
//
// Parameters:
//
//	w - The HTTP response writer
//	status - The HTTP status code
//	msg - The message to send in the response
func (app *application) ajaxResponse(w http.ResponseWriter, status int, msg envelope, err error) {
	
	// setting the response data
	var resData envelope
	
	// DEBUG
	app.logger.Debug(fmt.Sprintf("status: %d", status))
	
	// checking the status code
	if status < http.StatusBadRequest {
		
		// wrapping the message in a JSON object
		resData = envelope{"response": msg}
		
	} else {
		if nil == err {
			err = fmt.Errorf("no error sent to ajaxResponse")
		}
		// logging the error
		app.logger.Error(err.Error())
		
		// wrapping error in JSON object
		resData = envelope{"error": "internal server error"}
	}
	
	// marshalling the resData
	jsonData, err := json.Marshal(resData)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	
	// setting the Content-Type header to JSON
	w.Header().Set("Content-Type", "application/json")
	
	// setting the Status response
	w.WriteHeader(status)
	
	// send the response with the JSON data
	_, err = w.Write(jsonData)
	if err != nil {
		app.logger.Error(err.Error())
	}
}

// background runs a function in the background.
//
// Parameters:
//
//	fn - The function to run
func (app *application) background(fn func()) {
	
	app.wg.Add(1)
	go func() {
		
		defer app.wg.Done()
		
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		
		fn()
		
	}()
}

// isAuthenticated checks if the user is authenticated.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	bool - True if the user is authenticated, false otherwise
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	
	return isAuthenticated
}

// getUserID retrieves the user ID from the session.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	int - The user ID, or 0 if not found
func (app *application) getUserID(r *http.Request) int {
	id, ok := app.sessionManager.Get(r.Context(), authenticatedUserIDSessionManager).(int)
	if !ok {
		return 0
	}
	return id
}

// getUserRole retrieves the user role from the session.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	string - The user role, or an empty string if not found
func (app *application) getUserRole(r *http.Request) string {
	role, ok := app.sessionManager.Get(r.Context(), userRoleSessionManager).(string)
	if !ok {
		return ""
	}
	return role
}

// newTemplateData retrieves the template data for rendering a page.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	templateData - The template data containing various information
func (app *application) newTemplateData(r *http.Request) templateData {
	
	// retrieving the nonce
	nonce := app.getNonce(r)
	
	// returning the templateData with all information
	var tmplData = templateData{
		WebsocketURL: fmt.Sprintf("ws://%s:%d/ws", app.config.websocketURL, app.config.port),
		CurrentYear:  time.Now().Year(),
		Flash:        app.sessionManager.PopString(r.Context(), "flash"),
		Nonce:        nonce,
		CSRFToken:    nosurf.Token(r),
		Error: struct {
			Title   string
			Message string
		}{
			Title:   "Error 404",
			Message: "We didn't find what you were looking for :(",
		},
	}
	
	return tmplData
}

// render renders a template and writes the response to the HTTP writer.
//
// Parameters:
//
//	w - The HTTP response writer
//	r - The HTTP request
//	status - The HTTP status code
//	page - The template page to render
//	data - The template data
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	
	// retrieving the appropriate set of templates
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, r, fmt.Errorf("the template %s does not exist", page))
		return
	}
	
	// creating a bytes Buffer
	buf := new(bytes.Buffer)
	
	// executing the template in the buffer to catch any possible parsing error,
	// so that the user doesn't see a half-empty page
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// if it's all okay, write the status in the header and write the buffer in the ResponseWriter
	w.WriteHeader(status)
	
	buf.WriteTo(w)
}

// getPathID retrieves the integer ID from the URL path.
//
// Parameters:
//
//	r - The HTTP request
//
// Returns:
//
//	int - The integer ID, or 0 if not found
//	error - If any error occurs during the process
func getPathID(r *http.Request) (int, error) {
	
	// fetching the id param from the URL
	param := flow.Param(r.Context(), "id")
	
	// looking for errors
	if param == "" {
		return 0, fmt.Errorf("id param required")
	}
	
	// converting the param to int
	id, err := strconv.Atoi(param)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id param: %w", err)
	}
	
	// return the integer id
	return id, nil
}
