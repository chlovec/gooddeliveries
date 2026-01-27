package kitchen

import (
	css "challenge/client"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKitchen_PlaceOrder_PickUpOrder(t *testing.T) {
	hotOrder := css.Order{
		ID:          "hot1",
		Name:        "Hot Pizza",
		Temp: string(TemperatureHot),
		Price:       10,
		Freshness:   600,
	}

	coldOrder := css.Order{
		ID:          "cold1",
		Name:        "Cold Salad",
		Temp: string(TemperatureCold),
		Price:       5,
		Freshness:   900,
	}

	roomOrder := css.Order{
		ID:          "room1",
		Name:        "Room Sandwich",
		Temp: string(TemperatureRoom),
		Price:       7,
		Freshness:   180,
	}

	invalidOrder := css.Order{
		ID:        "room2",
		Name:      "Room Sandwich",
		Price:     7,
		Freshness: 1200,
	}

	coldOrder2 := css.Order{
		ID:          "cold2",
		Name:        "Cold Salad",
		Temp: string(TemperatureCold),
		Price:       5,
		Freshness:   0,
	}

	coldOrder3 := css.Order{
		ID:          "cold3",
		Name:        "Cold Salad",
		Temp: string(TemperatureCold),
		Price:       5,
		Freshness:   0,
	}

	coldOrder4 := css.Order{
		ID:          "cold4",
		Name:        "Cold Salad",
		Temp: string(TemperatureCold),
		Price:       5,
		Freshness:   30,
	}

	hotOrder2 := css.Order{
		ID:          "hot2",
		Name:        "Hot Pizza",
		Temp: string(TemperatureHot),
		Price:       10,
		Freshness:   600,
	}

	decay := 2
	f, err := os.OpenFile("myorders", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(f, nil))

	t.Run("PlaceOrder/PreferredStorage", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)

		// k.PlaceOrder(css.Order{})
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		// k.PlaceOrder(invalidOrder)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		// Verify cold order was stored
		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder, &pickColdOrder)
		require.Zero(t, k.cooler.len())

		// Verify hot order was stored
		pickupHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder, &pickupHotOrder)
		require.Zero(t, k.heater.len())

		// Verify hot order was stored
		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.True(t, ok)
		require.Equal(t, roomOrder, &pickupRoomOrder)
		require.Zero(t, k.shelf.len())

		// Verify invalidOrder was not stored
		pickupInvalidOrder, ok := k.PickUpOrder(invalidOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupInvalidOrder)
	})

	t.Run("PickUpOrder/PreferredStorage/Decay", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)
		k.PlaceOrder(coldOrder2)
		require.Equal(t, 1, k.cooler.len())

		require.Equal(t, 1, k.cooler.len())

		time.Sleep(1 * time.Millisecond)
		order, ok := k.PickUpOrder(coldOrder2.ID)

		require.False(t, ok)
		require.Zero(t, order)
		require.Equal(t, 0, k.cooler.len())
	})

	t.Run("PickUpOrder/SecondaryStorage/Decay", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)
		k.PlaceOrder(coldOrder2)
		require.Equal(t, 1, k.cooler.len())

		time.Sleep(1 * time.Millisecond)
		order, ok := k.PickUpOrder(coldOrder2.ID)

		require.False(t, ok)
		require.Zero(t, order)

		require.Equal(t, 0, k.cooler.len())
	})

	t.Run("PickUpOrder/ColdOrder/ShelfStorage/Decay", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)
		k.PlaceOrder(coldOrder2)
		k.PlaceOrder(coldOrder3)

		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		time.Sleep(2 * time.Millisecond)
		order, ok := k.PickUpOrder(coldOrder3.ID)

		require.Equal(t, 1, k.cooler.len())
		require.Zero(t, k.shelf.len())

		require.False(t, ok)
		require.Zero(t, order)
	})

	t.Run("PickUpOrder/ShelfStorage/Discard/ShelfOrder", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(coldOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder4)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		// Verify cold order was stored
		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder, &pickColdOrder)
		require.Zero(t, k.cooler.len())

		// Verify hot order was stored
		pickupHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder, &pickupHotOrder)
		require.Zero(t, k.heater.len())

		// Verify room order was discarded
		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupRoomOrder)
		require.Equal(t, 1, k.shelf.len())

		// Verify cold order 4 was picked from the shelf
		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder4, &pickupColdOrder4)
		require.Zero(t, k.shelf.len())
	})

	t.Run("PickUpOrder/MoveHotOrder/Shelf_Heater", func(t *testing.T) {
		k := NewKitchen(1, 1, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(hotOrder2)
		k.PlaceOrder(coldOrder)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 2, k.shelf.len())

		// Verify hotOrder was picked up from the Heater
		pickupHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder, &pickupHotOrder)
		require.Zero(t, k.heater.len())

		k.PlaceOrder(coldOrder2)

		// Verify hotOrder2 was picked up from the heater
		pickupHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder2, &pickupHotOrder2)
		require.Zero(t, k.heater.len())
	})

	t.Run("PickUpOrder/MoveColdOrder/Shelf_Cooler", func(t *testing.T) {
		// Move/FromShelf/Cold/Storage
		k := NewKitchen(1, 1, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 2, k.shelf.len())

		// Verify cold order was picked up from cold storage
		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder, &pickColdOrder)
		require.Zero(t, k.cooler.len())

		k.PlaceOrder(hotOrder2)

		// Verify coldOrder4 was moved to the cooler
		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder4, &pickupColdOrder4)
		require.Zero(t, k.cooler.len())

		// Verify hotOrder2 was stored on the shelf
		pickupHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder2, &pickupHotOrder2)
		require.Equal(t, 1, k.shelf.len())
	})

	t.Run("Placement/DiscardRoomOrderOnShelfToMakeRoomForColdOrder", func(t *testing.T) {
		// cooler is full
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		k.PlaceOrder(coldOrder4)

		// roomOrder is discarded
		// coldOrder4 is moved to the shelf

		// Verify shelf is still at capacity
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		// Verify roomOrder was discarded
		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupRoomOrder)
		require.Equal(t, 1, k.shelf.len())

		// Verify coldOrder is picked up from the cooler
		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder, &pickColdOrder)
		require.Zero(t, k.cooler.len())

		// Verify coldOrder4 is picked up from the shelf
		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.True(t, ok)
		require.Equal(t, coldOrder4, &pickupColdOrder4)
		require.Zero(t, k.shelf.len())
	})

	t.Run("Placement/DiscardRoomOrderOnShelfToMakeRoomForHotOrder", func(t *testing.T) {
		// Heater is full
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(roomOrder)
		k.PlaceOrder(hotOrder)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.shelf.len())

		k.PlaceOrder(hotOrder2)

		// roomOrder is discarded
		// hotOrder2 is moved to the shelf

		// Verify shelf is still at capacity
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.shelf.len())

		// Verify roomOrder was discarded
		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupRoomOrder)
		require.Equal(t, 1, k.shelf.len())

		// Verify hotOrder is picked up from the heater
		pickHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder, &pickHotOrder)
		require.Zero(t, k.heater.len())

		// Verify hotOrder2 is picked up from the shelf
		pickHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		require.Equal(t, hotOrder2, &pickHotOrder2)
		require.Zero(t, k.shelf.len())
	})

	t.Run("Placement/MoveColdOrderFromShelf_CoolerIsAlreadyFull", func(t *testing.T) {
		k := NewKitchen(1, 1, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)

		// Verify orders are placed in the right storage
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 2, k.shelf.len())

		k.PlaceOrder(hotOrder2)

		// Verify storages are still at capacity
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 2, k.shelf.len())

		// Verify coldOrder4 was discarded to place hotOrder2 in the shelf
		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.False(t, ok)
		require.Zero(t, pickupColdOrder4)

		// Verify storages are still at capacity
		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 2, k.shelf.len())

		pickUpHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)

		// Verify hotOrder2 is picked up from the shelf
		require.True(t, ok)
		require.Equal(t, hotOrder2, &pickUpHotOrder2)
		require.Equal(t, 1, k.shelf.len())
	})
}
