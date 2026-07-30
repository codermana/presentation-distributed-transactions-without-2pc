// Vendored from https://github.com/codermana/distributed-design-patterns-training
// examples/04-data-management-multiple/02_cqrs.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var ctx = context.Background()

// #region write-model
// Define the order struct for the write model (Command Side)
type Order struct {
	ID       uint    `json:"id"`
	Customer string  `json:"customer"`
	Total    float64 `json:"total"`
}

// Command side: Save order to the database (write model)
func createOrder(db *gorm.DB, order Order) {
	if err := db.Create(&order).Error; err != nil {
		log.Fatal("Error creating order:", err)
	}
	fmt.Println("Order created:", order)
}

// #endregion

// #region read-model
// Query side: Get order from the cache (read model)
func getOrderFromCache(client *redis.Client, orderID string) (string, error) {
	val, err := client.Get(ctx, orderID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("order not found in cache")
		}
		return "", err
	}
	return val, nil
}

// #endregion

// #region projection
// Query side: Cache order details
func cacheOrder(client *redis.Client, orderID string, orderDetails string) {
	err := client.Set(ctx, orderID, orderDetails, 0).Err()
	if err != nil {
		log.Fatal("Error caching order:", err)
	}
	fmt.Println("Order cached:", orderID)
}

// #endregion

func main() {
	// Set up SQLite database for write model (Order storage)
	db, err := gorm.Open("sqlite3", "./orders.db")
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	db.AutoMigrate(&Order{}) // Migrate the Order model

	// Set up Redis client for caching
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create an order (command side)
	order := Order{Customer: "Alice", Total: 150.00}
	createOrder(db, order)

	// Cache the order (query side)
	cacheOrder(client, fmt.Sprintf("%d", order.ID), fmt.Sprintf("Order %d for %s", order.ID, order.Customer))

	// Retrieve order from cache (query side)
	cachedOrder, err := getOrderFromCache(client, fmt.Sprintf("%d", order.ID))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Cached order:", cachedOrder)
}

// **Explanation:**
// - The **Command side** saves the order to a SQLite database.
// - The **Query side** retrieves order data from Redis to serve fast queries.
// - The system segregates write and read models for better performance and scalability.
