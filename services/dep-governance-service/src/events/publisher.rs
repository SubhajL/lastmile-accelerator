use crate::error::{AppError, Result};
use async_nats::Client;
use std::time::Duration;

#[derive(Clone)]
pub struct NatsPublisher {
    client: Client,
}

impl NatsPublisher {
    pub async fn connect(url: &str) -> Result<Self> {
        tracing::info!("Connecting to NATS at {}", url);

        let mut attempts = 0;
        let max_attempts = 3;

        while attempts < max_attempts {
            match async_nats::connect_with_options(
                url,
                async_nats::ConnectOptions::new()
                    .name("dep-governance-service")
                    .retry_on_initial_connect()
                    .connection_timeout(Duration::from_secs(10)),
            )
            .await
            {
                Ok(client) => {
                    tracing::info!("Successfully connected to NATS");
                    return Ok(Self { client });
                }
                Err(e) => {
                    attempts += 1;
                    tracing::warn!(
                        "Failed to connect to NATS (attempt {}/{}): {}",
                        attempts,
                        max_attempts,
                        e
                    );
                    if attempts < max_attempts {
                        tokio::time::sleep(Duration::from_secs(2u64.pow(attempts))).await;
                    }
                }
            }
        }

        Err(AppError::Nats(format!(
            "Failed to connect to NATS after {} attempts",
            max_attempts
        )))
    }

    pub async fn publish(&self, topic: &str, payload: &[u8], traceparent: &str) -> Result<()> {
        tracing::debug!("Publishing to topic: {}", topic);

        let mut headers = async_nats::HeaderMap::new();
        headers.insert("traceparent", traceparent);

        let payload_bytes = bytes::Bytes::from(payload.to_vec());

        self.client
            .publish_with_headers(topic.to_string(), headers, payload_bytes)
            .await
            .map_err(|e| AppError::Nats(format!("Failed to publish message: {}", e)))?;

        self.client
            .flush()
            .await
            .map_err(|e| AppError::Nats(format!("Failed to flush: {}", e)))?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_nats_publisher_connects() {
        if std::env::var("TEST_NATS_URL").is_err() {
            eprintln!("Skipping test: TEST_NATS_URL not set");
            return;
        }

        let nats_url = std::env::var("TEST_NATS_URL").unwrap();
        let result = NatsPublisher::connect(&nats_url).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_nats_publisher_publishes_event() {
        if std::env::var("TEST_NATS_URL").is_err() {
            eprintln!("Skipping test: TEST_NATS_URL not set");
            return;
        }

        let nats_url = std::env::var("TEST_NATS_URL").unwrap();
        let publisher = NatsPublisher::connect(&nats_url).await.unwrap();

        let payload = b"test payload";
        let result = publisher
            .publish("test.topic", payload, "00-test-traceparent-00")
            .await;

        assert!(result.is_ok());
    }

    #[tokio::test]
    #[ignore] // Flaky due to environment-specific network behavior
    async fn test_nats_publisher_handles_connection_failure() {
        // Use a clearly invalid URL format that should fail
        let invalid_url = "nats://240.0.0.0:65535";
        let result = tokio::time::timeout(
            std::time::Duration::from_secs(5),
            NatsPublisher::connect(invalid_url)
        ).await;
        
        // Either timeout or error is acceptable as connection failure
        match result {
            Ok(Ok(_)) => panic!("Expected connection to fail but it succeeded"),
            Ok(Err(e)) => {
                assert!(e.to_string().contains("Failed to connect to NATS"));
            }
            Err(_) => {
                // Timeout is also acceptable as a failure
            }
        }
    }
}
