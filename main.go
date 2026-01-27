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
	"challenge/kitchen"
)

var (
	endpoint = flag.String("endpoint", "https://api.cloudkitchens.com", "Problem server endpoint")
	auth     = flag.String("auth", "", "Authentication token (required)")
	name     = flag.String("name", "", "Problem name. Leave blank (optional)")
	seed     = flag.Int64("seed", 0, "Problem seed (random if zero)")

	rate = flag.Duration("rate", 500*time.Millisecond, "Inverse order rate")
	min  = flag.Duration("min", 4*time.Second, "Minimum pickup time")
	max  = flag.Duration("max", 8*time.Second, "Maximum pickup time")

	coolerCapacity = flag.Int("cooler", 6, "Cooler capacity")
	heaterCapacity = flag.Int("heater", 6, "Heater capacity")
	shelfCapacity  = flag.Int("shelf", 12, "Shelf capacity")

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
		go func (o css.Order)  {
			defer wg.Done()

			randomDelay := *min + rand.N(*max - *min)
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
