use anyhow::Result;
use opentelemetry::trace::TracerProvider;
use opentelemetry_otlp::HasExportConfig;
use opentelemetry_sdk::Resource;
use tokio::net::TcpListener;
use tonic::transport::Server;
use tracing::{info, error};
use tracing_subscriber::{EnvFilter, layer::SubscriberExt, util::SubscriberInitExt};

mod config;
mod database;
mod repository;
mod service;
mod health;
mod otel_middleware;

use service::InventoryServiceImpl;
use config::Config;
use database::Database;
use otel_middleware::OtelGrpcService;

fn init_telemetry() -> Result<()> {
    let endpoint = std::env::var("OTEL_EXPORTER_OTLP_ENDPOINT")
        .unwrap_or_else(|_| "http://localhost:4317".to_string());

    let service_name = std::env::var("OTEL_SERVICE_NAME")
        .unwrap_or_else(|_| "inventory-service".to_string());

    let resource = Resource::from_schema_url(
        vec![
            opentelemetry::KeyValue::new("service.name", service_name),
        ],
        "https://opentelemetry.io/schemas/1.26.0",
    );

    let mut exporter_builder = opentelemetry_otlp::SpanExporter::builder()
        .with_tonic();
    exporter_builder.export_config().endpoint = Some(endpoint);

    let exporter = exporter_builder.build()?;

    let provider = opentelemetry_sdk::trace::TracerProvider::builder()
        .with_resource(resource)
        .with_batch_exporter(exporter, opentelemetry_sdk::runtime::Tokio)
        .build();

    let tracer = provider.tracer("shinkansen-inventory");
    opentelemetry::global::set_tracer_provider(provider);
    opentelemetry::global::set_text_map_propagator(
        opentelemetry_sdk::propagation::TraceContextPropagator::new(),
    );
    tracing_log::LogTracer::init()?;
    let otel_layer = tracing_opentelemetry::layer().with_tracer(tracer);

    let env_filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new("info"));

    tracing_subscriber::registry()
        .with(env_filter)
        .with(otel_layer)
        .with(tracing_subscriber::fmt::layer().json())
        .init();

    Ok(())
}

fn init_fallback_logging() {
    let env_filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new("info"));
    tracing_subscriber::registry()
        .with(env_filter)
        .with(tracing_subscriber::fmt::layer().json())
        .init();
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    if let Err(e) = init_telemetry() {
        eprintln!("Warning: failed to init telemetry: {e}, continuing without");
        init_fallback_logging();
    }

    info!("Starting inventory service");

    let cfg = Config::load()?;
    info!("Config loaded: {:?}", cfg);

    let db = Database::new(&cfg.database_url).await?;
    info!("Database connected");

    info!("Running migrations...");
    db.run_migrations().await?;
    info!("Migrations completed");

    let inventory_service = InventoryServiceImpl::new(db);
    let inventory_grpc = shinkansen_proto::shinkansen::inventory::inventory_service_server::InventoryServiceServer::new(inventory_service);

    info!("Starting gRPC server on {}", cfg.grpc_server_address);
    let grpc_addr = cfg.grpc_bind_addr()?;
    let grpc_listener = TcpListener::bind(&grpc_addr).await?;

    let grpc_server = Server::builder()
        .layer(tower::layer::layer_fn(OtelGrpcService::new))
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
