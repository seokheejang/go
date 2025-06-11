package user

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents a role in the system
type UserRole struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `gorm:"unique"`
}

// User represents a user in the system
type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string
	RoleID    uint
	Role      *UserRole `gorm:"foreignKey:role_id;references:id"`
}
