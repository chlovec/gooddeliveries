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
$ go run main.go --auth=agbo6gdrs1wd
```

## Discard criteria

When there is no available storage for a new placement request, the system looks for an item in the shelf to discard to make room for the new request. It looks for the first cooler-order and first-hot order on the shelf and picks one with the least freshness and discards that. If both have the same freshness, it discards the oldest.

If there is no cold or hot item on the shelf, it chooses the oldest shelf item and discards it.
