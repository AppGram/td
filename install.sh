#!/bin/bash

set -e

BINARY_NAME="td"
INSTALL_DIR="${HOME}/.local/bin"

echo "Building ${BINARY_NAME}..."
go build -o "${BINARY_NAME}" .

echo "Installing to ${INSTALL_DIR}..."
mkdir -p "${INSTALL_DIR}"
mv "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo "Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""

# Check if install dir is in PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    echo "Note: ${INSTALL_DIR} is not in your PATH."
    echo "Add this to your shell config:"
    echo ""
    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
    echo ""
fi

echo "Run 'td' to start!"
