package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

var clients = make(map[*websocket.Conn]chan struct{})
var clientsMutex sync.Mutex

func (app *application) startPostgresFetcher(stopChan chan struct{}) {
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-stopChan:
			app.logger.Info("stopping Postgres fetcher")
			return
		case <-ticker.C:
			devices, err := app.Models.Device.GetAll()
			if err != nil {
				app.logger.Error(err.Error())
				return
			}
			
			if len(devices) > 0 {
				
				// Modify module names for user
				for i, device := range devices {
					for j, module := range device.Modules {
						devices[i].Modules[j].Name = moduleName(module.Name)
					}
				}
				
				jsonData, err := json.Marshal(devices)
				if err == nil {
					app.broadcastToClients(jsonData)
				}
			}
		default:
		}
	}
}

func (app *application) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // allow all origins
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.logger.Error(fmt.Sprintf("Websocket upgrade error: %s", err.Error()))
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			app.logger.Error(fmt.Sprintf("Websocket closing error: %s", err.Error()))
		}
	}(conn)
	
	// Create a stop channel for the client
	stopChan := make(chan struct{})
	
	// Register the client in the global clients map with the stop channel
	clientsMutex.Lock()
	clients[conn] = stopChan
	clientsMutex.Unlock()
	
	app.logger.Info("Websocket: new client connected")
	
	// Start the goroutine, passing the stop channel
	go app.startPostgresFetcher(stopChan)
	
	// Listen for client messages and handle disconnection
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	
	// When the client disconnects, close the stop channel to notify the goroutine
	clientsMutex.Lock()
	delete(clients, conn)
	clientsMutex.Unlock()
	app.logger.Info("Websocket: client disconnected")
	
	// Close the stop channel to notify the goroutine to stop
	close(stopChan)
}

func (app *application) broadcastToClients(message []byte) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			app.logger.Error(fmt.Sprintf("Websocket error: %s", err.Error()))
			err := client.Close()
			if err != nil {
				app.logger.Error(fmt.Sprintf("Websocket closing error: %s", err.Error()))
				return
			}
			delete(clients, client)
		}
	}
}
