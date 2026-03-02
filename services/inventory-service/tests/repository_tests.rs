#[cfg(test)]
mod tests {
    use chrono::Utc;
    use uuid::Uuid;

    use shinkansen_inventory::database::Database;
    use shinkansen_inventory::repository::Repository;

    async fn setup_test_db() -> Database {
        let db_url = std::env::var("TEST_DATABASE_URL")
            .unwrap_or_else(|_| "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen".to_string());

        let db = Database::new(&db_url).await.expect("Failed to connect to database");
        let repo = Repository::new(db.clone());
        repo.clear_all().await.expect("Failed to clear test data");

        db
    }

    #[tokio::test]
    async fn test_repository_get_stock_not_found() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let result = repo.get_stock(
            Uuid::new_v4(),
            None,
            Uuid::new_v4(),
        ).await;

        assert!(result.is_ok());
        assert!(result.unwrap().is_none());
    }

    #[tokio::test]
    async fn test_repository_create_stock() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let quantity = 100;

        let stock_id = repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            quantity,
        ).await.expect("Failed to create stock");

        assert_ne!(stock_id, Uuid::nil());

        let result = repo.get_stock(product_id, None, warehouse_id).await;
        
        let stock = result
            .expect("Failed to get stock")
            .expect("Stock not found");

        assert_eq!(stock.product_id, product_id);
        assert_eq!(stock.warehouse_id, warehouse_id);
        assert_eq!(stock.quantity, quantity);
        assert_eq!(stock.available_quantity, quantity);
        assert_eq!(stock.reserved_quantity, 0);
    }

    #[tokio::test]
    async fn test_repository_update_existing_stock() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create initial stock
        repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            50,
        ).await.expect("Failed to create stock");

        // Update with additional quantity
        repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            25,
        ).await.expect("Failed to update stock");

        let stock = repo.get_stock(product_id, None, warehouse_id).await
            .expect("Failed to get stock")
            .expect("Stock not found");

        assert_eq!(stock.quantity, 75);
        assert_eq!(stock.available_quantity, 75);
    }

    #[tokio::test]
    async fn test_repository_reserve_stock_success() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        // Create stock
        let stock_id = repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            100,
        ).await.expect("Failed to create stock");

        // Reserve stock
        let quantity = 10;
        let expires_at = Utc::now() + chrono::Duration::hours(1);

        repo.reserve_stock(
            order_id,
            stock_id,
            quantity,
            expires_at,
        ).await.expect("Failed to reserve stock");

        // Verify reservation
        let stock = repo.get_stock(product_id, None, warehouse_id).await
            .expect("Failed to get stock")
            .expect("Stock not found");

        assert_eq!(stock.quantity, 100);
        assert_eq!(stock.reserved_quantity, quantity);
        assert_eq!(stock.available_quantity, 90);
    }

    #[tokio::test]
    async fn test_repository_reserve_stock_insufficient() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        // Create stock with limited quantity
        let stock_id = repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            5,
        ).await.expect("Failed to create stock");

        // Try to reserve more than available
        let quantity = 10;
        let expires_at = Utc::now() + chrono::Duration::hours(1);

        let result = repo.reserve_stock(
            order_id,
            stock_id,
            quantity,
            expires_at,
        ).await;

        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("Insufficient stock"));
    }

    #[tokio::test]
    async fn test_repository_release_stock() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        // Create stock
        let stock_id = repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            100,
        ).await.expect("Failed to create stock");

        // Reserve stock
        let quantity = 10;
        let expires_at = Utc::now() + chrono::Duration::hours(1);

        repo.reserve_stock(
            order_id,
            stock_id,
            quantity,
            expires_at,
        ).await.expect("Failed to reserve stock");

        // Release stock
        repo.release_stock(order_id).await.expect("Failed to release stock");

        // Verify release
        let stock = repo.get_stock(product_id, None, warehouse_id).await
            .expect("Failed to get stock")
            .expect("Stock not found");

        assert_eq!(stock.quantity, 100);
        assert_eq!(stock.reserved_quantity, 0);
        assert_eq!(stock.available_quantity, 100);
    }

    #[tokio::test]
    async fn test_repository_create_stock_movement() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create stock
        let stock_id = repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            100,
        ).await.expect("Failed to create stock");

        // Create movement
        let movement_id = repo.create_stock_movement(
            stock_id,
            "MOVEMENT_TYPE_INBOUND",
            50,
            Some("Initial stocking"),
        ).await.expect("Failed to create movement");

        assert_ne!(movement_id, Uuid::nil());

        // List movements
        let movements = repo.list_stock_movements(stock_id, 10, 0).await
            .expect("Failed to list movements");

        assert_eq!(movements.len(), 1);
        assert_eq!(movements[0].stock_item_id, stock_id);
        assert_eq!(movements[0].quantity, 50);
        assert_eq!(movements[0].movement_type, "MOVEMENT_TYPE_INBOUND");
    }

    #[tokio::test]
    async fn test_repository_update_stock_quantity() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create stock
        repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            100,
        ).await.expect("Failed to create stock");

        // Update quantity
        repo.update_stock_quantity(
            product_id,
            None,
            warehouse_id,
            -30,
        ).await.expect("Failed to update stock");

        let stock = repo.get_stock(product_id, None, warehouse_id).await
            .expect("Failed to get stock")
            .expect("Stock not found");

        assert_eq!(stock.quantity, 70);
    }

    #[tokio::test]
    async fn test_repository_stock_with_variant() {
        let db = setup_test_db().await;
        let repo = Repository::new(db);

        let product_id = Uuid::new_v4();
        let variant_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create stock with variant
        let _stock_id = repo.create_or_update_stock(
            product_id,
            Some(variant_id),
            warehouse_id,
            50,
        ).await.expect("Failed to create stock");

        // Get stock with variant
        let stock = repo.get_stock(product_id, Some(variant_id), warehouse_id).await
            .expect("Failed to get stock")
            .expect("Stock not found");

        assert_eq!(stock.variant_id, Some(variant_id));
        assert_eq!(stock.quantity, 50);

        // Get stock without variant should return none
        let result = repo.get_stock(product_id, None, warehouse_id).await
            .expect("Failed to query stock");

        assert!(result.is_none());
    }
}
