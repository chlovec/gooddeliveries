package kitchenv3

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type Temperature string

const (
	TemperatureHot  Temperature = "hot"
	TemperatureCold Temperature = "cold"
	TemperatureRoom Temperature = "room"
)

type KitchenOrder struct {
	ID          string
	Name        string
	Temperature Temperature
	Price       int
	Freshness   time.Duration
	cookedAt    time.Time
	lastUpdated time.Time
}

// -- General storage for Cooler and Heater --

type Storage struct {
	capacity int64
	count    int64
	items    map[string]*KitchenOrder
	mu       sync.Mutex
}

func NewStorage(capacity int64) *Storage {
	return &Storage{
		capacity: capacity,
		items:    make(map[string]*KitchenOrder, int(capacity)),
	}
}

func (s *Storage) Add(order *KitchenOrder) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure there is space
	if s.count == s.capacity {
		return false
	}

	order.cookedAt = time.Now()

	// Assume every other is unique
	s.items[order.ID] = order
	s.count++
	return true
}

func (s *Storage) Remove(orderid string) (*KitchenOrder, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.items[orderid]
	if ok {
		delete(s.items, orderid)
		s.count--
	}
	return order, ok
}

func (s *Storage) HasSpace() bool {
	return atomic.LoadInt64(&s.count) < int64(s.capacity)
}

// -- Shelf storage --

type ShelfStorage struct {
	capacity  int64
	count     int64
	decay     int
	items     map[string]*list.Element
	coldItems *list.List
	hotItems  *list.List
	roomItems *list.List
	mu        sync.Mutex
}

func NewShelfStorage(capacity int64, decay int) *ShelfStorage {
	return &ShelfStorage{
		capacity:  capacity,
		coldItems: list.New(),
		hotItems:  list.New(),
		roomItems: list.New(),
		decay:     decay,
		items:     make(map[string]*list.Element, capacity),
	}
}

func (s *ShelfStorage) Add(order *KitchenOrder) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure there is space
	if s.count == s.capacity {
		return false
	}

	var el *list.Element
	order.cookedAt = time.Now()

	// Assume order's temperature is any of cold, hot, room
	switch order.Temperature {
	case TemperatureCold:
		el = s.coldItems.PushBack(order)
	case TemperatureHot:
		el = s.hotItems.PushBack(order)
	default:
		el = s.roomItems.PushBack(order)
	}

	s.items[order.ID] = el
	s.count++
	return true
}

func (s *ShelfStorage) Remove(orderid string) (*KitchenOrder, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	el, ok := s.items[orderid]
	if !ok {
		return nil, false
	}

	order := el.Value.(*KitchenOrder)
	switch order.Temperature {
	case TemperatureCold:
		s.coldItems.Remove(el)
	case TemperatureHot:
		s.hotItems.Remove(el)
	default:
		s.roomItems.Remove(el)
	}

	delete(s.items, orderid)

	order.lastUpdated = time.Now()
	order.Freshness -= time.Duration(s.decay) * time.Second * order.lastUpdated.Sub(order.cookedAt)

	s.count--
	return order, true
}

func (s *ShelfStorage) GetOrderToDiscard() *KitchenOrder {
	// 1. Get the oldest cold item and the oldest hot item
	// 2. Compare both and return the oldest
	// 3. If there is no cold or hot item, return the oldest shelf item

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.count == 0 {
		return nil
	}

	var oldestEl *list.Element
	findOldest := func(l *list.List) {
		front := l.Front()
		if front == nil {
			return
		}
		if oldestEl == nil || front.Value.(*KitchenOrder).cookedAt.Before(oldestEl.Value.(*KitchenOrder).cookedAt) {
			oldestEl = front
		}
	}

	findOldest(s.coldItems)
	findOldest(s.hotItems)
	if oldestEl == nil {
		findOldest(s.roomItems)
	}

	if oldestEl == nil {
		return nil
	}

	return oldestEl.Value.(*KitchenOrder)
}

func (s *ShelfStorage) GetFirstColdOrder() *KitchenOrder {
	el := s.coldItems.Front()
	if el == nil {
		return nil
	}

	return el.Value.(*KitchenOrder)
}

func (s *ShelfStorage) GetFirstHotOrder() *KitchenOrder {
	el := s.hotItems.Front()
	if el == nil {
		return nil
	}

	return el.Value.(*KitchenOrder)
}

func (s *ShelfStorage) GetFirstRoomOrder() *KitchenOrder {
	el := s.roomItems.Front()
	if el == nil {
		return nil
	}

	return el.Value.(*KitchenOrder)
}

func (s *ShelfStorage) HasSpace() bool {
	return atomic.LoadInt64(&s.count) < int64(s.capacity)
}
