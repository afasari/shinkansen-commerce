"""
Analytics Worker for Shinkansen Commerce

This module handles:
- Batch analytics processing
- Kafka event consumption
- Data warehouse ETL
- AI/ML insights generation
"""

import asyncio
import json
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
from dataclasses import dataclass, asdict
from decimal import Decimal

import click
from aiokafka import AIOKafkaConsumer
import asyncpg
import redis.asyncio as aioredis
from pydantic import BaseModel, Field


logger = logging.getLogger(__name__)


# Event Models
class OrderEvent(BaseModel):
    event_id: str
    event_type: str
    order_id: str
    user_id: str
    status: str
    timestamp: datetime
    data: Dict[str, Any]


class ProductViewEvent(BaseModel):
    event_id: str
    product_id: str
    user_id: Optional[str]
    session_id: str
    timestamp: datetime


# Analytics Data Models
@dataclass
class DailySalesMetrics:
    date: datetime.date
    total_orders: int
    total_revenue: Decimal
    total_items_sold: int
    unique_customers: int
    average_order_value: Decimal
    payment_method_breakdown: Dict[str, int]


@dataclass
class ProductPerformanceMetrics:
    product_id: str
    product_name: str
    views_last_7d: int
    views_last_30d: int
    orders_last_7d: int
    orders_last_30d: int
    conversion_rate: float
    revenue_last_7d: Decimal
    revenue_last_30d: Decimal


@dataclass
class UserBehaviorMetrics:
    user_id: str
    total_orders: int
    total_spent: Decimal
    favorite_categories: List[str]
    average_order_value: Decimal
    last_order_date: Optional[datetime.date]
    days_since_last_order: Optional[int]


class AnalyticsWorker:
    """Main analytics worker class"""

    def __init__(
        self,
        postgres_url: str,
        redis_url: str,
        kafka_brokers: List[str],
    ):
        self.postgres_url = postgres_url
        self.redis_url = redis_url
        self.kafka_brokers = kafka_brokers
        self.db_pool: Optional[asyncpg.Pool] = None
        self.redis: Optional[aioredis.Redis] = None
        self.consumer: Optional[AIOKafkaConsumer] = None
        self.running = False

    async def start(self):
        """Start the analytics worker"""
        logger.info("Starting analytics worker")

        # Initialize database connection
        self.db_pool = await asyncpg.create_pool(self.postgres_url)
        logger.info("Connected to PostgreSQL")

        # Initialize Redis connection
        self.redis = await aioredis.from_url(self.redis_url)
        logger.info("Connected to Redis")

        # Initialize Kafka consumer
        self.consumer = AIOKafkaConsumer(
            'orders',
            'product-views',
            'user-events',
            bootstrap_servers=','.join(self.kafka_brokers),
            group_id='analytics-worker',
            auto_offset_reset='earliest',
        )
        await self.consumer.start()
        logger.info("Connected to Kafka")

        self.running = True

    async def stop(self):
        """Stop the analytics worker"""
        logger.info("Stopping analytics worker")
        self.running = False

        if self.consumer:
            await self.consumer.stop()

        if self.db_pool:
            await self.db_pool.close()

        if self.redis:
            await self.redis.close()

        logger.info("Analytics worker stopped")

    async def process_events(self):
        """Process events from Kafka"""
        logger.info("Starting event processing")

        while self.running:
            try:
                async for msg in self.consumer:
                    await self.handle_message(msg)
            except Exception as e:
                logger.error(f"Error processing message: {e}")
                await asyncio.sleep(1)

    async def handle_message(self, msg):
        """Handle a single Kafka message"""
        try:
            topic = msg.topic
            value = json.loads(msg.value.decode('utf-8'))

            if topic == 'orders':
                event = OrderEvent(**value)
                await self.process_order_event(event)
            elif topic == 'product-views':
                event = ProductViewEvent(**value)
                await self.process_product_view_event(event)
            elif topic == 'user-events':
                await self.process_user_event(value)

        except Exception as e:
            logger.error(f"Error handling message: {e}")

    async def process_order_event(self, event: OrderEvent):
        """Process an order event"""
        logger.debug(f"Processing order event: {event.event_type}")

        if event.event_type == 'order.created':
            await self.track_new_order(event)
        elif event.event_type == 'order.paid':
            await self.track_payment(event)
        elif event.event_type == 'order.delivered':
            await self.track_delivery(event)

    async def process_product_view_event(self, event: ProductViewEvent):
        """Process a product view event"""
        logger.debug(f"Processing product view: {event.product_id}")

        # Store in analytics database
        async with self.db_pool.acquire() as conn:
            await conn.execute(
                """
                INSERT INTO analytics.product_views
                (product_id, user_id, session_id, viewed_at)
                VALUES ($1, $2, $3, $4)
                """,
                event.product_id,
                event.user_id,
                event.session_id,
                event.timestamp,
            )

        # Update view counter in Redis
        date_key = event.timestamp.strftime('%Y-%m-%d')
        await self.redis.incr(f"product_views:{event.product_id}:{date_key}")

    async def process_user_event(self, event: Dict[str, Any]):
        """Process a user event"""
        logger.debug(f"Processing user event: {event.get('event_type')}")
        # Handle user-specific events

    async def track_new_order(self, event: OrderEvent):
        """Track a new order"""
        async with self.db_pool.acquire() as conn:
            await conn.execute(
                """
                INSERT INTO analytics.orders_daily
                (order_date, order_id, user_id, total_amount, payment_method)
                VALUES ($1, $2, $3, $4, $5)
                """,
                event.timestamp.date(),
                event.order_id,
                event.user_id,
                Decimal(str(event.data.get('total_amount', 0))),
                event.data.get('payment_method'),
            )

    async def track_payment(self, event: OrderEvent):
        """Track a payment completion"""
        async with self.db_pool.acquire() as conn:
            await conn.execute(
                """
                UPDATE analytics.orders_daily
                SET paid_at = $1, payment_status = 'paid'
                WHERE order_id = $2
                """,
                event.timestamp,
                event.order_id,
            )

    async def track_delivery(self, event: OrderEvent):
        """Track a delivery completion"""
        async with self.db_pool.acquire() as conn:
            await conn.execute(
                """
                UPDATE analytics.orders_daily
                SET delivered_at = $1, delivery_status = 'delivered'
                WHERE order_id = $2
                """,
                event.timestamp,
                event.order_id,
            )

    async def generate_daily_report(self, date: datetime.date) -> DailySalesMetrics:
        """Generate daily sales report"""
        logger.info(f"Generating daily report for {date}")

        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                SELECT
                    COUNT(*) as total_orders,
                    COALESCE(SUM(total_amount), 0) as total_revenue,
                    COALESCE(SUM(item_count), 0) as total_items_sold,
                    COUNT(DISTINCT user_id) as unique_customers,
                    COALESCE(AVG(total_amount), 0) as average_order_value
                FROM analytics.orders_daily
                WHERE order_date = $1 AND payment_status = 'paid'
                """,
                date,
            )

            # Get payment method breakdown
            payment_rows = await conn.fetch(
                """
                SELECT payment_method, COUNT(*) as count
                FROM analytics.orders_daily
                WHERE order_date = $1 AND payment_status = 'paid'
                GROUP BY payment_method
                """,
                date,
            )

            payment_breakdown = {r['payment_method']: r['count'] for r in payment_rows}

        return DailySalesMetrics(
            date=date,
            total_orders=row['total_orders'],
            total_revenue=row['total_revenue'],
            total_items_sold=row['total_items_sold'],
            unique_customers=row['unique_customers'],
            average_order_value=row['average_order_value'],
            payment_method_breakdown=payment_breakdown,
        )

    async def generate_product_performance_report(
        self,
        product_ids: List[str]
    ) -> List[ProductPerformanceMetrics]:
        """Generate product performance report"""
        logger.info(f"Generating product performance report for {len(product_ids)} products")

        end_date = datetime.now().date()
        start_date_7d = end_date - timedelta(days=7)
        start_date_30d = end_date - timedelta(days=30)

        async with self.db_pool.acquire() as conn:
            products = []
            for product_id in product_ids:
                # Get views
                views_7d = 0
                views_30d = 0

                # Get orders and revenue
                orders_7d = await conn.fetchval(
                    """
                    SELECT COALESCE(COUNT(*), 0)
                    FROM analytics.orders_daily od
                    JOIN analytics.order_items oi ON od.order_id = oi.order_id
                    WHERE oi.product_id = $1
                        AND od.order_date >= $2
                        AND od.payment_status = 'paid'
                    """,
                    product_id,
                    start_date_7d,
                )

                revenue_7d = await conn.fetchval(
                    """
                    SELECT COALESCE(SUM(oi.quantity * oi.unit_price), 0)
                    FROM analytics.orders_daily od
                    JOIN analytics.order_items oi ON od.order_id = oi.order_id
                    WHERE oi.product_id = $1
                        AND od.order_date >= $2
                        AND od.payment_status = 'paid'
                    """,
                    product_id,
                    start_date_7d,
                )

                # Calculate conversion rate
                conversion_rate = 0
                if views_7d > 0:
                    conversion_rate = orders_7d / views_7d

                # Get product name
                product_name = await conn.fetchval(
                    "SELECT name FROM catalog.products WHERE id = $1",
                    product_id,
                ) or "Unknown"

                products.append(ProductPerformanceMetrics(
                    product_id=product_id,
                    product_name=product_name,
                    views_last_7d=views_7d,
                    views_last_30d=views_30d,
                    orders_last_7d=orders_7d,
                    orders_last_30d=0,
                    conversion_rate=conversion_rate,
                    revenue_last_7d=Decimal(str(revenue_7d)),
                    revenue_last_30d=Decimal('0'),
                ))

        return products

    async def generate_user_behavior_report(
        self,
        user_id: str
    ) -> UserBehaviorMetrics:
        """Generate user behavior report for a specific user"""
        logger.info(f"Generating user behavior report for {user_id}")

        async with self.db_pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                SELECT
                    COUNT(*) as total_orders,
                    COALESCE(SUM(total_amount), 0) as total_spent,
                    COALESCE(AVG(total_amount), 0) as average_order_value,
                    MAX(order_date) as last_order_date
                FROM analytics.orders_daily
                WHERE user_id = $1 AND payment_status = 'paid'
                """,
                user_id,
            )

            # Get favorite categories
            cat_rows = await conn.fetch(
                """
                SELECT c.name, COUNT(*) as order_count
                FROM analytics.orders_daily od
                JOIN analytics.order_items oi ON od.order_id = oi.order_id
                JOIN catalog.products p ON oi.product_id = p.id
                JOIN catalog.categories c ON p.category_id = c.id
                WHERE od.user_id = $1 AND od.payment_status = 'paid'
                GROUP BY c.name
                ORDER BY order_count DESC
                LIMIT 5
                """,
                user_id,
            )

            favorite_categories = [r['name'] for r in cat_rows]

        last_order = row['last_order_date']
        days_since = None
        if last_order:
            days_since = (datetime.now().date() - last_order).days

        return UserBehaviorMetrics(
            user_id=user_id,
            total_orders=row['total_orders'],
            total_spent=Decimal(str(row['total_spent'])),
            favorite_categories=favorite_categories,
            average_order_value=Decimal(str(row['average_order_value'])),
            last_order_date=last_order,
            days_since_last_order=days_since,
        )

    async def run_etl_pipeline(self):
        """Run the ETL pipeline to transfer data to data warehouse"""
        logger.info("Running ETL pipeline")

        while self.running:
            try:
                # ETL for orders
                await self.etl_orders()

                # ETL for product views
                await self.etl_product_views()

                # ETL for user behaviors
                await self.etl_user_behaviors()

                # Wait before next run
                await asyncio.sleep(300)  # 5 minutes

            except Exception as e:
                logger.error(f"ETL pipeline error: {e}")
                await asyncio.sleep(60)

    async def etl_orders(self):
        """ETL process for orders"""
        async with self.db_pool.acquire() as conn:
            # Transfer new orders to data warehouse
            await conn.execute("""
                INSERT INTO dw.fact_orders (order_id, order_date, user_id, total_amount, payment_method)
                SELECT o.id, o.created_at::date, o.user_id, o.total_units, o.payment_method
                FROM order_service.orders o
                LEFT JOIN dw.fact_orders d ON o.id = d.order_id
                WHERE d.order_id IS NULL
                LIMIT 1000
            """)
        logger.debug("Orders ETL completed")

    async def etl_product_views(self):
        """ETL process for product views"""
        async with self.db_pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO dw.fact_product_views (product_id, view_date, user_id, session_id)
                SELECT product_id, viewed_at::date, user_id, session_id
                FROM analytics.product_views pv
                LEFT JOIN dw.fact_product_views d ON
                    pv.product_id = d.product_id AND
                    pv.viewed_at::date = d.view_date
                WHERE d.product_id IS NULL
                LIMIT 10000
            """)
        logger.debug("Product views ETL completed")

    async def etl_user_behaviors(self):
        """ETL process for user behaviors"""
        async with self.db_pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO dw.fact_user_behaviors (user_id, behavior_date, total_orders, total_spent)
                SELECT user_id, order_date, COUNT(*), SUM(total_amount)
                FROM analytics.orders_daily
                WHERE payment_status = 'paid'
                GROUP BY user_id, order_date
                ON CONFLICT (user_id, behavior_date) DO UPDATE SET
                    total_orders = EXCLUDED.total_orders,
                    total_spent = EXCLUDED.total_spent
            """)
        logger.debug("User behaviors ETL completed")


# CLI Interface
@click.group()
def cli():
    """Analytics Worker CLI"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )


@cli.command()
@click.option('--postgres-url', default='postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen', help='PostgreSQL connection URL')
@click.option('--redis-url', default='redis://localhost:6379', help='Redis connection URL')
@click.option('--kafka-brokers', default='localhost:9092', help='Kafka broker addresses')
def consume(postgres_url, redis_url, kafka_brokers):
    """Start consuming Kafka events"""
    worker = AnalyticsWorker(
        postgres_url=postgres_url,
        redis_url=redis_url,
        kafka_brokers=kafka_brokers.split(','),
    )

    async def run():
        await worker.start()
        try:
            await worker.process_events()
        finally:
            await worker.stop()

    asyncio.run(run())


@cli.command()
@click.option('--postgres-url', default='postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen', help='PostgreSQL connection URL')
@click.option('--redis-url', default='redis://localhost:6379', help='Redis connection URL')
@click.option('--kafka-brokers', default='localhost:9092', help='Kafka broker addresses')
@click.option('--date', type=click.DateTime(format='%Y-%m-%d'), default=None, help='Date for report (default: today)')
def report(postgres_url, redis_url, kafka_brokers, date):
    """Generate daily analytics report"""
    if date is None:
        date = datetime.now().date()
    else:
        date = date.date()

    worker = AnalyticsWorker(
        postgres_url=postgres_url,
        redis_url=redis_url,
        kafka_brokers=kafka_brokers.split(','),
    )

    async def generate():
        await worker.start()
        try:
            report = await worker.generate_daily_report(date)
            print(json.dumps(asdict(report), indent=2, default=str))
        finally:
            await worker.stop()

    asyncio.run(generate())


@cli.command()
@click.option('--postgres-url', default='postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen', help='PostgreSQL connection URL')
@click.option('--redis-url', default='redis://localhost:6379', help='Redis connection URL')
@click.option('--kafka-brokers', default='localhost:9092', help='Kafka broker addresses')
def etl(postgres_url, redis_url, kafka_brokers):
    """Run ETL pipeline"""
    worker = AnalyticsWorker(
        postgres_url=postgres_url,
        redis_url=redis_url,
        kafka_brokers=kafka_brokers.split(','),
    )

    async def run():
        await worker.start()
        try:
            await worker.run_etl_pipeline()
        finally:
            await worker.stop()

    asyncio.run(run())


if __name__ == '__main__':
    cli()
