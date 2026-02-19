use std::sync::Arc;
use std::time::Duration;
use anyhow::Result;
use chrono::{DateTime, Utc};
use tonic::{Request, Response, Status};
use tracing::{info, error, instrument};
use prost_types::Timestamp;
use uuid::Uuid;

use shinkansen_proto::shinkansen::inventory::{
    inventory_service_server::InventoryService as InventoryServiceTrait,
    GetStockRequest, GetStockResponse, StockItem as ProtoStockItem,
    UpdateStockRequest,
    ReserveStockRequest, ReserveStockResponse,
    ReleaseStockRequest,
    GetStockMovementsRequest, GetStockMovementsResponse, StockMovement as ProtoStockMovement,
};
use shinkansen_proto::shinkansen::common::{Empty, Pagination};

use crate::repository::{Repository, StockItem, StockMovement};
use crate::database::Database;

pub struct InventoryService {
    repository: Arc<Repository>,
}

impl InventoryService {
    pub fn new(db: Database) -> Self {
        Self {
            repository: Arc::new(Repository::new(db)),
        }
    }
}

#[tonic::async_trait]
impl InventoryServiceTrait for InventoryService {
    #[instrument(skip(self))]
    async fn get_stock(
        &self,
        request: Request<GetStockRequest>,
    ) -> Result<Response<GetStockResponse>, Status> {
        let req = request.into_inner();
        info!("Getting stock for product_id: {:?}", req.product_id);

        let product_id = parse_uuid(&req.product_id)?;
        let variant_id = if req.variant_id.is_empty() {
            None
        } else {
            Some(parse_uuid(&req.variant_id)?)
        };
        let warehouse_id = parse_uuid(&req.warehouse_id)?;

        match self.repository.get_stock(product_id, variant_id, warehouse_id).await {
            Ok(Some(stock)) => {
                let proto_item = stock_to_proto(stock);
                Ok(Response::new(GetStockResponse { stock: Some(proto_item) }))
            }
            Ok(None) => {
                let proto_item = ProtoStockItem {
                    id: String::new(),
                    product_id: req.product_id.clone(),
                    variant_id: req.variant_id.clone(),
                    warehouse_id: req.warehouse_id.clone(),
                    quantity: 0,
                    reserved_quantity: 0,
                    available_quantity: 0,
                    updated_at: Some(timestamp_to_proto(Utc::now())),
                };
                Ok(Response::new(GetStockResponse { stock: Some(proto_item) }))
            }
            Err(e) => {
                error!("Failed to get stock: {:?}", e);
                Err(Status::internal(format!("Failed to get stock: {}", e)))
            }
        }
    }
    
    #[instrument(skip(self))]
    async fn update_stock(
        &self,
        request: Request<UpdateStockRequest>,
    ) -> Result<Response<Empty>, Status> {
        let req = request.into_inner();
        info!("Updating stock for product_id: {}, delta: {}", req.product_id, req.quantity_delta);

        let product_id = parse_uuid(&req.product_id)?;
        let variant_id = if req.variant_id.is_empty() {
            None
        } else {
            Some(parse_uuid(&req.variant_id)?)
        };
        let warehouse_id = parse_uuid(&req.warehouse_id)?;

        let stock = match self.repository.get_stock(product_id, variant_id, warehouse_id).await {
            Ok(Some(s)) => s,
            Ok(None) => {
                match self.repository.create_or_update_stock(
                    product_id, variant_id, warehouse_id, req.quantity_delta
                ).await {
                    Ok(id) => {
                        info!("Created new stock item: {}", id);
                        let reason = if req.reason.is_empty() { None } else { Some(req.reason.as_str()) };
                        if req.quantity_delta > 0 {
                            if let Err(e) = self.repository.create_stock_movement(
                                id, "MOVEMENT_TYPE_INBOUND", req.quantity_delta, reason
                            ).await {
                                error!("Failed to create movement: {:?}", e);
                            }
                        } else if let Err(e) = self.repository.create_stock_movement(
                            id, "MOVEMENT_TYPE_OUTBOUND", req.quantity_delta, reason
                        ).await {
                            error!("Failed to create movement: {:?}", e);
                        }
                        return Ok(Response::new(Empty {}));
                    }
                    Err(e) => {
                        error!("Failed to create stock: {:?}", e);
                        return Err(Status::internal(format!("Failed to create stock: {}", e)));
                    }
                }
            }
            Err(e) => {
                error!("Failed to get stock: {:?}", e);
                return Err(Status::internal(format!("Failed to get stock: {}", e)));
            }
        };

        if let Err(e) = self.repository.update_stock_quantity(
            product_id, variant_id, warehouse_id, req.quantity_delta
        ).await {
            error!("Failed to update stock: {:?}", e);
            return Err(Status::internal(format!("Failed to update stock: {}", e)));
        }

        let movement_type = if req.quantity_delta > 0 {
            "MOVEMENT_TYPE_INBOUND"
        } else {
            "MOVEMENT_TYPE_OUTBOUND"
        };
        let reason = if req.reason.is_empty() { None } else { Some(req.reason.as_str()) };
        if let Err(e) = self.repository.create_stock_movement(
            stock.id, movement_type, req.quantity_delta, reason
        ).await {
            error!("Failed to create movement: {:?}", e);
        }

        Ok(Response::new(Empty {}))
    }

    #[instrument(skip(self))]
    async fn reserve_stock(
        &self,
        request: Request<ReserveStockRequest>,
    ) -> Result<Response<ReserveStockResponse>, Status> {
        let req = request.into_inner();
        info!("Reserving stock for order_id: {}", req.order_id);
        
        let order_id = parse_uuid(&req.order_id)?;
        let expires_at = req.expires_at.as_ref()
            .map(|t| timestamp_from_proto(t.clone()))
            .unwrap_or_else(|| Utc::now() + Duration::from_secs(1800));
        
        let mut failed_items = Vec::new();
        let mut success = true;
        
        for item in &req.items {
            let product_id = parse_uuid(&item.product_id)?;
            let variant_id = if item.variant_id.is_empty() {
                None
            } else {
                Some(parse_uuid(&item.variant_id)?)
            };
            let warehouse_id = parse_uuid(&item.warehouse_id)?;
            
            match self.repository.get_stock(product_id, variant_id, warehouse_id).await {
                Ok(Some(stock)) => {
                    if stock.available_quantity < item.quantity {
                        failed_items.push(item.product_id.clone());
                        success = false;
                        continue;
                    }
                    
                    match self.repository.reserve_stock(
                        order_id, stock.id, item.quantity, expires_at
                    ).await {
                        Ok(_) => {
                            if let Err(e) = self.repository.create_stock_movement(
                                stock.id, "MOVEMENT_TYPE_RESERVATION", item.quantity,
                                Some(&format!("Order: {}", req.order_id))
                            ).await {
                                error!("Failed to create movement: {:?}", e);
                            }
                        }
                        Err(e) => {
                            error!("Failed to reserve stock: {:?}", e);
                            failed_items.push(item.product_id.clone());
                            success = false;
                        }
                    }
                }
                Ok(None) => {
                    failed_items.push(item.product_id.clone());
                    success = false;
                }
                Err(e) => {
                    error!("Failed to get stock: {:?}", e);
                    failed_items.push(item.product_id.clone());
                    success = false;
                }
            }
        }
        
        Ok(Response::new(ReserveStockResponse {
            reservation_id: Uuid::new_v4().to_string(),
            success,
            failed_items,
        }))
    }
    
    #[instrument(skip(self))]
    async fn release_stock(
        &self,
        request: Request<ReleaseStockRequest>,
    ) -> Result<Response<Empty>, Status> {
        let req = request.into_inner();
        info!("Releasing stock for reservation_id: {}", req.reservation_id);
        
        let order_id = parse_uuid(&req.reservation_id)?;
        
        if let Err(e) = self.repository.release_stock(order_id).await {
            error!("Failed to release stock: {:?}", e);
            return Err(Status::internal(format!("Failed to release stock: {}", e)));
        }
        
        Ok(Response::new(Empty {}))
    }
    
    #[instrument(skip(self))]
    async fn get_stock_movements(
        &self,
        request: Request<GetStockMovementsRequest>,
    ) -> Result<Response<GetStockMovementsResponse>, Status> {
        let req = request.into_inner();
        info!("Getting stock movements for stock_item_id: {}", req.stock_item_id);
        
        let stock_item_id = parse_uuid(&req.stock_item_id)?;

        let limit = req.pagination.as_ref().map(|p| p.limit).unwrap_or(50) as i64;
        let offset = req.pagination.as_ref()
            .map(|p| (p.page - 1) * limit as i32)
            .unwrap_or(0) as i64;

        match self.repository.list_stock_movements(stock_item_id, limit, offset).await {
            Ok(movements) => {
                let proto_movements: Vec<ProtoStockMovement> = movements
                    .into_iter()
                    .map(movement_to_proto)
                    .collect();
                let pagination = Some(Pagination {
                    page: req.pagination.as_ref().map(|p| p.page).unwrap_or(1),
                    limit: req.pagination.as_ref().map(|p| p.limit).unwrap_or(50),
                    total: proto_movements.len() as i32,
                });

                Ok(Response::new(GetStockMovementsResponse {
                    movements: proto_movements,
                    pagination,
                }))
            }
            Err(e) => {
                error!("Failed to get stock movements: {:?}", e);
                Err(Status::internal(format!("Failed to get stock movements: {}", e)))
            }
        }
    }
}

fn parse_uuid(s: &str) -> Result<uuid::Uuid, Status> {
    uuid::Uuid::parse_str(s)
        .map_err(|_| Status::invalid_argument(format!("Invalid UUID: {}", s)))
}

fn timestamp_to_proto(dt: DateTime<Utc>) -> Timestamp {
    Timestamp {
        seconds: dt.timestamp(),
        nanos: dt.timestamp_subsec_nanos() as i32,
    }
}

fn timestamp_from_proto(ts: Timestamp) -> DateTime<Utc> {
    DateTime::<Utc>::from_timestamp(ts.seconds, ts.nanos as u32)
        .unwrap_or_else(|| Utc::now())
}

fn stock_to_proto(stock: StockItem) -> ProtoStockItem {
    ProtoStockItem {
        id: stock.id.to_string(),
        product_id: stock.product_id.to_string(),
        variant_id: stock.variant_id.map(|v| v.to_string()).unwrap_or_default(),
        warehouse_id: stock.warehouse_id.to_string(),
        quantity: stock.quantity,
        reserved_quantity: stock.reserved_quantity,
        available_quantity: stock.available_quantity,
        updated_at: Some(timestamp_to_proto(stock.updated_at)),
    }
}

fn movement_to_proto(movement: StockMovement) -> ProtoStockMovement {
    ProtoStockMovement {
        id: movement.id.to_string(),
        stock_item_id: movement.stock_item_id.to_string(),
        r#type: 0,
        quantity: movement.quantity,
        reference: movement.reference.unwrap_or_default(),
        created_at: Some(timestamp_to_proto(movement.created_at)),
    }
}
