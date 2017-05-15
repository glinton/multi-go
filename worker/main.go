package main

import (
	"fmt"
	"time"
	"os"

	"github.com/garyburd/redigo/redis"
)

var redisConn redis.Conn
var debug bool = false

func init() {
	redisAddr := os.Getenv("DATA_CACHE_HOST")
	if redisAddr == "" {
		redisAddr = "127.0.0.1"
	}
	
	url := fmt.Sprintf("redis://%s:6379/0", redisAddr)

	var err error
	redisConn, err = redis.DialURL(url)
  	if err != nil {
    	fmt.Printf("Failed to reach redis - %s\n", err)
    	os.Exit(1)
  	}

  	if os.Getenv("DEBUG") == "true" {
  		debug = true
  	}
}

func main() {
	for range time.Tick(2 * time.Second) {
		// fmt.Println("Working")
		work()
	}
}

func work() {
	start := time.Now()

	var sold interface{}
	var err error
	count := 0

	sold, err = redisConn.Do("LPOP", "sold")
	for sold != nil {
		if err != nil {
			fmt.Printf("Failed to get purchase to process - %s\n", err)
			return
		}

		if sold == nil {
			return
		}

		// fmt.Printf("Shipping - %s\n", sold)
		_, err = redisConn.Do("RPUSH", "shipped", sold)
		if err != nil {
			fmt.Printf("Failed to process purchase - %s\n", err)
			return
		}
		// fmt.Printf("Shipped - %s\n", sold)

		count++
		sold, err = redisConn.Do("LPOP", "sold")
	}

	if count != 0 {
		fmt.Printf("Processed %d in %s\n", count, time.Since(start))
	}
}
