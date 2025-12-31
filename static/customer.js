// State
let accessToken = null;
let currentUserName = null;

// DOM Elements
const loginScreen = document.getElementById('login-screen');
const chatScreen = document.getElementById('chat-screen');
const loginForm = document.getElementById('login-form');
const messagesContainer = document.getElementById('messages');
const messageInput = document.getElementById('message-input');
const sendBtn = document.getElementById('send-btn');
const newChatBtn = document.getElementById('new-chat-btn');
const logoutBtn = document.getElementById('logout-btn');
const userInfo = document.getElementById('user-info');
const usernameDisplay = document.getElementById('username-display');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
  // Check for existing session
  const savedToken = sessionStorage.getItem('accessToken');
  const savedName = sessionStorage.getItem('customerName');
  if (savedToken && savedName) {
    accessToken = savedToken;
    currentUserName = savedName;
    showChatScreen();
  }

  // Event listeners
  loginForm.addEventListener('submit', handleLogin);
  sendBtn.addEventListener('click', sendMessage);
  newChatBtn.addEventListener('click', startNewChat);
  logoutBtn.addEventListener('click', handleLogout);

  // Input handling
  messageInput.addEventListener('input', handleInputChange);
  messageInput.addEventListener('keydown', handleKeyDown);

  // Quick actions
  document.querySelectorAll('.quick-action').forEach(btn => {
    btn.addEventListener('click', () => {
      const message = btn.dataset.message;
      messageInput.value = message;
      handleInputChange();
      sendMessage();
    });
  });
});

// Auto-resize textarea
function handleInputChange() {
  messageInput.style.height = 'auto';
  messageInput.style.height = Math.min(messageInput.scrollHeight, 200) + 'px';

  // Enable/disable send button
  if (messageInput.value.trim()) {
    sendBtn.classList.add('active');
    sendBtn.disabled = false;
  } else {
    sendBtn.classList.remove('active');
    sendBtn.disabled = true;
  }
}

function handleKeyDown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    if (messageInput.value.trim()) {
      sendMessage();
    }
  }
}

// Login
async function handleLogin(e) {
  e.preventDefault();
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;

  try {
    const res = await fetch('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });

    const data = await res.json();

    if (res.ok && data.status === 'success') {
      accessToken = data.access_token;
      currentUserName = data.customer_name || email;
      sessionStorage.setItem('accessToken', accessToken);
      sessionStorage.setItem('customerName', currentUserName);
      showChatScreen();
    } else {
      showError(data.message || 'Invalid email or password');
    }
  } catch (err) {
    console.error('Login error:', err);
    showError('Connection error. Please try again.');
  }
}

function showChatScreen() {
  loginScreen.classList.add('hidden');
  chatScreen.classList.remove('hidden');
  userInfo.classList.remove('hidden');
  logoutBtn.classList.remove('hidden');
  usernameDisplay.textContent = currentUserName;
  messageInput.focus();
}

function handleLogout() {
  accessToken = null;
  currentUserName = null;
  sessionStorage.removeItem('accessToken');
  sessionStorage.removeItem('customerName');
  chatScreen.classList.add('hidden');
  loginScreen.classList.remove('hidden');
  userInfo.classList.add('hidden');
  logoutBtn.classList.add('hidden');
  // Reset chat
  messagesContainer.innerHTML = getWelcomeHTML();
  bindQuickActions();
}

function startNewChat() {
  messagesContainer.innerHTML = getWelcomeHTML();
  bindQuickActions();
  messageInput.value = '';
  handleInputChange();
  messageInput.focus();
}

function getWelcomeHTML() {
  return `
    <div class="welcome-message">
      <img src="https://www.rbcroyalbank.com/dvl/v1.0/assets/images/logos/rbc-logo-shield.svg" alt="RBC" class="welcome-logo">
      <h2>How can I help you today?</h2>
      <p>Ask me about your accounts, transactions, cards, loans, or any banking questions.</p>
      <div class="quick-actions">
        <button class="quick-action" data-message="What's my account balance?">
          <i class="fas fa-wallet"></i> Check Balance
        </button>
        <button class="quick-action" data-message="Show my recent transactions">
          <i class="fas fa-history"></i> Recent Transactions
        </button>
        <button class="quick-action" data-message="What cards do I have?">
          <i class="fas fa-credit-card"></i> My Cards
        </button>
        <button class="quick-action" data-message="Do I have any loans?">
          <i class="fas fa-file-invoice-dollar"></i> My Loans
        </button>
      </div>
    </div>
  `;
}

function bindQuickActions() {
  document.querySelectorAll('.quick-action').forEach(btn => {
    btn.addEventListener('click', () => {
      const message = btn.dataset.message;
      messageInput.value = message;
      handleInputChange();
      sendMessage();
    });
  });
}

// Send message
async function sendMessage() {
  const message = messageInput.value.trim();
  if (!message || !accessToken) return;

  // Clear input
  messageInput.value = '';
  handleInputChange();

  // Remove welcome message if present
  const welcome = messagesContainer.querySelector('.welcome-message');
  if (welcome) welcome.remove();

  // Add user message
  appendMessage('user', message);

  // Show typing indicator
  const typingId = showTypingIndicator();

  try {
    const res = await fetch('/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`
      },
      body: JSON.stringify({ message })
    });

    removeTypingIndicator(typingId);

    if (res.status === 401) {
      handleLogout();
      showError('Session expired. Please login again.');
      return;
    }

    const data = await res.json();
    appendMessage('assistant', data.reply);
  } catch (err) {
    removeTypingIndicator(typingId);
    console.error('Chat error:', err);
    showError('Failed to send message. Please try again.');
  }
}

// Message rendering
function appendMessage(role, content) {
  const messageDiv = document.createElement('div');
  messageDiv.className = `message ${role}`;

  const avatar = role === 'user'
    ? `<div class="message-avatar"><i class="fas fa-user"></i></div>`
    : `<div class="message-avatar"><img src="https://www.rbcroyalbank.com/dvl/v1.0/assets/images/logos/rbc-logo-shield.svg" alt="RBC"></div>`;

  const sender = role === 'user' ? 'You' : 'RBC Assistant';

  messageDiv.innerHTML = `
    <div class="message-header">
      ${avatar}
      <span class="message-sender">${sender}</span>
    </div>
    <div class="message-content">${role === 'user' ? escapeHtml(content) : parseMarkdown(content)}</div>
  `;

  messagesContainer.appendChild(messageDiv);
  scrollToBottom();
}

function showTypingIndicator() {
  const id = 'typing-' + Date.now();
  const div = document.createElement('div');
  div.id = id;
  div.className = 'message assistant';
  div.innerHTML = `
    <div class="message-header">
      <div class="message-avatar"><img src="https://www.rbcroyalbank.com/dvl/v1.0/assets/images/logos/rbc-logo-shield.svg" alt="RBC"></div>
      <span class="message-sender">RBC Assistant</span>
    </div>
    <div class="typing-indicator">
      <span></span><span></span><span></span>
    </div>
  `;
  messagesContainer.appendChild(div);
  scrollToBottom();
  return id;
}

function removeTypingIndicator(id) {
  const el = document.getElementById(id);
  if (el) el.remove();
}

function scrollToBottom() {
  messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// Markdown parser
function parseMarkdown(text) {
  // Preserve sources section
  let sourcesDiv = '';
  const sourcesMatch = text.match(/<div class='sources-section'>[\s\S]*?<\/div>$/);
  if (sourcesMatch) {
    sourcesDiv = sourcesMatch[0];
    text = text.replace(sourcesDiv, '');
  }

  // Bold and italic
  text = text.replace(/\*\*([^*]+?)\*\*/g, '<strong>$1</strong>');
  text = text.replace(/\*([^*]+?)\*/g, '<em>$1</em>');

  // Ordered lists
  text = text.replace(/(\d+\.\s+.*(?:\n|$))+/g, match => {
    const items = match.split(/\n/).filter(item => /^\d+\.\s+/.test(item));
    const listItems = items.map(item => `<li>${item.replace(/^\d+\.\s+/, '')}</li>`).join('');
    return `<ol>${listItems}</ol>`;
  });

  // Unordered lists
  text = text.replace(/(\*\s+.*(?:\n|$))+/g, match => {
    const items = match.split(/\n/).filter(item => /^\*\s+/.test(item));
    const listItems = items.map(item => `<li>${item.replace(/^\*\s+/, '')}</li>`).join('');
    return `<ul>${listItems}</ul>`;
  });

  // Headers
  text = text.replace(/^### (.*?)$/gm, '<h3>$1</h3>');
  text = text.replace(/^## (.*?)$/gm, '<h2>$1</h2>');
  text = text.replace(/^# (.*?)$/gm, '<h1>$1</h1>');

  // Code
  text = text.replace(/`(.*?)`/g, '<code>$1</code>');

  // Paragraphs
  const paragraphs = text.split(/\n\n+/);
  text = paragraphs.map(p => p.trim() ? `<p>${p}</p>` : '').join('');
  text = text.replace(/\n/g, '<br>');

  return text + sourcesDiv;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Error handling
function showError(message) {
  const toast = document.createElement('div');
  toast.className = 'error-toast';
  toast.textContent = message;
  document.body.appendChild(toast);
  setTimeout(() => toast.remove(), 4000);
}
