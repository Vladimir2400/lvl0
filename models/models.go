package models

import "time"

// Order представляет главную структуру заказа
type Order struct {
	OrderUID          string    `gorm:"primaryKey" json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `gorm:"foreignKey:OrderUID;references:OrderUID" json:"delivery"`
	Payment           Payment   `gorm:"foreignKey:OrderUID;references:OrderUID" json:"payment"`
	Items             []Item    `gorm:"foreignKey:OrderUID;references:OrderUID" json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

// Delivery информация о доставке
type Delivery struct {
	ID       uint   `gorm:"primaryKey" json:"-"` // ID не приходит из JSON
	OrderUID string `gorm:"index" json:"-"`      // OrderUID тоже
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Zip      string `json:"zip"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Region   string `json:"region"`
	Email    string `json:"email"`
}

// Payment информация об оплате
type Payment struct {
	ID           uint   `gorm:"primaryKey" json:"-"`
	OrderUID     string `gorm:"index" json:"-"`
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

// Item информация о товаре в заказе
type Item struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	OrderUID    string `gorm:"index" json:"-"`
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
