package kitchenv3

import (
	"challenge/client"
	"fmt"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	return fmt.Sprintf("%d validation errors occurred", len(v))
}

func IsValidOrder(order client.Order) error {
	var errs ValidationErrors

	if order.ID == "" {
		errs = append(errs, ValidationError{Field: "ID", Message: "is required"})
	}

	if order.Name == "" {
		errs = append(errs, ValidationError{Field: "Name", Message: "is required"})
	}

	// Validate Temperature
	temp := Temperature(order.Temp)
	switch temp {
	case TemperatureHot, TemperatureCold, TemperatureRoom:
	default:
		errs = append(
			errs, ValidationError{Field: "Temp", Message: "must be one of hot, cold, or room"})
	}

	if order.Price <= 0 {
		errs = append(errs, ValidationError{Field: "Price", Message: "must be greater than zero"})
	}

	if order.Freshness <= 0 {
		errs = append(errs, ValidationError{Field: "Freshness", Message: "must be positive"})
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
