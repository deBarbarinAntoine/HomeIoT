package data

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	
	"gorm.io/gorm"
)

type Device struct {
	ID         string `gorm:"primaryKey;index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	LocationID uint
	Location   Location `gorm:"foreignKey:LocationID"`
	Type       string
	Name       string
	Modules    []Module `gorm:"foreignKey:DeviceID"`
}

func (d *Device) GetChannel(iModule IModule) string {
	channel := fmt.Sprintf("home/%s/%d/%s/%s/%s", d.Location.Type, d.LocationID, d.Type, d.ID, iModule.GetName())
	return channel
}

type DeviceModel struct {
	DB     *gorm.DB
	Broker *Broker
}

func (m *DeviceModel) GetByID(id string) (*Device, error) {
	var device Device
	err := m.DB.Joins("Location").Joins("Module").Where("device_id = ?", id).First(&device).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf("device with id %s not found", id)
		default:
			return nil, fmt.Errorf("failed to get device with id %s: %w", id, err)
		}
	}
	
	return &device, nil
}

func (m *DeviceModel) GetByLocationID(locationID uint) ([]*Device, error) {
	var devices []*Device
	err := m.DB.Joins("Location").Joins("Module").Where("location_id = ?", locationID).First(&devices).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf("device in location with id %d not found", locationID)
		default:
			return nil, fmt.Errorf("failed to get device in location with id %d: %w", locationID, err)
		}
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("device in location with id %d not found", locationID)
	}
	
	return devices, nil
}

func (m *DeviceModel) GetAll() ([]*Device, error) {
	var devices []*Device
	err := m.DB.Joins("Location").Preload("Modules").Find(&devices).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("devices not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("0 devices found")
	}
	return devices, nil
}

func (m *DeviceModel) UpdateLocation(device *Device) error {
	
	result := m.DB.FirstOrCreate(&device.Location, &Location{Name: device.Location.Name, Type: device.Location.Type})
	if result.Error != nil {
		return fmt.Errorf("error updating location: %w", result.Error)
	}
	err := m.DB.Model(&device).Where("id = ?", device.ID).Update("location_id", device.Location.ID).Error
	if err != nil {
		return fmt.Errorf("error updating device locationID: %w", err)
	}
	
	err = m.Reset(device)
	if err != nil {
		return err
	}
	
	return nil
}

func (m *DeviceModel) Reset(device *Device) error {
	resetModule, err := NewResetModule()
	if err != nil {
		return err
	}
	channel := device.GetChannel(resetModule)
	resetValue, err := ToBool(resetModule.GetValue())
	if err != nil {
		return fmt.Errorf("error getting value for reset module %s: %w", resetModule.GetName(), err)
	}
	
	m.Broker.Pub(channel, strconv.FormatBool(resetValue))
	
	return nil
}

func (m *DeviceModel) CheckOrCreate(device *Device) error {
	
	// Fetch device from Database by ID
	err := m.DB.Joins("Module").First(&device, "id = ?", device.ID).Error
	
	// Handle errors
	if err != nil {
		switch {
		
		// Device does not exist
		case errors.Is(err, gorm.ErrRecordNotFound):
			
			// Check or create the location
			err = m.CheckOrCreateLocation(&device.Location)
			if err != nil {
				return err
			}
			
			// Create the device with its modules
			result := m.DB.Create(device)
			if result.Error != nil {
				return fmt.Errorf("could not create device %v: %w", device.ID, result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("could not create device %v: %d rows affected", device.ID, result.RowsAffected)
			}
		default:
			return fmt.Errorf("error fetching device: %w", err)
		}
	}
	
	return nil
}

func (m *DeviceModel) CheckOrCreateLocation(location *Location) error {
	err := m.DB.First(&location, location.ID).Error
	if err == nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			result := m.DB.Create(&location)
			if result.Error != nil {
				return fmt.Errorf("could not create location: %w", err)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("could not create location: %d rows affected", result.RowsAffected)
			}
		default:
			return fmt.Errorf("error fetching location: %w", err)
		}
	}
	return nil
}
