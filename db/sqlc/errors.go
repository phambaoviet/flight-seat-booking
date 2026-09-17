package db

import "errors"

var (
	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatAlreadyBooked = errors.New("seat is already booked")
	ErrSeatAlreadyOnHold = errors.New("seat is already on hold")
)