package repository

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"wb-service/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Item{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func generateUniqueOrderUID() string {
	return fmt.Sprintf("test_order_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

func createTestOrder() *models.Order {
	return &models.Order{
		OrderUID:          generateUniqueOrderUID(),
		TrackNumber:       "TRACK123",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "customer_123",
		DeliveryService:   "cdek",
		Shardkey:          "1",
		SmID:              100,
		DateCreated:       time.Now(),
		OofShard:          "1",
		Delivery: models.Delivery{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "New York",
			Address: "123 Main St",
			Region:  "NY",
			Email:   "john@example.com",
		},
		Payment: models.Payment{
			Transaction:  "txn_123",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "stripe",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Chase",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      12345,
				TrackNumber: "TRACK123",
				Price:       500,
				Rid:         "rid_123",
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
}

func TestGormDatabase_CreateOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDatabase(db)

	t.Run("create order successfully", func(t *testing.T) {
		order := createTestOrder()

		err := repo.CreateOrder(order)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify order was created
		var count int64
		db.Model(&models.Order{}).Where("order_uid = ?", order.OrderUID).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 order, got %d", count)
		}
	})

	t.Run("create order with duplicate UID should fail", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewGormDatabase(db)

		order1 := createTestOrder()
		order2 := createTestOrder()
		order2.OrderUID = order1.OrderUID // Same UID

		// Create first order
		err := repo.CreateOrder(order1)
		if err != nil {
			t.Errorf("Expected no error for first order, got: %v", err)
		}

		// Try to create second order with same UID
		err = repo.CreateOrder(order2)
		if err == nil {
			t.Error("Expected error for duplicate UID, got nil")
		}
	})
}

func TestGormDatabase_GetOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDatabase(db)

	t.Run("get existing order", func(t *testing.T) {
		// Create test order
		order := createTestOrder()
		err := repo.CreateOrder(order)
		if err != nil {
			t.Fatalf("Failed to create test order: %v", err)
		}

		// Get the order
		retrieved, err := repo.GetOrder(order.OrderUID)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected order, got nil")
		}

		if retrieved.OrderUID != order.OrderUID {
			t.Errorf("Expected OrderUID %s, got %s", order.OrderUID, retrieved.OrderUID)
		}

		if retrieved.TrackNumber != order.TrackNumber {
			t.Errorf("Expected TrackNumber %s, got %s", order.TrackNumber, retrieved.TrackNumber)
		}

		// Check relations are loaded
		if retrieved.Delivery.Name != order.Delivery.Name {
			t.Errorf("Expected delivery name %s, got %s", order.Delivery.Name, retrieved.Delivery.Name)
		}

		if retrieved.Payment.Amount != order.Payment.Amount {
			t.Errorf("Expected payment amount %d, got %d", order.Payment.Amount, retrieved.Payment.Amount)
		}

		if len(retrieved.Items) != len(order.Items) {
			t.Errorf("Expected %d items, got %d", len(order.Items), len(retrieved.Items))
		}
	})

	t.Run("get non-existing order", func(t *testing.T) {
		_, err := repo.GetOrder("non_existing_order")
		if err == nil {
			t.Error("Expected error for non-existing order, got nil")
		}
	})

	t.Run("get order with empty UID", func(t *testing.T) {
		_, err := repo.GetOrder("")
		if err == nil {
			t.Error("Expected error for empty UID, got nil")
		}
	})
}

func TestGormDatabase_GetAllOrders(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDatabase(db)

	t.Run("get all orders when none exist", func(t *testing.T) {
		orders, err := repo.GetAllOrders()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(orders) != 0 {
			t.Errorf("Expected 0 orders, got %d", len(orders))
		}
	})

	t.Run("get all orders with multiple orders", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewGormDatabase(db)

		// Create multiple test orders
		order1 := createTestOrder()
		order2 := createTestOrder()
		order3 := createTestOrder()

		// Insert orders
		err := repo.CreateOrder(order1)
		if err != nil {
			t.Fatalf("Failed to create order1: %v", err)
		}

		err = repo.CreateOrder(order2)
		if err != nil {
			t.Fatalf("Failed to create order2: %v", err)
		}

		err = repo.CreateOrder(order3)
		if err != nil {
			t.Fatalf("Failed to create order3: %v", err)
		}

		// Get all orders
		orders, err := repo.GetAllOrders()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(orders) != 3 {
			t.Errorf("Expected 3 orders, got %d", len(orders))
		}

		// Check that relations are loaded
		for _, order := range orders {
			if order.Delivery.Name == "" {
				t.Error("Expected delivery to be loaded")
			}
			if order.Payment.Amount == 0 {
				t.Error("Expected payment to be loaded")
			}
			if len(order.Items) == 0 {
				t.Error("Expected items to be loaded")
			}
		}
	})
}

func TestGormDatabase_Close(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDatabase(db)

	t.Run("close database connection", func(t *testing.T) {
		err := repo.Close()
		if err != nil {
			t.Errorf("Expected no error when closing database, got: %v", err)
		}

		// Try to use database after close - should fail
		_, err = repo.GetAllOrders()
		if err == nil {
			t.Error("Expected error when using database after close, got nil")
		}
	})
}

func TestNewGormDatabase(t *testing.T) {
	t.Run("create new gorm database instance", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewGormDatabase(db)

		if repo == nil {
			t.Error("Expected repository instance, got nil")
		}

		// Test that it implements the Database interface
		_, ok := repo.(interface {
			CreateOrder(*models.Order) error
			GetOrder(string) (*models.Order, error)
			GetAllOrders() ([]models.Order, error)
			Close() error
		})

		if !ok {
			t.Error("Repository does not implement expected interface")
		}
	})
}

func TestGormDatabase_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormDatabase(db)

	t.Run("full integration test", func(t *testing.T) {
		// 1. Initially no orders
		orders, err := repo.GetAllOrders()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(orders) != 0 {
			t.Errorf("Expected 0 initial orders, got %d", len(orders))
		}

		// 2. Create an order
		order := createTestOrder()
		err = repo.CreateOrder(order)
		if err != nil {
			t.Errorf("Expected no error creating order, got: %v", err)
		}

		// 3. Get the order
		retrieved, err := repo.GetOrder(order.OrderUID)
		if err != nil {
			t.Errorf("Expected no error getting order, got: %v", err)
		}
		if retrieved.OrderUID != order.OrderUID {
			t.Errorf("Retrieved order UID mismatch")
		}

		// 4. Get all orders should return 1
		orders, err = repo.GetAllOrders()
		if err != nil {
			t.Errorf("Expected no error getting all orders, got: %v", err)
		}
		if len(orders) != 1 {
			t.Errorf("Expected 1 order after creation, got %d", len(orders))
		}

		// 5. Create another order
		order2 := createTestOrder()
		err = repo.CreateOrder(order2)
		if err != nil {
			t.Errorf("Expected no error creating second order, got: %v", err)
		}

		// 6. Get all orders should return 2
		orders, err = repo.GetAllOrders()
		if err != nil {
			t.Errorf("Expected no error getting all orders, got: %v", err)
		}
		if len(orders) != 2 {
			t.Errorf("Expected 2 orders after second creation, got %d", len(orders))
		}
	})
}