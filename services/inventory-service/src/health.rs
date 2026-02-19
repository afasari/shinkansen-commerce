use axum::{
    extract::State,
    response::{IntoResponse, Json},
    routing::get,
};
use serde::Serialize;
use tracing::info;

#[derive(Debug, Clone, Serialize)]
pub struct HealthResponse {
    pub status: String,
}

pub async fn health_check() -> impl IntoResponse {
    info!("Health check requested");
    Json(HealthResponse {
        status: "healthy".to_string(),
    })
}

pub fn router() -> axum::Router {
    axum::Router::new()
        .route("/health", get(health_check))
}
