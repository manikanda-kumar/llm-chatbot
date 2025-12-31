"""Simple Flask web interface to test Go Banking API directly."""

import os
from flask import Flask, render_template_string, request, jsonify
import httpx

app = Flask(__name__)

BANKING_API_URL = os.getenv("BANKING_API_URL", "http://localhost:8080")

HTML_TEMPLATE = '''
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Banking API Test Interface</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #003168; margin-bottom: 20px; }
        .card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .card h2 { color: #003168; margin-bottom: 15px; font-size: 18px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: 500; color: #333; }
        input, select { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
        button { background: #003168; color: white; border: none; padding: 12px 24px; border-radius: 4px; cursor: pointer; font-size: 14px; }
        button:hover { background: #004a9f; }
        .result { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 4px; padding: 15px; margin-top: 15px; font-family: monospace; white-space: pre-wrap; max-height: 400px; overflow-y: auto; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(350px, 1fr)); gap: 20px; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 10px; border-radius: 4px; margin-top: 10px; }
        .error { background: #f8d7da; border: 1px solid #f5c6cb; padding: 10px; border-radius: 4px; margin-top: 10px; color: #721c24; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; padding: 10px; border-radius: 4px; margin-top: 10px; color: #155724; }
        .tabs { display: flex; gap: 10px; margin-bottom: 20px; }
        .tab { padding: 10px 20px; background: #e9ecef; border-radius: 4px; cursor: pointer; }
        .tab.active { background: #003168; color: white; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏦 Banking API Test Interface</h1>

        <div class="tabs">
            <div class="tab active" onclick="showTab('customers')">Customers</div>
            <div class="tab" onclick="showTab('accounts')">Accounts</div>
            <div class="tab" onclick="showTab('verification')">Identity Verification</div>
            <div class="tab" onclick="showTab('transfers')">Transfers</div>
        </div>

        <!-- Customers Tab -->
        <div id="customers" class="tab-content active">
            <div class="grid">
                <div class="card">
                    <h2>Search Customer by Name</h2>
                    <div class="form-group">
                        <label>First Name</label>
                        <input type="text" id="search-first-name" placeholder="e.g., John">
                    </div>
                    <div class="form-group">
                        <label>Last Name</label>
                        <input type="text" id="search-last-name" placeholder="e.g., Smith">
                    </div>
                    <button onclick="searchByName()">Search</button>
                    <div id="name-result" class="result" style="display:none;"></div>
                </div>

                <div class="card">
                    <h2>Get Customer by ID</h2>
                    <div class="form-group">
                        <label>Customer ID</label>
                        <input type="text" id="customer-id" placeholder="e.g., cust-001">
                    </div>
                    <button onclick="getCustomer()">Get Details</button>
                    <div id="customer-result" class="result" style="display:none;"></div>
                </div>

                <div class="card">
                    <h2>List Customers by Status</h2>
                    <div class="form-group">
                        <label>Status</label>
                        <select id="customer-status">
                            <option value="active">Active</option>
                            <option value="inactive">Inactive</option>
                            <option value="suspended">Suspended</option>
                        </select>
                    </div>
                    <button onclick="listByStatus()">List</button>
                    <div id="status-result" class="result" style="display:none;"></div>
                </div>
            </div>
        </div>

        <!-- Accounts Tab -->
        <div id="accounts" class="tab-content">
            <div class="grid">
                <div class="card">
                    <h2>Get Customer Accounts</h2>
                    <div class="form-group">
                        <label>Customer ID</label>
                        <input type="text" id="acc-customer-id" placeholder="e.g., cust-001">
                    </div>
                    <button onclick="getCustomerAccounts()">Get Accounts</button>
                    <div id="accounts-result" class="result" style="display:none;"></div>
                </div>

                <div class="card">
                    <h2>Get Account Balance</h2>
                    <div class="form-group">
                        <label>Account ID</label>
                        <input type="text" id="balance-account-id" placeholder="e.g., acc-001">
                    </div>
                    <button onclick="getBalance()">Get Balance</button>
                    <div id="balance-result" class="result" style="display:none;"></div>
                </div>

                <div class="card">
                    <h2>Get Customer Loans</h2>
                    <div class="form-group">
                        <label>Customer ID</label>
                        <input type="text" id="loan-customer-id" placeholder="e.g., cust-001">
                    </div>
                    <button onclick="getLoans()">Get Loans</button>
                    <div id="loans-result" class="result" style="display:none;"></div>
                </div>
            </div>
        </div>

        <!-- Verification Tab -->
        <div id="verification" class="tab-content">
            <div class="grid">
                <div class="card">
                    <h2>Identity Verification (Name + DOB)</h2>
                    <p style="color:#666; margin-bottom:15px;">Tests the critical identity verification flow</p>
                    <div class="form-group">
                        <label>First Name</label>
                        <input type="text" id="verify-first-name" placeholder="e.g., Michael">
                    </div>
                    <div class="form-group">
                        <label>Last Name</label>
                        <input type="text" id="verify-last-name" placeholder="e.g., Johnson">
                    </div>
                    <div class="form-group">
                        <label>Date of Birth</label>
                        <input type="date" id="verify-dob" value="1992-05-20">
                    </div>
                    <button onclick="verifyIdentity()">Verify Identity</button>
                    <div id="verify-result" class="result" style="display:none;"></div>
                </div>

                <div class="card">
                    <h2>Search by Date of Birth</h2>
                    <p style="color:#666; margin-bottom:15px;">Find all customers with same DOB</p>
                    <div class="form-group">
                        <label>Date of Birth</label>
                        <input type="date" id="dob-search" value="1988-09-12">
                    </div>
                    <button onclick="searchByDOB()">Search</button>
                    <div id="dob-result" class="result" style="display:none;"></div>
                </div>
            </div>

            <div class="card" style="margin-top:20px;">
                <h2>Test Scenarios</h2>
                <p style="margin-bottom:15px;">Click to test pre-configured identity verification scenarios:</p>
                <button onclick="testScenario('john-smith')" style="margin-right:10px;">3x John Smith (diff DOB)</button>
                <button onclick="testScenario('michael-johnson')" style="margin-right:10px;">2x Michael Johnson (same DOB, diff gender)</button>
                <button onclick="testScenario('sarah-williams')" style="margin-right:10px;">2x Sarah Williams (same DOB)</button>
                <button onclick="testScenario('dob-cluster')">DOB Cluster (1988-09-12)</button>
                <div id="scenario-result" class="result" style="display:none;"></div>
            </div>
        </div>

        <!-- Transfers Tab -->
        <div id="transfers" class="tab-content">
            <div class="grid">
                <div class="card">
                    <h2>Transfer Funds</h2>
                    <div class="form-group">
                        <label>From Account ID</label>
                        <input type="text" id="from-account" placeholder="e.g., acc-001">
                    </div>
                    <div class="form-group">
                        <label>To Account ID</label>
                        <input type="text" id="to-account" placeholder="e.g., acc-002">
                    </div>
                    <div class="form-group">
                        <label>Amount</label>
                        <input type="number" id="transfer-amount" placeholder="e.g., 100.00" step="0.01">
                    </div>
                    <button onclick="transfer()">Transfer</button>
                    <div id="transfer-result" class="result" style="display:none;"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        function showTab(tabId) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            document.querySelector(`.tab[onclick="showTab('${tabId}')"]`).classList.add('active');
            document.getElementById(tabId).classList.add('active');
        }

        async function apiCall(endpoint, resultId) {
            const resultDiv = document.getElementById(resultId);
            resultDiv.style.display = 'block';
            resultDiv.textContent = 'Loading...';
            resultDiv.className = 'result';

            try {
                const response = await fetch(endpoint);
                const data = await response.json();

                // Check for warnings/verification requirements
                if (data.requires_verification || data.warning) {
                    resultDiv.innerHTML = '<div class="warning">⚠️ ' + (data.warning || 'Multiple matches - verification required') + '</div>' +
                        '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
                } else if (data.error) {
                    resultDiv.innerHTML = '<div class="error">❌ ' + data.error + '</div>';
                } else {
                    resultDiv.textContent = JSON.stringify(data, null, 2);
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="error">❌ ' + error.message + '</div>';
            }
        }

        async function apiPost(endpoint, body, resultId) {
            const resultDiv = document.getElementById(resultId);
            resultDiv.style.display = 'block';
            resultDiv.textContent = 'Loading...';

            try {
                const response = await fetch(endpoint, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(body)
                });
                const data = await response.json();

                if (data.status === 'completed') {
                    resultDiv.innerHTML = '<div class="success">✅ Transfer completed!</div>' +
                        '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
                } else {
                    resultDiv.textContent = JSON.stringify(data, null, 2);
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="error">❌ ' + error.message + '</div>';
            }
        }

        function searchByName() {
            const firstName = document.getElementById('search-first-name').value;
            const lastName = document.getElementById('search-last-name').value;
            apiCall(`/api/customers/search/name?first_name=${firstName}&last_name=${lastName}`, 'name-result');
        }

        function getCustomer() {
            const id = document.getElementById('customer-id').value;
            apiCall(`/api/customers/${id}`, 'customer-result');
        }

        function listByStatus() {
            const status = document.getElementById('customer-status').value;
            apiCall(`/api/customers/status/${status}`, 'status-result');
        }

        function getCustomerAccounts() {
            const id = document.getElementById('acc-customer-id').value;
            apiCall(`/api/customers/${id}/accounts`, 'accounts-result');
        }

        function getBalance() {
            const id = document.getElementById('balance-account-id').value;
            apiCall(`/api/accounts/${id}/balance`, 'balance-result');
        }

        function getLoans() {
            const id = document.getElementById('loan-customer-id').value;
            apiCall(`/api/customers/${id}/loans`, 'loans-result');
        }

        function verifyIdentity() {
            const firstName = document.getElementById('verify-first-name').value;
            const lastName = document.getElementById('verify-last-name').value;
            const dob = document.getElementById('verify-dob').value;
            apiCall(`/api/customers/verify?first_name=${firstName}&last_name=${lastName}&dob=${dob}`, 'verify-result');
        }

        function searchByDOB() {
            const dob = document.getElementById('dob-search').value;
            apiCall(`/api/customers/search/dob?dob=${dob}`, 'dob-result');
        }

        function transfer() {
            const fromAcc = document.getElementById('from-account').value;
            const toAcc = document.getElementById('to-account').value;
            const amount = parseFloat(document.getElementById('transfer-amount').value);
            apiPost('/api/transfers', { from_account_id: fromAcc, to_account_id: toAcc, amount: amount }, 'transfer-result');
        }

        function testScenario(scenario) {
            const resultDiv = document.getElementById('scenario-result');
            resultDiv.style.display = 'block';

            switch(scenario) {
                case 'john-smith':
                    apiCall('/api/customers/search/name?first_name=John&last_name=Smith', 'scenario-result');
                    break;
                case 'michael-johnson':
                    apiCall('/api/customers/verify?first_name=Michael&last_name=Johnson&dob=1992-05-20', 'scenario-result');
                    break;
                case 'sarah-williams':
                    apiCall('/api/customers/verify?first_name=Sarah&last_name=Williams&dob=1988-09-12', 'scenario-result');
                    break;
                case 'dob-cluster':
                    apiCall('/api/customers/search/dob?dob=1988-09-12', 'scenario-result');
                    break;
            }
        }
    </script>
</body>
</html>
'''

@app.route('/')
def index():
    return render_template_string(HTML_TEMPLATE)

# Proxy routes to Go Banking API
@app.route('/api/<path:path>', methods=['GET', 'POST'])
def proxy(path):
    url = f"{BANKING_API_URL}/api/v1/{path}"

    try:
        with httpx.Client() as client:
            if request.method == 'GET':
                resp = client.get(url, params=request.args)
            else:
                resp = client.post(url, json=request.get_json())

            return jsonify(resp.json()), resp.status_code
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    print(f"Banking API URL: {BANKING_API_URL}")
    print("Starting test interface on http://localhost:5001")
    app.run(host='0.0.0.0', port=5001, debug=True)
