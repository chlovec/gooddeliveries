package kitchenv3

import (
	"challenge/client"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Kitchen struct {
	heater *Storage
	cooler *Storage
	shelf  *ShelfStorage
	logger *slog.Logger
	mu     sync.Mutex
}

func NewKitchen(
	hotCapacity int64,
	coldCapacity int64,
	shelfCapacity int64,
	decay int,
	logger *slog.Logger,
) *Kitchen {
	return &Kitchen{
		heater: NewStorage(hotCapacity),
		cooler: NewStorage(coldCapacity),
		shelf:  NewShelfStorage(shelfCapacity, decay),
		logger: logger,
	}
}

func (k *Kitchen) PlaceOrder(newOrder client.Order) error {
	// validate order
	if err := IsValidOrder(newOrder); err != nil {
		return err
	}

	order := &KitchenOrder{
		ID:          newOrder.ID,
		Name:        newOrder.Name,
		Temperature: Temperature(newOrder.Temp),
		Price:       newOrder.Price,
		Freshness:   time.Duration(newOrder.Freshness) * time.Second,
		cookedAt:    time.Now(),
	}

	var placed bool
	switch order.Temperature {
	case TemperatureHot:
		placed = k.heater.Add(order)
	case TemperatureCold:
		placed = k.cooler.Add(order)
	default:
		placed = k.shelf.Add(order)
	}

	if !placed {
		placed = k.placeInShelf(order)
	}

	return nil
}

func (k *Kitchen) PickUpOrder(orderID string) (client.Order, error) {
	var foundOrder *KitchenOrder

	// Try to find and remove the order from  any of the three storages
	if order, ok := k.heater.Remove(orderID); ok {
		foundOrder = order
	} else if order, ok := k.cooler.Remove(orderID); ok {
		foundOrder = order
	} else if order, ok := k.shelf.Remove(orderID); ok {
		foundOrder = order
	}

	if foundOrder == nil {
		return client.Order{}, errors.New("order not found")
	} else if foundOrder.Freshness <= 0 {
		return client.Order{}, fmt.Errorf("order has expired: %+v", foundOrder.Freshness)
	}

	return client.Order{
		ID:        foundOrder.ID,
		Name:      foundOrder.Name,
		Temp:      string(foundOrder.Temperature),
		Price:     foundOrder.Price,
		Freshness: int(foundOrder.Freshness),
	}, nil
}

// -- Helper Functions --

func (k *Kitchen) placeInShelf(order *KitchenOrder) bool {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.shelf.HasSpace() {
		return k.shelf.Add(order)
	}

	if order.Temperature == TemperatureCold && k.moveShelfHotOrder() {
		k.moveShelfColdOrder()
	} else if order.Temperature == TemperatureHot && k.moveShelfColdOrder() {
		k.moveShelfHotOrder()
	} else {
		toDiscard := k.shelf.GetOrderToDiscard()
		k.shelf.Remove(toDiscard.ID)
	}

	return k.shelf.Add(order)
}

func (k *Kitchen) moveShelfColdOrder() bool {
	if !k.cooler.HasSpace() {
		return false
	}

	order := k.shelf.GetFirstColdOrder()
	if order == nil {
		return false
	}

	if _, ok := k.shelf.Remove(order.ID); !ok {
		return false
	}
	return k.cooler.Add(order)
}

func (k *Kitchen) moveShelfHotOrder() bool {
	if !k.heater.HasSpace() {
		return false
	}

	order := k.shelf.GetFirstHotOrder()
	if order == nil {
		return false
	}

	if _, ok := k.shelf.Remove(order.ID); !ok {
		return false
	}
	return k.heater.Add(order)
}
