package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"wb-service/models"

	"github.com/segmentio/kafka-go"
)

// FakeDataGenerator генерирует фейковые данные для тестирования
type FakeDataGenerator struct {
	rand *rand.Rand
}

// NewFakeDataGenerator создает новый генератор фейковых данных
func NewFakeDataGenerator() *FakeDataGenerator {
	return &FakeDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateOrder генерирует фейковый заказ
func (g *FakeDataGenerator) GenerateOrder() *models.Order {
	orderUID := g.generateOrderUID()
	trackNumber := g.generateTrackNumber()

	order := &models.Order{
		OrderUID:          orderUID,
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Delivery:          g.generateDelivery(orderUID),
		Payment:           g.generatePayment(orderUID),
		Items:             g.generateItems(trackNumber, g.rand.Intn(3)+1), // 1-3 товара
		Locale:            g.generateLocale(),
		InternalSignature: "",
		CustomerID:        g.generateCustomerID(),
		DeliveryService:   g.generateDeliveryService(),
		Shardkey:          strconv.Itoa(g.rand.Intn(10)),
		SmID:              g.rand.Intn(100) + 1,
		DateCreated:       g.generateDateCreated(),
		OofShard:          strconv.Itoa(g.rand.Intn(5)),
	}

	return order
}

func (g *FakeDataGenerator) generateOrderUID() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	uid := make([]byte, 19)
	for i := range uid {
		uid[i] = chars[g.rand.Intn(len(chars))]
	}
	return string(uid) + "test"
}

func (g *FakeDataGenerator) generateTrackNumber() string {
	return "WBILMTESTTRACK"
}

func (g *FakeDataGenerator) generateDelivery(orderUID string) models.Delivery {
	names := []string{"Ivan Ivanov", "Petr Petrov", "Anna Sidorova", "Maria Komarova", "Test Testov"}
	cities := []string{"Moscow", "Saint Petersburg", "Kiryat Mozkin", "Kazan", "Novosibirsk"}
	regions := []string{"Moscow Region", "Leningrad Region", "Kraiot", "Tatarstan", "Novosibirsk Region"}
	addresses := []string{"Ploshad Mira 15", "Lenina 10", "Pushkina 5", "Gagarina 20", "Sovetskaya 1"}

	return models.Delivery{
		OrderUID: orderUID,
		Name:     names[g.rand.Intn(len(names))],
		Phone:    g.generatePhone(),
		Zip:      g.generateZip(),
		City:     cities[g.rand.Intn(len(cities))],
		Address:  addresses[g.rand.Intn(len(addresses))],
		Region:   regions[g.rand.Intn(len(regions))],
		Email:    g.generateEmail(),
	}
}

func (g *FakeDataGenerator) generatePayment(orderUID string) models.Payment {
	currencies := []string{"USD", "RUB", "EUR"}
	providers := []string{"wbpay", "sberbank", "tinkoff", "yandex"}
	banks := []string{"alpha", "sberbank", "tinkoff", "vtb"}

	amount := g.rand.Intn(5000) + 100       // 100-5100
	deliveryCost := g.rand.Intn(2000) + 100 // 100-2100
	goodsTotal := amount - deliveryCost
	if goodsTotal < 0 {
		goodsTotal = amount / 2
		deliveryCost = amount - goodsTotal
	}

	return models.Payment{
		OrderUID:     orderUID,
		Transaction:  orderUID,
		RequestID:    "",
		Currency:     currencies[g.rand.Intn(len(currencies))],
		Provider:     providers[g.rand.Intn(len(providers))],
		Amount:       amount,
		PaymentDt:    time.Now().Unix(),
		Bank:         banks[g.rand.Intn(len(banks))],
		DeliveryCost: deliveryCost,
		GoodsTotal:   goodsTotal,
		CustomFee:    g.rand.Intn(100),
	}
}

func (g *FakeDataGenerator) generateItems(trackNumber string, count int) []models.Item {
	items := make([]models.Item, count)
	brands := []string{"Nike", "Adidas", "Vivienne Sabo", "Apple", "Samsung"}
	names := []string{"T-shirt", "Sneakers", "Mascaras", "Phone", "Laptop"}
	sizes := []string{"XS", "S", "M", "L", "XL", "0"}

	for i := 0; i < count; i++ {
		price := g.rand.Intn(1000) + 50 // 50-1050
		sale := g.rand.Intn(50)         // 0-50%
		totalPrice := price * (100 - sale) / 100

		items[i] = models.Item{
			ChrtID:      g.rand.Intn(10000000) + 1000000,
			TrackNumber: trackNumber,
			Price:       price,
			Rid:         g.generateRid(),
			Name:        names[g.rand.Intn(len(names))],
			Sale:        sale,
			Size:        sizes[g.rand.Intn(len(sizes))],
			TotalPrice:  totalPrice,
			NmID:        g.rand.Intn(10000000) + 1000000,
			Brand:       brands[g.rand.Intn(len(brands))],
			Status:      g.generateStatus(),
		}
	}

	return items
}

func (g *FakeDataGenerator) generatePhone() string {
	prefixes := []string{"+7", "+1", "+44", "+49", "+33"}
	prefix := prefixes[g.rand.Intn(len(prefixes))]
	number := fmt.Sprintf("%d", g.rand.Int63n(9000000000)+1000000000)
	return prefix + number[:10]
}

func (g *FakeDataGenerator) generateZip() string {
	return fmt.Sprintf("%d", g.rand.Intn(900000)+100000)
}

func (g *FakeDataGenerator) generateEmail() string {
	domains := []string{"gmail.com", "yandex.ru", "mail.ru", "example.com"}
	usernames := []string{"test", "user", "admin", "client", "customer"}

	username := usernames[g.rand.Intn(len(usernames))]
	domain := domains[g.rand.Intn(len(domains))]
	number := g.rand.Intn(1000)

	return fmt.Sprintf("%s%d@%s", username, number, domain)
}

func (g *FakeDataGenerator) generateLocale() string {
	locales := []string{"en", "ru", "de", "fr", "es"}
	return locales[g.rand.Intn(len(locales))]
}

func (g *FakeDataGenerator) generateCustomerID() string {
	return "test" + strconv.Itoa(g.rand.Intn(1000))
}

func (g *FakeDataGenerator) generateDeliveryService() string {
	services := []string{"meest", "cdek", "boxberry", "pickpoint", "dhl"}
	return services[g.rand.Intn(len(services))]
}

func (g *FakeDataGenerator) generateDateCreated() time.Time {
	// Генерируем дату в течение последних 30 дней
	daysAgo := g.rand.Intn(30)
	return time.Now().AddDate(0, 0, -daysAgo)
}

func (g *FakeDataGenerator) generateRid() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	rid := make([]byte, 20)
	for i := range rid {
		rid[i] = chars[g.rand.Intn(len(chars))]
	}
	return string(rid) + "test"
}

func (g *FakeDataGenerator) generateStatus() int {
	statuses := []int{200, 201, 202, 400, 404}
	return statuses[g.rand.Intn(len(statuses))]
}

func main() {
	var (
		brokers = flag.String("brokers", "localhost:9092", "Kafka brokers")
		topic   = flag.String("topic", "orders", "Kafka topic")
		count   = flag.Int("count", 10, "Number of orders to generate")
		delay   = flag.Duration("delay", time.Second, "Delay between messages")
	)
	flag.Parse()

	log.Printf("Запуск генератора данных...")
	log.Printf("Brokers: %s, Topic: %s, Count: %d, Delay: %s", *brokers, *topic, *count, *delay)

	// Создаем Kafka writer
	w := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(*brokers, ",")...),
		Topic:        *topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("Ошибка при закрытии Kafka writer: %v", err)
		}
	}()

	generator := NewFakeDataGenerator()

	for i := 0; i < *count; i++ {
		// Генерируем заказ
		order := generator.GenerateOrder()

		// Сериализуем в JSON
		orderJSON, err := json.Marshal(order)
		if err != nil {
			log.Printf("Ошибка сериализации заказа: %v", err)
			continue
		}

		// Отправляем в Kafka
		err = w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(order.OrderUID),
				Value: orderJSON,
			},
		)

		if err != nil {
			log.Printf("Ошибка отправки заказа: %v", err)
			continue
		}

		log.Printf("Отправлен заказ %d/%d: %s", i+1, *count, order.OrderUID)

		if i < *count-1 {
			time.Sleep(*delay)
		}
	}

	log.Printf("Генерация завершена. Отправлено %d заказов", *count)
}