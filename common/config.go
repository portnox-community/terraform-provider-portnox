package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	APIKey        string
	BaseURL       string
	Logger        *log.Logger
	Retries       int // Number of retries for API requests
	RetryInterval int // Retry interval in seconds between retries
}

func NewConfig(apiKey string, baseURL string, retries int, retryInterval int, logger *log.Logger) *Config {
	if apiKey == "" {
		apiKey = os.Getenv("TF_VAR_PORTNOX_API_KEY")
	}

	return &Config{
		APIKey:        apiKey,
		BaseURL:       baseURL,
		Retries:       retries,
		RetryInterval: retryInterval,
		Logger:        logger,
	}
}

func (c *Config) MakeRequest(method, endpoint string, payload interface{}) ([]byte, error) {
	url := c.BaseURL + endpoint

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	maskedAPIKey := c.APIKey[:1] + "*************************" + c.APIKey[len(c.APIKey)-1:]

	requestLog := map[string]interface{}{
		"method": method,
		"url":    url,
		"headers": map[string]string{
			"Authorization": "Bearer " + maskedAPIKey,
			"Content-Type":  "application/json",
		},
		"body": string(body),
	}

	if logJSON, err := json.MarshalIndent(requestLog, "", "  "); err == nil {
		if c.Logger != nil {
			c.Logger.Printf("[DEBUG] Full API Request:\n%s", logJSON)
		} else {
			log.Printf("[DEBUG] Full API Request:\n%s", logJSON)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Printf("[ERROR] HTTP request failed: %v", err)
		} else {
			log.Printf("[ERROR] HTTP request failed: %v", err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseLog := map[string]interface{}{
		"status":  resp.Status,
		"headers": resp.Header,
		"body":    string(responseBody),
	}

	if logJSON, err := json.MarshalIndent(responseLog, "", "  "); err == nil {
		if c.Logger != nil {
			c.Logger.Printf("[DEBUG] Full API Response:\n%s", logJSON)
		} else {
			log.Printf("[DEBUG] Full API Response:\n%s", logJSON)
		}
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	return responseBody, nil
}

// IsNotFoundError checks if an error corresponds to a 404 Not Found response
func (c *Config) IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for 404 status in the error message
	if strings.Contains(err.Error(), "404") {
		return true
	}

	// Check for specific 400 error with InternalErrorCode 5357
	if strings.Contains(err.Error(), "400") {
		var errorResponse struct {
			InternalErrorCode int `json:"InternalErrorCode"`
		}
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			if errorResponse.InternalErrorCode == 5357 {
				return true
			}
		}
	}

	return false
}

func (c *Config) MakeRequestWithRetry(method, endpoint string, payload interface{}) ([]byte, error) {
	var responseBody []byte
	var err error
	backoff := c.RetryInterval // Initial backoff in seconds, based on RetryInterval

	if c.Logger != nil {
		c.Logger.Printf("[DEBUG] Starting MakeRequestWithRetry with maxRetries=%d and retry_interval=%d", c.Retries, c.RetryInterval)
	} else {
		log.Printf("[DEBUG] Starting MakeRequestWithRetry with maxRetries=%d and retry_interval=%d", c.Retries, c.RetryInterval)
	}

	for attempt := 1; attempt <= c.Retries; attempt++ {
		if c.Logger != nil {
			c.Logger.Printf("[DEBUG] Attempt %d/%d: Making request to %s", attempt, c.Retries, endpoint)
		} else {
			log.Printf("[DEBUG] Attempt %d/%d: Making request to %s", attempt, c.Retries, endpoint)
		}

		responseBody, err = c.MakeRequest(method, endpoint, payload)
		if err == nil {
			if c.Logger != nil {
				c.Logger.Printf("[DEBUG] Request succeeded on attempt %d", attempt)
			} else {
				log.Printf("[DEBUG] Request succeeded on attempt %d", attempt)
			}
			return responseBody, nil
		}

		// Check if the error is a 429 Too Many Requests
		if strings.Contains(err.Error(), "429") {
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond // Add random jitter up to 1 second
			if c.Logger != nil {
				c.Logger.Printf("[WARN] Received 429 Too Many Requests. Retrying in %d seconds with jitter (attempt %d/%d)...", backoff, attempt, c.Retries)
			} else {
				log.Printf("[WARN] Received 429 Too Many Requests. Retrying in %d seconds with jitter (attempt %d/%d)...", backoff, attempt, c.Retries)
			}
			time.Sleep(time.Duration(backoff)*time.Second + jitter)
			backoff *= 2 // Exponential backoff
			continue
		}

		// If the error is not retryable, log and break the loop
		if c.Logger != nil {
			c.Logger.Printf("[ERROR] Non-retryable error encountered: %v", err)
		} else {
			log.Printf("[ERROR] Non-retryable error encountered: %v", err)
		}
		break
	}

	if c.Logger != nil {
		c.Logger.Printf("[ERROR] All retry attempts failed. Returning last error: %v", err)
	} else {
		log.Printf("[ERROR] All retry attempts failed. Returning last error: %v", err)
	}

	return responseBody, err
}
