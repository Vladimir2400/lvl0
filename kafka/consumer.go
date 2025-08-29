package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
	"wb-service/database"
	"wb-service/models"

	"github.com/segmentio/kafka-go"
)

// OrderCache - кэш для хранения заказов в памяти.
var OrderCache = struct {
	sync.RWMutex
	Data map[string]models.Order // Поле Data теперь экспортируемое (публичное)
}{Data: make(map[string]models.Order)}

// LoadCacheFromDB загружает все заказы из базы данных в кэш
func LoadCacheFromDB() {
	var orders []models.Order

	result := database.DB.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders)

	if result.Error != nil {
		log.Printf("Ошибка при загрузке кэша из БД: %v", result.Error)
		return
	}

	OrderCache.Lock()
	defer OrderCache.Unlock()

	for _, order := range orders {
		OrderCache.Data[order.OrderUID] = order // Обращаемся к OrderCache.Data
	}

	log.Printf("Кэш успешно загружен. %d записей.", len(orders))
}

// StartConsumer запускает процесс прослушивания топика Kafka.
func StartConsumer() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "orders",
		GroupID:        "order-group",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})

	log.Println("Kafka Consumer запущен...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("ошибка при чтении сообщения: %v", err)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("Ошибка десериализации JSON: %v. Сообщение: %s", err, string(m.Value))
			continue
		}

		if result := database.DB.Create(&order); result.Error != nil {
			log.Printf("Ошибка сохранения заказа в БД: %v", result.Error)
			continue
		}

		OrderCache.Lock()
		OrderCache.Data[order.OrderUID] = order // Обращаемся к OrderCache.Data
		OrderCache.Unlock()

		log.Printf("Заказ %s успешно обработан и сохранен.", order.OrderUID)
	}
}
