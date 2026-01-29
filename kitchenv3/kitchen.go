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
	var storageName string
	switch order.Temperature {
	case TemperatureHot:
		placed = k.heater.Add(order)
		storageName = client.Heater
	case TemperatureCold:
		placed = k.cooler.Add(order)
		storageName = client.Cooler
	default:
		placed = k.shelf.Add(order)
		storageName = client.Shelf
	}

	if !placed {
		placed = k.placeInShelf(order)
		storageName = client.Shelf
	}

	// Log pickup and return results
	if !placed {
		return errors.New("unable to place order")
	}

	k.logger.Info(client.Place, "order id", order.ID, "target", storageName)
	return nil
}

func (k *Kitchen) PickUpOrder(orderID string) (client.Order, error) {
	var foundOrder *KitchenOrder

	// Try to find and remove the order from  any of the three storages
	var storageName string

	if order, ok := k.heater.Remove(orderID); ok {
		foundOrder = order
		storageName = client.Heater
	} else if order, ok := k.cooler.Remove(orderID); ok {
		foundOrder = order
		storageName = client.Cooler
	} else if order, ok := k.shelf.Remove(orderID); ok {
		storageName = client.Shelf
		foundOrder = order
	}

	if foundOrder == nil {
		return client.Order{}, errors.New("order not found")
	}

	k.logger.Info(client.Pickup, "order id", foundOrder.ID, "target", storageName)

	if foundOrder.Freshness <= 0 {
		// Should this also be logged as discarded?
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
		k.logger.Info(client.Discard, "order id", toDiscard.ID, "target", client.Shelf)
	}

	return k.shelf.Add(order)
}

func (k *Kitchen) moveShelfColdOrder() bool {
	if !k.cooler.HasSpace() {
		return false
	}

	order := k.shelf.GetFirstColdOrder()
	if order == nil || !k.cooler.Add(order) {
		return false
	}

	if _, ok := k.shelf.Remove(order.ID); !ok {
		return false
	}

	k.logger.Info(client.Move, "order id", order.ID, "target", client.Cooler)
	return true
}

func (k *Kitchen) moveShelfHotOrder() bool {
	if !k.heater.HasSpace() {
		return false
	}

	order := k.shelf.GetFirstHotOrder()
	if order == nil || !k.heater.Add(order) {
		return false
	}

	if _, ok := k.shelf.Remove(order.ID); !ok {
		return false
	}

	k.logger.Info(client.Move, "order id", order.ID, "target", client.Heater)
	return false
}
