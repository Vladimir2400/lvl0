package interfaces

import (
	"context"
	"wb-service/models"
)

// Database интерфейс для работы с базой данных
type Database interface {
	CreateOrder(order *models.Order) error
	GetOrder(orderUID string) (*models.Order, error)
	GetAllOrders() ([]models.Order, error)
	Close() error
}

// Cache интерфейс для работы с кэшем
type Cache interface {
	Get(key string) (*models.Order, bool)
	Set(key string, order *models.Order)
	LoadFromDB(db Database) error
	Size() int
	Clear()
}

// MessageConsumer интерфейс для получения сообщений из очереди
type MessageConsumer interface {
	Start(ctx context.Context) error
	Stop() error
}

// OrderValidator интерфейс для валидации заказов
type OrderValidator interface {
	Validate(order *models.Order) error
}

// OrderService основной интерфейс сервиса заказов
type OrderService interface {
	GetOrder(orderUID string) (*models.Order, error)
	ProcessOrder(order *models.Order) error
}