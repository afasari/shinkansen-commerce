use sqlx::postgres::PgPool;
use std::sync::Arc;

pub struct Database {
    pool: Arc<PgPool>,
}

impl Database {
    pub async fn new(database_url: &str) -> Result<Self, sqlx::Error> {
        let pool = PgPool::connect(database_url).await?;
        Ok(Self {
            pool: Arc::new(pool),
        })
    }
    
    pub fn pool(&self) -> &PgPool {
        &self.pool
    }
    
    pub async fn run_migrations(&self) -> Result<(), sqlx::migrate::MigrateError> {
        sqlx::migrate!("./migrations").run(&*self.pool).await
    }
}

#[cfg(test)]
impl Database {
    pub async fn for_test() -> Result<Self, sqlx::Error> {
        let pool = PgPool::connect("postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen_test?sslmode=disable").await?;
        Ok(Self {
            pool: Arc::new(pool),
        })
    }
}
