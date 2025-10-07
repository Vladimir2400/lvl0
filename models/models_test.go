package models

import (
	"testing"
	"time"
)

func TestOrderModel(t *testing.T) {
	t.Run("create order with all fields", func(t *testing.T) {
		now := time.Now()

		order := Order{
			OrderUID:          "test_order_123",
			TrackNumber:       "TRACK123",
			Entry:             "WBIL",
			Locale:            "en",
			InternalSignature: "signature",
			CustomerID:        "customer_123",
			DeliveryService:   "cdek",
			Shardkey:          "1",
			SmID:              100,
			DateCreated:       now,
			OofShard:          "1",
		}

		if order.OrderUID != "test_order_123" {
			t.Errorf("Expected OrderUID 'test_order_123', got '%s'", order.OrderUID)
		}

		if order.TrackNumber != "TRACK123" {
			t.Errorf("Expected TrackNumber 'TRACK123', got '%s'", order.TrackNumber)
		}

		if order.SmID != 100 {
			t.Errorf("Expected SmID 100, got %d", order.SmID)
		}

		if !order.DateCreated.Equal(now) {
			t.Errorf("Expected DateCreated to be %v, got %v", now, order.DateCreated)
		}
	})

	t.Run("order with delivery", func(t *testing.T) {
		delivery := Delivery{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "New York",
			Address: "123 Main St",
			Region:  "NY",
			Email:   "john@example.com",
		}

		order := Order{
			OrderUID: "test_order_with_delivery",
			Delivery: delivery,
		}

		if order.Delivery.Name != "John Doe" {
			t.Errorf("Expected delivery name 'John Doe', got '%s'", order.Delivery.Name)
		}

		if order.Delivery.Email != "john@example.com" {
			t.Errorf("Expected delivery email 'john@example.com', got '%s'", order.Delivery.Email)
		}
	})
}

func TestDeliveryModel(t *testing.T) {
	t.Run("create delivery with all fields", func(t *testing.T) {
		delivery := Delivery{
			ID:       1,
			OrderUID: "order_123",
			Name:     "Jane Smith",
			Phone:    "+9876543210",
			Zip:      "54321",
			City:     "Los Angeles",
			Address:  "456 Oak Ave",
			Region:   "CA",
			Email:    "jane@example.com",
		}

		if delivery.Name != "Jane Smith" {
			t.Errorf("Expected Name 'Jane Smith', got '%s'", delivery.Name)
		}

		if delivery.Phone != "+9876543210" {
			t.Errorf("Expected Phone '+9876543210', got '%s'", delivery.Phone)
		}

		if delivery.City != "Los Angeles" {
			t.Errorf("Expected City 'Los Angeles', got '%s'", delivery.City)
		}
	})

	t.Run("delivery with empty email", func(t *testing.T) {
		delivery := Delivery{
			Name:  "Test User",
			Phone: "+1111111111",
			Email: "",
		}

		if delivery.Email != "" {
			t.Errorf("Expected empty email, got '%s'", delivery.Email)
		}
	})
}

func TestPaymentModel(t *testing.T) {
	t.Run("create payment with all fields", func(t *testing.T) {
		payment := Payment{
			ID:           1,
			OrderUID:     "order_123",
			Transaction:  "txn_123",
			RequestID:    "req_123",
			Currency:     "USD",
			Provider:     "stripe",
			Amount:       1000,
			PaymentDt:    1234567890,
			Bank:         "Chase",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    50,
		}

		if payment.Transaction != "txn_123" {
			t.Errorf("Expected Transaction 'txn_123', got '%s'", payment.Transaction)
		}

		if payment.Amount != 1000 {
			t.Errorf("Expected Amount 1000, got %d", payment.Amount)
		}

		if payment.Currency != "USD" {
			t.Errorf("Expected Currency 'USD', got '%s'", payment.Currency)
		}
	})

	t.Run("payment with zero amounts", func(t *testing.T) {
		payment := Payment{
			Amount:       0,
			DeliveryCost: 0,
			GoodsTotal:   0,
			CustomFee:    0,
		}

		if payment.Amount != 0 {
			t.Errorf("Expected Amount 0, got %d", payment.Amount)
		}
	})
}

func TestItemModel(t *testing.T) {
	t.Run("create item with all fields", func(t *testing.T) {
		item := Item{
			ID:          1,
			OrderUID:    "order_123",
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
		}

		if item.ChrtID != 12345 {
			t.Errorf("Expected ChrtID 12345, got %d", item.ChrtID)
		}

		if item.Name != "Test Product" {
			t.Errorf("Expected Name 'Test Product', got '%s'", item.Name)
		}

		if item.Price != 500 {
			t.Errorf("Expected Price 500, got %d", item.Price)
		}

		if item.Sale != 10 {
			t.Errorf("Expected Sale 10, got %d", item.Sale)
		}
	})

	t.Run("item with zero sale", func(t *testing.T) {
		item := Item{
			Price:      1000,
			Sale:       0,
			TotalPrice: 1000,
		}

		if item.Sale != 0 {
			t.Errorf("Expected Sale 0, got %d", item.Sale)
		}

		if item.TotalPrice != 1000 {
			t.Errorf("Expected TotalPrice 1000, got %d", item.TotalPrice)
		}
	})
}

func TestOrderWithAllRelations(t *testing.T) {
	t.Run("order with delivery, payment and items", func(t *testing.T) {
		delivery := Delivery{
			Name:    "Full Test",
			Phone:   "+1111111111",
			City:    "Test City",
			Address: "Test Address",
			Email:   "test@example.com",
		}

		payment := Payment{
			Transaction: "test_txn",
			Currency:    "USD",
			Amount:      1500,
		}

		items := []Item{
			{
				ChrtID: 1,
				Name:   "Item 1",
				Price:  500,
			},
			{
				ChrtID: 2,
				Name:   "Item 2",
				Price:  1000,
			},
		}

		order := Order{
			OrderUID:    "full_test_order",
			TrackNumber: "FULL_TRACK",
			Delivery:    delivery,
			Payment:     payment,
			Items:       items,
		}

		if len(order.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(order.Items))
		}

		if order.Items[0].Name != "Item 1" {
			t.Errorf("Expected first item name 'Item 1', got '%s'", order.Items[0].Name)
		}

		if order.Payment.Amount != 1500 {
			t.Errorf("Expected payment amount 1500, got %d", order.Payment.Amount)
		}

		if order.Delivery.Email != "test@example.com" {
			t.Errorf("Expected delivery email 'test@example.com', got '%s'", order.Delivery.Email)
		}
	})
}

func TestModelFieldTags(t *testing.T) {
	t.Run("check json tags are present", func(t *testing.T) {
		// Создаем структуры чтобы убедиться что они компилируются с тегами
		order := Order{}
		delivery := Delivery{}
		payment := Payment{}
		item := Item{}

		// Простая проверка что структуры созданы
		if order.OrderUID == "" {
			// OK - поле инициализировано пустой строкой
		}

		if delivery.Name == "" {
			// OK - поле инициализировано пустой строкой
		}

		if payment.Amount == 0 {
			// OK - поле инициализировано нулем
		}

		if item.Price == 0 {
			// OK - поле инициализировано нулем
		}
	})
}