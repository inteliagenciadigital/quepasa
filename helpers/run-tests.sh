#!/bin/bash

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}=== Starting QuePasa Workspace Test Suite ===${NC}"
echo ""

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC_DIR="$(dirname "$SCRIPT_DIR")/src"

if [ ! -d "$SRC_DIR" ]; then
    echo -e "${RED}Error: src directory not found at $SRC_DIR${NC}"
    exit 1
fi

cd "$SRC_DIR"

echo -e "${YELLOW}Running tests across all workspace modules...${NC}"
echo ""

go test github.com/nocodeleaks/quepasa/... 2>&1

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All workspace tests passed successfully!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Workspace tests failed!${NC}"
    exit 1
fi

