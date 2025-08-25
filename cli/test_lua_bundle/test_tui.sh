#!/bin/bash
# Test script to simulate TUI navigation for lua-utils bundle

echo "Testing Harlequin TUI Lua Utils Bundle Flow..."

# The TUI flow would be:
# 1. Start harlequin (shows main menu with "Build Project" and "Lua Utils")
# 2. Navigate down to "Lua Utils" and press Enter
# 3. Select "Bundle" command and press Enter
# 4. Select main.lua as entrypoint and press Enter
# 5. Enter output path (or use default) and press Enter
# 6. Watch bundling progress
# 7. See success result

echo "Manual TUI flow:"
echo "1. Run: ../harlequin"
echo "2. Use ↓ arrow to select 'Lua Utils'"
echo "3. Press Enter"
echo "4. Select 'Bundle' (should be highlighted)"
echo "5. Press Enter"
echo "6. Select 'main.lua' from the file list"
echo "7. Press Enter"
echo "8. Review/edit output path (default: main.bundled.lua)"
echo "9. Press Enter to start bundling"
echo "10. See success message and press Enter to exit"

echo ""
echo "Files available for testing:"
ls -la *.lua

echo ""
echo "To test manually, run: ../harlequin"
