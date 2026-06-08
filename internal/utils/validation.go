package utils

import (
	"errors"
	"strings"
)

func ValidateImageType(value string) (string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "", errors.New("Field cannot be empty")
	}

	if trimmedValue != "jpg" && trimmedValue != "png" && trimmedValue != "jpeg" && trimmedValue != "webp" {
		return "", errors.New("Type is not valid (jpg,png,jpeg)")
	}
	return trimmedValue, nil
}
