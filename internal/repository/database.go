package repository

import (
	"wb-service/internal/interfaces"
	"wb-service/models"

	"gorm.io/gorm"
)

// GormDatabase реализует интерфейс Database для GORM
type GormDatabase struct {
	db *gorm.DB
}

// NewGormDatabase создает новый экземпляр GormDatabase
func NewGormDatabase(db *gorm.DB) interfaces.Database {
	return &GormDatabase{db: db}
}

// CreateOrder создает новый заказ в базе данных
func (g *GormDatabase) CreateOrder(order *models.Order) error {
	return g.db.Create(order).Error
}

// GetOrder получает заказ по UID
func (g *GormDatabase) GetOrder(orderUID string) (*models.Order, error) {
	var order models.Order
	err := g.db.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		First(&order, "order_uid = ?", orderUID).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// GetAllOrders получает все заказы из базы данных
func (g *GormDatabase) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	err := g.db.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error

	return orders, err
}

// Close закрывает соединение с базой данных
func (g *GormDatabase) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}