// Vendored from https://github.com/codermana/distributed-design-patterns-training
// examples/03-data-management-single/04_saga.go
package main

import (
	"fmt"
	"log"
)

type Step func() error

// #region steps
func OrderBooking() error {
	fmt.Println("Booking order...")
	return nil
}

func Payment() error {
	fmt.Println("Processing payment...")
	// return fmt.Errorf("card declined") // uncomment to watch compensations run
	return nil
}

func Shipping() error {
	fmt.Println("Shipping order...")
	return nil
}

// #endregion

// #region compensations
// compensations[i] undoes steps[i]: a semantic undo, not a database rollback
func CancelOrderBooking() error {
	fmt.Println("Cancelling order booking...")
	return nil
}

func CancelPayment() error {
	fmt.Println("Refunding payment...")
	return nil
}

func CancelShipping() error {
	fmt.Println("Cancelling shipping...")
	return nil
}

// #endregion

// #region run-saga
func RunSaga(steps []Step, compensations []Step) error {
	for i, step := range steps {
		err := step()
		if err != nil {
			fmt.Println("Error occurred, starting compensations...")
			// Execute compensating actions in reverse order
			for j := i - 1; j >= 0; j-- {
				compensations[j]()
			}
			return err
		}
	}
	return nil
}

// #endregion

// #region main
func main() {
	steps := []Step{OrderBooking, Payment, Shipping}
	compensations := []Step{CancelOrderBooking, CancelPayment, CancelShipping}

	err := RunSaga(steps, compensations)
	if err != nil {
		log.Println("Saga failed:", err)
	} else {
		fmt.Println("Saga completed successfully.")
	}
}

// #endregion
