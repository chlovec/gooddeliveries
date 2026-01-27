package kitchen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShellStorageA_AddRemoveLen(t *testing.T) {
	decay := 2
	sh := newShelfStorage(decay)
	now := time.Now()

	hotOrder := &KitchenOrder{
		ID:          "hot1",
		Name:        "Hot Pizza",
		Temperature: TemperatureHot,
		Price:       10,
		Freshness:   10 * time.Minute,
		cookedAt:    now.Add(-5 * time.Minute),
	}

	coldOrder := &KitchenOrder{
		ID:          "cold1",
		Name:        "Cold Salad",
		Temperature: TemperatureCold,
		Price:       5,
		Freshness:   15 * time.Minute,
		cookedAt:    now.Add(-10 * time.Minute),
	}

	roomOrder := &KitchenOrder{
		ID:          "room1",
		Name:        "Room Sandwich",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
		cookedAt:    now.Add(-15 * time.Minute),
	}

	noTempOrder := &KitchenOrder{
		ID:        "room1",
		Name:      "Room Sandwich",
		Price:     7,
		Freshness: 20 * time.Minute,
		cookedAt:  now.Add(-15 * time.Minute),
	}

	// Add orders
	hotEl := sh.add(hotOrder)
	coldEl := sh.add(coldOrder)
	roomEl := sh.add(roomOrder)
	noTempEl := sh.add(noTempOrder)

	require.Equal(t, 3, sh.len(), "storage should contain exactly 3 items")
	require.Nil(t, noTempEl, "order without a valid temperature should not be added to the storage")

	require.NotNil(t, hotEl, "failed to add hot order to storage")
	require.Equal(t, hotOrder, hotEl.Value, "hot order should match what's in storage")

	require.NotNil(t, coldEl, "failed to add cold order to storage")
	require.Equal(t, coldOrder, coldEl.Value, "cold order should match what's in storage")

	require.NotNil(t, roomEl, "failed to add room order to storage")
	require.Equal(t, roomOrder, roomEl.Value, "room order should match what's in storage")

	sh.remove(nil)
	require.Equal(t, 3, sh.len(), "storage should contain exactly 3 items")

	sh.remove(hotEl)
	require.Equal(t, 2, sh.len(), "storage should contain exactly 2 items")

	sh.remove(coldEl)
	require.Equal(t, 1, sh.len(), "storage should contain exactly 1 items")

	sh.remove(roomEl)
	require.Zero(t, sh.len(), "storage should contain exactly 0 items")

	el := sh.add(nil)
	require.Nil(t, el, "storage should not add nil order")
}

func TestShellStorage_GetOrderToDiscard_GetLeastFreshOrOldest(t *testing.T) {
	sh := newShelfStorage(2)
	now := time.Now()

	hotOrder := &KitchenOrder{
		ID:          "hot1",
		Name:        "Hot Pizza",
		Temperature: TemperatureHot,
		Price:       10,
		Freshness:   10 * time.Minute,
		cookedAt:    now.Add(-5 * time.Minute),
	}

	coldOrder := &KitchenOrder{
		ID:          "cold1",
		Name:        "Cold Salad",
		Temperature: TemperatureCold,
		Price:       5,
		Freshness:   10 * time.Minute,
		cookedAt:    now.Add(-10 * time.Minute),
	}

	roomOrder := &KitchenOrder{
		ID:          "room1",
		Name:        "Room Sandwich",
		Temperature: TemperatureRoom,
		Price:       7,
		Freshness:   1 * time.Minute,
		cookedAt:    now.Add(-15 * time.Minute),
	}

	roomEl := sh.add(roomOrder)
	hotEl := sh.add(hotOrder)
	coldEl := sh.add(coldOrder)

	t.Run("Discard/PreferColdOverHot_WhenColdIsLessFresh", func(t *testing.T) {
		el := sh.getOrderToDiscard()
		require.Equal(t, el, coldEl)
	})

	t.Run("Discard/ReturnHot_WhenColdShelfIsEmpty", func(t *testing.T) {
		sh.remove(coldEl)
		el := sh.getOrderToDiscard()
		require.Equal(t, el, hotEl)
	})

	t.Run("Discard/ReturnRoom_WhenColdAndHotShelvesAreEmpty", func(t *testing.T) {
		sh.remove(hotEl)
		el := sh.getOrderToDiscard()
		require.Equal(t, el, roomEl)

		sh.remove(roomEl)
	})

	t.Run("Discard/UpdatePrecedence_WhenNewColdOrderIsLeastFresh", func(t *testing.T) {
		newColdEl := sh.add(coldOrder)
		el := sh.getOrderToDiscard()
		require.Equal(t, el, newColdEl)
		require.NotEqual(t, el, coldEl)
		sh.remove(newColdEl)
	})

	t.Run("Comparison/PickHot_WhenHotHasLowerCalculatedFreshness", func(t *testing.T) {
		coldOrder.cookedAt = now.Add(-5 * time.Minute)
		coldOrder.Freshness = 30 * time.Second

		hotOrder.cookedAt = now.Add(-5 * time.Minute)
		hotOrder.Freshness = 20 * time.Second

		newElCold := sh.add(coldOrder)
		newElHot := sh.add(hotOrder)
		el := sh.getLeastFreshOrOldest(newElCold, newElHot)
		require.Equal(t, el, newElHot)

		sh.remove(newElHot)
		sh.remove(newElCold)
	})

	t.Run("Comparison/PickCold_WhenColdIsOlderDespiteHigherFreshness", func(t *testing.T) {
		coldOrder.cookedAt = now.Add(-30 * time.Minute)
		coldOrder.Freshness = 60 * time.Minute

		hotOrder.cookedAt = now.Add(-20 * time.Minute)
		hotOrder.Freshness = 40 * time.Minute

		newElCold := sh.add(coldOrder)
		newElHot := sh.add(hotOrder)
		el := sh.getLeastFreshOrOldest(newElCold, newElHot)

		require.Equal(t, el, newElCold)

		sh.remove(newElHot)
		sh.remove(newElCold)
	})

	t.Run("TieBreaking/PreferHot_WhenStatsAreIdentical", func(t *testing.T) {
		// Matches the 'default' case in getLeastFreshOrOldest switch
		require.Zero(t, sh.len(), 0)
		coldOrder.cookedAt = now.Add(-30 * time.Minute)
		coldOrder.Freshness = 30 * time.Minute

		hotOrder.cookedAt = now.Add(-30 * time.Minute)
		hotOrder.Freshness = 30 * time.Minute

		newElCold := sh.add(coldOrder)
		newElHot := sh.add(hotOrder)
		el := sh.getLeastFreshOrOldest(newElCold, newElHot)

		require.Equal(t, el, newElHot)

		sh.remove(newElHot)
		sh.remove(newElCold)
	})
}