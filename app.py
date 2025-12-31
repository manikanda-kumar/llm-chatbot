import os, threading, asyncio, json
import requests
from flask import Flask, render_template, request, jsonify, abort
import jwt
from datetime import datetime, timedelta
from chatbot.database import auth_user, init_db
from chatbot.mcp.client_sse import InteractiveBankingAssistant
from chatbot.config import BANKING_API_URL

# Initialize Flask app pointing to local templates/ and static/
app = Flask(__name__, template_folder="templates", static_folder="static")

# JWT configuration
SECRET_KEY = os.getenv("SECRET_KEY", "your_secret_key")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 15

# Instantiate & initialize the agent at startup
assistant = InteractiveBankingAssistant()

# Spin up a dedicated loop in a background thread
background_loop = asyncio.new_event_loop()
def _start_background_loop(loop):
    asyncio.set_event_loop(loop)
    loop.run_until_complete(assistant.initialize_session())
    loop.run_forever()

t = threading.Thread(target=_start_background_loop, args=(background_loop,), daemon=True)
t.start()

# Helpers for JWT
def create_access_token(email: str, customer_id: str) -> str:
    expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    payload = {"sub": email, "customer_id": customer_id, "exp": expire}
    return jwt.encode(payload, SECRET_KEY, algorithm=ALGORITHM)


def lookup_customer_by_email(email: str) -> dict | None:
    """Lookup customer from Go Banking API by email."""
    try:
        resp = requests.post(
            f"{BANKING_API_URL}/api/v1/customers/search",
            json={"email": email},
            timeout=5
        )
        if resp.status_code == 200:
            data = resp.json()
            if data.get("customers") and len(data["customers"]) == 1:
                return data["customers"][0]
        return None
    except Exception as e:
        print(f"[ERROR] Customer lookup failed: {e}")
        return None


def verify_access_token(token: str) -> dict:
    """Returns dict with 'email' and 'customer_id'."""
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        email = payload.get("sub")
        customer_id = payload.get("customer_id")
        if not email or not customer_id:
            abort(401, "Invalid token payload")
        return {"email": email, "customer_id": customer_id}
    except jwt.ExpiredSignatureError:
        abort(401, "Token has expired")
    except jwt.InvalidTokenError:
        abort(401, "Invalid token")

# Routes
@app.route("/", methods=["GET"])
def index():
    return render_template("customer_chat.html")

@app.route("/demo", methods=["GET"])
def demo():
    return render_template("chat.html")

@app.route("/auth/login", methods=["POST"])
def auth_login():
    data = request.get_json() or {}
    email = data.get("email") or data.get("username")  # Support both for backward compat
    password = data.get("password")
    if not email or not password:
        abort(400, 'Missing "email" or "password"')

    # Lookup customer by email from Go Banking API
    customer = lookup_customer_by_email(email)
    if not customer:
        return jsonify({"status": "fail", "message": "Customer not found"}), 401

    # For demo: accept any password (in production, validate against auth service)
    # TODO: Add proper password validation
    if password != "password1":
        return jsonify({"status": "fail", "message": "Invalid password"}), 401

    customer_id = customer["id"]
    token = create_access_token(email, customer_id)
    return jsonify({
        "status": "success",
        "access_token": token,
        "token_type": "bearer",
        "customer_name": f"{customer['first_name']} {customer['last_name']}"
    }), 200

@app.route("/chat", methods=["POST"])
def chat():
    # 3) auth as before…
    auth_header = request.headers.get("Authorization", "")
    if not auth_header.startswith("Bearer "):
        return jsonify({"reply": "🔒 Please login to continue."}), 401
    token = auth_header.split(" ", 1)[1]
    user_info = verify_access_token(token)
    customer_id = user_info["customer_id"]

    # 4) Grab the incoming message
    msg = request.json.get("message", "").strip()
    if not msg:
        return jsonify({"reply": "💡 I didn't get any text."}), 400

    # 5) Schedule your send_message onto the background loop
    future = asyncio.run_coroutine_threadsafe(
        assistant.send_message(msg, customer_id),
        background_loop
    )
    try:
        result = future.result(timeout=30)   # wait up to 30s
    except Exception as e:
        print(f"[ERROR] Chat request failed: {e}")  # Log for debugging
        return jsonify({"reply": "I'm sorry, I'm having trouble processing your request right now. Please try again in a moment."}), 500

    # 6) Handle the two possible return types
    #    - A string → that’s your model’s reply
    #    - A dict/list → that’s raw tool output, so jsonify it or summarize
    if isinstance(result, str):
        return jsonify({"reply": result})

    # If it's a dict with an "error" key, log it and return generic message:
    if isinstance(result, dict) and "error" in result:
        print(f"[ERROR] Tool returned error: {result['error']}")
        return jsonify({"reply": "I'm sorry, I couldn't complete that request. Please try again or rephrase your question."})

    # Otherwise, just stringify the payload
    return jsonify({"reply": json.dumps(result, indent=2)})

if __name__ == "__main__":
    init_db()
    # Turn off the reloader
    app.run(host="0.0.0.0", port=3000, debug=True, use_reloader=False)
