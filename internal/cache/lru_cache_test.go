package cache

import (
	"testing"
	"time"
	"wb-service/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLRUCache_BasicOperations(t *testing.T) {
	cache := NewLRUCache(3, time.Hour) // Capacity 3, TTL 1 hour

	// Test Set and Get
	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	cache.Set("test1", order1)

	if got, found := cache.Get("test1"); !found || got.OrderUID != "test1" {
		t.Errorf("Expected to find order test1")
	}

	// Test cache miss
	if _, found := cache.Get("nonexistent"); found {
		t.Errorf("Expected cache miss for nonexistent key")
	}

	// Test Size
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}
}

func TestLRUCache_CapacityLimit(t *testing.T) {
	cache := NewLRUCache(2, time.Hour) // Capacity 2

	// Add items to exceed capacity
	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	order2 := &models.Order{OrderUID: "test2", TrackNumber: "track2"}
	order3 := &models.Order{OrderUID: "test3", TrackNumber: "track3"}

	cache.Set("test1", order1)
	cache.Set("test2", order2)
	cache.Set("test3", order3)

	// test1 should be evicted (least recently used)
	if _, found := cache.Get("test1"); found {
		t.Errorf("Expected test1 to be evicted")
	}

	// test2 and test3 should still be in cache
	if _, found := cache.Get("test2"); !found {
		t.Errorf("Expected test2 to be in cache")
	}
	if _, found := cache.Get("test3"); !found {
		t.Errorf("Expected test3 to be in cache")
	}

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}
}

func TestLRUCache_LRUEviction(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	order2 := &models.Order{OrderUID: "test2", TrackNumber: "track2"}
	order3 := &models.Order{OrderUID: "test3", TrackNumber: "track3"}

	cache.Set("test1", order1)
	cache.Set("test2", order2)

	// Access test1 to make it more recently used
	cache.Get("test1")

	// Add test3, should evict test2 (least recently used)
	cache.Set("test3", order3)

	if _, found := cache.Get("test2"); found {
		t.Errorf("Expected test2 to be evicted")
	}

	if _, found := cache.Get("test1"); !found {
		t.Errorf("Expected test1 to still be in cache")
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(3, time.Hour)

	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	cache.Set("test1", order1)

	// Update the same key
	order1Updated := &models.Order{OrderUID: "test1", TrackNumber: "updated"}
	cache.Set("test1", order1Updated)

	if got, found := cache.Get("test1"); !found || got.TrackNumber != "updated" {
		t.Errorf("Expected updated order")
	}

	// Size should remain 1
	if cache.Size() != 1 {
		t.Errorf("Expected size 1 after update, got %d", cache.Size())
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(3, 100*time.Millisecond) // Very short TTL for testing

	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	cache.Set("test1", order1)

	// Should be available immediately
	if _, found := cache.Get("test1"); !found {
		t.Errorf("Expected to find order immediately after set")
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	if _, found := cache.Get("test1"); found {
		t.Errorf("Expected order to be expired after TTL")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(3, time.Hour)

	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	order2 := &models.Order{OrderUID: "test2", TrackNumber: "track2"}

	cache.Set("test1", order1)
	cache.Set("test2", order2)

	if cache.Size() != 2 {
		t.Errorf("Expected size 2 before clear, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	if _, found := cache.Get("test1"); found {
		t.Errorf("Expected cache to be empty after clear")
	}
}

func TestLRUCache_ZeroTTL(t *testing.T) {
	cache := NewLRUCache(3, 0) // No TTL

	order1 := &models.Order{OrderUID: "test1", TrackNumber: "track1"}
	cache.Set("test1", order1)

	// Should always be available with zero TTL
	time.Sleep(10 * time.Millisecond)
	if _, found := cache.Get("test1"); !found {
		t.Errorf("Expected order to be available with zero TTL")
	}
}

type mockRepository struct {
	db *gorm.DB
}

func (m *mockRepository) CreateOrder(order *models.Order) error {
	return m.db.Create(order).Error
}

func (m *mockRepository) GetOrder(orderUID string) (*models.Order, error) {
	var order models.Order
	err := m.db.Preload("Delivery").Preload("Payment").Preload("Items").
		First(&order, "order_uid = ?", orderUID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (m *mockRepository) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	err := m.db.Preload("Delivery").Preload("Payment").Preload("Items").
		Find(&orders).Error
	return orders, err
}

func (m *mockRepository) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func setupTestDatabase() *mockRepository {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Item{})

	return &mockRepository{db: db}
}

func createTestOrderForLoadFromDB(uid string) *models.Order {
	return &models.Order{
		OrderUID:          uid,
		TrackNumber:       "LOAD_TRACK_" + uid,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "customer_" + uid,
		DeliveryService:   "test_service",
		Shardkey:          "1",
		SmID:              100,
		DateCreated:       time.Now(),
		OofShard:          "1",
		Delivery: models.Delivery{
			Name:    "Test User " + uid,
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "Test City",
			Address: "123 Test St",
			Region:  "Test Region",
			Email:   "test" + uid + "@example.com",
		},
		Payment: models.Payment{
			Transaction:  "txn_" + uid,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "test_provider",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Test Bank",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      12345,
				TrackNumber: "LOAD_TRACK_" + uid,
				Price:       500,
				Rid:         "rid_" + uid,
				Name:        "Test Product " + uid,
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

func TestLRUCache_LoadFromDB(t *testing.T) {
	db := setupTestDatabase()
	cache := NewLRUCache(10, time.Hour)

	t.Run("load orders from database", func(t *testing.T) {
		// Create test orders in database
		order1 := createTestOrderForLoadFromDB("load1")
		order2 := createTestOrderForLoadFromDB("load2")

		err := db.CreateOrder(order1)
		if err != nil {
			t.Fatalf("Failed to create test order 1: %v", err)
		}

		err = db.CreateOrder(order2)
		if err != nil {
			t.Fatalf("Failed to create test order 2: %v", err)
		}

		// Load from database
		cache.LoadFromDB(db)

		// Verify orders are in cache
		cached1, found1 := cache.Get(order1.OrderUID)
		if !found1 {
			t.Error("Expected order1 to be in cache")
		} else if cached1.OrderUID != order1.OrderUID {
			t.Errorf("Cached order1 UID mismatch")
		}

		cached2, found2 := cache.Get(order2.OrderUID)
		if !found2 {
			t.Error("Expected order2 to be in cache")
		} else if cached2.OrderUID != order2.OrderUID {
			t.Errorf("Cached order2 UID mismatch")
		}

		// Verify cache size
		expectedSize := 2
		if cache.Size() != expectedSize {
			t.Errorf("Expected cache size %d, got %d", expectedSize, cache.Size())
		}
	})

	t.Run("load from empty database", func(t *testing.T) {
		emptyDB := setupTestDatabase()
		emptyCache := NewLRUCache(10, time.Hour)

		emptyCache.LoadFromDB(emptyDB)

		// Cache should be empty
		if emptyCache.Size() != 0 {
			t.Errorf("Expected empty cache after loading from empty DB, got size %d", emptyCache.Size())
		}
	})

	t.Run("load with relations intact", func(t *testing.T) {
		db := setupTestDatabase()
		cache := NewLRUCache(10, time.Hour)

		// Create order with full relations
		order := createTestOrderForLoadFromDB("relations")

		err := db.CreateOrder(order)
		if err != nil {
			t.Fatalf("Failed to create test order: %v", err)
		}

		// Load from database
		cache.LoadFromDB(db)

		// Get from cache and verify relations
		cached, found := cache.Get(order.OrderUID)
		if !found {
			t.Fatal("Expected order to be in cache")
		}

		// Check delivery relation
		if cached.Delivery.Name != order.Delivery.Name {
			t.Error("Delivery relation not properly loaded")
		}

		// Check payment relation
		if cached.Payment.Amount != order.Payment.Amount {
			t.Error("Payment relation not properly loaded")
		}

		// Check items relation
		if len(cached.Items) != len(order.Items) {
			t.Error("Items relation not properly loaded")
		}

		if len(cached.Items) > 0 {
			if cached.Items[0].Name != order.Items[0].Name {
				t.Error("Item details not properly loaded")
			}
		}
	})

	t.Run("load respects cache capacity", func(t *testing.T) {
		db := setupTestDatabase()
		smallCache := NewLRUCache(2, time.Hour) // Small capacity

		// Create 3 orders in database
		order1 := createTestOrderForLoadFromDB("cap1")
		order2 := createTestOrderForLoadFromDB("cap2")
		order3 := createTestOrderForLoadFromDB("cap3")

		db.CreateOrder(order1)
		db.CreateOrder(order2)
		db.CreateOrder(order3)

		// Load from database
		smallCache.LoadFromDB(db)

		// Cache should contain only 2 items (capacity limit)
		if smallCache.Size() > 2 {
			t.Errorf("Expected cache size <= 2, got %d", smallCache.Size())
		}

		// At least some orders should be loaded
		if smallCache.Size() == 0 {
			t.Error("Expected some orders to be loaded despite capacity limit")
		}
	})

	t.Run("load error handling", func(t *testing.T) {
		cache := NewLRUCache(10, time.Hour)

		// Test with nil database - should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("LoadFromDB should not panic with nil DB: %v", r)
			}
		}()

		cache.LoadFromDB(nil)

		// Cache should remain empty
		if cache.Size() != 0 {
			t.Errorf("Expected cache to remain empty after nil DB, got size %d", cache.Size())
		}
	})
}