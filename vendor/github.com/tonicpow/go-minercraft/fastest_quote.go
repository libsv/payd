package minercraft

import (
	"context"
	"errors"
	"sync"
)

// FastestQuote will check all known miners and return the fastest quote response
//
// Note: this might return different results each time if miners have the same rates as
// it's a race condition on which results come back first
func (c *Client) FastestQuote() (*FeeQuoteResponse, error) {

	// Get the fastest quote
	result := c.fetchFastestQuote()
	if result == nil {
		return nil, errors.New("no quotes found")
	}

	// Check for error?
	if result.Response.Error != nil {
		return nil, result.Response.Error
	}

	// Parse the response
	quote, err := result.parseQuote()
	if err != nil {
		return nil, err
	}

	// Return the quote
	return &quote, nil
}

// fetchFastestQuote will return a quote that is the quickest to resolve
func (c *Client) fetchFastestQuote() *internalResult {

	// The channel for the internal results
	resultsChannel := make(chan *internalResult, len(c.Miners))

	// Create a context (to cancel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Loop each miner (break into a Go routine for each quote request)
	var wg sync.WaitGroup
	for _, miner := range c.Miners {
		wg.Add(1)
		go func(ctx context.Context, client *Client, miner *Miner) {
			defer wg.Done()
			res := getQuote(ctx, client, miner)
			if res.Response.Error == nil {
				resultsChannel <- res
			}
		}(ctx, c, miner)
	}

	// Waiting for all requests to finish
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	return <-resultsChannel
}
