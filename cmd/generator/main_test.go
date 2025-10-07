package main

import (
	"encoding/json"
	"testing"
	"time"
	"wb-service/models"
)

func TestFakeDataGenerator_GenerateOrder(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generate order with all required fields", func(t *testing.T) {
		order := generator.GenerateOrder()

		if order == nil {
			t.Fatal("Expected order, got nil")
		}

		// Test required fields
		if order.OrderUID == "" {
			t.Error("OrderUID should not be empty")
		}

		if order.TrackNumber == "" {
			t.Error("TrackNumber should not be empty")
		}

		if order.Entry == "" {
			t.Error("Entry should not be empty")
		}

		if order.CustomerID == "" {
			t.Error("CustomerID should not be empty")
		}

		if order.DeliveryService == "" {
			t.Error("DeliveryService should not be empty")
		}

		if order.SmID <= 0 {
			t.Error("SmID should be positive")
		}

		// Test that DateCreated is not in the future
		if order.DateCreated.After(time.Now()) {
			t.Error("DateCreated should not be in the future")
		}

		// Test that DateCreated is within last 30 days
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		if order.DateCreated.Before(thirtyDaysAgo) {
			t.Error("DateCreated should be within last 30 days")
		}
	})

	t.Run("generate order with valid delivery", func(t *testing.T) {
		order := generator.GenerateOrder()

		if order.Delivery.Name == "" {
			t.Error("Delivery name should not be empty")
		}

		if order.Delivery.Phone == "" {
			t.Error("Delivery phone should not be empty")
		}

		if order.Delivery.City == "" {
			t.Error("Delivery city should not be empty")
		}

		if order.Delivery.Address == "" {
			t.Error("Delivery address should not be empty")
		}

		if order.Delivery.Region == "" {
			t.Error("Delivery region should not be empty")
		}

		if order.Delivery.Zip == "" {
			t.Error("Delivery zip should not be empty")
		}

		if order.Delivery.Email == "" {
			t.Error("Delivery email should not be empty")
		}
	})

	t.Run("generate order with valid payment", func(t *testing.T) {
		order := generator.GenerateOrder()

		if order.Payment.Transaction == "" {
			t.Error("Payment transaction should not be empty")
		}

		if order.Payment.Currency == "" {
			t.Error("Payment currency should not be empty")
		}

		if order.Payment.Provider == "" {
			t.Error("Payment provider should not be empty")
		}

		if order.Payment.Bank == "" {
			t.Error("Payment bank should not be empty")
		}

		if order.Payment.Amount <= 0 {
			t.Error("Payment amount should be positive")
		}

		if order.Payment.PaymentDt <= 0 {
			t.Error("Payment datetime should be positive")
		}

		if order.Payment.DeliveryCost < 0 {
			t.Error("Delivery cost should not be negative")
		}

		if order.Payment.GoodsTotal < 0 {
			t.Error("Goods total should not be negative")
		}
	})

	t.Run("generate order with valid items", func(t *testing.T) {
		order := generator.GenerateOrder()

		if len(order.Items) == 0 {
			t.Error("Order should have at least one item")
		}

		if len(order.Items) > 3 {
			t.Error("Order should have at most 3 items")
		}

		for i, item := range order.Items {
			if item.ChrtID <= 0 {
				t.Errorf("Item %d ChrtID should be positive", i)
			}

			if item.TrackNumber == "" {
				t.Errorf("Item %d TrackNumber should not be empty", i)
			}

			if item.Price <= 0 {
				t.Errorf("Item %d price should be positive", i)
			}

			if item.Rid == "" {
				t.Errorf("Item %d rid should not be empty", i)
			}

			if item.Name == "" {
				t.Errorf("Item %d name should not be empty", i)
			}

			if item.Size == "" {
				t.Errorf("Item %d size should not be empty", i)
			}

			if item.TotalPrice <= 0 {
				t.Errorf("Item %d total price should be positive", i)
			}

			if item.NmID <= 0 {
				t.Errorf("Item %d NmID should be positive", i)
			}

			if item.Brand == "" {
				t.Errorf("Item %d brand should not be empty", i)
			}

			if item.Sale < 0 || item.Sale > 50 {
				t.Errorf("Item %d sale should be between 0 and 50", i)
			}
		}
	})
}

func TestFakeDataGenerator_OrderUIDGeneration(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generate unique order UIDs", func(t *testing.T) {
		uids := make(map[string]bool)

		for i := 0; i < 100; i++ {
			uid := generator.generateOrderUID()
			if uids[uid] {
				t.Errorf("Duplicate UID generated: %s", uid)
			}
			uids[uid] = true

			// Check format
			if len(uid) != 23 { // 19 random chars + "test"
				t.Errorf("Expected UID length 23, got %d for UID: %s", len(uid), uid)
			}

			if uid[19:] != "test" {
				t.Errorf("Expected UID to end with 'test', got: %s", uid)
			}
		}
	})
}

func TestFakeDataGenerator_PhoneGeneration(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generate valid phone numbers", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			phone := generator.generatePhone()

			if phone == "" {
				t.Error("Phone should not be empty")
			}

			if phone[0] != '+' {
				t.Errorf("Phone should start with '+', got: %s", phone)
			}

			if len(phone) < 8 || len(phone) > 20 {
				t.Errorf("Phone length should be between 8-20 chars, got %d: %s", len(phone), phone)
			}
		}
	})
}

func TestFakeDataGenerator_EmailGeneration(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generate valid email addresses", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			email := generator.generateEmail()

			if email == "" {
				t.Error("Email should not be empty")
			}

			// Basic email format check
			atIndex := -1
			dotIndex := -1
			for j, char := range email {
				if char == '@' {
					atIndex = j
				}
				if char == '.' && j > atIndex {
					dotIndex = j
				}
			}

			if atIndex == -1 {
				t.Errorf("Email should contain '@': %s", email)
			}

			if dotIndex == -1 {
				t.Errorf("Email should contain '.' after '@': %s", email)
			}

			if atIndex == 0 || atIndex == len(email)-1 {
				t.Errorf("Email '@' should not be at start or end: %s", email)
			}
		}
	})
}

func TestFakeDataGenerator_ZipGeneration(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generate valid zip codes", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			zip := generator.generateZip()

			if zip == "" {
				t.Error("Zip should not be empty")
			}

			if len(zip) != 6 {
				t.Errorf("Expected zip length 6, got %d: %s", len(zip), zip)
			}

			// Check all digits
			for _, char := range zip {
				if char < '0' || char > '9' {
					t.Errorf("Zip should contain only digits: %s", zip)
					break
				}
			}
		}
	})
}

func TestFakeDataGenerator_ItemGeneration(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generate items with correct track number", func(t *testing.T) {
		trackNumber := "TEST_TRACK_123"
		items := generator.generateItems(trackNumber, 2)

		if len(items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(items))
		}

		for i, item := range items {
			if item.TrackNumber != trackNumber {
				t.Errorf("Item %d track number mismatch: expected %s, got %s", i, trackNumber, item.TrackNumber)
			}
		}
	})

	t.Run("generate variable number of items", func(t *testing.T) {
		trackNumber := "TEST_TRACK"

		for count := 1; count <= 5; count++ {
			items := generator.generateItems(trackNumber, count)

			if len(items) != count {
				t.Errorf("Expected %d items, got %d", count, len(items))
			}
		}
	})

	t.Run("item sale calculation", func(t *testing.T) {
		trackNumber := "TEST_TRACK"
		items := generator.generateItems(trackNumber, 1)

		item := items[0]
		expectedTotal := item.Price * (100 - item.Sale) / 100

		if item.TotalPrice != expectedTotal {
			t.Errorf("Item total price calculation error: price=%d, sale=%d%%, expected=%d, got=%d",
				item.Price, item.Sale, expectedTotal, item.TotalPrice)
		}
	})
}

func TestFakeDataGenerator_PaymentCalculation(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("payment amounts are consistent", func(t *testing.T) {
		payment := generator.generatePayment("test_order")

		if payment.Amount <= 0 {
			t.Error("Payment amount should be positive")
		}

		if payment.DeliveryCost < 0 {
			t.Error("Delivery cost should not be negative")
		}

		if payment.GoodsTotal < 0 {
			t.Error("Goods total should not be negative")
		}

		if payment.Amount != payment.DeliveryCost + payment.GoodsTotal {
			t.Errorf("Payment amount (%d) should equal delivery cost (%d) + goods total (%d)",
				payment.Amount, payment.DeliveryCost, payment.GoodsTotal)
		}
	})

	t.Run("payment has correct order reference", func(t *testing.T) {
		orderUID := "test_payment_order"
		payment := generator.generatePayment(orderUID)

		if payment.OrderUID != orderUID {
			t.Errorf("Expected payment OrderUID %s, got %s", orderUID, payment.OrderUID)
		}

		if payment.Transaction != orderUID {
			t.Errorf("Expected payment Transaction %s, got %s", orderUID, payment.Transaction)
		}
	})
}

func TestFakeDataGenerator_JSONSerialization(t *testing.T) {
	generator := NewFakeDataGenerator()

	t.Run("generated order serializes to valid JSON", func(t *testing.T) {
		order := generator.GenerateOrder()

		jsonData, err := json.Marshal(order)
		if err != nil {
			t.Errorf("Failed to marshal order to JSON: %v", err)
		}

		// Try to unmarshal back
		var unmarshaled models.Order
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal order from JSON: %v", err)
		}

		// Basic checks
		if unmarshaled.OrderUID != order.OrderUID {
			t.Errorf("JSON round-trip failed for OrderUID")
		}

		if len(unmarshaled.Items) != len(order.Items) {
			t.Errorf("JSON round-trip failed for Items count")
		}
	})
}

func TestNewFakeDataGenerator(t *testing.T) {
	t.Run("create new generator", func(t *testing.T) {
		generator := NewFakeDataGenerator()

		if generator == nil {
			t.Error("Expected generator instance, got nil")
		}

		if generator.rand == nil {
			t.Error("Expected random generator to be initialized")
		}
	})

	t.Run("generators produce different results", func(t *testing.T) {
		generator1 := NewFakeDataGenerator()
		generator2 := NewFakeDataGenerator()

		// Generate orders with both generators
		order1 := generator1.GenerateOrder()
		order2 := generator2.GenerateOrder()

		// They should be different (very high probability)
		if order1.OrderUID == order2.OrderUID {
			t.Error("Different generators produced same OrderUID (very unlikely)")
		}
	})
}