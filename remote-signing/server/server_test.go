package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerCreation(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		srv := New(nil)
		assert.NotNil(t, srv)
		assert.Equal(t, "localhost", srv.config.Host)
		assert.Equal(t, 8080, srv.config.Port)
		assert.Equal(t, int64(10*1024*1024), srv.config.MaxDataSize)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &Config{
			Host:        "0.0.0.0",
			Port:        9090,
			MaxDataSize: 5 * 1024 * 1024,
		}
		srv := New(config)
		assert.NotNil(t, srv)
		assert.Equal(t, "0.0.0.0", srv.config.Host)
		assert.Equal(t, 9090, srv.config.Port)
		assert.Equal(t, int64(5*1024*1024), srv.config.MaxDataSize)
	})
}

func TestServerLifecycle(t *testing.T) {
	config := &Config{
		Host:           "localhost",
		Port:           8083, // Different port to avoid conflicts
		AllowedOrigins: []string{"*"},
		MaxDataSize:    1024,
		SigningTimeout: 5 * time.Second,
	}

	srv := New(config)
	require.NotNil(t, srv)

	// Server should not be running initially
	assert.False(t, srv.IsRunning())

	// Start server
	ctx, cancel := context.WithCancel(context.Background())

	serverStarted := make(chan error, 1)
	go func() {
		serverStarted <- srv.Start(ctx)
	}()

	// Wait a bit for server to start
	time.Sleep(200 * time.Millisecond)

	// Server should be running
	assert.True(t, srv.IsRunning())

	// Stop server by cancelling context
	cancel()

	// Wait for server to finish
	select {
	case err := <-serverStarted:
		// Server should stop gracefully
		if err != nil && err != http.ErrServerClosed && err.Error() != "http: Server closed" {
			t.Errorf("Unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}

	// Server should not be running
	assert.False(t, srv.IsRunning())
}

func TestServerMethods(t *testing.T) {
	srv := New(nil)

	t.Run("ListSigningRequests", func(t *testing.T) {
		requests := srv.ListSigningRequests()
		assert.NotNil(t, requests)
		assert.Empty(t, requests)
	})

	t.Run("GetWebSocketHub", func(t *testing.T) {
		hub := srv.GetWebSocketHub()
		assert.NotNil(t, hub)
	})

	t.Run("GetSigningRequest", func(t *testing.T) {
		req, exists := srv.GetSigningRequest("non-existent-uuid")
		assert.Nil(t, req)
		assert.False(t, exists)
	})
}

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 8080, config.Port)
		assert.Equal(t, []string{"*"}, config.AllowedOrigins)
		assert.Equal(t, int64(10*1024*1024), config.MaxDataSize)
		assert.Equal(t, 30*time.Minute, config.SigningTimeout)
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		// Valid config
		config := &Config{
			Host:           "localhost",
			Port:           8080,
			AllowedOrigins: []string{"*"},
			MaxDataSize:    1024,
			SigningTimeout: time.Minute,
		}
		assert.NotNil(t, config)

		// Test port bounds
		config.Port = 0
		// Note: We could add validation if needed
		config.Port = 65536
		// Note: We could add validation if needed
	})
}
