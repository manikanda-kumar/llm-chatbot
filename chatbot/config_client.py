"""Configuration settings for the chatbot client."""

# Command patterns
COMMANDS = {
    "exit": ["exit", "quit", "q", "bye", "goodbye"],
    "clear": ["clear", "clear history", "start over", "reset"],
    "user": ["user", "switch user", "change user"],
}


# System instructions template
SYSTEM_INSTRUCTIONS = """
You are an RBC Banking Agent helping user {user_id}.

CRITICAL: YOU MUST CALL A FUNCTION FOR EVERY BANKING REQUEST. Do not just respond with text when a function call is needed.

INTENT → ACTION MAPPING (follow this exactly):
- "balance" / "how much" / "what's in" → call get_account_balance
- "accounts" / "list" / "show my" → call list_user_accounts
- "transfer" / "send" / "move money" → call transfer_funds
- "transactions" / "history" / "recent" / "spending" → call get_transaction_history
- "show my cards" / "card details" / "card status" → call get_customer_cards
- "block card" / "freeze card" / "cancel card" → call block_card (get cards first)
- "loans" / "mortgage" / "auto loan" → call get_customer_loans
- "loan payment" / "payment schedule" → call get_loan_schedule (get loans first)
- General banking questions about RBC → call answer_banking_question

DISAMBIGUATION (common confusions):
- "credit card balance" → get_account_balance with account_number="3456789012" (NOT get_customer_cards)
- "debit card balance" → get_account_balance with account_number="1234567890" (checking account)
- "show my cards" → get_customer_cards (card details, NOT balance)
- "block my card" → First get_customer_cards to find card_id, then block_card
- "loan balance" → get_customer_loans (shows balance in response)
- "loan payment schedule" → First get_customer_loans to find loan_id, then get_loan_schedule

ACCOUNT NUMBER MAPPING:
- Checking/Chequing → "1234567890"
- Savings → "2345678901"
- Credit Card → "3456789012"

RULES:
1. ALWAYS call a function when the user asks about their accounts, balances, transactions, cards, or loans.
2. ONLY use one function per request.
3. The user_id parameter is automatically filled - do not worry about it.
4. For transfers: use exact account numbers and amount as string without $ (e.g., "50.00").
5. REFUSE non-banking questions (fitness, travel, cooking, etc.) - explain you only help with banking.
6. If user gives a short reply like "checking" or "savings" after you asked a question, use that to complete the previous request.
7. Only respond conversationally (without function call) for pure greetings like "hi" or "hello" with no banking context.
8. NEVER assume user has a product. If they ask about loans, FIRST call get_customer_loans to verify. If empty, tell them they have no loans - don't ask for loan details.

EXAMPLES:
- User: "What's my balance?" → Call get_account_balance (returns all if no account specified)
- User: "Show my accounts" → Call list_user_accounts
- User: "checking" (after you asked which account) → Call get_account_balance with "1234567890"
- User: "hi" → Respond with greeting (no function call)
- User: "hi, what's my balance?" → Call get_account_balance (not just a greeting!)
- User: "What's my credit card balance?" → Call get_account_balance with account_number="3456789012"
- User: "Show me my cards" → Call get_customer_cards
- User: "Block my debit card" → Call get_customer_cards first, then block_card with the card_id
"""

# Tool definitions
TOOL_DEFINITIONS = [
    {
        "name": "answer_banking_question",
        "description": "Answer a banking question using the RAG system with RBC documentation. ONLY use for questions about banking, financial services, or RBC products and services. DO NOT use for non-banking topics like fitness, travel, cooking, etc.",
        "parameters": {
            "type": "object",
            "properties": {
                "question": {
                    "type": "string",
                    "description": "The banking-related question to answer (must be about banking, finance, or RBC services)",
                }
            },
            "required": ["question"],
        },
    },
    {
        "name": "list_user_accounts",
        "description": "CALL THIS when user asks to see accounts, list accounts, show my accounts, what accounts do I have. Returns all accounts for the user.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "Auto-filled by system",
                }
            },
            "required": ["user_id"],
        },
    },
    {
        "name": "list_target_accounts",
        "description": "CALL THIS when user asks 'where can I transfer to' or 'what accounts can I send money to'. Lists transfer destinations.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
                "from_account": {
                    "type": "string",
                    "description": "The source account number (must be exact account number, not name)",
                },
            },
            "required": ["user_id", "from_account"],
        },
    },
    {
        "name": "transfer_funds",
        "description": "CALL THIS when user asks to transfer, send, move, or pay money between accounts. Requires from_account, to_account, and amount.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
                "from_account": {
                    "type": "string",
                    "description": "The source account number (must be exact account number, not name): 1234567890 for checking, 2345678901 for savings, 3456789012 for credit card",
                },
                "to_account": {
                    "type": "string",
                    "description": "The destination account number (must be exact account number, not name): 1234567890 for checking, 2345678901 for savings, 3456789012 for credit card",
                },
                "amount": {
                    "type": "string",
                    "description": "The amount to transfer as a string without $ or commas (e.g., '50.00')",
                },
            },
            "required": ["user_id", "from_account", "to_account", "amount"],
        },
    },
    {
        "name": "get_account_balance",
        "description": "CALL THIS when user asks about balance, how much money, account amount, or funds available. Returns balance. If no account specified, returns ALL balances.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "Auto-filled by system",
                },
                "account_number": {
                    "type": "string",
                    "description": "Optional. Account number: 1234567890=checking, 2345678901=savings, 3456789012=credit. If omitted, returns all balances.",
                },
            },
            "required": ["user_id"],
        },
    },
    {
        "name": "get_transaction_history",
        "description": "CALL THIS when user asks about transactions, recent activity, spending, purchases, deposits, or statement. Shows transaction history for an account.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
                "account_number": {
                    "type": "string",
                    "description": "The account number (must be exact account number, not name): 1234567890 for checking, 2345678901 for savings, 3456789012 for credit card",
                },
                "days": {
                    "type": "integer",
                    "description": "Number of days of history to retrieve (default: 30)",
                },
            },
            "required": ["user_id", "account_number"],
        },
    },
    {
        "name": "get_customer_cards",
        "description": "CALL THIS when user asks about their cards, card details, card status, card number, expiry date, or 'show my cards'. Returns card info (NOT balance). NOTE: For 'credit card balance', use get_account_balance with account 3456789012. For 'debit card balance', use get_account_balance with account 1234567890 (checking).",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
            },
            "required": ["user_id"],
        },
    },
    {
        "name": "block_card",
        "description": "CALL THIS when user wants to block, freeze, disable, or cancel a card. WORKFLOW: First call get_customer_cards to get the card_id, then call this with that card_id.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
                "card_id": {
                    "type": "string",
                    "description": "The card ID to block",
                },
            },
            "required": ["user_id", "card_id"],
        },
    },
    {
        "name": "get_customer_loans",
        "description": "CALL THIS when user asks about loans, mortgages, auto loans, personal loans, loan balance, or 'my loans'. Returns all active loans. ALWAYS call this first before asking for loan details.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
            },
            "required": ["user_id"],
        },
    },
    {
        "name": "get_loan_schedule",
        "description": "CALL THIS when user asks about loan payments, payment schedule, amortization, or 'when is my loan payment due'. WORKFLOW: First call get_customer_loans to get loan_id, then call this.",
        "parameters": {
            "type": "object",
            "properties": {
                "user_id": {
                    "type": "string",
                    "description": "The ID of the user (will be automatically filled)",
                },
                "loan_id": {
                    "type": "string",
                    "description": "The loan ID to get schedule for",
                },
            },
            "required": ["user_id", "loan_id"],
        },
    },
]

# Model configuration
MODEL_CONFIG = {
    "model_name": "openai/gpt-oss-20b",
    "temperature": 0.1,
    "tool_calling_config": {"mode": "AUTO"},
    "base_url": "https://openrouter.ai/api/v1",
    "api_key_env": "OPENROUTER_API_KEY",
}
