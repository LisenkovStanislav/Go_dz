package handler

import (
	"example.com/orders/internal/model"
)

type OrderCreator interface {
	Create(customer string) (model.Order, error)
}

type Handler struct {
	orders OrderCreator
}

func NewHandler(orders OrderCreator) *Handler {
	return &Handler{orders: orders}
}

func (h *Handler) Create(customer string) (model.Order, error) {
	return h.orders.Create(customer)
}
