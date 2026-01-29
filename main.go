package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	css "challenge/client"
	kitchen "challenge/kitchenv3"
)

var (
	endpoint = flag.String("endpoint", "https://api.cloudkitchens.com", "Problem server endpoint")
	auth     = flag.String("auth", "", "Authentication token (required)")
	name     = flag.String("name", "", "Problem name. Leave blank (optional)")
	seed     = flag.Int64("seed", 0, "Problem seed (random if zero)")

	rate = flag.Duration("rate", 500*time.Millisecond, "Inverse order rate")
	min  = flag.Duration("min", 4*time.Second, "Minimum pickup time")
	max  = flag.Duration("max", 8*time.Second, "Maximum pickup time")

	// coolerCapacity = flag.Int("cooler", 6, "Cooler capacity")
	// heaterCapacity = flag.Int("heater", 6, "Heater capacity")
	// shelfCapacity  = flag.Int("shelf", 12, "Shelf capacity")

	coolerCapacity = flag.Int64("cooler", 6, "Cooler capacity")
	heaterCapacity = flag.Int64("heater", 6, "Heater capacity")
	shelfCapacity  = flag.Int64("shelf", 12, "Shelf capacity")

	decayFactor = flag.Int("decay", 2, "Shelf decay multiplier")
)

func parseLogsToActions(buf *bytes.Buffer) ([]css.Action, error) {
	var actions []css.Action
	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		var raw map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			return nil, err
		}

		var ts int64
		if t, ok := raw["time"].(string); ok {
			parsed, err := time.Parse(time.RFC3339Nano, t)
			if err != nil {
				return nil, err
			}
			ts = parsed.UnixMicro()
		}

		action := css.Action{
			Timestamp: ts,
			ID:        fmt.Sprint(raw["order id"]),
			Action:    fmt.Sprint(raw["msg"]),
			Target:    fmt.Sprint(raw["target"]),
		}

		actions = append(actions, action)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return actions, nil
}

func main() {
	flag.Parse()

	client := css.NewClient(*endpoint, *auth)
	id, orders, err := client.New(*name, *seed)
	if err != nil {
		log.Fatalf("Failed to fetch test problem: %v", err)
	}

	// ------ Execution harness logic goes here using rate, min and max ------
	var buf bytes.Buffer
	kitchen := kitchen.NewKitchen(
		*heaterCapacity,
		*coolerCapacity,
		*shelfCapacity,
		*decayFactor,
		slog.New(slog.NewJSONHandler(&buf, nil)),
	)

	ticker := time.NewTicker(*rate)
	defer ticker.Stop()

	var wg sync.WaitGroup

	// var actions []css.Action
	for _, order := range orders {
		<-ticker.C

		log.Printf("Received: %+v", order)

		kitchen.PlaceOrder(order)
		wg.Add(1)
		go func(o css.Order) {
			defer wg.Done()

			randomDelay := *min + rand.N(*max-*min)
			time.Sleep(randomDelay)

			kitchen.PickUpOrder(o.ID)
		}(order)
	}

	wg.Wait()

	actions, err := parseLogsToActions(&buf)
	if err != nil {
		print(err)
		log.Fatalf("Failed to parse logs: %v", err)
	}

	// ------------------------------------------------------------------------

	result, err := client.Solve(id, *rate, *min, *max, actions)
	if err != nil {
		log.Fatalf("Failed to submit test solution: %v", err)
	}
	log.Printf("Test result: %v", result)
}

func printLogs(actions []css.Action) {
	fmt.Printf("%-20s | %-10s | %-10s | %-10s\n", "TIMESTAMP", "ACTION", "ORDER ID", "TARGET")
	fmt.Println(strings.Repeat("-", 60))

	for _, a := range actions {
		fmt.Printf("%-20d | %-10s | %-10s | %-10s\n",
			a.Timestamp,
			a.Action,
			a.ID,
			a.Target,
		)
	}
}

// import (
// 	"challenge/client"
// 	kitchen "challenge/kitchenv3"
// 	// kitchen "challenge/kitchen"
// 	"log/slog"
// 	"math/rand/v2"
// 	"os"
// 	"sync"
// 	"time"
// )

// type LogContent struct {
// 	Time    time.Time `json:"time"`
// 	Level   string    `json:"level"`
// 	Msg     string    `json:"msg"`
// 	OrderID string    `json:"order id"`
// 	Target  string    `json:"target"`
// }

// func getData() []client.Order {
// 	return []client.Order{
// 		{ID: "ngcnw", Name: "Coconut", Temp: "room", Price: 20, Freshness: 34},
// 		{ID: "ku5pn", Name: "Cookie Dough", Temp: "cold", Price: 18, Freshness: 39},
// 		{ID: "kxhxa", Name: "Turkey Sandwich", Temp: "cold", Price: 18, Freshness: 29},
// 		{ID: "j667k", Name: "BBQ Pizza", Temp: "hot", Price: 7, Freshness: 55},
// 		{ID: "89kh5", Name: "Danish Pastry", Temp: "room", Price: 6, Freshness: 54},
// 		{ID: "543au", Name: "Burrito", Temp: "hot", Price: 8, Freshness: 26},
// 		{ID: "oexb9", Name: "Coconut", Temp: "room", Price: 8, Freshness: 49},
// 		{ID: "9sf4x", Name: "Yoghurt", Temp: "cold", Price: 19, Freshness: 24},
// 		{ID: "i9bzh", Name: "Gas Station Sushi", Temp: "room", Price: 13, Freshness: 34},
// 		{ID: "y7ccb", Name: "Pork Chop", Temp: "hot", Price: 20, Freshness: 22},
// 		{ID: "z76hy", Name: "Chicken Tacos", Temp: "hot", Price: 18, Freshness: 52},
// 		{ID: "gpwif", Name: "Raspberries", Temp: "room", Price: 8, Freshness: 58},
// 		{ID: "nmq6g", Name: "Mixed Greens", Temp: "cold", Price: 10, Freshness: 59},
// 		{ID: "83g19", Name: "Burrito", Temp: "hot", Price: 8, Freshness: 56},
// 		{ID: "btj85", Name: "Spaghetti", Temp: "hot", Price: 8, Freshness: 22},
// 		{ID: "kkppd", Name: "Mixed Greens", Temp: "cold", Price: 7, Freshness: 21},
// 		{ID: "xrz34", Name: "Gas Station Sushi", Temp: "room", Price: 17, Freshness: 32},
// 		{ID: "ys71g", Name: "Chocolate Gelato", Temp: "cold", Price: 9, Freshness: 59},
// 		{ID: "paicj", Name: "Hamburger", Temp: "hot", Price: 13, Freshness: 41},
// 		{ID: "884uz", Name: "Tuna Sandwich", Temp: "cold", Price: 19, Freshness: 43},
// 		{ID: "sjswr", Name: "Sushi", Temp: "cold", Price: 16, Freshness: 53},
// 		{ID: "6tnab", Name: "Lukewarm Coke", Temp: "room", Price: 9, Freshness: 30},
// 		{ID: "a16b3", Name: "Cookie Dough", Temp: "cold", Price: 13, Freshness: 46},
// 		{ID: "j31df", Name: "Vanilla Ice Cream", Temp: "cold", Price: 15, Freshness: 26},
// 		{ID: "kygry", Name: "Coconut", Temp: "room", Price: 14, Freshness: 45},
// 		{ID: "5mc9i", Name: "Pressed Juice", Temp: "cold", Price: 5, Freshness: 39},
// 		{ID: "8h6bb", Name: "Italian Meatballs", Temp: "hot", Price: 6, Freshness: 48},
// 		{ID: "x1x84", Name: "Whole Wheat Bread", Temp: "room", Price: 6, Freshness: 25},
// 		{ID: "psuno", Name: "Pad Thai", Temp: "hot", Price: 8, Freshness: 32},
// 		{ID: "br7hh", Name: "Tomato Soup", Temp: "hot", Price: 4, Freshness: 38},
// 		{ID: "nz638", Name: "Burrito", Temp: "hot", Price: 9, Freshness: 40},
// 		{ID: "7eo3m", Name: "Gas Station Sushi", Temp: "room", Price: 16, Freshness: 59},
// 		{ID: "u989e", Name: "Orange Sherbet", Temp: "cold", Price: 16, Freshness: 37},
// 		{ID: "kibgn", Name: "Stale Bread", Temp: "room", Price: 10, Freshness: 52},
// 		{ID: "qghg3", Name: "Dry Biscuits", Temp: "room", Price: 6, Freshness: 39},
// 		{ID: "hhuz1", Name: "Lasagna", Temp: "hot", Price: 18, Freshness: 35},
// 		{ID: "8yrfb", Name: "Stale Bread", Temp: "room", Price: 18, Freshness: 28},
// 		{ID: "qzw54", Name: "Gas Station Sushi", Temp: "room", Price: 19, Freshness: 24},
// 		{ID: "cn3bc", Name: "Mixed Greens", Temp: "cold", Price: 18, Freshness: 58},
// 		{ID: "8ccs3", Name: "Turkey Sandwich", Temp: "cold", Price: 20, Freshness: 39},
// 		{ID: "dyp5f", Name: "Fish Tacos", Temp: "hot", Price: 20, Freshness: 42},
// 		{ID: "q7u9e", Name: "Sushi", Temp: "cold", Price: 20, Freshness: 27},
// 		{ID: "aqbz8", Name: "Gas Station Sushi", Temp: "room", Price: 14, Freshness: 29},
// 		{ID: "ahcf7", Name: "Popsicle", Temp: "cold", Price: 15, Freshness: 41},
// 		{ID: "c8nm3", Name: "Mac & Cheese", Temp: "hot", Price: 11, Freshness: 23},
// 		{ID: "qr6nf", Name: "Fizzed-Out Pepsi", Temp: "room", Price: 15, Freshness: 55},
// 		{ID: "ssrts", Name: "Pork Chop", Temp: "hot", Price: 7, Freshness: 45},
// 		{ID: "3qepk", Name: "Kebab", Temp: "hot", Price: 13, Freshness: 37},
// 	}
// }

// // const logFile string = "kitchen_service_log_v1"
// // const heaterCapacity int = 6
// // const coolerCapacity int = 6
// // const shelfCapacity int = 12

// const logFile string = "kitchen_service_log_v3"
// const heaterCapacity int64 = 6
// const coolerCapacity int64 = 6
// const shelfCapacity int64 = 12

// const decayFactor int = 2
// const rate time.Duration = 500*time.Millisecond
// const min time.Duration = 4*time.Second
// const max time.Duration = 8*time.Second

// func PlacePickupOrder() {
// 	f, _ := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	logger := slog.New(slog.NewJSONHandler(f, nil))

// 	kitchen := kitchen.NewKitchen(
// 		heaterCapacity,
// 		coolerCapacity,
// 		shelfCapacity,
// 		decayFactor,
// 		logger,
// 	)

// 	ticker := time.NewTicker(rate)
// 	defer ticker.Stop()

// 	var wg sync.WaitGroup

// 	data := getData()
// 	for _, order := range data {
// 		<-ticker.C

// 		kitchen.PlaceOrder(order)
// 		wg.Add(1)
// 		go func(o client.Order) {
// 			defer wg.Done()

// 			randomDelay := min + rand.N(max-min)
// 			time.Sleep(randomDelay)

// 			kitchen.PickUpOrder(o.ID)
// 		}(order)
// 	}

// 	wg.Wait()
// }

// func main() {
// 	PlacePickupOrder()
// }
