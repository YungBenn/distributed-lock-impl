package main

import (
    "context"
    "fmt"
    "time"

    "github.com/go-redsync/redsync/v4"
    "github.com/go-redsync/redsync/v4/redis/goredis/v9"
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
)

var (
    ctx = context.Background()
)

func main() {
    // Create a Redis client
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    pool := goredis.NewPool(client)

    // Create a Redsync instance
    rs := redsync.New(pool)

    userID := uuid.NewString()
    user := fmt.Sprintf("User-%s", userID)
    bookSeat(client, rs, "flight123", "seat45", user)
}

// SET Book data
func setBookData(client *redis.Client, flightID, seatID, userID string) {
	key := fmt.Sprintf("flight:%s:seat:%s", flightID, seatID)
	client.Set(ctx, key, userID, 5 * time.Minute)
}

// GET Book data
func getBookData(client *redis.Client, flightID, seatID string) string {
	key := fmt.Sprintf("flight:%s:seat:%s", flightID, seatID)
	orderID, err := client.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	return orderID
}

func bookSeat(client *redis.Client, rs *redsync.Redsync, flightID, seatID, user string) {
    // Create a lock key specific to the flight and seat
    lockKey := fmt.Sprintf("seat:%s:%s", flightID, seatID)
    mutex := rs.NewMutex(lockKey, redsync.WithExpiry(60*time.Second)) // Increased expiry time

    // Attempt to acquire the lock
    fmt.Printf("%s: %s attempting to acquire lock for seat %s\n", user, flightID, seatID)
    if err := mutex.LockContext(ctx); err != nil {
        fmt.Printf("%s: %s Error acquiring lock: %v\n", user, flightID, err)
        return
    }
    defer func() {
        fmt.Printf("%s: %s releasing lock for seat %s\n", user, flightID, seatID)
        if ok, err := mutex.UnlockContext(ctx); !ok || err != nil {
            fmt.Printf("%s: %s Error releasing lock: %v\n", user, flightID, err)
        } else {
            fmt.Printf("%s: %s Lock released successfully\n", user, flightID)
        }
    }()

	
    fmt.Printf("%s: %s Lock acquired, booking seat %s\n", user, flightID, seatID)

	bookData := getBookData(client, flightID, seatID)
	if bookData != "" {
		fmt.Printf("%s: %s Seat %s already booked with order ID: %s\n", user, flightID, seatID, bookData)
		return
	}
	
	setBookData(client, flightID, seatID, uuid.NewString())
	
	checkData := getBookData(client, flightID, seatID)
	if checkData == "" {
		fmt.Printf("%s: %s Error booking seat %s\n", user, flightID, seatID)
		return
	}

    // Simulate seat booking process
    time.Sleep(2 * time.Second)

    // Simulate updating the seat status in the database
    orderID := uuid.NewString()
    fmt.Printf("%s: %s Seat %s booked with order ID: %s\n", user, flightID, seatID, orderID)
}
