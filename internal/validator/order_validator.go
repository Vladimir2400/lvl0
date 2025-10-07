package validator

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"wb-service/internal/interfaces"
	"wb-service/models"
)

// OrderValidator реализует интерфейс OrderValidator
type OrderValidator struct{}

// NewOrderValidator создает новый валидатор заказов
func NewOrderValidator() interfaces.OrderValidator {
	return &OrderValidator{}
}

// Validate валидирует заказ
func (v *OrderValidator) Validate(order *models.Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	if err := v.validateOrder(order); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := v.validateDelivery(&order.Delivery); err != nil {
		return fmt.Errorf("delivery validation failed: %w", err)
	}

	if err := v.validatePayment(&order.Payment); err != nil {
		return fmt.Errorf("payment validation failed: %w", err)
	}

	if err := v.validateItems(order.Items); err != nil {
		return fmt.Errorf("items validation failed: %w", err)
	}

	return nil
}

// validateOrder валидирует основные поля заказа
func (v *OrderValidator) validateOrder(order *models.Order) error {
	if order.OrderUID == "" {
		return errors.New("order_uid is required")
	}

	if len(order.OrderUID) < 5 {
		return errors.New("order_uid is too short")
	}

	if len(order.OrderUID) > 100 {
		return errors.New("order_uid is too long")
	}

	if order.TrackNumber == "" {
		return errors.New("track_number is required")
	}

	if order.Entry == "" {
		return errors.New("entry is required")
	}

	if order.CustomerID == "" {
		return errors.New("customer_id is required")
	}

	if order.DeliveryService == "" {
		return errors.New("delivery_service is required")
	}

	if order.SmID <= 0 {
		return errors.New("sm_id must be positive")
	}

	// Проверяем, что дата создания не в будущем
	if order.DateCreated.After(time.Now()) {
		return errors.New("date_created cannot be in the future")
	}

	// Проверяем, что дата создания не слишком старая (например, не более 10 лет назад)
	tenYearsAgo := time.Now().AddDate(-10, 0, 0)
	if order.DateCreated.Before(tenYearsAgo) {
		return errors.New("date_created is too old")
	}

	// Проверяем locale
	validLocales := []string{"en", "ru", "fr", "de", "es"}
	localeValid := false
	for _, validLocale := range validLocales {
		if order.Locale == validLocale {
			localeValid = true
			break
		}
	}
	if !localeValid {
		return errors.New("invalid locale")
	}

	return nil
}

// validateDelivery валидирует информацию о доставке
func (v *OrderValidator) validateDelivery(delivery *models.Delivery) error {
	if delivery.Name == "" {
		return errors.New("delivery name is required")
	}

	if len(delivery.Name) < 2 {
		return errors.New("delivery name is too short")
	}

	if delivery.Phone == "" {
		return errors.New("delivery phone is required")
	}

	// Строгая валидация телефона (должен начинаться с + и содержать достаточно цифр)
	phoneRegex := regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
	if !phoneRegex.MatchString(delivery.Phone) {
		return errors.New("invalid phone format")
	}

	if delivery.Email != "" {
		if _, err := mail.ParseAddress(delivery.Email); err != nil {
			return fmt.Errorf("invalid email format: %w", err)
		}
	}

	if delivery.City == "" {
		return errors.New("delivery city is required")
	}

	if delivery.Address == "" {
		return errors.New("delivery address is required")
	}

	if delivery.Region == "" {
		return errors.New("delivery region is required")
	}

	if delivery.Zip == "" {
		return errors.New("delivery zip is required")
	}

	// Простая валидация почтового индекса (только цифры)
	zipRegex := regexp.MustCompile(`^[0-9]{5,10}$`)
	if !zipRegex.MatchString(delivery.Zip) {
		return errors.New("invalid zip format")
	}

	return nil
}

// validatePayment валидирует информацию об оплате
func (v *OrderValidator) validatePayment(payment *models.Payment) error {
	if payment.Transaction == "" {
		return errors.New("payment transaction is required")
	}

	if payment.Currency == "" {
		return errors.New("payment currency is required")
	}

	// Проверяем, что валюта - это трехбуквенный код
	if len(payment.Currency) != 3 {
		return errors.New("currency must be 3 characters long")
	}

	payment.Currency = strings.ToUpper(payment.Currency)

	if payment.Provider == "" {
		return errors.New("payment provider is required")
	}

	if payment.Amount <= 0 {
		return errors.New("payment amount must be positive")
	}

	if payment.DeliveryCost < 0 {
		return errors.New("delivery cost cannot be negative")
	}

	if payment.GoodsTotal < 0 {
		return errors.New("goods total cannot be negative")
	}

	if payment.CustomFee < 0 {
		return errors.New("custom fee cannot be negative")
	}

	if payment.Bank == "" {
		return errors.New("payment bank is required")
	}

	if payment.PaymentDt <= 0 {
		return errors.New("payment_dt must be positive")
	}

	return nil
}

// validateItems валидирует товары в заказе
func (v *OrderValidator) validateItems(items []models.Item) error {
	if len(items) == 0 {
		return errors.New("order must contain at least one item")
	}

	for i, item := range items {
		if err := v.validateItem(&item); err != nil {
			return fmt.Errorf("item %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateItem валидирует отдельный товар
func (v *OrderValidator) validateItem(item *models.Item) error {
	if item.ChrtID <= 0 {
		return errors.New("chrt_id must be positive")
	}

	if item.TrackNumber == "" {
		return errors.New("item track_number is required")
	}

	if item.Price <= 0 {
		return errors.New("item price must be positive")
	}

	if item.Rid == "" {
		return errors.New("item rid is required")
	}

	if item.Name == "" {
		return errors.New("item name is required")
	}

	if item.Sale < 0 || item.Sale > 100 {
		return errors.New("item sale must be between 0 and 100")
	}

	if item.Size == "" {
		return errors.New("item size is required")
	}

	if item.TotalPrice <= 0 {
		return errors.New("item total_price must be positive")
	}

	if item.NmID <= 0 {
		return errors.New("item nm_id must be positive")
	}

	if item.Brand == "" {
		return errors.New("item brand is required")
	}

	if item.Status < 100 {
		return errors.New("item status must be at least 100")
	}

	return nil
}