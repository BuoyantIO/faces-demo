use phf;
use rand::Rng;
use serde;
use serde_json;
use std::str::FromStr;
use std::thread::sleep;
use std::time::Duration;
use wasmcloud_component::http;
use wasmcloud_component::http::StatusCode;
use wasmcloud_component::wasi;

/// A static map of smiley names to their corresponding HTML-encoded
/// Unicode values.
static SMILEYS: phf::Map<&'static str, &'static str> = phf::phf_map! {
    "Grinning"    => "&#x1F603;",
    "Sleeping"    => "&#x1F634;",
    "Cursing"     => "&#x1F92C;",
    "Kaboom"      => "&#x1F92F;",
    "HeartEyes"   => "&#x1F60D;",
    "Neutral"     => "&#x1F610;",
    "RollingEyes" => "&#x1F644;",
    "Screaming"   => "&#x1F631;",
    "Vomiting"    => "&#x1F92E;",
    "Rusty"       => "&#x1F980;",
};

/// Returns the HTML-encoded Unicode value for a given smiley name, or None if not found.
fn get_smiley(name: &str) -> &'static str {
    SMILEYS.get(name).copied().unwrap_or("&#x1F92E;")
}

struct Component {
    smiley: &'static str,
    error_fraction: u32,
    delay_buckets: Vec<u32>,
}

/// Represents the structure of the HTTP response body returned by the component.
///
/// This struct is serialized to JSON and includes information about the request,
/// any errors that occurred, and additional metadata.
///
/// # Fields
/// - `headers`: A map of HTTP request headers, with header names as lowercase strings.
/// - `errors`: An optional list of error messages, present only if an error occurred.
/// - `client_address`: The address of the client making the request (currently set to "unknown").
/// - `method`: The HTTP method used for the request (e.g., "GET", "POST").
/// - `path`: The path portion of the request URI.
/// - `status`: The HTTP status code returned in the response.
/// - `smiley`: An HTML-encoded smiley, e.g. "&#x1F603;" for a grinning face.
#[derive(serde::Serialize)]
struct ResponseBody {
    headers: std::collections::HashMap<String, String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    errors: Option<Vec<String>>,
    client_address: String,
    method: String,
    path: String,
    status: u16,
    smiley: String,
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

        // Read SMILEY from wasi:config (default to "Rusty")
        let smiley_name = wasi::config::store::get("SMILEY")
            .ok()
            .and_then(|opt| opt)
            .unwrap_or("Rusty".to_string());

        let smiley = get_smiley(&smiley_name);

        println!("Configured with ERROR_FRACTION: {}", error_fraction);
        println!("Configured with DELAY_BUCKETS: {:?}", delay_buckets);
        println!("Configured with SMILEY {}: {}", smiley_name, smiley);

        Component {
            smiley,
            error_fraction,
            delay_buckets,
        }
    }
}

/// Implements behavior for the `Component` struct, including error simulation and artificial delays.
///
/// # Methods
///
/// - `should_error(&self) -> bool`
///   Determines whether the component should simulate an error based on the configured `error_fraction`.
///   If `error_fraction` is nonzero, generates a random number between 1 and 100 (inclusive) and returns
///   `true` if the number is less than or equal to `error_fraction`.
///
/// - `apply_delay(&self)`
///   Applies a random artificial delay based on the configured `delay_buckets`. If `delay_buckets` is empty,
///   no delay is applied. Otherwise, randomly selects one of the buckets and sleeps for the specified duration
///   in milliseconds.
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

        // Default status to OK...
        let mut status = StatusCode::OK;

        // ...and errors to an optional list of strings that is currently None.
        let mut errors: Option<Vec<String>> = None;

        // Check if we should return an error based on ERROR_FRACTION
        if component.should_error() {
            // If we should error, set status to 500
            status = StatusCode::INTERNAL_SERVER_ERROR;

            errors = Some(vec![format!(
                "Smiley error! (error fraction {}%)",
                component.error_fraction
            )]);

            // // Create error response using the builder pattern
            // let response = http::Response::builder()
            //     .status(StatusCode::INTERNAL_SERVER_ERROR)
            //     .header("Content-Type", "text/plain")
            //     .body(error_message.clone())
            //     .unwrap_or_else(|_| {
            //         // Fallback in case of builder error
            //         let mut resp = http::Response::new(error_message.clone());
            //         *resp.status_mut() = StatusCode::INTERNAL_SERVER_ERROR;
            //         resp
            //     });

            // return Ok(response);
        }

        // This response is for all cases.
        let headers: std::collections::HashMap<_, _> = _request
            .headers()
            .iter()
            .map(|(k, v)| {
                (
                    k.to_string().to_lowercase(),
                    v.to_str().unwrap_or("").to_string(),
                )
            })
            .collect();

        let body = ResponseBody {
            headers,
            errors,
            client_address: "unknown".to_string(),
            method: _request.method().to_string(),
            path: _request.uri().path().to_string(),
            status: status.as_u16(),
            smiley: component.smiley.to_string(),
        };

        let body_str = serde_json::to_string(&body).unwrap();
        let mut response = http::Response::new(body_str);
        *response.status_mut() = status;
        response
            .headers_mut()
            .insert("content-type", "application/json".parse().unwrap());
        Ok(response)
    }
}
