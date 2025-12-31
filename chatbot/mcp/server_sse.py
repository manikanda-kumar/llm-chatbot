from mcp.server.fastmcp import FastMCP
from dotenv import load_dotenv
from decimal import Decimal
import os
import sys

# Add the parent directory to the Python path to import from src and chatbot
parent_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), '../..'))
sys.path.append(parent_dir)
print(f"Added to Python path: {parent_dir}")

# Import RAG components (optional - may fail due to chromadb compatibility)
try:
    from chatbot.rag.rag_chatbot import RBCChatbot
    RAG_AVAILABLE = True
except ImportError as e:
    print(f"Warning: RAG not available - {e}")
    RBCChatbot = None
    RAG_AVAILABLE = False

# Import configuration
from chatbot.config import (
    MCP_NAME, MCP_HOST, MCP_PORT, DEFAULT_USER_ID,
    ACCOUNT_ID_MAPPINGS, CUSTOMER_ID_MAPPINGS
)

# Import banking client for Go API
from chatbot.banking_client import get_banking_client

# Load environment variables from .env file
load_dotenv("../../.env")

# Initialize the RAG chatbot (for FAQ/policy questions)
chatbot = None
if RAG_AVAILABLE:
    try:
        chatbot = RBCChatbot()
    except Exception as e:
        print(f"Warning: Could not initialize RAG chatbot - {e}")
        RAG_AVAILABLE = False

# Get banking client instance
banking_client = get_banking_client()

# Create the MCP server
mcp = FastMCP(name=MCP_NAME, host=MCP_HOST, port=MCP_PORT)


def _get_api_account_id(account_number: str) -> str:
    """Map legacy account number to Go API account ID."""
    return ACCOUNT_ID_MAPPINGS.get(account_number, account_number)


def _get_api_customer_id(user_id: str) -> str:
    """Return customer ID as-is (no mapping needed - we pass actual customer_id now)."""
    return user_id


# RAG Tool: Answer questions using the RAG system
@mcp.tool()
def answer_banking_question(question: str = "") -> dict:
    """
    Answer a banking question using the RAG system with RBC documentation.
    Only for banking, financial services, or RBC-related questions.
    Returns the answer and sources.
    """
    if not question or question.strip() == "":
        return {
            "answer": "I need a specific question to help you. What would you like to know about banking?",
            "sources": []
        }
    if not RAG_AVAILABLE or chatbot is None:
        return {
            "answer": "RAG system is not available. Please use the banking tools for account-related queries.",
            "sources": []
        }
    print(f"[RAG] Processing question: {question}")
    result = chatbot.answer_question(question)
    print(f"[RAG] Found answer with {len(result['sources'])} sources")
    return {
        "answer": result["answer"],
        "sources": result["sources"]
    }


# Tool 1: List all accounts belonging to a user
@mcp.tool()
def list_user_accounts(user_id: str) -> list[dict]:
    """List all accounts for a given user."""
    print(f"[DEBUG] list_user_accounts called with user_id={user_id}")

    try:
        customer_id = _get_api_customer_id(user_id)
        accounts = banking_client.get_customer_accounts(customer_id)

        # Transform to legacy format for compatibility
        result = []
        for acc in accounts:
            result.append({
                "account_number": acc["id"],
                "account_name": f"{acc['account_type'].title()} Account",
                "balance": str(acc["balance"]),
                "currency": acc.get("currency", "CAD"),
                "status": acc.get("status", "active")
            })

        print(f"[DEBUG] Accounts: {result}")
        return result
    except Exception as e:
        print(f"[ERROR] Failed to list accounts: {e}")
        return []


# Tool 2: List target accounts that can receive transfers
@mcp.tool()
def list_target_accounts(user_id: str, from_account: str) -> list[dict]:
    """List all other accounts this user can transfer to."""
    print(f"[DEBUG] list_target_accounts called with user_id={user_id}, from_account={from_account}")

    try:
        customer_id = _get_api_customer_id(user_id)
        from_account_id = _get_api_account_id(from_account)
        accounts = banking_client.get_customer_accounts(customer_id)

        # Filter out the source account
        result = []
        for acc in accounts:
            if acc["id"] != from_account_id:
                result.append({
                    "account_number": acc["id"],
                    "account_name": f"{acc['account_type'].title()} Account",
                    "balance": str(acc["balance"]),
                    "currency": acc.get("currency", "CAD")
                })

        print(f"[DEBUG] Transfer targets: {result}")
        return result
    except Exception as e:
        print(f"[ERROR] Failed to list target accounts: {e}")
        return []


# Tool 3: Transfer funds between two accounts
@mcp.tool()
def transfer_funds(user_id: str, from_account: str, to_account: str, amount: str) -> str:
    """Transfer funds from one account to another."""
    print(f"[DEBUG] transfer_funds called with user_id={user_id}, from_account={from_account}, to_account={to_account}, amount={amount}")

    try:
        # Clean amount
        clean_amount = amount.replace('$', '').replace(',', '')
        decimal_amount = float(clean_amount)

        # Map to API IDs
        from_account_id = _get_api_account_id(from_account)
        to_account_id = _get_api_account_id(to_account)

        print(f"[DEBUG] Transferring {decimal_amount} from {from_account_id} to {to_account_id}")

        # Call Go API
        result = banking_client.transfer(from_account_id, to_account_id, decimal_amount)

        if result.get("status") == "completed":
            return f"✅ Transferred ${clean_amount} from {from_account} to {to_account}. Transfer ID: {result['id']}"
        else:
            return f"⏳ Transfer initiated. Status: {result.get('status', 'pending')}"

    except Exception as e:
        print(f"[ERROR] Transfer failed: {str(e)}")
        return f"❌ Transfer failed: {str(e)}"


# Tool 4: Get account balance
@mcp.tool()
def get_account_balance(user_id: str, account_number: str = "") -> dict:
    """Get the balance of a specific account. If account_number is empty, returns all account balances."""
    print(f'[DEBUG] get_account_balance called with user_id={user_id}, account_number={account_number}')

    try:
        customer_id = _get_api_customer_id(user_id)

        # If no account specified, return all balances
        if not account_number or account_number.strip() == "":
            accounts = banking_client.get_customer_accounts(customer_id)
            balances = []
            for acc in accounts:
                balances.append({
                    "account_number": acc["id"],
                    "account_name": f"{acc['account_type'].title()} Account",
                    "balance": str(acc["balance"]),
                    "currency": acc.get("currency", "CAD")
                })
            return {"accounts": balances, "message": "Here are all your account balances"}

        # Specific account requested
        account_id = _get_api_account_id(account_number)
        balance = banking_client.get_balance(account_id)

        # Get account details for name
        account = banking_client.get_account(account_id)

        return {
            "account_number": account_number,
            "account_name": f"{account['account_type'].title()} Account",
            "balance": str(balance["balance"]),
            "currency": balance.get("currency", "CAD")
        }
    except Exception as e:
        print(f"[ERROR] Failed to get balance: {e}")
        return {"error": f"Account {account_number} not found."}


# Tool 5: Get transaction history
@mcp.tool()
def get_transaction_history(user_id: str, account_number: str, days: int = 30) -> list[dict]:
    """Get the transaction history for a specific account."""
    print(f"[DEBUG] get_transaction_history called with user_id={user_id}, account_number={account_number}, days={days}")

    try:
        account_id = _get_api_account_id(account_number)
        transactions = banking_client.get_transactions(account_id, limit=50)

        # Transform to legacy format
        result = []
        for txn in transactions:
            # Determine transaction type
            txn_type = txn.get("type", "unknown")
            if txn_type in ["deposit", "transfer_in"]:
                transaction_type = "credit"
                amount = str(txn["amount"])
            else:
                transaction_type = "debit"
                amount = str(-txn["amount"])

            result.append({
                "transaction_id": txn["id"],
                "date": txn["created_at"].split("T")[0],
                "description": txn.get("description", txn_type.replace("_", " ").title()),
                "amount": amount,
                "transaction_type": transaction_type
            })

        print(f"[DEBUG] Returning: {len(result)} transactions")
        return result
    except Exception as e:
        print(f"[ERROR] Failed to get transactions: {e}")
        return []


# Tool 6: Get customer cards
@mcp.tool()
def get_customer_cards(user_id: str) -> list[dict]:
    """Get all cards for a customer."""
    print(f"[DEBUG] get_customer_cards called with user_id={user_id}")

    try:
        customer_id = _get_api_customer_id(user_id)
        cards = banking_client.get_customer_cards(customer_id)

        result = []
        for card in cards:
            result.append({
                "card_id": card["id"],
                "card_number": card["card_number"],
                "card_type": card["card_type"],
                "expiry_date": card["expiry_date"],
                "status": card["status"]
            })

        print(f"[DEBUG] Cards: {result}")
        return result
    except Exception as e:
        print(f"[ERROR] Failed to get cards: {e}")
        return []


# Tool 7: Block a card
@mcp.tool()
def block_card(user_id: str, card_id: str) -> str:
    """Block a card."""
    print(f"[DEBUG] block_card called with user_id={user_id}, card_id={card_id}")

    try:
        result = banking_client.block_card(card_id)
        return f"✅ Card {card_id} has been blocked successfully."
    except Exception as e:
        print(f"[ERROR] Failed to block card: {e}")
        return f"❌ Failed to block card: {str(e)}"


# Tool 8: Get customer loans
@mcp.tool()
def get_customer_loans(user_id: str) -> list[dict]:
    """Get all loans for a customer."""
    print(f"[DEBUG] get_customer_loans called with user_id={user_id}")

    try:
        customer_id = _get_api_customer_id(user_id)
        loans = banking_client.get_customer_loans(customer_id)

        result = []
        for loan in loans:
            result.append({
                "loan_id": loan["id"],
                "loan_type": loan["loan_type"],
                "principal": str(loan["principal"]),
                "balance": str(loan["balance"]),
                "interest_rate": f"{loan['interest_rate']}%",
                "monthly_payment": str(loan["monthly_payment"]),
                "status": loan["status"]
            })

        print(f"[DEBUG] Loans: {result}")
        return result
    except Exception as e:
        print(f"[ERROR] Failed to get loans: {e}")
        return []


# Tool 9: Get loan payment schedule
@mcp.tool()
def get_loan_schedule(user_id: str, loan_id: str) -> list[dict]:
    """Get the payment schedule for a loan."""
    print(f"[DEBUG] get_loan_schedule called with user_id={user_id}, loan_id={loan_id}")

    try:
        schedule = banking_client.get_loan_schedule(loan_id)

        # Return first 12 months for readability
        return schedule[:12]
    except Exception as e:
        print(f"[ERROR] Failed to get loan schedule: {e}")
        return []


# Run the MCP server using SSE transport
if __name__ == "__main__":
    print("[INFO] Starting MCP server on http://127.0.0.1:8050 using SSE transport...")
    print("[INFO] Using Go Banking API at", os.environ.get("BANKING_API_URL", "http://localhost:8080"))
    mcp.run(transport="sse")
