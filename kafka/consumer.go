package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"wb-service/config"
	"wb-service/database"
	"wb-service/internal/cache"
	"wb-service/internal/interfaces"
	"wb-service/internal/validator"
	"wb-service/models"

	"github.com/segmentio/kafka-go"
)

// OrderCache - кэш для хранения заказов в памяти.
var OrderCache interfaces.Cache

// InitCache инициализирует кэш
func InitCache(cfg *config.Config) {
	ttl := time.Duration(cfg.Cache.TTL) * time.Second
	OrderCache = cache.NewLRUCache(cfg.Cache.MaxSize, ttl)
}

// LoadCacheFromDB загружает все заказы из базы данных в кэш
func LoadCacheFromDB(db interfaces.Database) error {
	if OrderCache == nil {
		log.Println("Кэш не инициализирован")
		return nil
	}

	if err := OrderCache.LoadFromDB(db); err != nil {
		log.Printf("Ошибка при загрузке кэша из БД: %v", err)
		return err
	}

	log.Printf("Кэш успешно загружен. %d записей.", OrderCache.Size())
	return nil
}

// StartConsumer запускает процесс прослушивания топика Kafka.
func StartConsumer(cfg *config.Config, ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topic,
		GroupID:        cfg.Kafka.GroupID,
		MinBytes:       cfg.Kafka.MinBytes,
		MaxBytes:       cfg.Kafka.MaxBytes,
		CommitInterval: time.Second,
	})
	defer r.Close()

	// Создаем валидатор
	orderValidator := validator.NewOrderValidator()

	log.Println("Kafka Consumer запущен...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Получен сигнал остановки Kafka Consumer")
			return
		default:
			// Устанавливаем таймаут для чтения сообщения
			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			m, err := r.FetchMessage(msgCtx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					continue
				}
				log.Printf("ошибка при чтении сообщения: %v", err)
				continue
			}

			var order models.Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("Ошибка десериализации JSON: %v. Сообщение: %s", err, string(m.Value))
				// Коммитим сообщение даже если не смогли его распарсить
				r.CommitMessages(context.Background(), m)
				continue
			}

			// Валидируем заказ
			if err := orderValidator.Validate(&order); err != nil {
				log.Printf("Ошибка валидации заказа %s: %v", order.OrderUID, err)
				// Коммитим сообщение даже если валидация не прошла
				r.CommitMessages(context.Background(), m)
				continue
			}

			if result := database.DB.Create(&order); result.Error != nil {
				log.Printf("Ошибка сохранения заказа в БД: %v", result.Error)
				continue
			}

			// Добавляем в кэш
			OrderCache.Set(order.OrderUID, &order)

			// Коммитим сообщение после успешной обработки
			if err := r.CommitMessages(context.Background(), m); err != nil {
				log.Printf("Ошибка коммита сообщения: %v", err)
			} else {
				log.Printf("Заказ %s успешно обработан и сохранен.", order.OrderUID)
			}
		}
	}
}
