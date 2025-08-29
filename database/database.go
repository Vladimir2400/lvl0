package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn := "host=127.0.0.1 user=wb_user password=wb_password dbname=wb_db port=5434 sslmode=disable"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных: %v", err)
	}

	fmt.Println("Успешное подключение к базе данных!")
}
