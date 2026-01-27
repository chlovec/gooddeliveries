package kitchen

import (
	"container/list"
	"time"
)

type Temperature string

const (
	TemperatureHot  Temperature = "hot"
	TemperatureCold Temperature = "cold"
	TemperatureRoom Temperature = "room"
)

type KitchenOrder struct {
	ID          string        `validate:"required"`
	Name        string        `validate:"required"`
	Temperature Temperature   `validate:"temp"`
	Price       float64       `validate:"gt=0"`
	Freshness   time.Duration `validate:"gt=0"`
	cookedAt    time.Time
	lastUpdated time.Time
}

type storage struct {
	store *list.List
}

func newStorage() *storage {
	return &storage{
		store: list.New(),
	}
}

func (s *storage) add(order *KitchenOrder) *list.Element {
	el := s.store.PushBack(order)
	return el
}

func (s *storage) getFirstElement() *list.Element {
	return s.store.Front()
}

func (s *storage) remove(el *list.Element) {
	if el == nil || el.Value == nil {
		return
	}

	order := el.Value.(*KitchenOrder)
	s.store.Remove(el)

	if order.lastUpdated.IsZero() {
		order.lastUpdated = order.cookedAt
	}
	order.Freshness -= time.Since(order.lastUpdated)
}

func (s *storage) len() int {
	return s.store.Len()
}

type shelfStorage struct {
	shelf     *storage
	coldShelf *storage
	hotShelf  *storage
	decay     int
}

func newShelfStorage(decay int) *shelfStorage {
	return &shelfStorage{
		shelf:     newStorage(),
		coldShelf: newStorage(),
		hotShelf:  newStorage(),
		decay:     decay,
	}
}

func (shs *shelfStorage) add(order *KitchenOrder) *list.Element {
	if order == nil {
		return nil
	}

	// Add order to the appropriate shelf based on the preferred storage requirement
	// This makes finding a hot/cold order to move back to hot/cold storage O(1)
	switch order.Temperature {
	case TemperatureRoom:
		return shs.shelf.add(order)
	case TemperatureHot:
		return shs.hotShelf.add(order)
	case TemperatureCold:
		return shs.coldShelf.add(order)
	default:
		return nil
	}
}

func (shs *shelfStorage) remove(el *list.Element) {
	if el == nil {
		return
	}

	// Remove order from the appropriate shelf
	order := el.Value.(*KitchenOrder)
	switch order.Temperature {
	case TemperatureRoom:
		shs.shelf.remove(el)
	case TemperatureHot:
		shs.hotShelf.remove(el)
	default:
		shs.coldShelf.remove(el)
	}

	// Track time in shelf
	order.lastUpdated = time.Now()
	totalDecay := float64(shs.decay) * float64(order.lastUpdated.Sub(order.cookedAt))
	order.Freshness -= time.Duration(totalDecay)
}

func (shs *shelfStorage) getFirstColdOrder() *list.Element {
	// returns nil if no element exist
	return shs.coldShelf.getFirstElement()
}

func (shs *shelfStorage) getFirstHotOrder() *list.Element {
	// returns nil if no element exist
	return shs.hotShelf.getFirstElement()
}

func (shs *shelfStorage) getOrderToDiscard() *list.Element {
	elCold := shs.getFirstColdOrder()
	elHot := shs.getFirstHotOrder()

	switch {
	case elCold != nil && elHot != nil:
		return shs.getLeastFreshOrOldest(elCold, elHot)

	case elHot != nil:
		return elHot

	case elCold != nil:
		return elCold

	default:
		return shs.shelf.getFirstElement()
	}
}

func (shs *shelfStorage) getLeastFreshOrOldest(el1, el2 *list.Element) *list.Element {
	now := time.Now()
	order1 := el1.Value.(*KitchenOrder)
	order2 := el2.Value.(*KitchenOrder)

	// Formula: Value = (Freshness - (DecayFactor * TimeInStorage))
	val1 := shs.getOrderFreshness(order1, now)
	val2 := shs.getOrderFreshness(order2, now)

	switch {
	case val1 < val2:
		return el1
	case val2 < val1:
		return el2
	// Tie-breaker: If calculated freshness is equal, discard the one cooked earlier
	case order1.cookedAt.Before(order2.cookedAt):
		return el1
	default:
		return el2
	}
}

func (shs *shelfStorage) getOrderFreshness(order *KitchenOrder, refTime time.Time) time.Duration {
	// Convert all components to float64 to maintain precision during decay multiplication
	freshnessValue := float64(order.Freshness)
	timeInStorage := float64(refTime.Sub(order.cookedAt))

	// Formula: Remaining Freshness = Max Freshness - (Decay Factor * Time Passed)
	remaining := freshnessValue - (timeInStorage * float64(shs.decay))
	return time.Duration(remaining)
}

func (shs *shelfStorage) len() int {
	return shs.shelf.len() + shs.hotShelf.len() + shs.coldShelf.len()
}
