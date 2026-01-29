package kitchenv3

import (
	css "challenge/client"
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
		Temp:      "bad",
		Price:     7,
		Freshness: 1200,
	}

	coldOrder2 := css.Order{
		ID:        "cold2",
		Name:      "Cold Salad",
		Temp:      string(TemperatureCold),
		Price:     5,
		Freshness: 1,
	}

	coldOrder3 := css.Order{
		ID:        "cold3",
		Name:      "Cold Salad",
		Temp:      string(TemperatureCold),
		Price:     5,
		Freshness: 2,
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
	f, _ := os.OpenFile("myorders_v3", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	logger := slog.New(slog.NewTextHandler(f, nil))

	const one int64 = 1

	t.Run("PlaceOrder/RoutesOrdersToPreferredStorage_WhenCapacityAvailable", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)

		k.PlaceOrder(coldOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)

		err := k.PlaceOrder(invalidOrder)
		vErrs, ok := err.(ValidationErrors)
		require.True(t, ok, "Error should be of type ValidationErrors")
		require.Len(t, vErrs, 1)
		require.Equal(t, "1 validation errors occurred", err.Error())
		require.Equal(t, "must be one of hot, cold, or room", vErrs[0].Message)
		require.Equal(t, "Temp", vErrs[0].Field)

		require.Equal(t, one, k.heater.Len())
		require.Equal(t, one, k.cooler.Len())
		require.Equal(t, one, k.shelf.Len())

		// Verify cold order was stored
		pickColdOrder, err := k.PickUpOrder(coldOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, coldOrder, pickColdOrder)
		require.Zero(t, k.cooler.Len())

		// Verify hot order was stored
		pickupHotOrder, err := k.PickUpOrder(hotOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder, pickupHotOrder)
		require.Zero(t, k.heater.Len())

		// Verify room order was stored
		pickupRoomOrder, err := k.PickUpOrder(roomOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, roomOrder, pickupRoomOrder)
		require.Zero(t, k.shelf.Len())

		// Verify invalidOrder was not stored
		pickupInvalidOrder, err := k.PickUpOrder(invalidOrder.ID)
		require.Error(t, err)
		require.Zero(t, pickupInvalidOrder)
		require.Equal(t, "order not found", err.Error())
	})

	t.Run(
		"PlaceOrder/Shelf_DiscardPolicy_DiscardsShelfOrder_WhenAllStoragesAreFull",
		func(t *testing.T) {
			k := NewKitchen(one, one, one, decay, logger)

			k.PlaceOrder(coldOrder)
			k.PlaceOrder(hotOrder)
			k.PlaceOrder(roomOrder)
			k.PlaceOrder(coldOrder4)

			require.Equal(t, one, k.heater.Len())
			require.Equal(t, one, k.cooler.Len())
			require.Equal(t, one, k.shelf.Len())

			pickColdOrder, err := k.PickUpOrder(coldOrder.ID)
			require.Nil(t, err)
			assertOrderMatch(t, coldOrder, pickColdOrder)

			pickupHotOrder, err := k.PickUpOrder(hotOrder.ID)
			require.Nil(t, err)
			assertOrderMatch(t, hotOrder, pickupHotOrder)

			pickupRoomOrder, err := k.PickUpOrder(roomOrder.ID)
			require.Error(t, err)
			require.Zero(t, pickupRoomOrder)

			pickupColdOrder4, err := k.PickUpOrder(coldOrder4.ID)
			require.Nil(t, err)
			assertOrderMatch(t, coldOrder4, pickupColdOrder4)
		})

	t.Run("PlaceOrder/MovesHotOrderFromShelfToHeater_WhenHeaterHasCapacity", func(t *testing.T) {
		k := NewKitchen(one, one, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(hotOrder2)
		k.PlaceOrder(coldOrder)

		pickupHotOrder, err := k.PickUpOrder(hotOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder, pickupHotOrder)

		k.PlaceOrder(coldOrder2)

		pickupHotOrder2, err := k.PickUpOrder(hotOrder2.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder2, pickupHotOrder2)
	})

	t.Run("PlaceOrder/MovesColdOrderFromShelfToCooler_WhenCoolerHasCapacity", func(t *testing.T) {
		k := NewKitchen(one, one, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)

		pickColdOrder, err := k.PickUpOrder(coldOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, coldOrder, pickColdOrder)

		k.PlaceOrder(hotOrder2)

		pickupColdOrder4, err := k.PickUpOrder(coldOrder4.ID)
		require.Nil(t, err)
		assertOrderMatch(t, coldOrder4, pickupColdOrder4)

		pickupHotOrder2, err := k.PickUpOrder(hotOrder2.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder2, pickupHotOrder2)
	})

	t.Run("PlaceOrder/Shelf_DiscardPolicy_DiscardsRoomOrder_ToPlaceColdOrder", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)

		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)

		pickupRoomOrder, err := k.PickUpOrder(roomOrder.ID)
		require.Error(t, err)
		require.Zero(t, pickupRoomOrder)

		pickColdOrder, err := k.PickUpOrder(coldOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, coldOrder, pickColdOrder)

		pickupColdOrder4, err := k.PickUpOrder(coldOrder4.ID)
		require.Nil(t, err)
		assertOrderMatch(t, coldOrder4, pickupColdOrder4)
	})

	t.Run("PlaceOrder/Shelf_DiscardPolicy_DiscardsRoomOrder_ToPlaceHotOrder", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)

		k.PlaceOrder(roomOrder)
		k.PlaceOrder(hotOrder)
		k.PlaceOrder(hotOrder2)

		pickupRoomOrder, err := k.PickUpOrder(roomOrder.ID)
		require.Error(t, err)
		require.Zero(t, pickupRoomOrder)

		pickHotOrder, err := k.PickUpOrder(hotOrder.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder, pickHotOrder)

		pickHotOrder2, err := k.PickUpOrder(hotOrder2.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder2, pickHotOrder2)
	})

	t.Run("PlaceOrder/DoesNotMoveColdOrderFromShelf_WhenCoolerIsFull", func(t *testing.T) {
		k := NewKitchen(one, one, 2, decay, logger)

		k.PlaceOrder(hotOrder)
		k.PlaceOrder(roomOrder)
		k.PlaceOrder(coldOrder)
		k.PlaceOrder(coldOrder4)
		k.PlaceOrder(hotOrder2)

		pickupColdOrder4, err := k.PickUpOrder(coldOrder4.ID)
		require.Error(t, err)
		require.Zero(t, pickupColdOrder4)

		pickUpHotOrder2, err := k.PickUpOrder(hotOrder2.ID)
		require.Nil(t, err)
		assertOrderMatch(t, hotOrder2, pickUpHotOrder2)
	})

	t.Run("PickUpOrder/Fails_WhenOrderExpiredInPreferredStorage", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)
		k.PlaceOrder(coldOrder2)
		require.Equal(t, one, k.cooler.Len())

		time.Sleep(1 * time.Second)
		order, err := k.PickUpOrder(coldOrder2.ID)

		require.Error(t, err)
		require.Zero(t, order)
		require.Zero(t, k.cooler.Len())
	})

	t.Run("PickUpOrder/Fails_WhenOrderExpiredInSecondaryStorage", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)
		k.PlaceOrder(coldOrder2)
		require.Equal(t, one, k.cooler.Len())

		time.Sleep(1 * time.Second)
		order, err := k.PickUpOrder(coldOrder2.ID)

		require.Error(t, err)
		require.Zero(t, order)
		require.Zero(t, k.cooler.Len())
	})

	t.Run("PickUpOrder/Fails_WhenColdOrderExpiresOnShelf", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)
		k.PlaceOrder(coldOrder2)
		k.PlaceOrder(coldOrder3)

		require.Equal(t, one, k.cooler.Len())
		require.Equal(t, one, k.shelf.Len())

		time.Sleep(2 * time.Second)
		order, err := k.PickUpOrder(coldOrder3.ID)

		require.Equal(t, one, k.cooler.Len())
		require.Zero(t, k.shelf.Len())

		require.Error(t, err)
		require.Zero(t, order)
	})

	t.Run("PlaceOrder/ReturnsValidationError_WhenOrderIsInvalid", func(t *testing.T) {
		k := NewKitchen(one, one, one, decay, logger)
		invalidOrder := css.Order{}

		err := k.PlaceOrder(invalidOrder)

		require.Error(t, err)

		require.Zero(t, k.heater.Len())
		require.Zero(t, k.cooler.Len())
		require.Zero(t, k.shelf.Len())

		vErrs, ok := err.(ValidationErrors)
		require.True(t, ok, "Error should be of type ValidationErrors")
		require.Len(t, vErrs, 5)

		require.Equal(t, "ID", vErrs[0].Field)
		require.Contains(t, vErrs[0].Message, "is required")

		require.Equal(t, "Name", vErrs[1].Field)
		require.Contains(t, vErrs[1].Message, "is required")

		require.Equal(t, "Temp", vErrs[2].Field)
		require.Contains(t, vErrs[2].Message, "must be one of hot, cold, or room")

		require.Equal(t, "Price", vErrs[3].Field)
		require.Equal(t, "must be greater than zero", vErrs[3].Message)

		require.Equal(t, "Freshness", vErrs[4].Field)
		require.Equal(t, "must be positive", vErrs[4].Message)

		require.ErrorContains(t, err, "5 validation errors occurred")
	})
}
