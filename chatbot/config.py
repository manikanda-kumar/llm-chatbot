"""Configuration settings for the chatbot application."""

import os
from pathlib import Path

# Database settings
DB_FILE = os.environ.get("CHATBOT_DB_FILE", "bank.db")
DB_INIT_SQL = Path(__file__).parent / "init.sql"

# Account number mappings (for client-side account name resolution)
ACCOUNT_MAPPINGS = {
    "checking": "1234567890",
    "chequing": "1234567890",
    "cheque": "1234567890",
    "saving": "2345678901",
    "savings": "2345678901",
    "credit": "3456789012",
    "credit card": "3456789012",
}

# Go Banking API settings
BANKING_API_URL = os.environ.get("BANKING_API_URL", "http://localhost:8080")
BANKING_API_KEY = os.environ.get("BANKING_API_KEY", "")

# Mapping from legacy account numbers to Go API account IDs
ACCOUNT_ID_MAPPINGS = {
    "1234567890": "acc-001",  # Checking -> John Doe's checking
    "2345678901": "acc-002",  # Savings -> John Doe's savings
    "3456789012": "acc-003",  # Credit -> Jane Smith's checking (for demo)
}

# Mapping from legacy user IDs to Go API customer IDs
CUSTOMER_ID_MAPPINGS = {
    "test1": "cust-001",  # Default test user -> John Doe
    "test2": "cust-002",  # Secondary test user -> Jane Smith
}

# Default user for testing
DEFAULT_USER_ID = "test1"

# Vector database settings
VECTOR_DB_DIR = os.environ.get("VECTOR_DB_DIR", "./chroma_db")
DOCS_DIRECTORY = os.environ.get("DOCS_DIRECTORY", "./rbc_documents")

# OpenRouter API settings
OPENROUTER_API_KEY = os.environ.get("OPENROUTER_API_KEY")
OPENROUTER_BASE_URL = "https://openrouter.ai/api/v1"
OPENROUTER_MODEL = "openai/gpt-oss-20b"

# MCP settings
MCP_HOST = os.environ.get("MCP_HOST", "127.0.0.1")
MCP_PORT = int(os.environ.get("MCP_PORT", "8050"))
MCP_NAME = os.environ.get("MCP_NAME", "RBC-RAG-MCP")
