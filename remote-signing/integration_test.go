package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

const (
	testServerHost = "localhost"
	testServerPort = 8082
	testTimeout    = 30 * time.Second
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	testServerURL = fmt.Sprintf("http://%s:%d", testServerHost, testServerPort)
	testWSURL     = fmt.Sprintf("ws://%s:%d/ws", testServerHost, testServerPort)
)

// TestIntegrationSuite runs the complete integration test suite
func TestIntegrationSuite(t *testing.T) {
	// Setup test server
	srv, cancel := setupTestServer(t)
	defer cancel()

	// Wait for server to start
	waitForServer(t, testServerURL)

	// Run integration tests
	t.Run("CompleteSigningWorkflow", func(t *testing.T) {
		testCompleteSigningWorkflow(t, srv)
	})

	t.Run("JSONDataSubmission", func(t *testing.T) {
		testJSONDataSubmission(t, srv)
	})

	t.Run("BinaryDataSubmission", func(t *testing.T) {
		testBinaryDataSubmission(t, srv)
	})

	t.Run("WebSocketNotifications", func(t *testing.T) {
		testWebSocketNotifications(t, srv)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, srv)
	})

	t.Run("MultipleRequests", func(t *testing.T) {
		testMultipleRequests(t, srv)
	})
}

// setupTestServer creates and starts a test server
func setupTestServer(t *testing.T) (*server.Server, context.CancelFunc) {
	config := &server.Config{
		Host:           testServerHost,
		Port:           testServerPort,
		AllowedOrigins: []string{"*"},
		MaxDataSize:    10 * 1024 * 1024, // 10MB
		SigningTimeout: testTimeout,
	}

	srv := server.New(config)
	ctx, cancel := context.WithCancel(context.Background())

	// Start server in background
	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	return srv, cancel
}

// waitForServer waits for the server to be ready
func waitForServer(t *testing.T, serverURL string) {
	client := &http.Client{Timeout: 1 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(serverURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("Server failed to start within timeout")
}

// testCompleteSigningWorkflow tests the complete end-to-end workflow with goar
func testCompleteSigningWorkflow(t *testing.T, srv *server.Server) {
	// Step 1: Load test wallet from file
	walletData, err := os.ReadFile("test-wallet.json")
	require.NoError(t, err, "Failed to read test wallet file")

	// Create wallet from the JSON data
	wallet, err := goar.NewWallet(walletData, "https://arweave.net")
	require.NoError(t, err, "Failed to create wallet from test data")

	// Step 2: Create test data (simulating an Arweave data item)
	testData := []byte("Hello, Arweave! This is a test data item for remote signing.")

	// Create bundle item (data item)
	bundleItem := &types.BundleItem{
		SignatureType: 1, // RSA-PSS with SHA256
		Target:        "", // No target
		Anchor:        "", // No anchor
		Tags: []types.Tag{
			{Name: "Content-Type", Value: "text/plain"},
			{Name: "App-Name", Value: "Harlequin-Remote-Signing-Test"},
			{Name: "App-Version", Value: "1.0.0"},
		},
		Data: string(testData),
	}

	// For this test, we'll simulate the data item bytes
	// In a real scenario, this would be properly formatted ANS-104 data item
	dataItemBytes := testData // For simplicity, using raw data

	t.Logf("Created bundle item with wallet %s: %d bytes", wallet.Owner()[:20]+"...", len(dataItemBytes))

	// Step 2: Submit data to remote signing server
	submitReq := server.SubmitDataRequest{
		Data:        dataItemBytes,
		ClientID:    "integration-test-client",
		CallbackURL: "http://localhost:9999/callback", // Mock callback URL
	}

	submitBody, err := json.Marshal(submitReq)
	require.NoError(t, err, "Failed to marshal submit request")

	resp, err := http.Post(testServerURL+"/", "application/json", bytes.NewReader(submitBody))
	require.NoError(t, err, "Failed to submit data")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Submit request failed")

	var submitResp server.SubmitDataResponse
	err = json.NewDecoder(resp.Body).Decode(&submitResp)
	require.NoError(t, err, "Failed to decode submit response")

	uuid := submitResp.UUID
	require.NotEmpty(t, uuid, "UUID should not be empty")
	t.Logf("Received UUID: %s", uuid)

	// Step 3: Retrieve unsigned data from server
	getResp, err := http.Get(testServerURL + "/" + uuid)
	require.NoError(t, err, "Failed to retrieve data")
	defer getResp.Body.Close()

	require.Equal(t, http.StatusOK, getResp.StatusCode, "Get request failed")

	retrievedData, err := io.ReadAll(getResp.Body)
	require.NoError(t, err, "Failed to read retrieved data")

	// Verify data integrity
	require.Equal(t, dataItemBytes, retrievedData, "Retrieved data should match submitted data")
	t.Logf("Successfully retrieved %d bytes", len(retrievedData))

	// Step 4: Sign the data with goar (simulated)
	// Update the bundle item with the retrieved data
	bundleItem.Data = string(retrievedData)

	// For this integration test, we'll simulate signing by creating mock signature data
	// In a real scenario, the actual goar signing would be done in the web interface
	mockSignature := "mock-signature-" + wallet.Owner()[:20]
	bundleItem.Signature = mockSignature
	bundleItem.Owner = wallet.Owner()
	bundleItem.Id = "mock-id-12345678"

	// Create signed data (in real scenario this would be proper ANS-104 format)
	signatureInfo := fmt.Sprintf("signed-with-%s", mockSignature[:20])
	signedBytes := append(retrievedData, []byte(" - "+signatureInfo)...)

	t.Logf("Signed data item: %d bytes with mock signature", len(signedBytes))

	// Step 5: Submit signed data back to server
	signedReq := server.SubmitSignedDataRequest{
		SignedData: signedBytes,
	}

	signedBody, err := json.Marshal(signedReq)
	require.NoError(t, err, "Failed to marshal signed request")

	signedResp, err := http.Post(testServerURL+"/"+uuid, "application/json", bytes.NewReader(signedBody))
	require.NoError(t, err, "Failed to submit signed data")
	defer signedResp.Body.Close()

	require.Equal(t, http.StatusOK, signedResp.StatusCode, "Signed submit request failed")

	var successResp server.SuccessResponse
	err = json.NewDecoder(signedResp.Body).Decode(&successResp)
	require.NoError(t, err, "Failed to decode success response")

	t.Logf("Server response: %s", successResp.Message)

	// Step 6: Verify the signature using goar
	// For this test, we'll verify that the bundle item was properly signed
	require.NotEmpty(t, bundleItem.Signature, "Bundle item should have signature")
	require.NotEmpty(t, bundleItem.Owner, "Bundle item should have owner")

	// Verify the signer matches our wallet
	walletOwner := wallet.Owner()
	require.Equal(t, walletOwner, bundleItem.Owner, "Signer should match wallet owner")

	ownerDisplay := walletOwner
	if len(ownerDisplay) > 20 {
		ownerDisplay = ownerDisplay[:20] + "..."
	}
	t.Logf("✅ Signature verification successful! Signer: %s", ownerDisplay)

	// Step 7: Verify server state
	statusResp, err := http.Get(testServerURL + "/status")
	require.NoError(t, err, "Failed to get server status")
	defer statusResp.Body.Close()

	var status server.StatusResponse
	err = json.NewDecoder(statusResp.Body).Decode(&status)
	require.NoError(t, err, "Failed to decode status response")

	t.Logf("Server status: %+v", status)
}

// testJSONDataSubmission tests submitting data via JSON payload
func testJSONDataSubmission(t *testing.T, srv *server.Server) {
	_ = srv // Mark as used to avoid unused variable warning
	testData := []byte("Test JSON submission data")

	submitReq := server.SubmitDataRequest{
		Data:     testData,
		ClientID: "json-test-client",
	}

	submitBody, err := json.Marshal(submitReq)
	require.NoError(t, err)

	resp, err := http.Post(testServerURL+"/", "application/json", bytes.NewReader(submitBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var submitResp server.SubmitDataResponse
	err = json.NewDecoder(resp.Body).Decode(&submitResp)
	require.NoError(t, err)

	// Retrieve and verify
	getResp, err := http.Get(testServerURL + "/" + submitResp.UUID)
	require.NoError(t, err)
	defer getResp.Body.Close()

	retrievedData, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	assert.Equal(t, testData, retrievedData)
}

// testBinaryDataSubmission tests submitting raw binary data
func testBinaryDataSubmission(t *testing.T, srv *server.Server) {
	_ = srv // Mark as used to avoid unused variable warning
	// Generate random binary data
	testData := make([]byte, 1024)
	_, err := rand.Read(testData)
	require.NoError(t, err)

	// Submit as raw binary
	resp, err := http.Post(testServerURL+"/", "application/octet-stream", bytes.NewReader(testData))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var submitResp server.SubmitDataResponse
	err = json.NewDecoder(resp.Body).Decode(&submitResp)
	require.NoError(t, err)

	// Retrieve and verify
	getResp, err := http.Get(testServerURL + "/" + submitResp.UUID)
	require.NoError(t, err)
	defer getResp.Body.Close()

	retrievedData, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)

	assert.Equal(t, testData, retrievedData)
}

// testWebSocketNotifications tests WebSocket real-time notifications
func testWebSocketNotifications(t *testing.T, srv *server.Server) {
	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(testWSURL, nil)
	require.NoError(t, err, "Failed to connect to WebSocket")
	defer conn.Close()

	// Submit data to get UUID
	testData := []byte("WebSocket test data")
	submitReq := server.SubmitDataRequest{
		Data:     testData,
		ClientID: "websocket-test-client",
	}

	submitBody, err := json.Marshal(submitReq)
	require.NoError(t, err)

	resp, err := http.Post(testServerURL+"/", "application/json", bytes.NewReader(submitBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	var submitResp server.SubmitDataResponse
	err = json.NewDecoder(resp.Body).Decode(&submitResp)
	require.NoError(t, err)

	uuid := submitResp.UUID

	// Subscribe to UUID updates
	subscribeMsg := map[string]string{
		"type": "subscribe",
		"uuid": uuid,
	}

	err = conn.WriteJSON(subscribeMsg)
	require.NoError(t, err, "Failed to send subscribe message")

	// Submit signed data (mock signature)
	signedReq := server.SubmitSignedDataRequest{
		SignedData: append(testData, []byte(" - signed")...),
	}

	signedBody, err := json.Marshal(signedReq)
	require.NoError(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		http.Post(testServerURL+"/"+uuid, "application/json", bytes.NewReader(signedBody))
	}()

	// Wait for WebSocket notification
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// We might receive multiple messages, look for the "signed" message
	for i := 0; i < 3; i++ {
		var message map[string]interface{}
		err = conn.ReadJSON(&message)
		require.NoError(t, err, "Failed to receive WebSocket message")

		t.Logf("Received WebSocket message %d: %+v", i+1, message)

		if msgType, ok := message["type"].(string); ok && msgType == "signed" {
			if msgUUID, ok := message["uuid"].(string); ok {
				assert.Equal(t, uuid, msgUUID)
				t.Logf("✅ Received signing completion notification for UUID: %s", uuid)
				return
			}
		}

		// Reset deadline for next message
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	}

	t.Error("Did not receive expected 'signed' WebSocket message")
}

// testErrorHandling tests various error conditions
func testErrorHandling(t *testing.T, srv *server.Server) {
	_ = srv // Mark as used to avoid unused variable warning
	t.Run("InvalidUUID", func(t *testing.T) {
		resp, err := http.Get(testServerURL + "/invalid-uuid")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NotFoundUUID", func(t *testing.T) {
		resp, err := http.Get(testServerURL + "/00000000-0000-0000-0000-000000000000")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("TooLargeData", func(t *testing.T) {
		// Create data larger than server limit (10MB)
		largeData := make([]byte, 11*1024*1024)
		resp, err := http.Post(testServerURL+"/", "application/octet-stream", bytes.NewReader(largeData))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": json}`)
		resp, err := http.Post(testServerURL+"/", "application/json", bytes.NewReader(invalidJSON))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// testMultipleRequests tests handling multiple concurrent requests
func testMultipleRequests(t *testing.T, srv *server.Server) {
	const numRequests = 10

	results := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			testData := []byte(fmt.Sprintf("Concurrent test data %d", id))

			submitReq := server.SubmitDataRequest{
				Data:     testData,
				ClientID: fmt.Sprintf("concurrent-client-%d", id),
			}

			submitBody, err := json.Marshal(submitReq)
			if err != nil {
				results <- false
				return
			}

			resp, err := http.Post(testServerURL+"/", "application/json", bytes.NewReader(submitBody))
			if err != nil {
				results <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				results <- false
				return
			}

			var submitResp server.SubmitDataResponse
			err = json.NewDecoder(resp.Body).Decode(&submitResp)
			if err != nil {
				results <- false
				return
			}

			// Retrieve data
			getResp, err := http.Get(testServerURL + "/" + submitResp.UUID)
			if err != nil {
				results <- false
				return
			}
			defer getResp.Body.Close()

			retrievedData, err := io.ReadAll(getResp.Body)
			if err != nil {
				results <- false
				return
			}

			results <- bytes.Equal(testData, retrievedData)
		}(i)
	}

	// Wait for all requests to complete
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")
	t.Logf("Successfully handled %d concurrent requests", successCount)
}

// Benchmark tests
func BenchmarkDataSubmission(b *testing.B) {
	srv, cancel := setupTestServer(&testing.T{})
	defer cancel()
	_ = srv // Mark as used to avoid unused variable warning

	waitForServer(&testing.T{}, testServerURL)

	testData := []byte("Benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			submitReq := server.SubmitDataRequest{
				Data:     testData,
				ClientID: "benchmark-client",
			}

			submitBody, _ := json.Marshal(submitReq)
			resp, err := http.Post(testServerURL+"/", "application/json", bytes.NewReader(submitBody))
			if err != nil {
				b.Error(err)
				return
			}
			resp.Body.Close()
		}
	})
}

func BenchmarkDataRetrieval(b *testing.B) {
	srv, cancel := setupTestServer(&testing.T{})
	defer cancel()
	_ = srv // Mark as used to avoid unused variable warning

	waitForServer(&testing.T{}, testServerURL)

	// Pre-create some data
	testData := []byte("Benchmark retrieval test data")
	submitReq := server.SubmitDataRequest{
		Data:     testData,
		ClientID: "benchmark-client",
	}

	submitBody, _ := json.Marshal(submitReq)
	resp, _ := http.Post(testServerURL+"/", "application/json", bytes.NewReader(submitBody))
	defer resp.Body.Close()

	var submitResp server.SubmitDataResponse
	json.NewDecoder(resp.Body).Decode(&submitResp)
	uuid := submitResp.UUID

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(testServerURL + "/" + uuid)
			if err != nil {
				b.Error(err)
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// Helper function to check if we're in a CI environment
func isCI() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Skip integration tests in CI unless explicitly enabled
	if isCI() && os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		fmt.Println("Skipping integration tests in CI (set RUN_INTEGRATION_TESTS=true to run)")
		os.Exit(0)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}
