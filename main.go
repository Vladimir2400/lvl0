package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wb-service/config"
	"wb-service/database"
	"wb-service/internal/repository"
	"wb-service/kafka"
	"wb-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getOrder обрабатывает запрос на получение заказа по его UID
func getOrder(c *gin.Context) {
	orderUID := c.Param("order_uid")

	// --- Новая логика: сначала ищем в кэше ---
	if order, found := kafka.OrderCache.Get(orderUID); found {
		log.Printf("Заказ %s НАЙДЕН В КЭШЕ!", orderUID)
		c.JSON(http.StatusOK, order)
		return
	}

	// --- Старая логика: если в кэше нет, идем в БД ---
	log.Printf("Заказ %s не найден в кэше, ищем в БД...", orderUID)

	// Проверяем, что БД инициализирована
	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var order models.Order
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
	kafka.OrderCache.Set(order.OrderUID, &order)

	// Отправляем найденный заказ в виде JSON
	c.JSON(http.StatusOK, order)
}

func main() {
	fmt.Println("Сервис запускается...")

	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем подключение к базе данных
	database.Init(cfg)

	// Инициализируем кэш
	kafka.InitCache(cfg)

	// Создаем репозиторий для работы с базой данных
	dbRepo := repository.NewGormDatabase(database.DB)

	// Загружаем кэш из базы данных при старте
	if err := kafka.LoadCacheFromDB(dbRepo); err != nil {
		log.Printf("Ошибка загрузки кэша: %v", err)
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем Kafka Consumer в отдельной горутине
	go kafka.StartConsumer(cfg, ctx)

	r := gin.Default()

	// Добавляем маршрут для получения заказа
	r.GET("/order/:order_uid", getOrder)

	// Добавляем health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Добавляем маршрут для отдачи нашей веб-страницы
	r.StaticFile("/", "./web/index.html")

	// Создаем HTTP сервер
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Сервер запускается на %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Создаем канал для получения сигналов ОС
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал завершения, начинается graceful shutdown...")

	// Отменяем контекст для остановки consumer
	cancel()

	// Создаем контекст с таймаутом для graceful shutdown сервера
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка graceful shutdown сервера: %v", err)
	} else {
		log.Println("HTTP сервер успешно остановлен")
	}

	// Закрываем соединение с базой данных
	if sqlDB, err := database.DB.DB(); err == nil {
		sqlDB.Close()
		log.Println("Соединение с базой данных закрыто")
	}

	log.Println("Сервис успешно завершен")
}
