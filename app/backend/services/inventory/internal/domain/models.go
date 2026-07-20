package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Location    string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewWarehouse(id uuid.UUID, code, name, location string, now time.Time) *Warehouse {
	return &Warehouse{ID: id, Code: strings.TrimSpace(strings.ToUpper(code)), Name: strings.TrimSpace(name), Location: strings.TrimSpace(location), Active: true, CreatedAt: now, UpdatedAt: now}
}

type StockItem struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	OnHand      int64
	Reserved    int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewStockItem(id, productID, warehouseID uuid.UUID, onHand int64, now time.Time) *StockItem {
	return &StockItem{ID: id, ProductID: productID, WarehouseID: warehouseID, OnHand: onHand, Reserved: 0, CreatedAt: now, UpdatedAt: now}
}

func (s *StockItem) Available() int64 { return s.OnHand - s.Reserved }

func (s *StockItem) Adjust(delta int64, now time.Time) {
	s.OnHand += delta
	if s.OnHand < 0 {
		s.OnHand = 0
	}
	if s.Reserved > s.OnHand {
		s.Reserved = s.OnHand
	}
	s.UpdatedAt = now
}

func (s *StockItem) Reserve(qty int64, now time.Time) error {
	if qty <= 0 || s.Available() < qty {
		return ErrInsufficientStock
	}
	s.Reserved += qty
	s.UpdatedAt = now
	return nil
}

func (s *StockItem) Release(qty int64, now time.Time) {
	if qty <= 0 {
		return
	}
	s.Reserved -= qty
	if s.Reserved < 0 {
		s.Reserved = 0
	}
	s.UpdatedAt = now
}

type ReservationStatus string

const (
	ReservationActive   ReservationStatus = "active"
	ReservationReleased ReservationStatus = "released"
)

type Reservation struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Reference   string
	Quantity    int64
	Status      ReservationStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewReservation(id, productID, warehouseID uuid.UUID, reference string, qty int64, now time.Time) *Reservation {
	return &Reservation{ID: id, ProductID: productID, WarehouseID: warehouseID, Reference: strings.TrimSpace(reference), Quantity: qty, Status: ReservationActive, CreatedAt: now, UpdatedAt: now}
}

func (r *Reservation) Release(now time.Time) {
	r.Status = ReservationReleased
	r.UpdatedAt = now
}
