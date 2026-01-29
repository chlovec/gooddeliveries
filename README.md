# README

Author: `<your name here>`

## How to run

The `Dockerfile` defines a self-contained Go reference environment. 
Build and run the program using [Docker](https://docs.docker.com/get-started/get-docker/):
```
$ docker build -t challenge .
$ docker run --rm -it challenge --auth=<token>
```
Feel free to modify the `Dockerfile` as you see fit.

If go `1.25` or later is locally installed, run the program directly for convenience:
```
$ go run main.go --auth=<token>
```

To run the tests and see the code coverage report, use the command below.
```
$ make test/rpt
```

## Discard criteria

When storage capacity is exhausted, the system performs a prioritized eviction from the overflow shelf. The selection process prioritizes Cold and Hot Temperature orders over Room Temperature orders. 
1. The system identifies the oldest Cold and Hot orders currently on the shelf.
2. If both exist, the older of the two is evicted; in the event of a tie, the Cold order is prioritized for eviction.
3. If only one type is present, that order is evicted regardless of age.
3. If no specialized orders are present, the system evicts the oldest Room Temperature order.

In creating this solution, I made assumption of what a valid order should be:
- ID is required
- Name is required
- Price must be greater than 0
- Temperature must be one of hot, cold or room (these are given)
- Freshness must be positive (we don't want to store food that has already decayed)
