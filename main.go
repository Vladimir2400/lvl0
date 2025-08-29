package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"wb-service/database"
	"wb-service/kafka"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getOrder обрабатывает запрос на получение заказа по его UID
func getOrder(c *gin.Context) {
	orderUID := c.Param("order_uid")

	// --- Новая логика: сначала ищем в кэше ---
	kafka.OrderCache.RLock()
	order, found := kafka.OrderCache.Data[orderUID] // Обращаемся к OrderCache.Data
	kafka.OrderCache.RUnlock()

	if found {
		log.Printf("Заказ %s НАЙДЕН В КЭШЕ!", orderUID)
		c.JSON(http.StatusOK, order)
		return
	}

	// --- Старая логика: если в кэше нет, идем в БД ---
	log.Printf("Заказ %s не найден в кэше, ищем в БД...", orderUID)
	result := database.DB.
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		First(&order, "order_uid = ?", orderUID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	// --- Новое: добавляем в кэш то, что нашли в БД ---
	log.Printf("Заказ %s найден в БД, добавляем в кэш...", orderUID)
	kafka.OrderCache.Lock() // Блокируем на запись
	kafka.OrderCache.Data[order.OrderUID] = order // Обращаемся к OrderCache.Data
	kafka.OrderCache.Unlock()

	// Отправляем найденный заказ в виде JSON
	c.JSON(http.StatusOK, order)
}

func main() {
	fmt.Println("Сервис запускается...")

	// Инициализируем подключение к базе данных
	database.Init()

	// Загружаем кэш из базы данных при старте
	kafka.LoadCacheFromDB()

	// Запускаем Kafka Consumer в отдельной горутине
	go kafka.StartConsumer()

	r := gin.Default()

	// Добавляем маршрут для получения заказа
	r.GET("/order/:order_uid", getOrder)

	// Добавляем маршрут для отдачи нашей веб-страницы
	r.StaticFile("/", "./web/index.html")

	// Запускаем сервер на порту 8080
	r.Run(":8080")
}
