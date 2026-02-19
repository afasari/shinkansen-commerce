use anyhow::Result;
use chrono::{DateTime, Utc};
use sqlx::postgres::PgPool;
use uuid::Uuid;
use crate::database::Database;

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct StockItem {
    pub id: Uuid,
    pub product_id: Uuid,
    pub variant_id: Option<Uuid>,
    pub warehouse_id: Uuid,
    pub quantity: i32,
    pub reserved_quantity: i32,
    pub available_quantity: i32,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct StockMovement {
    pub id: Uuid,
    pub stock_item_id: Uuid,
    pub movement_type: String,
    pub quantity: i32,
    pub reference: Option<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, sqlx::FromRow)]
pub struct StockReservation {
    pub id: Uuid,
    pub order_id: Uuid,
    pub stock_item_id: Uuid,
    pub quantity: i32,
    pub expires_at: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
}

pub struct Repository {
    db: Database,
}

impl Repository {
    pub fn new(db: Database) -> Self {
        Self { db }
    }
    
    pub fn pool(&self) -> &PgPool {
        self.db.pool()
    }
    
    pub async fn get_stock(
        &self,
        product_id: Uuid,
        variant_id: Option<Uuid>,
        warehouse_id: Uuid,
    ) -> Result<Option<StockItem>> {
        let stock = sqlx::query_as::<_, StockItem>(
            "SELECT id, product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at FROM inventory.stock_items WHERE product_id = $1 AND variant_id = $2 AND warehouse_id = $3"
        )
        .bind(product_id)
        .bind(variant_id)
        .bind(warehouse_id)
        .fetch_optional(self.pool())
        .await?;

        Ok(stock)
    }

    pub async fn get_stock_by_id(&self, stock_item_id: Uuid) -> Result<Option<StockItem>> {
        let stock = sqlx::query_as::<_, StockItem>(
            "SELECT id, product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at FROM inventory.stock_items WHERE id = $1"
        )
        .bind(stock_item_id)
        .fetch_optional(self.pool())
        .await?;

        Ok(stock)
    }
    
    pub async fn create_or_update_stock(
        &self,
        product_id: Uuid,
        variant_id: Option<Uuid>,
        warehouse_id: Uuid,
        quantity: i32,
    ) -> Result<Uuid> {
        let stock_id = sqlx::query_scalar(
            "INSERT INTO inventory.stock_items (product_id, variant_id, warehouse_id, quantity, reserved_quantity, available_quantity, created_at, updated_at)
             VALUES ($1, $2, $3, $4, 0, $4, NOW(), NOW())
             ON CONFLICT (product_id, variant_id, warehouse_id)
             DO UPDATE SET quantity = inventory.stock_items.quantity + $4, updated_at = NOW()
             RETURNING id"
        )
        .bind(product_id)
        .bind(variant_id)
        .bind(warehouse_id)
        .bind(quantity)
        .fetch_one(self.pool())
        .await?;
        
        Ok(stock_id)
    }
    
    pub async fn reserve_stock(
        &self,
        order_id: Uuid,
        stock_item_id: Uuid,
        quantity: i32,
        expires_at: DateTime<Utc>,
    ) -> Result<()> {
        let mut tx = self.db.pool().begin().await?;
        
        let updated = sqlx::query(
            "UPDATE inventory.stock_items
             SET reserved_quantity = reserved_quantity + $1, updated_at = NOW()
             WHERE id = $2 AND available_quantity >= $1
             RETURNING id"
        )
        .bind(quantity)
        .bind(stock_item_id)
        .fetch_optional(&mut *tx)
        .await?;
        
        if updated.is_none() {
            return Err(anyhow::anyhow!("Insufficient stock available for reservation"));
        }
        
        sqlx::query(
            "INSERT INTO inventory.stock_reservations (order_id, stock_item_id, quantity, expires_at, created_at)
             VALUES ($1, $2, $3, $4, NOW())
             ON CONFLICT (order_id, stock_item_id) DO UPDATE
                 SET quantity = stock_reservations.quantity + $3, expires_at = $4"
        )
        .bind(order_id)
        .bind(stock_item_id)
        .bind(quantity)
        .bind(expires_at)
        .execute(&mut *tx)
        .await?;
        
        tx.commit().await?;
        Ok(())
    }
    
    pub async fn release_stock(&self, order_id: Uuid) -> Result<()> {
        sqlx::query(
            "WITH released AS (
                 SELECT sr.stock_item_id, sr.quantity
                 FROM inventory.stock_reservations sr
                 WHERE sr.order_id = $1
             )
             UPDATE inventory.stock_items si
             SET reserved_quantity = si.reserved_quantity - released.quantity,
                 updated_at = NOW()
             FROM released
             WHERE si.id = released.stock_item_id"
        )
        .bind(order_id)
        .execute(self.pool())
        .await?;
        
        sqlx::query(
            "DELETE FROM inventory.stock_reservations WHERE order_id = $1"
        )
        .bind(order_id)
        .execute(self.pool())
        .await?;
        
        Ok(())
    }
    
    pub async fn update_stock_quantity(
        &self,
        product_id: Uuid,
        variant_id: Option<Uuid>,
        warehouse_id: Uuid,
        delta: i32,
    ) -> Result<()> {
        let new_quantity = if delta >= 0 {
            format!("GREATEST(0, inventory.stock_items.quantity + {})", delta)
        } else {
            format!("inventory.stock_items.quantity + {}", delta)
        };
        
        sqlx::query(&format!(
            "UPDATE inventory.stock_items
             SET quantity = {}, updated_at = NOW()
             WHERE product_id = $1 AND variant_id = $2 AND warehouse_id = $3",
            new_quantity
        ))
        .bind(product_id)
        .bind(variant_id)
        .bind(warehouse_id)
        .execute(self.pool())
        .await?;
        
        Ok(())
    }
    
    pub async fn create_stock_movement(
        &self,
        stock_item_id: Uuid,
        movement_type: &str,
        quantity: i32,
        reference: Option<&str>,
    ) -> Result<Uuid> {
        let movement_id = sqlx::query_scalar(
            "INSERT INTO inventory.stock_movements (stock_item_id, movement_type, quantity, reference, created_at)
             VALUES ($1, $2, $3, $4, NOW())
             RETURNING id"
        )
        .bind(stock_item_id)
        .bind(movement_type)
        .bind(quantity)
        .bind(reference)
        .fetch_one(self.pool())
        .await?;
        
        Ok(movement_id)
    }
    
    pub async fn list_stock_movements(
        &self,
        stock_item_id: Uuid,
        limit: i64,
        offset: i64,
    ) -> Result<Vec<StockMovement>> {
        let movements = sqlx::query_as(
            "SELECT id, stock_item_id, movement_type, quantity, reference, created_at
             FROM inventory.stock_movements
             WHERE stock_item_id = $1
             ORDER BY created_at DESC
             LIMIT $2 OFFSET $3"
        )
        .bind(stock_item_id)
        .bind(limit)
        .bind(offset)
        .fetch_all(self.pool())
        .await?;
        
        Ok(movements)
    }
}

#[cfg(test)]
impl Repository {
    pub async fn clear_all(&self) -> Result<()> {
        sqlx::query("DELETE FROM inventory.stock_reservations")
            .execute(self.pool())
            .await?;
        sqlx::query("DELETE FROM inventory.stock_movements")
            .execute(self.pool())
            .await?;
        sqlx::query("DELETE FROM inventory.stock_items")
            .execute(self.pool())
            .await?;
        Ok(())
    }
}
