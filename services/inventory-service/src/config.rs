use serde::Deserialize;
use std::env;
use std::net::SocketAddr;

fn default_grpc_address() -> String {
    "0.0.0.0:9105".to_string()
}

fn default_metrics_address() -> String {
    "0.0.0.0:8105".to_string()
}

fn default_database_url() -> String {
    "postgres://shinkansen:shinkansen_dev_password@postgres:5432/shinkansen?sslmode=disable"
        .to_string()
}

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    #[serde(default = "default_grpc_address")]
    pub grpc_server_address: String,

    #[serde(default = "default_metrics_address")]
    pub metrics_server_address: String,

    #[serde(default = "default_database_url")]
    pub database_url: String,
}

impl Config {
    pub fn load() -> Result<Self, Box<dyn std::error::Error>> {
        Ok(envy::from_env()?)
    }

    pub fn grpc_bind_addr(&self) -> Result<SocketAddr, Box<dyn std::error::Error>> {
        let addr = if self.grpc_server_address.starts_with(':') {
            format!("0.0.0.0{}", self.grpc_server_address)
        } else {
            self.grpc_server_address.clone()
        };
        addr.parse()
            .map_err(|e| Box::new(e) as Box<dyn std::error::Error>)
    }

    pub fn metrics_bind_addr(&self) -> Result<SocketAddr, Box<dyn std::error::Error>> {
        let addr = if self.metrics_server_address.starts_with(':') {
            format!("0.0.0.0{}", self.metrics_server_address)
        } else {
            self.metrics_server_address.clone()
        };
        addr.parse()
            .map_err(|e| Box::new(e) as Box<dyn std::error::Error>)
    }
}
