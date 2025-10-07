package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"wb-service/database"
	"wb-service/internal/cache"
	"wb-service/kafka"
	"wb-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/order/:order_uid", getOrder)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func setupTestCache() {
	kafka.OrderCache = cache.NewLRUCache(100, time.Hour)
}

func setupTestDatabase() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Item{})

	// Set global database for tests
	// Note: In real app, we should inject this dependency
	database.DB = db
}

func createTestOrderForCache() *models.Order {
	return &models.Order{
		OrderUID:          "test_cached_order",
		TrackNumber:       "CACHED_TRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "cached_customer",
		DeliveryService:   "cdek",
		Shardkey:          "1",
		SmID:              100,
		DateCreated:       time.Now(),
		OofShard:          "1",
		Delivery: models.Delivery{
			Name:    "Cached User",
			Phone:   "+1111111111",
			Zip:     "12345",
			City:    "Cache City",
			Address: "123 Cache St",
			Region:  "Cache Region",
			Email:   "cached@example.com",
		},
		Payment: models.Payment{
			Transaction:  "cached_txn",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "cache_pay",
			Amount:       500,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Cache Bank",
			DeliveryCost: 50,
			GoodsTotal:   450,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      54321,
				TrackNumber: "CACHED_TRACK",
				Price:       250,
				Rid:         "cached_rid",
				Name:        "Cached Product",
				Sale:        5,
				Size:        "L",
				TotalPrice:  450,
				NmID:        11111,
				Brand:       "CacheBrand",
				Status:      200,
			},
		},
	}
}

func createTestOrderForDB() *models.Order {
	return &models.Order{
		OrderUID:          "test_db_order",
		TrackNumber:       "DB_TRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "db_customer",
		DeliveryService:   "post",
		Shardkey:          "2",
		SmID:              200,
		DateCreated:       time.Now(),
		OofShard:          "2",
		Delivery: models.Delivery{
			Name:    "DB User",
			Phone:   "+2222222222",
			Zip:     "54321",
			City:    "DB City",
			Address: "456 DB Ave",
			Region:  "DB Region",
			Email:   "db@example.com",
		},
		Payment: models.Payment{
			Transaction:  "db_txn",
			RequestID:    "",
			Currency:     "EUR",
			Provider:     "db_pay",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "DB Bank",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    10,
		},
		Items: []models.Item{
			{
				ChrtID:      98765,
				TrackNumber: "DB_TRACK",
				Price:       500,
				Rid:         "db_rid",
				Name:        "DB Product",
				Sale:        10,
				Size:        "XL",
				TotalPrice:  900,
				NmID:        22222,
				Brand:       "DBBrand",
				Status:      202,
			},
		},
	}
}

func TestGetOrderFromCache(t *testing.T) {
	setupTestCache()
	router := setupTestRouter()

	// Add order to cache
	order := createTestOrderForCache()
	kafka.OrderCache.Set(order.OrderUID, order)

	t.Run("get order from cache successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/order/"+order.OrderUID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Order
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.OrderUID != order.OrderUID {
			t.Errorf("Expected OrderUID %s, got %s", order.OrderUID, response.OrderUID)
		}

		if response.Delivery.Name != order.Delivery.Name {
			t.Errorf("Expected delivery name %s, got %s", order.Delivery.Name, response.Delivery.Name)
		}
	})

	t.Run("get non-existing order from empty cache", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/order/non_existing_order", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 404 since we don't have DB setup in this test
		if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 404 or 500, got %d", w.Code)
		}
	})
}

func TestGetOrderFromDatabase(t *testing.T) {
	setupTestCache() // Empty cache
	setupTestDatabase()
	router := setupTestRouter()

	// Add order to database
	order := createTestOrderForDB()
	err := database.DB.Create(order).Error
	if err != nil {
		t.Fatalf("Failed to create test order in DB: %v", err)
	}

	t.Run("get order from database when not in cache", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/order/"+order.OrderUID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response models.Order
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.OrderUID != order.OrderUID {
			t.Errorf("Expected OrderUID %s, got %s", order.OrderUID, response.OrderUID)
		}

		// Verify order is now in cache
		cachedOrder, found := kafka.OrderCache.Get(order.OrderUID)
		if !found {
			t.Error("Expected order to be cached after DB retrieval")
		} else if cachedOrder.OrderUID != order.OrderUID {
			t.Errorf("Cached order UID mismatch: expected %s, got %s", order.OrderUID, cachedOrder.OrderUID)
		}
	})
}

func TestGetOrderNotFound(t *testing.T) {
	setupTestCache() // Empty cache
	setupTestDatabase()
	router := setupTestRouter()

	t.Run("get non-existing order returns 404", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/order/definitely_not_existing", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusNotFound, w.Code, w.Body.String())
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal error response: %v", err)
		}

		if response["error"] != "record not found" {
			t.Errorf("Expected error 'record not found', got '%s'", response["error"])
		}
	})

	t.Run("get order with empty UID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/order/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 404 because route doesn't match
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for empty UID, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	t.Run("health check returns ok", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal health response: %v", err)
		}

		if response["status"] != "ok" {
			t.Errorf("Expected status 'ok', got '%s'", response["status"])
		}
	})
}

func TestCacheAndDatabaseIntegration(t *testing.T) {
	setupTestCache()
	setupTestDatabase()
	router := setupTestRouter()

	// Create order in database only
	order := createTestOrderForDB()
	order.OrderUID = "integration_test_order"
	err := database.DB.Create(order).Error
	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	t.Run("first request loads from DB to cache", func(t *testing.T) {
		// First request - should load from DB
		req, _ := http.NewRequest("GET", "/order/"+order.OrderUID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify it's now in cache
		_, found := kafka.OrderCache.Get(order.OrderUID)
		if !found {
			t.Error("Expected order to be in cache after first request")
		}
	})

	t.Run("second request loads from cache", func(t *testing.T) {
		// Second request - should load from cache
		req, _ := http.NewRequest("GET", "/order/"+order.OrderUID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Order
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response.OrderUID != order.OrderUID {
			t.Errorf("Expected OrderUID %s, got %s", order.OrderUID, response.OrderUID)
		}
	})
}

func TestOrderParameterExtraction(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{"normal order UID", "/order/abc123", http.StatusNotFound}, // Not found is OK - means parameter was extracted
		{"order UID with special chars", "/order/abc-123_test", http.StatusNotFound},
		{"order UID with dots", "/order/order.123", http.StatusNotFound},
		{"long order UID", "/order/very_long_order_uid_12345678901234567890", http.StatusNotFound},
		{"short order UID", "/order/a", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// We expect 404 (not found) which means the parameter was extracted correctly
			// If the parameter extraction failed, we'd get a different error
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d for URL %s", tc.expectedStatus, w.Code, tc.url)
			}
		})
	}
}

func TestJSONResponseFormat(t *testing.T) {
	setupTestCache()
	router := setupTestRouter()

	// Add order to cache
	order := createTestOrderForCache()
	kafka.OrderCache.Set(order.OrderUID, order)

	t.Run("response has correct JSON structure", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/order/"+order.OrderUID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check Content-Type header
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got '%s'", contentType)
		}

		// Parse JSON to verify structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Response is not valid JSON: %v", err)
		}

		// Check required fields exist
		requiredFields := []string{"order_uid", "track_number", "delivery", "payment", "items"}
		for _, field := range requiredFields {
			if _, exists := response[field]; !exists {
				t.Errorf("Required field '%s' missing from JSON response", field)
			}
		}

		// Check delivery is an object
		if delivery, ok := response["delivery"].(map[string]interface{}); ok {
			if _, exists := delivery["name"]; !exists {
				t.Error("Delivery object should contain 'name' field")
			}
		} else {
			t.Error("Delivery should be an object")
		}

		// Check items is an array
		if items, ok := response["items"].([]interface{}); ok {
			if len(items) == 0 {
				t.Error("Items array should not be empty")
			}
		} else {
			t.Error("Items should be an array")
		}
	})
}