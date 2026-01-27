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

// Helper function to compare specific order fields
func assertOrderMatch(t *testing.T, expected css.Order, actual css.Order) {
	require.Equal(t, expected.ID, actual.ID, "ID mismatch")
	require.Equal(t, expected.Name, actual.Name, "Name mismatch")
	require.Equal(t, expected.Temp, actual.Temp, "Temperature mismatch")
	require.Equal(t, expected.Price, actual.Price, "Price mismatch")
}

func TestKitchen_PlaceOrder_PickUpOrder(t *testing.T) {
	hotOrder := css.Order{
		ID:        "hot1",
		Name:      "Hot Pizza",
		Temp:      string(TemperatureHot),
		Price:     10,
		Freshness: 600,
	}

	coldOrder := css.Order{
		ID:        "cold1",
		Name:      "Cold Salad",
		Temp:      string(TemperatureCold),
		Price:     5,
		Freshness: 900,
	}

	roomOrder := css.Order{
		ID:        "room1",
		Name:      "Room Sandwich",
		Temp:      string(TemperatureRoom),
		Price:     7,
		Freshness: 180,
	}

	invalidOrder := css.Order{
		ID:        "room2",
		Name:      "Room Sandwich",
		Price:     7,
		Freshness: 1200,
	}

	coldOrder2 := css.Order{
		ID:        "cold2",
		Name:      "Cold Salad",
		Temp:      string(TemperatureCold),
		Price:     5,
		Freshness: 0,
	}

	coldOrder3 := css.Order{
		ID:        "cold3",
		Name:      "Cold Salad",
		Temp:      string(TemperatureCold),
		Price:     5,
		Freshness: 0,
	}

	coldOrder4 := css.Order{
		ID:        "cold4",
		Name:      "Cold Salad",
		Temp:      string(TemperatureCold),
		Price:     5,
		Freshness: 30,
	}

	hotOrder2 := css.Order{
		ID:        "hot2",
		Name:      "Hot Pizza",
		Temp:      string(TemperatureHot),
		Price:     10,
		Freshness: 600,
	}

	decay := 2
	f, err := os.OpenFile("myorders", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(f, nil))

	t.Run("PlaceOrder/RoutesOrdersToPreferredStorage_WhenCapacityAvailable", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(coldOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)

		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		// Verify cold order was stored
		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder, pickColdOrder)
		require.Zero(t, k.cooler.len())

		// Verify hot order was stored
		pickupHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder, pickupHotOrder)
		require.Zero(t, k.heater.len())

		// Verify room order was stored
		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, roomOrder, pickupRoomOrder)
		require.Zero(t, k.shelf.len())

		// Verify invalidOrder was not stored
		pickupInvalidOrder, ok := k.PickUpOrder(invalidOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupInvalidOrder)
	})

	t.Run(
		"PlaceOrder/Shelf_DiscardPolicy_DiscardsShelfOrder_WhenAllStoragesAreFull", 
		func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(coldOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder4)

		require.Equal(t, 1, k.heater.len())
		require.Equal(t, 1, k.cooler.len())
		require.Equal(t, 1, k.shelf.len())

		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder, pickColdOrder)

		pickupHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder, pickupHotOrder)

		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupRoomOrder)

		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder4, pickupColdOrder4)
	})

	t.Run("PlaceOrder/MovesHotOrderFromShelfToHeater_WhenHeaterHasCapacity", func(t *testing.T) {
		k := NewKitchen(1, 1, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(hotOrder2)
		k.PlaceOrder(coldOrder)

		pickupHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder, pickupHotOrder)

		k.PlaceOrder(coldOrder2)

		pickupHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder2, pickupHotOrder2)
	})

	t.Run("PlaceOrder/MovesColdOrderFromShelfToCooler_WhenCoolerHasCapacity", func(t *testing.T) {
		k := NewKitchen(1, 1, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)

		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder, pickColdOrder)

		k.PlaceOrder(hotOrder2)

		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder4, pickupColdOrder4)

		pickupHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder2, pickupHotOrder2)
	})

	t.Run("PlaceOrder/Shelf_DiscardPolicy_DiscardsRoomOrder_ToPlaceColdOrder", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)

		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupRoomOrder)

		pickColdOrder, ok := k.PickUpOrder(coldOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder, pickColdOrder)

		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.True(t, ok)
		assertOrderMatch(t, coldOrder4, pickupColdOrder4)
	})

	t.Run("PlaceOrder/Shelf_DiscardPolicy_DiscardsRoomOrder_ToPlaceHotOrder", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)

		k.PlaceOrder(roomOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(hotOrder2)

		pickupRoomOrder, ok := k.PickUpOrder(roomOrder.ID)
		require.False(t, ok)
		require.Zero(t, pickupRoomOrder)

		pickHotOrder, ok := k.PickUpOrder(hotOrder.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder, pickHotOrder)

		pickHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder2, pickHotOrder2)
	})

	t.Run("PlaceOrder/DoesNotMoveColdOrderFromShelf_WhenCoolerIsFull", func(t *testing.T) {
		k := NewKitchen(1, 1, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)
		k.PlaceOrder(hotOrder2)

		pickupColdOrder4, ok := k.PickUpOrder(coldOrder4.ID)
		require.False(t, ok)
		require.Zero(t, pickupColdOrder4)

		pickUpHotOrder2, ok := k.PickUpOrder(hotOrder2.ID)
		require.True(t, ok)
		assertOrderMatch(t, hotOrder2, pickUpHotOrder2)
	})

	t.Run("PickUpOrder/Fails_WhenOrderExpiredInPreferredStorage", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)
		k.PlaceOrder(coldOrder2)
		require.Equal(t, 1, k.cooler.len())

		time.Sleep(1 * time.Millisecond)
		order, ok := k.PickUpOrder(coldOrder2.ID)

		require.False(t, ok)
		require.Zero(t, order)
		require.Equal(t, 0, k.cooler.len())
	})

	t.Run("PickUpOrder/Fails_WhenOrderExpiredInSecondaryStorage", func(t *testing.T) {
		k := NewKitchen(1, 1, 1, decay, logger)
		k.PlaceOrder(coldOrder2)
		require.Equal(t, 1, k.cooler.len())

		time.Sleep(1 * time.Millisecond)
		order, ok := k.PickUpOrder(coldOrder2.ID)

		require.False(t, ok)
		require.Zero(t, order)
		require.Equal(t, 0, k.cooler.len())
	})

	t.Run("PickUpOrder/Fails_WhenColdOrderExpiresOnShelf", func(t *testing.T) {
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
}
