import os
import sys
from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser
from langchain_core.runnables import RunnablePassthrough

# Handle imports whether called directly or from MCP
try:
    from chatbot.rag.vector_store import load_vector_store, create_vector_store
    from chatbot.rag.document_loader import load_documents, split_documents
except ImportError:
    from vector_store import load_vector_store, create_vector_store
    from document_loader import load_documents, split_documents

load_dotenv()

# Configure OpenRouter API
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY")
if not OPENROUTER_API_KEY:
    raise ValueError("OPENROUTER_API_KEY not found in environment variables")


class RBCChatbot:
    _instance = None

    def __new__(cls, *args, **kwargs):
        """Singleton pattern to ensure only one instance is created"""
        if cls._instance is None:
            cls._instance = super(RBCChatbot, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self, persist_directory=None):
        from chatbot.config import VECTOR_DB_DIR

        # Use config value if persist_directory is not provided
        if persist_directory is None:
            persist_directory = VECTOR_DB_DIR
        # Only initialize once
        if getattr(self, "_initialized", False):
            return

        # Initialize the vector store
        self._ensure_vector_store_exists(persist_directory)
        self.vector_store = load_vector_store(persist_directory)
        self.retriever = self.vector_store.as_retriever(search_kwargs={"k": 5})

        # Initialize the LLM with OpenRouter API
        api_key = os.getenv("OPENROUTER_API_KEY")
        self.llm = ChatOpenAI(
            model="openai/gpt-oss-20b",
            temperature=0.2,
            openai_api_key=api_key,
            openai_api_base="https://openrouter.ai/api/v1",
        )

        # System prompt
        self.system_prompt = """You are an AI agent for RBC Bank. Your purpose is to provide accurate information
about RBC's products, services, and policies based on the official documentation.
If you're unsure or the information isn't in the provided context, acknowledge that
and suggest the user contact RBC directly. Always be professional, helpful, and concise.

IMPORTANT: You must ONLY answer questions related to banking, financial services, or RBC products.
For any questions outside of these domains (like fitness, travel, cooking, etc.), politely decline
to answer and explain that you can only help with banking-related topics.

Context from documents:
{context}

Question: {question}

Answer based on the context above:"""

        # Create the prompt template
        self.prompt = ChatPromptTemplate.from_template(self.system_prompt)

        # Create the chain using LCEL
        self.chain = (
            {"context": self.retriever | self._format_docs, "question": RunnablePassthrough()}
            | self.prompt
            | self.llm
            | StrOutputParser()
        )

        self._initialized = True

    def _format_docs(self, docs):
        """Format retrieved documents into a string"""
        self._last_source_docs = docs  # Store for later retrieval
        return "\n\n".join(doc.page_content for doc in docs)

    def _ensure_vector_store_exists(self, persist_directory):
        """Make sure the vector store exists, create it if it doesn't"""
        from chatbot.config import DOCS_DIRECTORY

        if not os.path.exists(persist_directory):
            print("Vector store not found. Creating new vector store...")
            if os.path.exists(DOCS_DIRECTORY):
                documents = load_documents(DOCS_DIRECTORY)
                chunks = split_documents(documents)
                create_vector_store(chunks, persist_directory)
            else:
                print(f"Warning: Documents directory {DOCS_DIRECTORY} not found.")
                print("Creating empty vector store.")
                # Create an empty vector store
                create_vector_store([], persist_directory)

    def answer_question(self, question):
        """Answer a question using RAG"""
        try:
            self._last_source_docs = []  # Reset before query

            # Get the answer from the chain
            answer = self.chain.invoke(question)

            # Extract sources from stored docs
            sources = []
            for doc in self._last_source_docs:
                if hasattr(doc, "metadata") and "source" in doc.metadata:
                    sources.append(doc.metadata["source"])

            # Only include sources if the answer is actually about banking
            if (
                "I can only assist with banking" in answer
                or "I'm sorry, I can only answer" in answer
            ):
                sources = []

            # Return the answer and unique sources
            return {"answer": answer, "sources": list(set(sources))}
        except Exception as e:
            return {"answer": f"I encountered an error: {str(e)}", "sources": []}

    def get_relevant_documents(self, query):
        """Retrieve relevant documents for a query without generating an answer"""
        try:
            docs = self.vector_store.similarity_search(query, k=5)
            sources = []
            for doc in docs:
                if hasattr(doc, "metadata") and "source" in doc.metadata:
                    sources.append(doc.metadata["source"])
            return {"documents": docs, "sources": list(set(sources))}
        except Exception as e:
            return {"documents": [], "sources": [], "error": str(e)}
