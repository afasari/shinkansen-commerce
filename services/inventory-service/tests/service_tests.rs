#[cfg(test)]
mod tests {
    use std::time::Duration;
    use chrono::Utc;
    use prost_types::Timestamp;
    use tokio::time::sleep;
    use uuid::Uuid;

    use shinkansen_proto::shinkansen::inventory::{
        inventory_service_client::InventoryServiceClient,
        inventory_service_server::InventoryServiceServer,
        GetStockRequest, UpdateStockRequest, ReserveStockRequest, ReleaseStockRequest,
        StockReservationItem, GetStockMovementsRequest,
    };
    use shinkansen_proto::shinkansen::common::Pagination;
    use tonic::transport::Server;
    use tonic::Request;

    use shinkansen_inventory::service::InventoryServiceImpl;
    use shinkansen_inventory::database::Database;
    use shinkansen_inventory::repository::{Repository, StockItem};

    async fn setup_test_service() -> (Database, String) {
        let db_url = std::env::var("TEST_DATABASE_URL")
            .unwrap_or_else(|_| "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen".to_string());

        let db = Database::new(&db_url).await.expect("Failed to connect to database");
        let repo = Repository::new(db.clone());
        repo.clear_all().await.expect("Failed to clear test data");

        let service = InventoryServiceImpl::new(db.clone());
        let svc = InventoryServiceServer::new(service);

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        let server_url = format!("http://{}", addr);

        tokio::spawn(async move {
            Server::builder()
                .add_service(svc)
                .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(listener))
                .await
        });

        // Give server time to start
        sleep(Duration::from_millis(100)).await;

        (db, server_url)
    }

    #[allow(dead_code)]
    async fn create_test_stock(repo: &Repository, product_id: Uuid, quantity: i32) -> StockItem {
        let warehouse_id = Uuid::new_v4();
        repo.create_or_update_stock(
            product_id,
            None,
            warehouse_id,
            quantity,
        ).await.expect("Failed to create test stock");

        repo.get_stock(product_id, None, warehouse_id).await
            .expect("Failed to get created stock")
            .expect("Stock not found")
    }

    #[tokio::test]
    async fn test_get_stock_existing_item() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // First create a stock item
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 100,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        let _response = client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Now get the stock
        let request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
        });

        let response = client.get_stock(request)
            .await
            .expect("Failed to get stock");

        let stock = response.into_inner().stock.expect("No stock returned");
        assert_eq!(stock.product_id, product_id.to_string());
        assert_eq!(stock.quantity, 100);
        assert_eq!(stock.available_quantity, 100);
        assert_eq!(stock.reserved_quantity, 0);
    }

    #[tokio::test]
    async fn test_get_stock_nonexistent_item() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: Uuid::new_v4().to_string(),
            variant_id: String::new(),
            warehouse_id: Uuid::new_v4().to_string(),
        });

        let response = client.get_stock(request)
            .await
            .expect("Failed to get stock");

        let stock = response.into_inner().stock.expect("No stock returned");
        assert_eq!(stock.quantity, 0);
        assert_eq!(stock.available_quantity, 0);
        assert_eq!(stock.reserved_quantity, 0);
    }

    #[tokio::test]
    async fn test_update_stock_increase() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create initial stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 50,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Increase stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 25,
            reason: "Restock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to increase stock");

        // Verify
        let request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
        });

        let response = client.get_stock(request)
            .await
            .expect("Failed to get stock");

        let stock = response.into_inner().stock.expect("No stock returned");
        assert_eq!(stock.quantity, 75);
        assert_eq!(stock.available_quantity, 75);
    }

    #[tokio::test]
    async fn test_update_stock_decrease() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create initial stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 100,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Decrease stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: -30,
            reason: "Sale".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to decrease stock");

        // Verify
        let request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
        });

        let response = client.get_stock(request)
            .await
            .expect("Failed to get stock");

        let stock = response.into_inner().stock.expect("No stock returned");
        assert_eq!(stock.quantity, 70);
        assert_eq!(stock.available_quantity, 70);
    }

    #[tokio::test]
    async fn test_reserve_stock_success() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        // Create initial stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 100,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Reserve stock
        let request = Request::new(ReserveStockRequest {
            order_id: order_id.to_string(),
            items: vec![StockReservationItem {
                product_id: product_id.to_string(),
                variant_id: String::new(),
                warehouse_id: warehouse_id.to_string(),
                quantity: 10,
            }],
            expires_at: Some(Timestamp {
                seconds: (Utc::now() + chrono::Duration::hours(1)).timestamp(),
                nanos: 0,
            }),
        });

        let response = client.reserve_stock(request)
            .await
            .expect("Failed to reserve stock");

        let reserve_response = response.into_inner();
        assert!(reserve_response.success);
        assert!(reserve_response.failed_items.is_empty());

        // Verify stock levels
        let request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
        });

        let response = client.get_stock(request)
            .await
            .expect("Failed to get stock");

        let stock = response.into_inner().stock.expect("No stock returned");
        assert_eq!(stock.quantity, 100);
        assert_eq!(stock.reserved_quantity, 10);
        assert_eq!(stock.available_quantity, 90);
    }

    #[tokio::test]
    async fn test_reserve_stock_insufficient_quantity() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        // Create initial stock with limited quantity
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 5,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Try to reserve more than available
        let request = Request::new(ReserveStockRequest {
            order_id: order_id.to_string(),
            items: vec![StockReservationItem {
                product_id: product_id.to_string(),
                variant_id: String::new(),
                warehouse_id: warehouse_id.to_string(),
                quantity: 10,
            }],
            expires_at: None,
        });

        let response = client.reserve_stock(request)
            .await
            .expect("Failed to reserve stock");

        let reserve_response = response.into_inner();
        assert!(!reserve_response.success);
        assert_eq!(reserve_response.failed_items.len(), 1);
        assert_eq!(reserve_response.failed_items[0], product_id.to_string());
    }

    #[tokio::test]
    async fn test_release_stock() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();
        let order_id = Uuid::new_v4();

        // Create initial stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 100,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Reserve stock
        let request = Request::new(ReserveStockRequest {
            order_id: order_id.to_string(),
            items: vec![StockReservationItem {
                product_id: product_id.to_string(),
                variant_id: String::new(),
                warehouse_id: warehouse_id.to_string(),
                quantity: 10,
            }],
            expires_at: None,
        });

        client.reserve_stock(request)
            .await
            .expect("Failed to reserve stock");

        // Release stock
        let request = Request::new(ReleaseStockRequest {
            reservation_id: order_id.to_string(),
        });

        client.release_stock(request)
            .await
            .expect("Failed to release stock");

        // Verify stock levels
        let request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
        });

        let response = client.get_stock(request)
            .await
            .expect("Failed to get stock");

        let stock = response.into_inner().stock.expect("No stock returned");
        assert_eq!(stock.quantity, 100);
        assert_eq!(stock.reserved_quantity, 0);
        assert_eq!(stock.available_quantity, 100);
    }

    #[tokio::test]
    async fn test_get_stock_movements() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url)
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create initial stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 100,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Get stock item ID
        let get_request = Request::new(GetStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
        });

        let get_response = client.get_stock(get_request)
            .await
            .expect("Failed to get stock");
        let stock_item_id = get_response.into_inner().stock.expect("No stock").id;

        // Update stock to create a movement
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: -10,
            reason: "Sale".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to update stock");

        // Get movements
        let request = Request::new(GetStockMovementsRequest {
            stock_item_id: stock_item_id.clone(),
            pagination: Some(Pagination {
                page: 1,
                limit: 10,
                total: 0,
            }),
        });

        let response = client.get_stock_movements(request)
            .await
            .expect("Failed to get stock movements");

        let movements_response = response.into_inner();
        assert!(!movements_response.movements.is_empty());
    }

    #[tokio::test]
    async fn test_concurrent_reservations() {
        let (_db, server_url) = setup_test_service().await;
        let mut client = InventoryServiceClient::connect(server_url.clone())
            .await
            .expect("Failed to connect to service");

        let product_id = Uuid::new_v4();
        let warehouse_id = Uuid::new_v4();

        // Create initial stock
        let request = Request::new(UpdateStockRequest {
            stock_item_id: String::new(),
            product_id: product_id.to_string(),
            variant_id: String::new(),
            warehouse_id: warehouse_id.to_string(),
            quantity_delta: 20,
            reason: "Initial stock".to_string(),
            reference: String::new(),
        });

        client.update_stock(request)
            .await
            .expect("Failed to create initial stock");

        // Spawn multiple concurrent reservation requests
        let mut handles = vec![];
        for _i in 0..5 {
            let server_url_clone = server_url.clone();
            let product_id_str = product_id.to_string();
            let warehouse_id_str = warehouse_id.to_string();
            let order_id = Uuid::new_v4();

            let handle: tokio::task::JoinHandle<Result<tonic::Response<shinkansen_proto::shinkansen::inventory::ReserveStockResponse>, tonic::Status>> = tokio::spawn(async move {
                let mut client = InventoryServiceClient::connect(server_url_clone)
                    .await
                    .expect("Failed to connect");

                let request = Request::new(ReserveStockRequest {
                    order_id: order_id.to_string(),
                    items: vec![StockReservationItem {
                        product_id: product_id_str,
                        variant_id: String::new(),
                        warehouse_id: warehouse_id_str,
                        quantity: 5,
                    }],
                    expires_at: None,
                });

                client.reserve_stock(request).await
            });
            handles.push(handle);
        }

        // Wait for all reservations to complete
        let results: Vec<_> = futures::future::join_all(handles)
            .await
            .into_iter()
            .filter_map(|r| r.ok())
            .collect();

        // Count successful reservations
        let successful_count = results.iter()
            .filter(|r| r.as_ref().map(|resp| resp.get_ref().success).unwrap_or(false))
            .count();

        // At most 4 should succeed (4 * 5 = 20), at least some should fail
        assert!(successful_count <= 4);
        assert!(successful_count > 0);
    }
}
