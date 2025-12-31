import os
from langchain_community.vectorstores import Chroma
from langchain_openai import OpenAIEmbeddings
from dotenv import load_dotenv

load_dotenv()

# Configure OpenRouter API with the API key from .env
api_key = os.getenv("OPENROUTER_API_KEY")
if not api_key:
    raise ValueError("OPENROUTER_API_KEY not found in environment variables")

from chatbot.config import VECTOR_DB_DIR


def create_vector_store(documents, persist_directory=None):
    if persist_directory is None:
        persist_directory = VECTOR_DB_DIR
    """Create a vector store from document chunks"""
    # Initialize the embeddings using OpenAI-compatible embeddings (via OpenRouter)
    embeddings = OpenAIEmbeddings(
        model="text-embedding-3-small",
        openai_api_key=api_key,
        openai_api_base="https://openrouter.ai/api/v1",
        openai_api_type="openrouter",
    )

    # Create the vector store (persistence is automatic)
    vector_store = Chroma.from_documents(
        documents=documents, embedding=embeddings, persist_directory=persist_directory
    )
    print(f"Vector store created with {len(documents)} document chunks")
    print(f"Vector store persisted to {persist_directory}")
    return vector_store


def load_vector_store(persist_directory=None):
    if persist_directory is None:
        persist_directory = VECTOR_DB_DIR
    """Load an existing vector store"""
    # Use the same API key as in create_vector_store
    embeddings = OpenAIEmbeddings(
        model="text-embedding-3-small",
        openai_api_key=api_key,
        openai_api_base="https://openrouter.ai/api/v1",
        openai_api_type="openrouter",
    )
    vector_store = Chroma(
        persist_directory=persist_directory, embedding_function=embeddings
    )
    return vector_store
