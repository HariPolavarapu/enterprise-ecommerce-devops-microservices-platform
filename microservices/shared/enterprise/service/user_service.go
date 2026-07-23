package service

import (
	"errors"
	"fmt"
	"time"

	"enterprise/domain"
	"enterprise/repository"
)

type CreateUserInput struct {
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber string
	Role        domain.UserRole
}

type AddressInput struct {
	Line1       string
	Line2       string
	City        string
	State       string
	PostalCode  string
	Country     string
	AddressType domain.AddressType
	IsDefault   bool
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(input CreateUserInput) (*domain.User, error) {
	if input.Email == "" || input.FirstName == "" || input.LastName == "" {
		return nil, errors.New("email, first name and last name are required")
	}
	if input.Role == "" {
		input.Role = domain.RoleCustomer
	}
	user := &domain.User{
		Email:       input.Email,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		PhoneNumber: input.PhoneNumber,
		Role:        input.Role,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Preferences: domain.UserPreferences{DefaultCurrency: "USD", Locale: "en-US"},
	}
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUser(id string) (*domain.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) UpdateProfile(userID, firstName, lastName, phone string) (*domain.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	user.FirstName = firstName
	user.LastName = lastName
	user.PhoneNumber = phone
	user.UpdatedAt = time.Now()
	return user, s.repo.Save(user)
}

func (s *UserService) AddAddress(userID string, input AddressInput) (*domain.Address, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	address := &domain.Address{
		ID:          fmt.Sprintf("addr-%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Line1:       input.Line1,
		Line2:       input.Line2,
		City:        input.City,
		State:       input.State,
		PostalCode:  input.PostalCode,
		Country:     input.Country,
		AddressType: input.AddressType,
		IsDefault:   input.IsDefault,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	user.Addresses = append(user.Addresses, *address)
	if input.IsDefault {
		for i := range user.Addresses {
			if user.Addresses[i].ID != address.ID {
				user.Addresses[i].IsDefault = false
			}
		}
	}
	return address, s.repo.Save(user)
}

func (s *UserService) GetAddress(userID string) ([]domain.Address, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return user.Addresses, nil
}

func (s *UserService) SetDefaultAddress(userID, addressID string) (*domain.Address, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	for i := range user.Addresses {
		if user.Addresses[i].ID == addressID {
			user.Addresses[i].IsDefault = true
		} else {
			user.Addresses[i].IsDefault = false
		}
	}
	user.UpdatedAt = time.Now()
	return &user.Addresses[0], s.repo.Save(user)
}
