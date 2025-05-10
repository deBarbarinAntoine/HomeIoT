package data

import (
	"encoding/json"
	"fmt"
	
	"gorm.io/gorm"
)

type StartupMessage struct {
	DeviceID     string `json:"id"`
	Type         string `json:"type"`
	LocationID   uint   `json:"location_id"`
	LocationType string `json:"location_type"`
	LocationName string `json:"location_name"`
	Modules      []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"modules"`
}

func NewStartupMessage(payload []byte) (*StartupMessage, error) {
	var startupMessage StartupMessage
	if err := json.Unmarshal(payload, &startupMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return &startupMessage, nil
}

func NewResponseMessage(device *Device) *StartupMessage {
	
	// Create the responseMessage
	responseMessage := &StartupMessage{
		DeviceID:     device.ID,
		Type:         device.Type,
		LocationID:   device.LocationID,
		LocationType: device.Location.Type,
		LocationName: device.Location.Name,
	}
	
	// Add the modules
	for _, module := range device.Modules {
		responseMessage.Modules = append(responseMessage.Modules, struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}{Name: module.Name, Value: module.Value})
	}
	
	return responseMessage
}

func (startupMessage *StartupMessage) ToDevice() *Device {
	location := &Location{
		Model: gorm.Model{
			ID: startupMessage.LocationID,
		},
		Type: startupMessage.LocationType,
		Name: startupMessage.LocationName,
	}
	device := &Device{
		ID:         startupMessage.DeviceID,
		LocationID: startupMessage.LocationID,
		Location:   *location,
		Type:       startupMessage.Type,
		Modules:    nil,
	}
	
	for _, module := range startupMessage.Modules {
		device.Modules = append(device.Modules, Module{
			Name:  module.Name,
			Value: module.Value,
		})
	}
	return device
}
