#!/bin/bash

# 🎭 Harlequin Remote Signing Service Demo
# This script demonstrates the complete remote signing workflow

set -e

PORT=8090
echo "🎭 Harlequin Remote Signing Service Demo"
echo "========================================"
echo ""

# Start the server in background
echo "1. Starting Remote Signing Server on port $PORT..."
./remote-signing start --port $PORT &
SERVER_PID=$!

# Wait for server to start
sleep 3
echo "✅ Server started (PID: $SERVER_PID)"
echo ""

# Submit raw data
echo "2. Submitting raw data for signing..."
RESPONSE=$(curl -s -X POST http://localhost:$PORT/ \
  -H "Content-Type: application/json" \
  -d '{
    "data": "SGVsbG8gSGFybGVxdWluISBUaGlzIGlzIGEgZGVtbyBvZiBzaW1wbGlmaWVkIHJhdyBkYXRhIHNpZ25pbmcuIFRoZSBjbGllbnQgY2FuIGhhbmRsZSBhbnkgZGF0YSBzaGFwZS4=",
    "client_id": "demo-client",
    "callback_url": "http://localhost:3000/webhook"
  }')

echo "📤 Data submitted successfully!"
echo ""

# Parse response
UUID=$(echo $RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['uuid'])" 2>/dev/null)
SIGNING_URL=$(echo $RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['signing_url'])" 2>/dev/null)

echo "📋 Signing Details:"
echo "   UUID: $UUID"
echo "   Signing URL: $SIGNING_URL"
echo ""

# Retrieve the data
echo "3. Retrieving the unsigned data..."
curl -s http://localhost:$PORT/$UUID | python3 -m json.tool > /tmp/data.json
echo "✅ Data retrieved and saved to /tmp/data.json"
echo ""

# Show decoded data
ENCODED_DATA=$(cat /tmp/data.json | python3 -c "import sys, json, base64; data=json.load(sys.stdin); print(data['data'])")
DECODED_DATA=$(echo $ENCODED_DATA | base64 -d)
echo "📄 Decoded data content:"
echo "   \"$DECODED_DATA\""
echo ""

# Show server status
echo "4. Checking server status..."
curl -s http://localhost:$PORT/status | python3 -m json.tool
echo ""

echo "🌐 Next Steps:"
echo "   1. Open this URL in your browser: $SIGNING_URL"
echo "   2. Connect your Arweave wallet (ArConnect, etc.)"
echo "   3. Sign the data item"
echo "   4. The server will notify connected WebSocket clients"
echo ""

echo "📡 WebSocket Demo:"
echo "   You can connect to: ws://localhost:$PORT/ws"
echo "   Send: {\"type\": \"subscribe\", \"uuid\": \"$UUID\"}"
echo "   You'll receive real-time updates when the item is signed"
echo ""

echo "🔗 API Endpoints Available:"
echo "   • POST http://localhost:$PORT/                     (Submit data item)"
echo "   • GET http://localhost:$PORT/$UUID            (Get unsigned data)"
echo "   • POST http://localhost:$PORT/$UUID           (Submit signed data)"
echo "   • GET http://localhost:$PORT/sign/$UUID       (Signing interface)"
echo "   • WS ws://localhost:$PORT/ws                       (WebSocket)"
echo "   • GET http://localhost:$PORT/status                (Server status)"
echo "   • GET http://localhost:$PORT/health                (Health check)"
echo ""

echo "🛑 Press Enter to stop the demo server..."
read

# Cleanup
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
echo "✅ Demo completed!"
