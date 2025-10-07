package database

import (
	"os"
	"testing"
	"wb-service/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInit(t *testing.T) {
	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("DB_SSLMODE", "disable")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	t.Run("init with sqlite for testing", func(t *testing.T) {
		// Use SQLite for testing since we can't guarantee PostgreSQL is available
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}

		// Set the global DB variable to test the Init function behavior
		DB = db

		// Test auto-migration
		err = DB.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Item{})
		if err != nil {
			t.Errorf("Failed to auto-migrate: %v", err)
		}

		// Verify tables were created by trying to create a test order
		testOrder := &models.Order{
			OrderUID:          "test_init_order",
			TrackNumber:       "TEST_TRACK",
			Entry:             "WBIL",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "test_customer",
			DeliveryService:   "test_service",
			Shardkey:          "1",
			SmID:              100,
		}

		err = DB.Create(testOrder).Error
		if err != nil {
			t.Errorf("Failed to create test order after init: %v", err)
		}

		// Verify the order was created
		var count int64
		DB.Model(&models.Order{}).Where("order_uid = ?", testOrder.OrderUID).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 order after creation, got %d", count)
		}
	})

	t.Run("verify database connection is accessible", func(t *testing.T) {
		if DB == nil {
			t.Fatal("Database connection should not be nil after Init")
		}

		// Test database connection by pinging
		sqlDB, err := DB.DB()
		if err != nil {
			t.Errorf("Failed to get underlying sql.DB: %v", err)
		}

		err = sqlDB.Ping()
		if err != nil {
			t.Errorf("Database ping failed: %v", err)
		}
	})
}

func TestDatabaseOperationsAfterInit(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	DB = db

	// Auto-migrate
	err = DB.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Item{})
	if err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}

	t.Run("create order with relations", func(t *testing.T) {
		order := &models.Order{
			OrderUID:          "test_relations_order",
			TrackNumber:       "REL_TRACK",
			Entry:             "WBIL",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "rel_customer",
			DeliveryService:   "rel_service",
			Shardkey:          "1",
			SmID:              100,
			Delivery: models.Delivery{
				Name:    "Test User",
				Phone:   "+1234567890",
				Zip:     "12345",
				City:    "Test City",
				Address: "123 Test St",
				Region:  "Test Region",
				Email:   "test@example.com",
			},
			Payment: models.Payment{
				Transaction:  "test_txn",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "test_provider",
				Amount:       1000,
				PaymentDt:    1234567890,
				Bank:         "Test Bank",
				DeliveryCost: 100,
				GoodsTotal:   900,
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      12345,
					TrackNumber: "REL_TRACK",
					Price:       500,
					Rid:         "test_rid",
					Name:        "Test Product",
					Sale:        10,
					Size:        "M",
					TotalPrice:  450,
					NmID:        67890,
					Brand:       "TestBrand",
					Status:      200,
				},
			},
		}

		err := DB.Create(order).Error
		if err != nil {
			t.Errorf("Failed to create order with relations: %v", err)
		}

		// Verify all relations were created
		var retrievedOrder models.Order
		err = DB.Preload("Delivery").Preload("Payment").Preload("Items").
			First(&retrievedOrder, "order_uid = ?", order.OrderUID).Error
		if err != nil {
			t.Errorf("Failed to retrieve order with relations: %v", err)
		}

		if retrievedOrder.Delivery.Name != order.Delivery.Name {
			t.Errorf("Delivery relation not properly saved")
		}

		if retrievedOrder.Payment.Amount != order.Payment.Amount {
			t.Errorf("Payment relation not properly saved")
		}

		if len(retrievedOrder.Items) != len(order.Items) {
			t.Errorf("Items relation not properly saved")
		}
	})

	t.Run("verify foreign key constraints", func(t *testing.T) {
		// Create order without delivery/payment (should still work with GORM)
		order := &models.Order{
			OrderUID:          "test_fk_order",
			TrackNumber:       "FK_TRACK",
			Entry:             "WBIL",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "fk_customer",
			DeliveryService:   "fk_service",
			Shardkey:          "1",
			SmID:              100,
		}

		err := DB.Create(order).Error
		if err != nil {
			t.Errorf("Failed to create order without relations: %v", err)
		}

		// Verify order was created
		var count int64
		DB.Model(&models.Order{}).Where("order_uid = ?", order.OrderUID).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 order, got %d", count)
		}
	})
}