package data

import (
	"errors"
	"fmt"
	"slices"
	
	"gorm.io/gorm"
)

// When adding a module, be sure to add it in ModuleNames below and in *ModuleModels.Set(Module, any)

var ModuleNames = []string{
	LIGHT_CONTROLLER,
	LIGHT_SENSOR,
	PRESENCE_DETECTOR,
	LUMINOSITY_SENSOR,
	TEMPERATURE_SENSOR,
	CONSUMPTION_SENSOR,
	RESET,
}

func (m *ModuleModels) Set(module Module, value any) error {
	iModule, err := module.ToIModule()
	if err != nil {
		return err
	}
	
	device, err := m.GetDevice(module.DeviceID)
	if err != nil {
		return err
	}
	channel := device.GetChannel(iModule)
	
	// Set value according to Module type
	switch iModule.(type) {
	
	case LightController:
		err = m.LightController.Set(channel, value)
		if err != nil {
			return err
		}
	
	case LightSensor:
		err = m.LightSensor.Set(channel, value)
		if err != nil {
			return err
		}
	
	case PresenceDetector:
		err = m.PresenceDetector.Set(channel, value)
		if err != nil {
			return err
		}
	
	case LuminositySensor:
		err = m.LuminositySensor.Set(channel, value)
		if err != nil {
			return err
		}
	
	case TemperatureSensor:
		err = m.TemperatureSensor.Set(channel, value)
		if err != nil {
			return err
		}
	
	case ConsumptionSensor:
		err = m.ConsumptionSensor.Set(channel, value)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func (m *ModuleModels) GetDevice(deviceID string) (*Device, error) {
	var device Device
	
	err := m.DB.First(&device, "id = ?", deviceID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf("device %s not found", deviceID)
		default:
			return nil, fmt.Errorf("failed to query device: %w", err)
		}
	}
	
	return &device, nil
}

type ModuleModel struct {
	DB     *gorm.DB
	Broker *Broker
}

func (m *ModuleModel) GetByID(id uint) (*IModule, error) {
	var module Module
	m.DB.First(&module, id)
	
	iModule, err := module.ToIModule()
	if err != nil {
		return nil, err
	}
	
	return &iModule, nil
}

func (m *DataModel) UpdateModule(module *Module) error {
	err := m.DB.Model(module).Select("value").Updates(module).Error
	if err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}
	
	return nil
}

type Module struct {
	gorm.Model
	DeviceID string `gorm:"index"`
	Name     string
	Value    string
}

func (m *Module) GetValue() any {
	return m.Value
}

func (m *Module) GetName() string {
	return m.Name
}

func (m *Module) ToIModule() (IModule, error) {
	if !slices.Contains(ModuleNames, m.Name) {
		return nil, fmt.Errorf("module %s not found", m.Name)
	}
	switch m.Name {
	case PRESENCE_DETECTOR:
		value, err := ToBool(m.Value)
		if err != nil {
			return nil, err
		}
		return PresenceDetector{
			Model:      m.Model,
			DeviceID:   m.DeviceID,
			Name:       m.Name,
			IsPresence: value,
		}, nil
	
	case TEMPERATURE_SENSOR:
		value, err := ToFloat(m.Value)
		if err != nil {
			return nil, err
		}
		return TemperatureSensor{
			Model:        m.Model,
			DeviceID:     m.DeviceID,
			Name:         m.Name,
			ValueDegrees: value,
		}, nil
	
	case CONSUMPTION_SENSOR:
		value, err := ToFloat(m.Value)
		if err != nil {
			return nil, err
		}
		return ConsumptionSensor{
			Model:         m.Model,
			DeviceID:      m.DeviceID,
			Name:          m.Name,
			ValueWattHour: value,
		}, nil
	
	case LIGHT_CONTROLLER:
		value, err := ToBool(m.Value)
		if err != nil {
			return nil, err
		}
		return LightController{
			Model:    m.Model,
			DeviceID: m.DeviceID,
			Name:     m.Name,
			On:       value,
		}, nil
	
	case LIGHT_SENSOR:
		value, err := ToBool(m.Value)
		if err != nil {
			return nil, err
		}
		return LightSensor{
			Model:    m.Model,
			DeviceID: m.DeviceID,
			Name:     m.Name,
			IsOn:     value,
		}, nil
	
	case LUMINOSITY_SENSOR:
		value, err := ToFloat(m.Value)
		if err != nil {
			return nil, err
		}
		return LuminositySensor{
			Model:      m.Model,
			DeviceID:   m.DeviceID,
			Name:       m.Name,
			ValueLumen: value,
		}, nil
	
	case RESET:
		value, err := ToBool(m.Value)
		if err != nil {
			return nil, err
		}
		return Reset{
			Model:     m.Model,
			DeviceID:  m.DeviceID,
			Name:      m.Name,
			BoolValue: value,
		}, nil
	
	default:
		return nil, fmt.Errorf("module %s not found", m.Name)
	}
}
