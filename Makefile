.PHONY: help build start stop status logs clean \
        build-backend start-backend stop-backend \
        build-python start-mcp stop-mcp \
        start-web stop-web \
        start-all stop-all

# Colors
GREEN  := \033[0;32m
BLUE   := \033[0;34m
YELLOW := \033[0;33m
RED    := \033[0;31m
NC     := \033[0m

# Directories
LOGS_DIR := .logs
PIDS_DIR := .logs

# Default target
help:
	@echo ""
	@echo "$(BLUE)LLM Chatbot - Available Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Build:$(NC)"
	@echo "  make build          - Build all components"
	@echo "  make build-backend  - Build Go banking API"
	@echo "  make build-python   - Install Python dependencies"
	@echo ""
	@echo "$(GREEN)Start:$(NC)"
	@echo "  make start          - Start all services"
	@echo "  make start-backend  - Start Go banking API (port 8080)"
	@echo "  make start-mcp      - Start MCP server (port 8050)"
	@echo "  make start-web      - Start Flask web app (port 3000)"
	@echo ""
	@echo "$(GREEN)Stop:$(NC)"
	@echo "  make stop           - Stop all services"
	@echo "  make stop-backend   - Stop Go banking API"
	@echo "  make stop-mcp       - Stop MCP server"
	@echo "  make stop-web       - Stop Flask web app"
	@echo ""
	@echo "$(GREEN)Other:$(NC)"
	@echo "  make status         - Show running services"
	@echo "  make logs           - Tail all logs"
	@echo "  make clean          - Clean build artifacts and logs"
	@echo ""

# =============================================================================
# Build targets
# =============================================================================

build: build-backend build-python
	@echo "$(GREEN)✓ All components built$(NC)"

build-backend:
	@echo "$(BLUE)Building Go Banking API...$(NC)"
	@cd banking-core && go build -o bin/server .
	@echo "$(GREEN)✓ Backend built$(NC)"

build-python:
	@echo "$(BLUE)Setting up Python environment...$(NC)"
	@if [ ! -d ".venv" ]; then python3 -m venv .venv; fi
	@. .venv/bin/activate && pip install -q -r requirements.txt
	@echo "$(GREEN)✓ Python dependencies installed$(NC)"

# =============================================================================
# Start targets
# =============================================================================

start: start-all

start-all: $(LOGS_DIR)
	@echo "$(BLUE)Starting all services...$(NC)"
	@$(MAKE) -s start-backend
	@sleep 1
	@$(MAKE) -s start-mcp
	@sleep 2
	@$(MAKE) -s start-web
	@echo ""
	@echo "$(GREEN)All services running:$(NC)"
	@echo "  Banking API: http://localhost:8080"
	@echo "  MCP Server:  http://localhost:8050"
	@echo "  Chat UI:     http://localhost:3000"
	@echo ""
	@echo "Run $(YELLOW)make logs$(NC) to view logs"
	@echo "Run $(YELLOW)make stop$(NC) to stop all services"

start-backend: $(LOGS_DIR)
	@if [ -f $(PIDS_DIR)/backend.pid ] && kill -0 $$(cat $(PIDS_DIR)/backend.pid) 2>/dev/null; then \
		echo "$(YELLOW)Backend already running$(NC)"; \
	else \
		cd banking-core && ./bin/server > ../$(LOGS_DIR)/backend.log 2>&1 & echo $$! > ../$(PIDS_DIR)/backend.pid; \
		echo "$(GREEN)✓ Backend started (port 8080)$(NC)"; \
	fi

start-mcp: $(LOGS_DIR)
	@if [ -f $(PIDS_DIR)/mcp.pid ] && kill -0 $$(cat $(PIDS_DIR)/mcp.pid) 2>/dev/null; then \
		echo "$(YELLOW)MCP server already running$(NC)"; \
	else \
		. .venv/bin/activate && python chatbot/mcp/server_sse.py > $(LOGS_DIR)/mcp.log 2>&1 & echo $$! > $(PIDS_DIR)/mcp.pid; \
		echo "$(GREEN)✓ MCP server started (port 8050)$(NC)"; \
	fi

start-web: $(LOGS_DIR)
	@if [ -f $(PIDS_DIR)/web.pid ] && kill -0 $$(cat $(PIDS_DIR)/web.pid) 2>/dev/null; then \
		echo "$(YELLOW)Web app already running$(NC)"; \
	else \
		. .venv/bin/activate && python app.py > $(LOGS_DIR)/web.log 2>&1 & echo $$! > $(PIDS_DIR)/web.pid; \
		echo "$(GREEN)✓ Web app started (port 3000)$(NC)"; \
	fi

# =============================================================================
# Stop targets
# =============================================================================

stop: stop-all

stop-all:
	@echo "$(YELLOW)Stopping all services...$(NC)"
	@$(MAKE) -s stop-web
	@$(MAKE) -s stop-mcp
	@$(MAKE) -s stop-backend
	@echo "$(GREEN)✓ All services stopped$(NC)"

stop-backend:
	@if [ -f $(PIDS_DIR)/backend.pid ]; then \
		kill $$(cat $(PIDS_DIR)/backend.pid) 2>/dev/null && echo "$(GREEN)✓ Backend stopped$(NC)" || echo "  Backend not running"; \
		rm -f $(PIDS_DIR)/backend.pid; \
	else \
		lsof -ti:8080 | xargs -r kill 2>/dev/null || true; \
	fi

stop-mcp:
	@if [ -f $(PIDS_DIR)/mcp.pid ]; then \
		kill $$(cat $(PIDS_DIR)/mcp.pid) 2>/dev/null && echo "$(GREEN)✓ MCP server stopped$(NC)" || echo "  MCP server not running"; \
		rm -f $(PIDS_DIR)/mcp.pid; \
	else \
		lsof -ti:8050 | xargs -r kill 2>/dev/null || true; \
	fi

stop-web:
	@if [ -f $(PIDS_DIR)/web.pid ]; then \
		kill $$(cat $(PIDS_DIR)/web.pid) 2>/dev/null && echo "$(GREEN)✓ Web app stopped$(NC)" || echo "  Web app not running"; \
		rm -f $(PIDS_DIR)/web.pid; \
	else \
		lsof -ti:3000 | xargs -r kill 2>/dev/null || true; \
	fi

# =============================================================================
# Utility targets
# =============================================================================

$(LOGS_DIR):
	@mkdir -p $(LOGS_DIR)

status:
	@echo ""
	@echo "$(BLUE)Service Status:$(NC)"
	@echo ""
	@if lsof -ti:8080 >/dev/null 2>&1; then \
		echo "  $(GREEN)●$(NC) Backend    (port 8080) - running"; \
	else \
		echo "  $(RED)○$(NC) Backend    (port 8080) - stopped"; \
	fi
	@if lsof -ti:8050 >/dev/null 2>&1; then \
		echo "  $(GREEN)●$(NC) MCP Server (port 8050) - running"; \
	else \
		echo "  $(RED)○$(NC) MCP Server (port 8050) - stopped"; \
	fi
	@if lsof -ti:3000 >/dev/null 2>&1; then \
		echo "  $(GREEN)●$(NC) Web App    (port 3000) - running"; \
	else \
		echo "  $(RED)○$(NC) Web App    (port 3000) - stopped"; \
	fi
	@echo ""

logs:
	@tail -f $(LOGS_DIR)/*.log

logs-backend:
	@tail -f $(LOGS_DIR)/backend.log

logs-mcp:
	@tail -f $(LOGS_DIR)/mcp.log

logs-web:
	@tail -f $(LOGS_DIR)/web.log

clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	@rm -rf $(LOGS_DIR)
	@rm -rf banking-core/bin
	@echo "$(GREEN)✓ Cleaned$(NC)"

# =============================================================================
# Test targets
# =============================================================================

test: test-guardrails
	@echo "$(GREEN)All tests completed$(NC)"

test-guardrails:
	@echo "$(BLUE)Running guardrail tests...$(NC)"
	@. .venv/bin/activate && python -m pytest tests/test_guardrails.py -v --tb=short

test-security:
	@echo "$(BLUE)Running security tests...$(NC)"
	@. .venv/bin/activate && python -m pytest tests/test_guardrails.py::TestPromptInjection tests/test_guardrails.py::TestPIIProtection -v

test-functional:
	@echo "$(BLUE)Running functional tests...$(NC)"
	@. .venv/bin/activate && python -m pytest tests/test_guardrails.py::TestFunctionalCoverage -v
