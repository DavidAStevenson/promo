# promo

Toy golang project simulating a "promotion" feature (e.g. for an e-commerce site), including tests that enable time mocking

## Considerations

* I use goroutines and Timers to await the start and end of promo display periods
* For speedy testing of time related tests, I decided to use the "github.com/benbjohnson/clock" mock package
	* It supports mocking of time and it has Timer support as well
	* There are surely other options; this is just the first I found that met my needs
* Promotion model is very simple, just a string to represent a name for the promotion. Otherwise, it's a time period.

## Usage

Go get the time mocking package
```
go get -u github.com/benbjohnson/clock
```

Then import the "promo" package 

```go
import "promo"
```

Use func `promo.New` to create a promo. `New` needs a name, and a pair of `time.Time`s, indicating the start and end of the display period, as arguments. The following example creates a promo with a (very) short duration of 3 seconds, starting 2 seconds from now:

```go
now := time.Now()
start := now.Add(2 * time.Second)
end := now.Add(5 * time.Second)

p, _ := promo.New("Demo promo", start, end) // error return ignored for brevity
```

Use func `promo.Choose` to select the appropriate promo to be displayed, for the given IP address. `promo.Choose` will select from those promos currently within their display period (except when QA IP addresses are specified).
```
p := promo.Choose("192.168.0.1")
```

A promo's name and current status can be displayed with `Name()` and `Status()`, and whether it may be displayed or not is indicated by `AllowDisplay`. 

A simple, single promo example program (runs for 10 seconds) is found in main.go:

```
go run cmd/main.go
```

## Running tests

Run tests as follows:
```
go test -v promo

```
For simpler pass/fail output:
```
go test promo
```

## Issues

* This package is not currently thread-safe. (One might employ Mutexes around shared data structures to make that possible, but considering other concurrency models is another idea)
* For very large numbers of promos, func Choose would become slow due to linear search algo used
