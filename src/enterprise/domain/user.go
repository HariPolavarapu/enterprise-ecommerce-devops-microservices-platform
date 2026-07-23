package domain

import "time"

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
)

type User struct {
	ID           string
	Email        string
	FirstName    string
	LastName     string
	PhoneNumber  string
	Role         UserRole
	IsActive     bool
	IsDeleted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Addresses    []Address
	Preferences  UserPreferences
}

type UserPreferences struct {
	DefaultCurrency string
	Newsletter      bool
	Locale          string
}

type AddressType string

const (
	AddressTypeShipping AddressType = "shipping"
	AddressTypeBilling  AddressType = "billing"
)

type Address struct {
	ID           string
	UserID       string
	Line1        string
	Line2        string
	City         string
	State        string
	PostalCode   string
	Country      string
	AddressType  AddressType
	IsDefault    bool
	IsDeleted    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
