package validator

import (
	"testing"
	"time"
	"wb-service/models"
)

func TestOrderValidator_Validate(t *testing.T) {
	validator := NewOrderValidator()

	t.Run("valid order", func(t *testing.T) {
		order := createValidOrder()
		err := validator.Validate(order)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("empty order_uid", func(t *testing.T) {
		order := createValidOrder()
		order.OrderUID = ""
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for empty order_uid")
		}
	})

	t.Run("short order_uid", func(t *testing.T) {
		order := createValidOrder()
		order.OrderUID = "123"
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for short order_uid")
		}
	})

	t.Run("empty track_number", func(t *testing.T) {
		order := createValidOrder()
		order.TrackNumber = ""
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for empty track_number")
		}
	})

	t.Run("invalid sm_id", func(t *testing.T) {
		order := createValidOrder()
		order.SmID = 0
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for invalid sm_id")
		}
	})

	t.Run("future date", func(t *testing.T) {
		order := createValidOrder()
		order.DateCreated = time.Now().Add(24 * time.Hour)
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for future date")
		}
	})
}

func TestOrderValidator_ValidateDelivery(t *testing.T) {
	validator := &OrderValidator{}

	t.Run("valid delivery", func(t *testing.T) {
		delivery := createValidDelivery()
		err := validator.validateDelivery(delivery)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Name = ""
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for empty name")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Email = "invalid-email"
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for invalid email")
		}
	})

	t.Run("invalid zip", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Zip = "abc"
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for invalid zip")
		}
	})
}

func TestOrderValidator_ValidatePayment(t *testing.T) {
	validator := &OrderValidator{}

	t.Run("valid payment", func(t *testing.T) {
		payment := createValidPayment()
		err := validator.validatePayment(payment)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("empty transaction", func(t *testing.T) {
		payment := createValidPayment()
		payment.Transaction = ""
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for empty transaction")
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		payment := createValidPayment()
		payment.Amount = -100
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for negative amount")
		}
	})

	t.Run("invalid currency length", func(t *testing.T) {
		payment := createValidPayment()
		payment.Currency = "US"
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for invalid currency length")
		}
	})
}

func TestOrderValidator_ValidateItems(t *testing.T) {
	validator := &OrderValidator{}

	t.Run("valid items", func(t *testing.T) {
		items := []models.Item{createValidItem()}
		err := validator.validateItems(items)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		items := []models.Item{}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for empty items")
		}
	})

	t.Run("invalid sale", func(t *testing.T) {
		item := createValidItem()
		item.Sale = 150
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for invalid sale")
		}
	})

	t.Run("negative sale", func(t *testing.T) {
		item := createValidItem()
		item.Sale = -10
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for negative sale")
		}
	})

	t.Run("zero price", func(t *testing.T) {
		item := createValidItem()
		item.Price = 0
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for zero price")
		}
	})

	t.Run("empty item name", func(t *testing.T) {
		item := createValidItem()
		item.Name = ""
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for empty item name")
		}
	})

	t.Run("empty track number", func(t *testing.T) {
		item := createValidItem()
		item.TrackNumber = ""
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for empty track number")
		}
	})
}

func TestOrderValidator_EdgeCases(t *testing.T) {
	validator := NewOrderValidator()

	t.Run("nil order", func(t *testing.T) {
		err := validator.Validate(nil)
		if err == nil {
			t.Error("Expected error for nil order")
		}
	})

	t.Run("very long order_uid", func(t *testing.T) {
		order := createValidOrder()
		order.OrderUID = "this_is_a_very_long_order_uid_that_exceeds_normal_limits_and_should_be_rejected_by_validator_because_it_is_too_long_for_practical_use"
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for very long order_uid")
		}
	})

	t.Run("invalid locale", func(t *testing.T) {
		order := createValidOrder()
		order.Locale = "invalid"
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for invalid locale")
		}
	})

	t.Run("empty entry", func(t *testing.T) {
		order := createValidOrder()
		order.Entry = ""
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for empty entry")
		}
	})

	t.Run("empty customer_id", func(t *testing.T) {
		order := createValidOrder()
		order.CustomerID = ""
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for empty customer_id")
		}
	})

	t.Run("empty delivery_service", func(t *testing.T) {
		order := createValidOrder()
		order.DeliveryService = ""
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for empty delivery_service")
		}
	})

	t.Run("date too old", func(t *testing.T) {
		order := createValidOrder()
		order.DateCreated = time.Now().AddDate(-15, 0, 0) // 15 years ago
		err := validator.Validate(order)
		if err == nil {
			t.Error("Expected error for very old date")
		}
	})
}

func TestOrderValidator_DeliveryEdgeCases(t *testing.T) {
	validator := &OrderValidator{}

	t.Run("invalid phone format", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Phone = "123456789"
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for invalid phone format")
		}
	})

	t.Run("empty city", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.City = ""
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for empty city")
		}
	})

	t.Run("empty address", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Address = ""
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for empty address")
		}
	})

	t.Run("empty region", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Region = ""
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for empty region")
		}
	})

	t.Run("invalid email domain", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Email = "test@"
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for invalid email domain")
		}
	})

	t.Run("email without @ symbol", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Email = "testgmail.com"
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for email without @ symbol")
		}
	})

	t.Run("short name", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Name = "A"
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for short name")
		}
	})

	t.Run("empty zip", func(t *testing.T) {
		delivery := createValidDelivery()
		delivery.Zip = ""
		err := validator.validateDelivery(delivery)
		if err == nil {
			t.Error("Expected error for empty zip")
		}
	})
}

func TestOrderValidator_PaymentEdgeCases(t *testing.T) {
	validator := &OrderValidator{}

	t.Run("zero amount", func(t *testing.T) {
		payment := createValidPayment()
		payment.Amount = 0
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for zero amount")
		}
	})

	t.Run("empty currency", func(t *testing.T) {
		payment := createValidPayment()
		payment.Currency = ""
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for empty currency")
		}
	})

	t.Run("invalid currency too long", func(t *testing.T) {
		payment := createValidPayment()
		payment.Currency = "USDD"
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for currency too long")
		}
	})

	t.Run("empty provider", func(t *testing.T) {
		payment := createValidPayment()
		payment.Provider = ""
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for empty provider")
		}
	})

	t.Run("empty bank", func(t *testing.T) {
		payment := createValidPayment()
		payment.Bank = ""
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for empty bank")
		}
	})

	t.Run("negative delivery cost", func(t *testing.T) {
		payment := createValidPayment()
		payment.DeliveryCost = -100
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for negative delivery cost")
		}
	})

	t.Run("negative goods total", func(t *testing.T) {
		payment := createValidPayment()
		payment.GoodsTotal = -50
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for negative goods total")
		}
	})

	t.Run("negative payment_dt", func(t *testing.T) {
		payment := createValidPayment()
		payment.PaymentDt = -1
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for negative payment_dt")
		}
	})

	t.Run("zero payment_dt", func(t *testing.T) {
		payment := createValidPayment()
		payment.PaymentDt = 0
		err := validator.validatePayment(payment)
		if err == nil {
			t.Error("Expected error for zero payment_dt")
		}
	})
}

func TestOrderValidator_ItemEdgeCases(t *testing.T) {
	validator := &OrderValidator{}

	t.Run("zero chrt_id", func(t *testing.T) {
		item := createValidItem()
		item.ChrtID = 0
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for zero chrt_id")
		}
	})

	t.Run("negative chrt_id", func(t *testing.T) {
		item := createValidItem()
		item.ChrtID = -1
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for negative chrt_id")
		}
	})

	t.Run("zero nm_id", func(t *testing.T) {
		item := createValidItem()
		item.NmID = 0
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for zero nm_id")
		}
	})

	t.Run("empty rid", func(t *testing.T) {
		item := createValidItem()
		item.Rid = ""
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for empty rid")
		}
	})

	t.Run("empty size", func(t *testing.T) {
		item := createValidItem()
		item.Size = ""
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for empty size")
		}
	})

	t.Run("empty brand", func(t *testing.T) {
		item := createValidItem()
		item.Brand = ""
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for empty brand")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		item := createValidItem()
		item.Status = -1
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for invalid status")
		}
	})

	t.Run("zero total_price", func(t *testing.T) {
		item := createValidItem()
		item.TotalPrice = 0
		items := []models.Item{item}
		err := validator.validateItems(items)
		if err == nil {
			t.Error("Expected error for zero total_price")
		}
	})
}

// Helper functions to create valid test data

func createValidOrder() *models.Order {
	return &models.Order{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Delivery:          *createValidDelivery(),
		Payment:           *createValidPayment(),
		Items:             []models.Item{createValidItem()},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now().Add(-24 * time.Hour),
		OofShard:          "1",
	}
}

func createValidDelivery() *models.Delivery {
	return &models.Delivery{
		Name:    "Test Testov",
		Phone:   "+9720000000",
		Zip:     "2639809",
		City:    "Kiryat Mozkin",
		Address: "Ploshad Mira 15",
		Region:  "Kraiot",
		Email:   "test@gmail.com",
	}
}

func createValidPayment() *models.Payment {
	return &models.Payment{
		Transaction:  "b563feb7b2b84b6test",
		RequestID:    "",
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       1817,
		PaymentDt:    1637907727,
		Bank:         "alpha",
		DeliveryCost: 1500,
		GoodsTotal:   317,
		CustomFee:    0,
	}
}

func createValidItem() models.Item {
	return models.Item{
		ChrtID:      9934930,
		TrackNumber: "WBILMTESTTRACK",
		Price:       453,
		Rid:         "ab4219087a764ae0btest",
		Name:        "Mascaras",
		Sale:        30,
		Size:        "0",
		TotalPrice:  317,
		NmID:        2389232,
		Brand:       "Vivienne Sabo",
		Status:      202,
	}
}