// Vendored from https://github.com/codermana/distributed-design-patterns-training
// examples/04-data-management-multiple/03_transactional_outbox.go
package main

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// #region models
// Order and Outbox event structs
type Order struct {
	ID       uint    `json:"id"`
	Customer string  `json:"customer"`
	Total    float64 `json:"total"`
}

type OutboxEvent struct {
	ID        uint   `json:"id"`
	EventType string `json:"event_type"`
	OrderID   uint   `json:"order_id"`
	Published bool   `json:"published"`
}

// #endregion

// #region atomic-write
func createOrderWithEvent(db *gorm.DB, order Order) {
	tx := db.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		log.Fatal("Error creating order:", err)
	}

	event := OutboxEvent{EventType: "ORDER_PLACED", OrderID: order.ID}
	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		log.Fatal("Error creating outbox event:", err)
	}

	tx.Commit()
	fmt.Println("Order created and outbox event saved:", order)
}

// #endregion

// #region relay
// The relay: poll for unpublished events, publish them, then mark them done.
// Real systems run this in a separate process, or tail the database log
// with CDC (e.g. Debezium) instead of polling.
func relayOutboxEvents(db *gorm.DB, publish func(OutboxEvent)) {
	var events []OutboxEvent
	if err := db.Where("published = ?", false).Order("id").Find(&events).Error; err != nil {
		log.Fatal("Error fetching outbox events:", err)
	}

	for _, event := range events {
		publish(event) // a crash here means a redelivery later: at-least-once
		db.Model(&event).Update("published", true)
	}
}

// #endregion

func main() {
	// Set up SQLite database for both orders and outbox
	db, err := gorm.Open("sqlite3", "./order_outbox.db")
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	db.AutoMigrate(&Order{}, &OutboxEvent{}) // Migrate models

	// Create order and event in a transactional manner
	order := Order{Customer: "Alice", Total: 150.00}
	createOrderWithEvent(db, order)

	// Run one relay pass, standing in for the background publisher
	relayOutboxEvents(db, func(event OutboxEvent) {
		fmt.Println("Publishing event:", event.EventType, "for order", event.OrderID)
	})
}
