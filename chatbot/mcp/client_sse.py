"""Interactive Banking Agent using OpenRouter and MCP."""

import os
import asyncio
import sys
import json
import random
from typing import Dict, List, Any, Optional, Tuple

# Add the parent directory to the Python path to import from src and chatbot
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "../..")))

from mcp import ClientSession
from mcp.client.sse import sse_client
from openai import AsyncOpenAI
from dotenv import load_dotenv

# Import custom modules
from chatbot.config import DEFAULT_USER_ID, ACCOUNT_MAPPINGS
from chatbot.config_client import SYSTEM_INSTRUCTIONS, TOOL_DEFINITIONS, MODEL_CONFIG
from chatbot.response_formatter import ResponseFormatter
from chatbot.intent_detector import IntentDetector

# Load environment variables
load_dotenv("../../.env")

# Configure OpenRouter client
openrouter_client = AsyncOpenAI(
    api_key=os.getenv("OPENROUTER_API_KEY"), base_url="https://openrouter.ai/api/v1"
)


class InteractiveBankingAssistant:
    """Interactive banking agent using OpenRouter and MCP."""

    def __init__(self):
        """Initialize the banking assistant."""
        self.conversation_history = []
        self.user_id = DEFAULT_USER_ID
        self.session = None
        self.read_stream = None
        self.write_stream = None
        self.account_mappings = ACCOUNT_MAPPINGS

    async def initialize_session(self):
        """Initialize the MCP session."""
        from chatbot.config import MCP_HOST, MCP_PORT

        mcp_url = f"http://{MCP_HOST}:{MCP_PORT}/sse"
        self.sse_client = sse_client(mcp_url)
        self.read_stream, self.write_stream = await self.sse_client.__aenter__()
        self.session = ClientSession(self.read_stream, self.write_stream)
        await self.session.__aenter__()
        await self.session.initialize()
        print("\n🔄 Connected to RBC Banking Agent")

    async def close_session(self):
        """Close the MCP session."""
        if self.session:
            await self.session.__aexit__(None, None, None)
        if hasattr(self, "sse_client"):
            await self.sse_client.__aexit__(None, None, None)

    async def _process_response(self, response):
        """Process the response from OpenRouter, handling function calls."""
        try:
            result = []
            has_function_call = False

            # Get the assistant message from response
            assistant_message = response.choices[0].message

            # Handle text content
            if assistant_message.content:
                text = assistant_message.content.strip()
                if text:
                    result.append(text)

            # Handle tool calls (function calls)
            if assistant_message.tool_calls:
                has_function_call = True
                for tool_call in assistant_message.tool_calls:
                    function_name = tool_call.function.name

                    # Skip empty function calls
                    if not function_name or function_name.strip() == "":
                        continue

                    # Parse function arguments
                    import json

                    try:
                        args_dict = json.loads(tool_call.function.arguments)
                    except:
                        args_dict = {}

                    # Execute the function call through MCP and wait for result
                    try:
                        function_result = await self._execute_function_call(
                            function_name, args_dict
                        )

                        # Parse the function result to extract actual data
                        parsed_result = self._parse_function_result(function_result)

                        # Format the result using the ResponseFormatter
                        formatted_result = ResponseFormatter.format_response(
                            function_name, parsed_result
                        )
                        if formatted_result:
                            result.append(formatted_result)

                    except Exception as e:
                        print(f"[ERROR] Tool execution failed: {e}")
                        error_msg = "I'm sorry, I'm having trouble processing your request right now. Please try again."
                        result.append(error_msg)

            # For simple greetings with no function calls, provide a friendly response
            if not has_function_call and not result:
                return "Hello! How can I help with your banking needs today?"

            # If there's an empty function call, provide a generic response
            if has_function_call and not result:
                return "How can I help you with your banking needs today?"

            return (
                "\n".join(result)
                if result
                else "Hello! How can I help with your banking needs today?"
            )
        except Exception as e:
            return f"Error processing response: {str(e)}"

    def _parse_function_result(self, result):
        """Parse the function result to extract the actual data."""
        try:
            # Check if result has content attribute (MCP response format)
            if hasattr(result, "content") and result.content:
                # Extract text content
                text_contents = []
                for content in result.content:
                    if hasattr(content, "text"):
                        text_contents.append(content.text)

                # Try to parse each text content as JSON
                parsed_contents = []
                for text in text_contents:
                    try:
                        parsed_contents.append(json.loads(text))
                    except:
                        parsed_contents.append(text)

                # Special handling for get_transaction_history
                # If we have a single transaction object, wrap it in a list
                if len(parsed_contents) == 1:
                    content = parsed_contents[0]
                    # Check if this looks like a transaction object
                    if isinstance(content, dict) and all(
                        key in content
                        for key in ["transaction_id", "date", "description"]
                    ):
                        return [content]
                    return content

                return parsed_contents

            # If it's a dictionary or list, return as is
            if isinstance(result, (dict, list)):
                return result

            # If it's a string that looks like JSON, parse it
            if isinstance(result, str):
                try:
                    if result.strip().startswith("{") or result.strip().startswith("["):
                        parsed = json.loads(result)
                        # Special handling for transaction objects
                        if isinstance(parsed, dict) and all(
                            key in parsed
                            for key in ["transaction_id", "date", "description"]
                        ):
                            return [parsed]
                        return parsed
                except:
                    pass

            # Return as is if we couldn't parse it
            return result
        except Exception as e:
            print(f"Error parsing function result: {e}")
            return result

    async def _execute_function_call(self, function_name, args):
        """Execute a function call through the MCP session."""
        try:
            # Check if function name is empty or invalid - this should never happen now
            if not function_name or function_name.strip() == "":
                # Just return silently without warning
                return {
                    "content": "No valid function specified",
                    "skip_response": True,
                    "error": True,
                }

            # Create a new dict from args to avoid modifying the original
            mcp_args = dict(args) if args else {}

            # ALWAYS override user_id with the actual logged-in user (don't trust model)
            if function_name != "answer_banking_question":
                mcp_args["user_id"] = self.user_id

            print(f"\n🔧 Executing function: {function_name} with args: {mcp_args}")

            # Call the function through MCP
            result = await self.session.call_tool(function_name, mcp_args)

            # Format the result for logging
            result_str = self._format_result_for_logging(result)

            # Print the result for debugging
            print(f"\n🔧 Function Result ({function_name}):")
            print(result_str)

            return result
        except Exception as e:
            error_msg = f"Error executing function {function_name}: {str(e)}"
            print(f"\n❌ {error_msg}")
            return {"error": error_msg}

    def _format_result_for_logging(self, result):
        """Format a result object for logging."""
        try:
            if isinstance(result, (dict, list)):
                return json.dumps(result, indent=2, default=str)
            return str(result)
        except Exception as e:
            print(f"Error formatting result for logging: {e}")
            return str(result)

    def build_prompt(self, user_input):
        """Build the prompt with conversation history."""
        # Get system instructions from config and format with user_id
        system_prompt = SYSTEM_INSTRUCTIONS.format(user_id=self.user_id)

        # Add conversation history
        history = "\n\n".join(
            [
                f"User: {msg['content']}"
                if msg["role"] == "user"
                else f"Assistant: {msg['content']}"
                for msg in self.conversation_history
            ]
        )

        # Add current user input
        full_prompt = (
            f"{system_prompt}\n\n{history}\n\nUser: {user_input}\n\nAssistant:"
        )
        return full_prompt

    async def send_message(self, user_input, customer_id=None):
        """Send a message to the assistant and get a response."""
        # Use provided customer_id or fall back to default
        if customer_id:
            self.user_id = customer_id

        # Print user input for debugging
        print(f"\n💬 User ({self.user_id}): {user_input}")

        # Add user message to history
        self.conversation_history.append({"role": "user", "content": user_input})

        # Check for commands first
        command, arg = IntentDetector.detect_command(user_input)
        if command:
            if command == "exit":
                return "Goodbye! Thank you for using RBC Banking Agent."
            elif command == "clear":
                self.conversation_history = []
                return "Conversation history cleared."
            elif command == "user" and arg:
                self.user_id = arg
                return f"User ID changed to: {self.user_id}"

        # For non-greetings, build the prompt with history
        try:
            # Generate content with system instructions from config
            system_instructions = SYSTEM_INSTRUCTIONS.format(user_id=self.user_id)

            # Prepare messages with system instruction and conversation history
            messages = [{"role": "system", "content": system_instructions}]
            messages.extend(
                self.conversation_history[-5:]
            )  # Keep last 5 messages for context

            # Call OpenRouter API with tools
            response = await openrouter_client.chat.completions.create(
                model=MODEL_CONFIG["model_name"],
                messages=messages,
                temperature=MODEL_CONFIG["temperature"],
                tools=[
                    {"type": "function", "function": tool_def}
                    for tool_def in TOOL_DEFINITIONS
                ],
                tool_choice="auto",
            )

            # Process and print response
            assistant_response = await self._process_response(response)

            print("\n🔁 Assistant:")
            print(assistant_response)

            # Add assistant response to history
            self.conversation_history.append(
                {"role": "assistant", "content": assistant_response}
            )

            return assistant_response

        except Exception as e:
            print(f"[ERROR] Message processing failed: {e}")
            return "I'm sorry, I'm having trouble right now. Please try again in a moment."

    async def run_interactive(self):
        """Run the assistant in interactive mode."""
        try:
            await self.initialize_session()

            print("\n🏦 RBC Banking Agent")
            print("Type 'exit' to quit, 'clear' to clear conversation history")
            print("Type 'user <id>' to change user ID (current: " + self.user_id + ")")

            while True:
                try:
                    user_input = input("\n💬 You: ")

                    # Check for exit command
                    command, _ = IntentDetector.detect_command(user_input)
                    if command == "exit":
                        print("Goodbye! Thank you for using RBC Banking Agent.")
                        break

                    # Process the message
                    response = await self.send_message(user_input)

                    # If the response indicates exit, break the loop
                    if command == "exit":
                        break

                except KeyboardInterrupt:
                    print("\nInterrupted. Exiting...")
                    break
                except Exception as e:
                    print(f"\n❌ Error: {str(e)}")
                    print("Continuing...")
        finally:
            print("\nClosing connection...")
            await self.close_session()


async def main():
    """Main entry point for the interactive banking agent."""
    try:
        assistant = InteractiveBankingAssistant()
        await assistant.run_interactive()
    except Exception as e:
        print(f"\n❌ Fatal error: {str(e)}")
        print("The assistant has encountered an unrecoverable error and must exit.")


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nProgram terminated by user.")
