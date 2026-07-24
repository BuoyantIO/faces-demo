use std::str::FromStr;
use std::thread::sleep;
use std::time::Duration;
use rand::Rng;
use wasmcloud_component::http;
use wasmcloud_component::http::StatusCode;
use wasmcloud_component::wasi;

struct Component {
    error_fraction: u32,
    delay_buckets: Vec<u32>,
}

impl Default for Component {
    fn default() -> Self {

        // Read ERROR_FRACTION from wasi:config (default to 0)
        let error_fraction = wasi::config::store::get("ERROR_FRACTION")
            .ok()
            .and_then(|opt| opt)
            .and_then(|val| u32::from_str(&val).ok())
            .unwrap_or(0)
            .min(100);

        // Read DELAY_BUCKETS from wasi:config (default to empty vec)
        let delay_buckets = wasi::config::store::get("DELAY_BUCKETS")
            .ok()
            .and_then(|opt| opt)
            .map(|val| {
                val.split(',')
                    .filter_map(|s| u32::from_str(s.trim()).ok())
                    .collect()
            })
            .unwrap_or_else(Vec::new);

        println!("Configured with ERROR_FRACTION: {}", error_fraction);
        println!("Configured with DELAY_BUCKETS: {:?}", delay_buckets);

        Component {
            error_fraction,
            delay_buckets,
        }
    }
}

impl Component {
    fn should_error(&self) -> bool {
        if self.error_fraction == 0 {
            return false;
        }
        
        let mut rng = rand::thread_rng();
        let roll = rng.gen_range(1..=100);
        roll <= self.error_fraction
    }

    fn apply_delay(&self) {
        if self.delay_buckets.is_empty() {
            return;
        }

        let mut rng = rand::thread_rng();
        let bucket_idx = rng.gen_range(0..self.delay_buckets.len());
        let delay_ms = self.delay_buckets[bucket_idx];
        
        if delay_ms > 0 {
            sleep(Duration::from_millis(delay_ms as u64));
        }
    }
}

// Export the component
http::export!(Component);

impl http::Server for Component {
    fn handle(
        _request: http::IncomingRequest,
    ) -> http::Result<http::Response<impl http::OutgoingBody>> {
        // Get a reference to the component instance
        let component = Component::default();
        
        // Apply configured delay
        component.apply_delay();

        // Check if we should return an error based on ERROR_FRACTION
        if component.should_error() {
            let error_message = format!("Smiley error! (error fraction {}%)", component.error_fraction);
            
            // Create error response using the builder pattern
            let response = http::Response::builder()
                .status(StatusCode::INTERNAL_SERVER_ERROR)
                .header("Content-Type", "text/plain")
                .body(error_message.clone())
                .unwrap_or_else(|_| {
                    // Fallback in case of builder error
                    let mut resp = http::Response::new(error_message.clone());
                    *resp.status_mut() = StatusCode::INTERNAL_SERVER_ERROR;
                    resp
                });

            return Ok(response);
        }

        // Normal response
        Ok(http::Response::new("Hello from Rust!\n".to_string()))
    }
}
