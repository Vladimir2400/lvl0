package cache

import (
	"container/list"
	"sync"
	"time"
	"wb-service/internal/interfaces"
	"wb-service/models"
)

// CacheItem представляет элемент кэша с временной меткой
type CacheItem struct {
	Key       string
	Value     *models.Order
	Timestamp time.Time
}

// LRUCache реализует LRU кэш с поддержкой TTL
type LRUCache struct {
	mutex    sync.RWMutex
	capacity int
	ttl      time.Duration
	items    map[string]*list.Element
	evictList *list.List
}

// NewLRUCache создает новый LRU кэш
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	c := &LRUCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}

	// Запускаем горутину для очистки устаревших элементов
	go c.cleanupExpired()

	return c
}

// Get получает значение из кэша
func (c *LRUCache) Get(key string) (*models.Order, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*CacheItem)

		// Проверяем, не истек ли TTL
		if c.ttl > 0 && time.Since(item.Timestamp) > c.ttl {
			c.removeElement(elem)
			return nil, false
		}

		// Перемещаем элемент в начало списка (most recently used)
		c.evictList.MoveToFront(elem)
		return item.Value, true
	}

	return nil, false
}

// Set добавляет значение в кэш
func (c *LRUCache) Set(key string, order *models.Order) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Если элемент уже существует, обновляем его
	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*CacheItem)
		item.Value = order
		item.Timestamp = time.Now()
		c.evictList.MoveToFront(elem)
		return
	}

	// Создаем новый элемент
	item := &CacheItem{
		Key:       key,
		Value:     order,
		Timestamp: time.Now(),
	}

	// Добавляем в начало списка
	elem := c.evictList.PushFront(item)
	c.items[key] = elem

	// Если превышена емкость, удаляем последний элемент
	if c.evictList.Len() > c.capacity {
		c.removeOldest()
	}
}

// LoadFromDB загружает данные из базы данных в кэш
func (c *LRUCache) LoadFromDB(db interfaces.Database) error {
	if db == nil {
		return nil // Просто возвращаем без ошибки для nil db
	}

	orders, err := db.GetAllOrders()
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, order := range orders {
		if len(c.items) >= c.capacity {
			break
		}

		item := &CacheItem{
			Key:       order.OrderUID,
			Value:     &order,
			Timestamp: time.Now(),
		}

		elem := c.evictList.PushFront(item)
		c.items[order.OrderUID] = elem
	}

	return nil
}

// Size возвращает текущий размер кэша
func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Clear очищает кэш
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList.Init()
}

// removeOldest удаляет самый старый элемент
func (c *LRUCache) removeOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement удаляет элемент из кэша
func (c *LRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	item := elem.Value.(*CacheItem)
	delete(c.items, item.Key)
}

// cleanupExpired периодически очищает устаревшие элементы
func (c *LRUCache) cleanupExpired() {
	if c.ttl <= 0 {
		return
	}

	ticker := time.NewTicker(c.ttl / 2) // Проверяем каждые ttl/2
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()

		var toRemove []*list.Element
		for elem := c.evictList.Back(); elem != nil; elem = elem.Prev() {
			item := elem.Value.(*CacheItem)
			if time.Since(item.Timestamp) > c.ttl {
				toRemove = append(toRemove, elem)
			} else {
				break // Так как список отсортирован по времени, можем остановиться
			}
		}

		for _, elem := range toRemove {
			c.removeElement(elem)
		}

		c.mutex.Unlock()
	}
}