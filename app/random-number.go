package app

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sync"
)

func GenRandNumber(w http.ResponseWriter, r *http.Request) {
	// Generate a random integer
	// var randomInt uint64
	// err := binary.Read(rand.Reader, binary.BigEndian, &randomInt)
	// if err != nil {
	// 	fmt.Fprintf(w, "Error generating random integer: %s\n", err.Error())
	// 	return
	// }
	gen_number, _ := rand.Int(rand.Reader, big.NewInt(999999999999))
	fmt.Fprintf(w, "%d", gen_number)
}

// These a re new way to sue golang for that cast however it seems not to be effective (test using weather forcasts)
// In theory it should be the most close simulation of the real casting coins but not sure why?
func CastOneline(w http.ResponseWriter, r *http.Request) {
	gen_number, err := CastTrigram()
	if err != nil {
		fmt.Fprintf(w, "Error generating random number: %s\n", err.Error())
		return
	}
	fmt.Fprintf(w, "%d", gen_number)
}

// CastCoin simulates casting a single coin and returns the result:
// 0 for "letter face" up, 1 for "no letter face" up
func CastCoin() (int, error) {
	// Generate a random byte - this is the most efficient way to get a random bit
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %v", err)
	}

	// Use just the least significant bit for an unbiased coin flip
	return int(b[0] & 1), nil
}

// CastTrigram simulates casting three coins concurrently and returns the count
// of "no letter face" up coins (value between 0-3)
func CastTrigram() (int, error) {
	var wg sync.WaitGroup
	results := make(chan int, 3)
	errors := make(chan error, 3)

	// Create ready channel to synchronize the start
	ready := make(chan struct{})

	// Set up all three goroutines at once
	wg.Add(3)

	// Create all goroutines first but they wait for the signal
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()

			// Wait for the signal - all goroutines will be unblocked simultaneously
			<-ready

			result, err := CastCoin()
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	// Now send the signal to start all goroutines simultaneously
	close(ready)

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Check for errors
	for err := range errors {
		if err != nil {
			return 0, err
		}
	}

	// Count the number of "no letter face" up (1s)
	count := 0
	for result := range results {
		count += result
	}

	return count, nil
}

// CastHexagram casts a complete hexagram by performing six trigram casts
// Returns a slice of 6 integers (0-3) representing the count of "no letter face" up coins for each line
func CastHexagram() ([]int, error) {
	hexagram := make([]int, 6)

	for i := 0; i < 6; i++ {
		result, err := CastTrigram()
		if err != nil {
			return nil, fmt.Errorf("error casting line %d: %v", i+1, err)
		}
		hexagram[i] = result
	}

	return hexagram, nil
}
