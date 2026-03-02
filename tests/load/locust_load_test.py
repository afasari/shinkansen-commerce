"""
Load testing script for Shinkansen Commerce using Locust.

This simulates realistic user behavior on the e-commerce platform.
"""

import random
import time
from locust import HttpUser, task, between, events
from locust.runners import MasterRunner


# Configuration
BASE_URL = "http://localhost:8080"
DEFAULT_PASSWORD = "password123"


class ShinkansenUser(HttpUser):
    """Simulated user behavior for Shinkansen Commerce"""

    wait_time = between(1, 5)
    host = BASE_URL

    def on_start(self):
        """Called when a user starts"""
        self.login()

    @task(3)
    def browse_products(self):
        """Browse product listings"""
        with self.client.get("/v1/products", params={
            "page": random.randint(1, 5),
            "limit": random.choice([20, 50, 100]),
        }, catch_response=True) as response:
            if response.status_code == 200:
                data = response.json()
                if data.get("products"):
                    product = random.choice(data["products"])
                    self.view_product_details(product["id"])

    @task(2)
    def search_products(self):
        """Search for products"""
        search_terms = ["electronics", "books", "food", "clothing", "home"]

        with self.client.get("/v1/products/search", params={
            "query": random.choice(search_terms),
            "page": 1,
            "limit": 20,
        }, name="/v1/products/search", catch_response=True) as response:
            if response.status_code == 200:
                data = response.json()
                # Occasionally view a product from search results
                if data.get("products") and random.random() < 0.3:
                    product = random.choice(data["products"])
                    self.view_product_details(product["id"])

    def view_product_details(self, product_id):
        """View detailed product information"""
        with self.client.get(f"/v1/products/{product_id}",
                            name="/v1/products/[id]",
                            catch_response=True) as response:
            if response.status_code == 200:
                # 20% chance to add to cart
                if random.random() < 0.2:
                    self.add_to_cart(product_id)

    @task(1)
    def view_categories(self):
        """View products by category"""
        categories = ["electronics", "books", "food", "home"]

        with self.client.get("/v1/products", params={
            "category_id": random.choice(categories),
            "page": 1,
            "limit": 20,
        }, catch_response=True) as response:
            if response.status_code == 200:
                data = response.json()
                if data.get("products"):
                    product = random.choice(data["products"])
                    self.view_product_details(product["id"])

    @task(2)
    def view_cart(self):
        """View shopping cart"""
        self.client.get("/v1/cart", catch_response=True)

    def add_to_cart(self, product_id):
        """Add product to cart"""
        with self.client.post("/v1/cart/items",
                           json={
                               "product_id": product_id,
                               "quantity": random.randint(1, 3),
                           },
                           catch_response=True) as response:
            if response.status_code in [200, 201]:
                # Maybe view cart after adding
                if random.random() < 0.5:
                    self.view_cart()

    @task(1)
    def checkout(self):
        """Go through checkout process (without completing payment)"""
        # View cart first
        with self.client.get("/v1/cart", catch_response=True) as response:
            if response.status_code != 200:
                return

        # Create order (cart to order conversion)
        with self.client.post("/v1/orders",
                           json={
                               "items": [],  # Would populate from cart
                               "shipping_address": {
                                   "name": "Test User",
                                   "phone": "090-1234-5678",
                                   "postal_code": "100-0001",
                                   "prefecture": "Tokyo",
                                   "city": "Chiyoda-ku",
                                   "address_line1": "1-1-1 Otemachi",
                               },
                               "payment_method": random.choice([1, 2, 5]),  # Credit Card, Konbini, PayPay
                           },
                           catch_response=True) as response:
            # Order creation may fail without actual cart items
            pass

    @task(1)
    def view_orders(self):
        """View order history"""
        with self.client.get("/v1/orders", params={
            "page": random.randint(1, 3),
            "limit": 10,
        }, catch_response=True):
            pass

    def login(self):
        """Simulate user login"""
        username = f"user{random.randint(1, 1000)}@example.com"

        with self.client.post("/v1/users/login",
                           json={
                               "email": username,
                               "password": DEFAULT_PASSWORD,
                           },
                           catch_response=True) as response:
            if response.status_code == 200:
                data = response.json()
                # Store auth token for future requests
                if data.get("access_token"):
                    self.client.headers.update({
                        "Authorization": f"Bearer {data['access_token']}"
                    })


class AdminUser(HttpUser):
    """Simulated admin user behavior"""

    wait_time = between(2, 10)
    weight = 1  # Only 1 admin per 100 regular users

    def on_start(self):
        """Admin login"""
        with self.client.post("/v1/users/login",
                           json={
                               "email": "admin@shinkansen.com",
                               "password": "admin123",
                           },
                           catch_response=True) as response:
            if response.status_code == 200:
                data = response.json()
                if data.get("access_token"):
                    self.client.headers.update({
                        "Authorization": f"Bearer {data['access_token']}"
                    })

    @task
    def view_analytics(self):
        """View analytics dashboard"""
        self.client.get("/v1/admin/analytics", catch_response=True)

    @task
    def view_recent_orders(self):
        """View recent orders"""
        with self.client.get("/v1/admin/orders",
                          params={"limit": 50},
                          catch_response=True):
            pass


# Custom event handlers for better reporting
class Request(events.Request):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)


class ShinkansenCustomer(ShinkansenUser):
    """Regular customer user"""


class ShinkansenBrowser(ShinkansenUser):
    """User who mostly browses without purchasing"""

    @task(5)
    def browse_products(self):
        super().browse_products()

    @task(1)
    def view_product_details(self):
        pass


class ShinkansenBuyer(ShinkansenUser):
    """User focused on purchasing"""

    @task(1)
    def browse_products(self):
        super().browse_products()

    @task(5)
    def checkout(self):
        super().checkout()


# Event listeners for custom metrics
@events.init.add_listener
def on_locust_init(environment, runner, **kwargs):
    """Run on Locust initialization"""
    if isinstance(runner, MasterRunner):
        print("Starting master node")
    else:
        print(f"Starting worker node")


# Test configuration
if __name__ == "__main__":
    import os

    # Allow command line override of host
    host = os.getenv("TARGET_HOST", "http://localhost:8080")

    # Run with command: locust -f load_test.py --host=http://localhost:8080
    # Or with headless mode: locust -f load_test.py --headless --host=http://localhost:8080 -u 100 -r 10
    print(f"Starting load test against {host}")
