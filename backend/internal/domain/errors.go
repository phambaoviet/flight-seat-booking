package domain

import "errors"

var (
	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatAlreadyBooked = errors.New("seat is already booked")
	ErrSeatAlreadyOnHold = errors.New("seat is already on hold")
	ErrHoldNotFound      = errors.New("hold not found")
	ErrSeatHoldNotFound = errors.New("seat hold not found")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInValidPassword = errors.New("password must be at least 6 characters long")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
	ErrSeatAlreadyHeld = errors.New("seat is already held")
	ErrUnauthorizedRelease = errors.New("unauthorized to release this hold")
	ErrHoldTokenMismatch = errors.New("hold token does not match")
)
