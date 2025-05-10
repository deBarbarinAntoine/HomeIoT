package data

import (
	"fmt"
	
	"gorm.io/gorm"
)

type Location struct {
	gorm.Model
	Type string
	Name string `gorm:"unique"`
}

type LocationModel struct {
	DB *gorm.DB
}

func (m *LocationModel) GetByID(id uint) *Location {
	var location *Location
	m.DB.Model(&Location{}).Where("id = ?", id).First(&location)
	return location
}

func (m *LocationModel) Delete(id uint) error {
	var usages int64
	err := m.DB.Model(&Device{}).Where("location_id = ?", id).Count(&usages).Error
	if err != nil {
		return fmt.Errorf("error counting devices in location with id %d: %w", id, err)
	}
	if usages > 0 {
		return fmt.Errorf("location with id %d is not empty", id)
	}
	
	err = m.DB.Delete(&Location{}, id).Error
	if err != nil {
		return fmt.Errorf("error deleting location with id %d: %w", id, err)
	}
	
	return nil
}

func (m *LocationModel) UpdateName(location *Location) error {
	err := m.DB.Model(&Location{}).Where("id = ?", location.ID).Update("name", location.Name).Error
	if err != nil {
		return fmt.Errorf("error updating location: %w", err)
	}
	
	return nil
}

func (m *LocationModel) UpdateType(location *Location) error {
	err := m.DB.Model(&Location{}).Where("id = ?", location.ID).Update("type", location.Type).Error
	if err != nil {
		return fmt.Errorf("error updating location: %w", err)
	}
	
	return nil
}
