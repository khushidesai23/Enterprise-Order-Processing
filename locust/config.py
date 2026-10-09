"""Optional Locust defaults. Credentials and entity IDs are environment-only."""
import os

BASE_URL = os.getenv("LOCUST_HOST", "http://localhost:8080")
ADMIN_EMAIL = os.getenv("LOCUST_EMAIL", "")
ADMIN_PASSWORD = os.getenv("LOCUST_PASSWORD", "")
PRODUCT_ID = os.getenv("LOCUST_PRODUCT_ID", "")
