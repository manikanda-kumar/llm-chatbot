"""
HTTP client for banking-core Go backend API.
"""
import os
from typing import Optional
import httpx

BANKING_API_URL = os.getenv("BANKING_API_URL", "http://localhost:8080")
BANKING_API_KEY = os.getenv("BANKING_API_KEY", "")


class BankingClient:
    def __init__(self, base_url: str = None, api_key: str = None):
        self.base_url = (base_url or BANKING_API_URL).rstrip("/")
        self.api_key = api_key or BANKING_API_KEY
        self.headers = {"Content-Type": "application/json"}
        if self.api_key:
            self.headers["X-API-Key"] = self.api_key

    def _get(self, path: str) -> dict:
        url = f"{self.base_url}{path}"
        with httpx.Client() as client:
            response = client.get(url, headers=self.headers)
            response.raise_for_status()
            return response.json()

    def _post(self, path: str, data: dict) -> dict:
        url = f"{self.base_url}{path}"
        with httpx.Client() as client:
            response = client.post(url, headers=self.headers, json=data)
            response.raise_for_status()
            return response.json()

    # Health
    def health(self) -> dict:
        return self._get("/health")

    # Customers
    def get_customer(self, customer_id: str) -> dict:
        return self._get(f"/api/v1/customers/{customer_id}")

    def list_customers(self) -> list:
        return self._get("/api/v1/customers")

    def get_customer_accounts(self, customer_id: str) -> list:
        return self._get(f"/api/v1/customers/{customer_id}/accounts")

    def get_customer_cards(self, customer_id: str) -> list:
        return self._get(f"/api/v1/customers/{customer_id}/cards")

    def get_customer_loans(self, customer_id: str) -> list:
        return self._get(f"/api/v1/customers/{customer_id}/loans")

    # Accounts
    def get_account(self, account_id: str) -> dict:
        return self._get(f"/api/v1/accounts/{account_id}")

    def get_balance(self, account_id: str) -> dict:
        return self._get(f"/api/v1/accounts/{account_id}/balance")

    def get_transactions(self, account_id: str, limit: int = 50) -> list:
        return self._get(f"/api/v1/accounts/{account_id}/transactions?limit={limit}")

    # Transactions
    def create_transaction(
        self,
        account_id: str,
        txn_type: str,
        amount: float,
        description: str = ""
    ) -> dict:
        return self._post("/api/v1/transactions", {
            "account_id": account_id,
            "type": txn_type,
            "amount": amount,
            "description": description
        })

    # Transfers
    def transfer(
        self,
        from_account_id: str,
        to_account_id: str,
        amount: float
    ) -> dict:
        return self._post("/api/v1/transfers", {
            "from_account_id": from_account_id,
            "to_account_id": to_account_id,
            "amount": amount
        })

    def get_transfer(self, transfer_id: str) -> dict:
        return self._get(f"/api/v1/transfers/{transfer_id}")

    # Cards
    def get_card(self, card_id: str) -> dict:
        return self._get(f"/api/v1/cards/{card_id}")

    def block_card(self, card_id: str) -> dict:
        return self._post(f"/api/v1/cards/{card_id}/block", {})

    # Loans
    def get_loan(self, loan_id: str) -> dict:
        return self._get(f"/api/v1/loans/{loan_id}")

    def get_loan_schedule(self, loan_id: str) -> list:
        return self._get(f"/api/v1/loans/{loan_id}/schedule")


# Singleton instance
_client: Optional[BankingClient] = None


def get_banking_client() -> BankingClient:
    global _client
    if _client is None:
        _client = BankingClient()
    return _client
