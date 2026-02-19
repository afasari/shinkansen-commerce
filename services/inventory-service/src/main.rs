use std::sync::Arc;
use anyhow::Result;
use tokio::net::TcpListener;
use tonic::transport::Server;
use tracing::{info, error};
use tracing_subscriber::{EnvFilter, layer::SubscriberExt, util::SubscriberInitExt};
use tower_http::trace::TraceLayer;
use axum::Router;

mod config;
mod database;
mod repository;
mod service;
mod health;

use shinkansen_proto::shinkansen::inventory::inventory_service_server::InventoryServiceServer;
use service::InventoryService;
use config::Config;
use database::Database;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let env_filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new("info"));
    
    tracing_subscriber::registry()
        .with(env_filter)
        .with(tracing_subscriber::fmt::layer())
        .init();
    
    info!("Starting inventory service");
    
    let cfg = Config::load()?;
    info!("Config loaded: {:?}", cfg);
    
    let db = Database::new(&cfg.database_url).await?;
    info!("Database connected");
    
    info!("Running migrations...");
    db.run_migrations().await?;
    info!("Migrations completed");
    
    let inventory_service = InventoryService::new(db);
    let inventory_grpc = shinkansen_proto::shinkansen::inventory::inventory_service_server::InventoryServiceServer::new(inventory_service);
    
    info!("Starting gRPC server on {}", cfg.grpc_server_address);
    let grpc_addr = cfg.grpc_bind_addr()?;
    let grpc_listener = TcpListener::bind(&grpc_addr).await?;
    
    let grpc_server = Server::builder()
        .add_service(inventory_grpc)
        .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(grpc_listener));
    
    info!("Starting HTTP server on {}", cfg.metrics_server_address);
    let http_app = health::router();
    let http_addr = cfg.metrics_bind_addr()?;
    let http_listener = TcpListener::bind(&http_addr).await?;
    let http_server = axum::serve(http_listener, http_app);
    
    tokio::select! {
        result = grpc_server => {
            if let Err(e) = result {
                error!("gRPC server error: {}", e);
            }
        }
        result = http_server => {
            if let Err(e) = result {
                error!("HTTP server error: {}", e);
            }
        }
        _ = tokio::signal::ctrl_c() => {
            info!("Received shutdown signal");
        }
    }
    
    info!("Server shutting down");
    Ok(())
}
