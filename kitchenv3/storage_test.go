package kitchenv3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStorage_AddRemoveHasSpace(t *testing.T) {
	order1 := &KitchenOrder{
		ID:          "order1",
		Name:        "Hot Pizza",
		Temperature: TemperatureHot,
		Price:       10,
		Freshness:   10 * time.Minute,
	}

	order2 := &KitchenOrder{
		ID:          "order2",
		Name:        "Cold Salad",
		Temperature: TemperatureCold,
		Price:       5,
		Freshness:   15 * time.Minute,
	}

	order3 := &KitchenOrder{
		ID:          "order3",
		Name:        "Room Sandwich",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
	}

	s := NewStorage(2)
	require.True(t, s.HasSpace())

	ok := s.Add(order1)
	require.True(t, ok)
	require.True(t, s.HasSpace())

	ok = s.Add(order2)
	require.True(t, ok)
	require.False(t, s.HasSpace())

	ok = s.Add(order3)
	require.False(t, ok)
	require.False(t, s.HasSpace())

	storedOrder1, ok := s.Remove(order1.ID)
	require.True(t, ok)
	require.Equal(t, order1, storedOrder1)

	storedOrder2, ok := s.Remove(order2.ID)
	require.True(t, ok)
	require.Equal(t, order2, storedOrder2)

	storedOrder3, ok := s.Remove(order3.ID)
	require.False(t, ok)
	require.Nil(t, storedOrder3)
}

func TestShellStorage_AddRemoveHasSpace(t *testing.T) {
	order1 := &KitchenOrder{
		ID:          "order1",
		Name:        "Hot Pizza",
		Temperature: TemperatureHot,
		Price:       10,
		Freshness:   10 * time.Minute,
	}

	order2 := &KitchenOrder{
		ID:          "order2",
		Name:        "Cold Salad",
		Temperature: TemperatureCold,
		Price:       5,
		Freshness:   15 * time.Minute,
	}

	order3 := &KitchenOrder{
		ID:          "order3",
		Name:        "Room Sandwich",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
	}

	order4 := &KitchenOrder{
		ID:          "order4",
		Name:        "Room Sandwich",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
	}

	s := NewShelfStorage(3, 2)
	require.True(t, s.HasSpace())

	ok := s.Add(order1)
	require.True(t, ok)
	require.True(t, s.HasSpace())
	require.Equal(t, 1, s.hotItems.Len())
	require.Zero(t, s.coldItems.Len())
	require.Zero(t, s.roomItems.Len())

	ok = s.Add(order2)
	require.True(t, ok)
	require.True(t, s.HasSpace())
	require.Equal(t, 1, s.hotItems.Len())
	require.Equal(t, 1, s.coldItems.Len())
	require.Zero(t, s.roomItems.Len())

	ok = s.Add(order3)
	require.True(t, ok)
	require.False(t, s.HasSpace())
	require.Equal(t, 1, s.hotItems.Len())
	require.Equal(t, 1, s.coldItems.Len())
	require.Equal(t, 1, s.roomItems.Len())

	ok = s.Add(order4)
	require.False(t, ok)
	require.False(t, s.HasSpace())
	require.Equal(t, 1, s.hotItems.Len())
	require.Equal(t, 1, s.coldItems.Len())
	require.Equal(t, 1, s.roomItems.Len())

	storedOrder1, ok := s.Remove(order1.ID)
	require.True(t, ok)
	require.Equal(t, order1, storedOrder1)
	require.Zero(t, s.hotItems.Len())
	require.Equal(t, 1, s.coldItems.Len())
	require.Equal(t, 1, s.roomItems.Len())

	storedOrder2, ok := s.Remove(order2.ID)
	require.True(t, ok)
	require.Equal(t, order2, storedOrder2)
	require.Zero(t, s.hotItems.Len())
	require.Zero(t, s.coldItems.Len())
	require.Equal(t, 1, s.roomItems.Len())

	storedOrder3, ok := s.Remove(order3.ID)
	require.True(t, ok)
	require.Equal(t, order3, storedOrder3)
	require.Zero(t, s.hotItems.Len())
	require.Zero(t, s.coldItems.Len())
	require.Zero(t, s.roomItems.Len())

	storedOrder4, ok := s.Remove(order4.ID)
	require.False(t, ok)
	require.Nil(t, storedOrder4)
}

func TestShellStorage_GetOrderToDiscard(t *testing.T) {
	coldOrder1 := &KitchenOrder{
		ID:          "coldOrder1",
		Name:        "Cold Salad",
		Temperature: TemperatureCold,
		Price:       5,
		Freshness:   15 * time.Minute,
	}

	coldOrder2 := &KitchenOrder{
		ID:          "coldOrder2",
		Name:        "Cold Salad",
		Temperature: TemperatureCold,
		Price:       5,
		Freshness:   5 * time.Minute,
	}

	hotOrder1 := &KitchenOrder{
		ID:          "hotOrder1",
		Name:        "hot soup",
		Temperature: TemperatureHot,
		Price:       5,
		Freshness:   15 * time.Minute,
	}

	hotOrder2 := &KitchenOrder{
		ID:          "hotOrder2",
		Name:        "hot pizza",
		Temperature: TemperatureHot,
		Price:       5,
		Freshness:   5 * time.Minute,
	}

	shelfOrder1 := &KitchenOrder{
		ID:          "shelfOrder1",
		Name:        "Room Sandwich",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
	}

	shelfOrder2 := &KitchenOrder{
		ID:          "shelfOrder2",
		Name:        "Room Water",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
	}

	const capacity int64 = 6
	const decay = 2

	t.Run("ReturnsFirstColdItem_WhenColdItems", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(coldOrder1)
		s.Add(coldOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, coldOrder1, actual)
		require.Equal(t, 2, s.coldItems.Len())
		require.Zero(t, s.hotItems.Len())
		require.Zero(t, s.roomItems.Len())
	})

	t.Run("ReturnsFirstHotItem_WhenHotItems", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(hotOrder2)
		s.Add(hotOrder1)

		actual := s.GetOrderToDiscard()

		require.Equal(t, hotOrder2, actual)
		require.Equal(t, 2, s.hotItems.Len())
		require.Zero(t, s.coldItems.Len())
		require.Zero(t, s.roomItems.Len())

	})

	t.Run("ReturnsFirstShelfItem_WhenShelfItems", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(shelfOrder1)
		s.Add(shelfOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, shelfOrder1, actual)
		require.Equal(t, 2, s.roomItems.Len())
		require.Zero(t, s.coldItems.Len())
		require.Zero(t, s.hotItems.Len())
	})

	t.Run("ReturnsFirstHotItem_WhenHotItemsAndShelfItems", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(shelfOrder1)
		s.Add(shelfOrder2)
		s.Add(hotOrder1)
		s.Add(hotOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, hotOrder1, actual)
		require.Equal(t, 2, s.hotItems.Len())
		require.Equal(t, 2, s.roomItems.Len())
		require.Zero(t, s.coldItems.Len())
	})

	t.Run("ReturnsFirstColdItem_WhenColdItemsAndShelfItems", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(shelfOrder1)
		s.Add(shelfOrder2)
		s.Add(coldOrder1)
		s.Add(coldOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, coldOrder1, actual)
		require.Equal(t, 2, s.coldItems.Len())
		require.Equal(t, 2, s.roomItems.Len())
		require.Zero(t, s.hotItems.Len())
	})

	t.Run("ReturnsFirstHotItem_WhenColdItemsAndHotItems_AndHotItemIsFirst", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(hotOrder1)
		s.Add(coldOrder1)
		s.Add(hotOrder2)
		s.Add(coldOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, hotOrder1, actual)
		require.Equal(t, 2, s.coldItems.Len())
		require.Equal(t, 2, s.hotItems.Len())
		require.Zero(t, s.roomItems.Len())
	})

	t.Run("ReturnsFirstColdItem_WhenColdItemsAndHotItems_AndColdItemIsFirst", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(coldOrder1)
		s.Add(hotOrder1)
		s.Add(hotOrder2)
		s.Add(coldOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, coldOrder1, actual)
		require.Equal(t, 2, s.coldItems.Len())
		require.Equal(t, 2, s.hotItems.Len())
		require.Zero(t, s.roomItems.Len())
	})

	t.Run("ReturnsFirstColdItem_WhenAllItemsItems_AndColdItemIsBeforeHotItem", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(shelfOrder1)
		s.Add(shelfOrder2)
		s.Add(coldOrder1)
		s.Add(hotOrder1)
		s.Add(hotOrder2)
		s.Add(coldOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, coldOrder1, actual)
		require.Equal(t, 2, s.coldItems.Len())
		require.Equal(t, 2, s.hotItems.Len())
		require.Equal(t, 2, s.roomItems.Len())
	})

	t.Run("ReturnsFirstHotItem_WhenAllItemsItems_AndHotItemIsBeforeColdItem", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		s.Add(shelfOrder1)
		s.Add(shelfOrder2)
		s.Add(hotOrder1)
		s.Add(coldOrder1)
		s.Add(hotOrder2)
		s.Add(coldOrder2)

		actual := s.GetOrderToDiscard()

		require.Equal(t, hotOrder1, actual)
		require.Equal(t, 2, s.coldItems.Len())
		require.Equal(t, 2, s.hotItems.Len())
		require.Equal(t, 2, s.roomItems.Len())
	})

	t.Run("ReturnsNil_WhenEmpty", func(t *testing.T) {
		s := NewShelfStorage(capacity, decay)
		actual := s.GetOrderToDiscard()
		require.Nil(t, actual)
	})
}
