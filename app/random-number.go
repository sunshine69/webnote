package app

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func GenerateRandom(max uint64) uint64 {
	f, err := os.Open("/dev/random")
	if err != nil {
		return 0
	}
	defer f.Close()

	var b [8]byte
	if _, err := io.ReadFull(f, b[:]); err != nil {
		return 0
	}

	num := binary.BigEndian.Uint64(b[:])

	if max == 0 {
		return num
	}
	if num <= max {
		return num
	} else {
		return num % (max + 1)
	}
}

func GenRandNumber(w http.ResponseWriter, r *http.Request) {
	gen_number := GenerateRandom(999999999999)
	fmt.Fprintf(w, "%d", gen_number)
}

// These a re new way to use golang for that cast however it seems not to be effective (test using weather forcasts)
// In theory it should be the most close simulation of the real casting coins but not sure why?
// This is simpler compare with solution given by Claude.ai
// It seems the universe is nothing 'concurrent' as the random gen device can only truly randome per each request anyway confirmed by ai
func CastOneLine(w http.ResponseWriter, r *http.Request) {
	startCh := make(chan struct{}) // Unbuffered channel for synchronization
	results := make([]chan int, 3)

	// Launch three goroutines that wait on startCh
	for i := range results {
		results[i] = make(chan int)
		go func(ch chan int) {
			<-startCh // Block until the channel is closed
			// Generate a random bit (0 or 1)
			b := make([]byte, 1)
			if _, err := rand.Read(b); err != nil {
				fmt.Fprintf(w, "Error: %v\n", err)
				return
			}
			ch <- int(b[0] & 1) // 0 for no letter, 1 for letter
		}(results[i])
	}

	// Close the channel to unblock all goroutines at once
	close(startCh)

	// Collect results and sum them up (0-3)
	sum := 0
	for _, ch := range results {
		sum += <-ch
	}

	fmt.Fprintf(w, "%d", sum)
}

func CastOneLineClaude(w http.ResponseWriter, r *http.Request) {
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
