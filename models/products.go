package models

import "github.com/google/uuid"

type Product struct {
	Id     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	QrCode string    `json:"qr_code"`
}
