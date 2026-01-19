#!/bin/bash

set -e

echo "Testing td application..."

cd /Users/nedimf/Documents/Projects/Personal/temp/td

rm -f /Users/nedimf/.config/td/td.db

echo "1. Testing version flag..."
./td -version

echo ""
echo "2. Testing add todo..."
./td -a "Buy groceries"
./td -a "Walk the dog"
./td -a "Finish report"

echo ""
echo "3. Testing install script..."
./install.sh

echo ""
echo "4. Testing binary after install..."
/Users/nedimf/.local/bin/td -version

echo ""
echo "5. Testing add from installed binary..."
/Users/nedimf/.local/bin/td -a "New task from installed binary"

echo ""
echo "All tests passed!"
echo ""
echo "To run the TUI:"
echo "  /Users/nedimf/.local/bin/td"
echo "  or add ~/bin to PATH and run 'td'"
