package kitchen

import (
	css "challenge/client"
	"container/list"
	"log/slog"
	"sync"
	"time"
)

// OrderAdder allows us to treat both simple storage and complex shelfStorage
// as valid targets for placement.
type OrderAdder interface {
	add(order *KitchenOrder) *list.Element
	len() int
}

type storageInfo struct {
	el        *list.Element
	storeTemp Temperature
}

type Kitchen struct {
	mu            sync.Mutex
	hotCapacity   int
	coldCapacity  int
	shelfCapacity int
	heater        *storage
	cooler        *storage
	shelf         *shelfStorage
	index         map[string]storageInfo
	logger        *slog.Logger
}

func NewKitchen(
	hotCapacity int,
	coldCapacity int,
	shelfCapacity int,
	decay int,
	logger *slog.Logger,
) *Kitchen {
	return &Kitchen{
		hotCapacity:   hotCapacity,
		coldCapacity:  coldCapacity,
		shelfCapacity: shelfCapacity,
		heater:        newStorage(),
		cooler:        newStorage(),
		shelf:         newShelfStorage(decay),
		index:         make(map[string]storageInfo),
		logger:        logger,
	}
}

func (k *Kitchen) PlaceOrder(order css.Order) error {
	// validate order
	if err := IsValidOrder(order); err != nil {
		return err
	}

	kOrder := &KitchenOrder{
		ID:          order.ID,
		Name:        order.Name,
		Temperature: Temperature(order.Temp),
		Price:       float64(order.Price),
		Freshness:   time.Duration(order.Freshness) * time.Second,
		cookedAt:    time.Now(),
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	kOrder.cookedAt = time.Now()

	switch kOrder.Temperature {
	case TemperatureHot:
		k.placeHotOrder(kOrder)
	case TemperatureCold:
		k.placeColdOrder(kOrder)
	default:
		k.placeShelfOrder(kOrder)
	}

	return nil
}

func (k *Kitchen) PickUpOrder(orderID string) (css.Order, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()

	orderMeta, ok := k.index[orderID]
	if !ok || orderMeta.el == nil || orderMeta.el.Value == nil {
		return css.Order{}, false
	}

	order := orderMeta.el.Value.(*KitchenOrder)
	k.removeFromStorage(orderMeta)

	// Log pickup
	k.logger.Info(css.Pickup, "order id", orderID, "target", k.getStorageName(orderMeta.storeTemp))

	if order == nil || order.Freshness <= 0 {
		return css.Order{}, false
	}

	return css.Order{
		ID:        order.ID,
		Name:      order.Name,
		Temp:      string(order.Temperature),
		Price:     int(order.Price),
		Freshness: int(order.Freshness),
	}, true
}

func (k *Kitchen) placeHotOrder(order *KitchenOrder) {
	k.placeInStorage(order, k.heater, TemperatureHot, k.moveShelfColdOrder)
}

func (k *Kitchen) placeColdOrder(order *KitchenOrder) {
	k.placeInStorage(order, k.cooler, TemperatureCold, k.moveShelfHotOrder)
}

func (k *Kitchen) placeShelfOrder(order *KitchenOrder) {
	// For room orders, primary is the shelf. If there no available space in the shelf
	// An item will be discarded from the shelf.
	k.placeInStorage(order, k.shelf, TemperatureRoom, nil)
}

func (k *Kitchen) placeInStorage(
	order *KitchenOrder, primary OrderAdder, primaryTemp Temperature, moveFunc func() bool,
) {
	var el *list.Element
	var currentTemp Temperature

	switch {
	// Preferred storage has room
	case primary.len() < k.getCapacity(primaryTemp):
		el = primary.add(order)
		currentTemp = primaryTemp

	// Shelf has space
	case k.shelf.len() < k.shelfCapacity:
		el = k.shelf.add(order)
		currentTemp = TemperatureRoom

	// Move hot / cold items out of the shelf
	case moveFunc != nil && moveFunc():
		el = k.shelf.add(order)
		currentTemp = TemperatureRoom

	// Forced eviction
	default:
		k.discardShelfOrder()
		el = k.shelf.add(order)
		currentTemp = TemperatureRoom
	}

	k.index[order.ID] = storageInfo{el: el, storeTemp: currentTemp}

	// Log order placement
	k.logger.Info(css.Place, "order id", order.ID, "target", k.getStorageName(currentTemp))
}

func (k *Kitchen) getCapacity(temp Temperature) int {
	switch temp {
	case TemperatureHot:
		return k.hotCapacity
	case TemperatureCold:
		return k.coldCapacity
	case TemperatureRoom:
		return k.shelfCapacity
	default:
		return 0
	}
}

func (k *Kitchen) moveShelfColdOrder() bool {
	if k.cooler.len() >= k.coldCapacity {
		return false
	}

	el := k.shelf.getFirstColdOrder()
	if el == nil {
		return false
	}

	order := el.Value.(*KitchenOrder)
	id := order.ID

	k.shelf.remove(el)
	delete(k.index, id)

	newEl := k.cooler.add(order)
	k.index[id] = storageInfo{el: newEl, storeTemp: TemperatureCold}

	// Log move order - cooler
	k.logger.Info(css.Move, "order id", id, "target", k.getStorageName(TemperatureCold))

	return true
}

func (k *Kitchen) moveShelfHotOrder() bool {
	if k.heater.len() >= k.hotCapacity {
		return false
	}
	el := k.shelf.getFirstHotOrder()
	if el == nil {
		return false
	}
	order := el.Value.(*KitchenOrder)
	id := order.ID

	k.shelf.remove(el)
	delete(k.index, id)

	newEl := k.heater.add(order)
	k.index[id] = storageInfo{el: newEl, storeTemp: TemperatureHot}

	// Log move order - heater
	k.logger.Info(css.Move, "order id", id, "target", k.getStorageName(TemperatureHot))

	return true
}

func (k *Kitchen) discardShelfOrder() {
	el := k.shelf.getOrderToDiscard()
	order := el.Value.(*KitchenOrder)
	meta := k.index[order.ID]
	k.removeFromStorage(meta)

	// Log discard order - shelf
	k.logger.Info(css.Discard, "order id", order.ID, "target", k.getStorageName(TemperatureRoom))
}

func (k *Kitchen) removeFromStorage(meta storageInfo) {
	order, ok := meta.el.Value.(*KitchenOrder)
	if !ok || order == nil {
		return
	}
	orderID := order.ID

	switch meta.storeTemp {
	case TemperatureCold:
		k.cooler.remove(meta.el)
	case TemperatureHot:
		k.heater.remove(meta.el)
	default:
		k.shelf.remove(meta.el)
	}

	delete(k.index, orderID)
}

func (k *Kitchen) getStorageName(temp Temperature) string {
	switch temp {
	case TemperatureCold:
		return css.Cooler
	case TemperatureHot:
		return css.Heater
	case TemperatureRoom:
		return css.Shelf
	default:
		return "unknown"
	}
}
