use std::{
    collections::HashMap,
    io::Write,
    net::SocketAddr,
    process::Stdio,
    sync::{
        atomic::{AtomicUsize, Ordering},
        Arc,
    },
    time::{Duration, Instant},
};

mod storage;

use anyhow::Context;
use axum::{
    extract::{Path, Query, State},
    http::{HeaderMap, Method, StatusCode},
    response::IntoResponse,
    routing::{any, get},
    Json, Router,
};
use clap::{Parser, Subcommand};
use serde_json::{json, Value};
use tokio::{net::TcpListener, sync::RwLock};
use tower_http::{
    cors::{Any, CorsLayer},
    trace::TraceLayer,
};
use tracing::{error, info, warn};
use webhook_bridge_core::{
    config::{BridgeConfig, ForwardRouteConfig, ScriptGroupConfig, ScriptRouteConfig},
    executor::{spawn_python_executor_on_port, ExecutorClient},
    proto::webhook::{ExecutePluginResponse, PluginInfo},
    VERSION,
};

use crate::storage::{ExecutionInsert, Storage};

#[derive(Parser)]
#[command(name = "webhook-bridge")]
#[command(version = VERSION)]
#[command(about = "Webhook Bridge 4.0 Rust service with Python hook execution")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Run {
        #[arg(short, long, default_value = "config.4.0.yaml")]
        config: String,
        #[arg(long, default_value_t = false)]
        no_python: bool,
    },
    Serve {
        #[arg(short, long, default_value = "config.4.0.yaml")]
        config: String,
        #[arg(long, default_value_t = false)]
        no_python: bool,
    },
    Admin {
        #[arg(short, long, default_value = "config.4.0.yaml")]
        config: String,
    },
    Worker {
        #[command(subcommand)]
        command: WorkerCommands,
    },
    CheckConfig {
        #[arg(short, long, default_value = "config.4.0.yaml")]
        config: String,
    },
}

#[derive(Subcommand)]
enum WorkerCommands {
    Start {
        #[arg(short, long, default_value = "config.4.0.yaml")]
        config: String,
        #[arg(long)]
        port: Option<u16>,
        #[arg(long, default_value_t = 0)]
        index: u16,
    },
}

#[derive(Clone)]
struct AppState {
    config: Arc<BridgeConfig>,
    executors: Arc<RwLock<Vec<ExecutorClient>>>,
    next_worker: Arc<AtomicUsize>,
    storage: Storage,
    started_at: Instant,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();

    match cli.command {
        Commands::Run { config, no_python } => serve(config, no_python).await,
        Commands::Serve { config, no_python } => serve(config, no_python).await,
        Commands::Admin { config } => admin(config).await,
        Commands::Worker { command } => match command {
            WorkerCommands::Start {
                config,
                port,
                index,
            } => worker_start(config, port, index).await,
        },
        Commands::CheckConfig { config } => {
            let cfg = load_config(&config)?;
            println!("{}", serde_json::to_string_pretty(&cfg)?);
            Ok(())
        }
    }
}

async fn serve(config_path: String, no_python: bool) -> anyhow::Result<()> {
    let cfg = load_config(&config_path)?;
    init_tracing(&cfg);

    let worker_count = cfg.executor.workers.max(1);
    let mut python_children = Vec::new();
    if cfg.python.auto_start && !no_python {
        for worker_index in 0..worker_count {
            let port = cfg.executor.port + worker_index;
            match spawn_python_executor_on_port(&cfg, port, Some(&config_path)) {
                Ok(child) => {
                    info!(
                        "started Python executor worker {} on {}:{}",
                        worker_index, cfg.executor.host, port
                    );
                    python_children.push(child);
                }
                Err(err) => warn!("failed to start Python executor worker {worker_index}: {err}"),
            }
        }
        tokio::time::sleep(Duration::from_millis(900)).await;
    }

    let mut executors = Vec::new();
    for worker_index in 0..worker_count {
        match ExecutorClient::connect_worker(&cfg, worker_index).await {
            Ok(client) => executors.push(client),
            Err(err) => warn!("Python executor worker {worker_index} is not connected yet: {err}"),
        }
    }

    let storage = Storage::open(&cfg.storage.sqlite_path)
        .with_context(|| format!("failed to open sqlite database {}", cfg.storage.sqlite_path))?;
    let _ = storage.insert_log(
        "info",
        "webhook-bridge-rust",
        "Rust bridge server started",
        None,
        None,
        json!({"workers": worker_count, "database": cfg.storage.sqlite_path}),
    );

    let state = AppState {
        config: Arc::new(cfg.clone()),
        executors: Arc::new(RwLock::new(executors)),
        next_worker: Arc::new(AtomicUsize::new(0)),
        storage,
        started_at: Instant::now(),
    };

    let app = router(state);
    let addr: SocketAddr = cfg
        .bind_addr()
        .parse()
        .with_context(|| format!("invalid bind address {}", cfg.bind_addr()))?;
    let listener = TcpListener::bind(addr).await?;

    info!("Webhook Bridge 4.0 Rust server listening on http://{addr}");
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    for mut child in python_children {
        let _ = child.kill().await;
    }

    Ok(())
}

async fn admin(config_path: String) -> anyhow::Result<()> {
    let cfg = load_config(&config_path)?;
    println!("Webhook Bridge Admin");
    println!("  API: http://{}", cfg.bind_addr());
    println!("  Dashboard: http://localhost:3002");
    println!("  SQLite: {}", cfg.storage.sqlite_path);
    println!("  Workers: {}", cfg.executor.workers.max(1));
    println!("  Script routes: {}", cfg.scripts.routes.len());
    println!("  Script groups: {}", cfg.scripts.groups.len());
    println!("  Forward routes: {}", cfg.forwarding.routes.len());
    println!("Run: webhook-bridge run --config {config_path}");
    Ok(())
}

async fn worker_start(config_path: String, port: Option<u16>, index: u16) -> anyhow::Result<()> {
    let cfg = load_config(&config_path)?;
    init_tracing(&cfg);
    let port = port.unwrap_or(cfg.executor.port + index);

    info!(
        "starting standalone Python worker {} on {}:{}",
        index, cfg.executor.host, port
    );
    let mut child = spawn_python_executor_on_port(&cfg, port, Some(&config_path))
        .with_context(|| format!("failed to start Python worker on port {port}"))?;
    let status = child.wait().await?;
    if !status.success() {
        anyhow::bail!("Python worker exited with status {status}");
    }
    Ok(())
}

fn router(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health))
        .route("/api/dashboard/status", get(dashboard_status))
        .route("/api/dashboard/stats", get(dashboard_stats))
        .route("/api/dashboard/plugins", get(dashboard_plugins))
        .route(
            "/api/dashboard/plugins/{plugin}",
            get(dashboard_plugin_info),
        )
        .route(
            "/api/dashboard/plugins/{plugin}/execute",
            axum::routing::post(dashboard_execute_plugin),
        )
        .route(
            "/api/dashboard/plugins/{plugin}/enable",
            axum::routing::post(dashboard_enable_plugin),
        )
        .route(
            "/api/dashboard/plugins/{plugin}/disable",
            axum::routing::post(dashboard_disable_plugin),
        )
        .route(
            "/api/dashboard/plugins/{plugin}/logs",
            get(dashboard_plugin_logs),
        )
        .route("/api/dashboard/workers", get(dashboard_workers))
        .route("/api/dashboard/logs", get(dashboard_logs))
        .route("/api/dashboard/config", get(dashboard_config))
        .route("/api/dashboard/python-env", get(dashboard_python_env))
        .route("/api/dashboard/interpreters", get(dashboard_interpreters))
        .route("/api/dashboard/connections", get(dashboard_connections))
        .route("/gateway", any(execute_gateway))
        .route("/webhook", any(execute_gateway))
        .route("/hooks", any(execute_gateway))
        .route("/api/webhook/{plugin}", any(execute_webhook))
        .layer(TraceLayer::new_for_http())
        .layer(
            CorsLayer::new()
                .allow_origin(Any)
                .allow_methods(Any)
                .allow_headers(Any),
        )
        .with_state(state)
}

fn load_config(path: &str) -> anyhow::Result<BridgeConfig> {
    match BridgeConfig::from_path(path) {
        Ok(cfg) => Ok(cfg),
        Err(err) if path == "config.4.0.yaml" => {
            eprintln!("config.4.0.yaml not found or invalid, using defaults: {err}");
            Ok(BridgeConfig::default())
        }
        Err(err) => Err(err).context("failed to load configuration"),
    }
}

fn init_tracing(cfg: &BridgeConfig) {
    let env_filter = tracing_subscriber::EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| cfg.logging.level.clone().into());
    let _ = tracing_subscriber::fmt()
        .with_env_filter(env_filter)
        .try_init();
}

async fn shutdown_signal() {
    if let Err(err) = tokio::signal::ctrl_c().await {
        error!("failed to install shutdown signal handler: {err}");
    }
}

async fn health() -> impl IntoResponse {
    Json(json!({
        "status": "healthy",
        "service": "webhook-bridge-rust",
        "version": VERSION,
    }))
}

async fn dashboard_status(State(state): State<AppState>) -> impl IntoResponse {
    let executor = ensure_executor(&state).await;
    let executor_health = match executor {
        Some(client) => client.health_check().await.ok(),
        None => None,
    };
    let grpc_connected = executor_health
        .as_ref()
        .map(|health| health.status == "healthy")
        .unwrap_or(false);

    let executor_json = executor_health.map(|health| {
        json!({
            "status": health.status,
            "message": health.message,
            "details": health.details,
        })
    });

    api_success(json!({
        "server_status": "running",
        "service": "webhook-bridge-rust",
        "status": if grpc_connected { "healthy" } else { "unknown" },
        "version": VERSION,
        "grpc_connected": grpc_connected,
        "worker_count": state.config.executor.workers.max(1),
        "active_workers": state.executors.read().await.len(),
        "total_jobs": state.storage.stats().map(|s| s["total_requests"].clone()).unwrap_or(json!(0)),
        "completed_jobs": state.storage.stats().map(|s| s["successful_requests"].clone()).unwrap_or(json!(0)),
        "failed_jobs": state.storage.stats().map(|s| s["failed_requests"].clone()).unwrap_or(json!(0)),
        "uptime": format_uptime(state.started_at.elapsed()),
        "executor": executor_json,
    }))
}

async fn dashboard_stats(State(state): State<AppState>) -> impl IntoResponse {
    let plugin_count = list_plugins(&state)
        .await
        .map(|p| p.len())
        .unwrap_or_default();
    let stored_stats = state.storage.stats().unwrap_or_else(|_| {
        json!({
            "total_requests": 0,
            "successful_requests": 0,
            "failed_requests": 0,
            "average_response_time": 0,
            "error_rate": 0,
        })
    });

    api_success(json!({
        "total_requests": stored_stats["total_requests"],
        "successful_requests": stored_stats["successful_requests"],
        "failed_requests": stored_stats["failed_requests"],
        "average_response_time": stored_stats["average_response_time"],
        "active_connections": state.executors.read().await.len(),
        "plugin_count": plugin_count,
        "error_rate": stored_stats["error_rate"],
        "uptime": format_uptime(state.started_at.elapsed()),
        "runtime": "rust",
        "executor": "python-grpc",
    }))
}

async fn dashboard_plugins(State(state): State<AppState>) -> impl IntoResponse {
    let mut items = state
        .config
        .scripts
        .groups
        .iter()
        .cloned()
        .map(script_group_to_json)
        .collect::<Vec<_>>();

    items.extend(
        state
            .config
            .scripts
            .routes
            .iter()
            .cloned()
            .map(script_route_to_json),
    );

    items.extend(
        state
            .config
            .forwarding
            .routes
            .iter()
            .cloned()
            .map(route_to_json),
    );

    match list_plugins(&state).await {
        Ok(plugins) => {
            items.extend(plugins.into_iter().map(plugin_to_json));
            api_success(Value::Array(items))
        }
        Err(_) if !items.is_empty() => api_success(Value::Array(items)),
        Err(err) => api_error(StatusCode::SERVICE_UNAVAILABLE, "executor_unavailable", err),
    }
}

async fn dashboard_plugin_info(
    State(state): State<AppState>,
    Path(plugin): Path<String>,
) -> impl IntoResponse {
    if let Some(route) = find_forward_route(&state, &plugin) {
        return api_success(route_to_json(route));
    }
    if let Some(route) = find_script_route(&state, &plugin) {
        return api_success(script_route_to_json(route));
    }
    if let Some(group) = find_script_group(&state, &plugin) {
        return api_success(script_group_to_json(group));
    }

    let Some(client) = ensure_executor(&state).await else {
        return api_error(
            StatusCode::SERVICE_UNAVAILABLE,
            "executor_unavailable",
            "Python executor is not connected",
        );
    };

    match client.get_plugin_info(plugin).await {
        Ok(response) if response.found => match response.plugin {
            Some(plugin) => api_success(plugin_to_json(plugin)),
            None => api_error(
                StatusCode::NOT_FOUND,
                "plugin_not_found",
                "Plugin not found",
            ),
        },
        Ok(_) => api_error(
            StatusCode::NOT_FOUND,
            "plugin_not_found",
            "Plugin not found",
        ),
        Err(err) => api_error(StatusCode::BAD_GATEWAY, "executor_error", err.to_string()),
    }
}

async fn dashboard_execute_plugin(
    State(state): State<AppState>,
    Path(plugin): Path<String>,
    body: Option<Json<Value>>,
) -> impl IntoResponse {
    let Some(client) = ensure_executor(&state).await else {
        return api_error(
            StatusCode::SERVICE_UNAVAILABLE,
            "executor_unavailable",
            "Python executor is not connected",
        );
    };

    let mut data = HashMap::new();
    if let Some(Json(value)) = body {
        flatten_json("", &value, &mut data);
    }

    let request_id = uuid::Uuid::new_v4().to_string();
    let started = Instant::now();
    let input_json = serde_json::to_string(&data).unwrap_or_else(|_| "{}".to_string());

    match client
        .execute_plugin(
            plugin.clone(),
            "POST".to_string(),
            data,
            HashMap::new(),
            String::new(),
        )
        .await
    {
        Ok(response) => {
            record_execution(
                &state,
                &request_id,
                &plugin,
                "POST",
                &input_json,
                &response,
                started.elapsed(),
            );
            plugin_response_with_request_id(response, request_id)
        }
        Err(err) => api_error(
            StatusCode::BAD_GATEWAY,
            "plugin_execution_failed",
            err.to_string(),
        ),
    }
}

async fn dashboard_enable_plugin(Path(plugin): Path<String>) -> impl IntoResponse {
    api_success(json!({
        "plugin": plugin,
        "enabled": true,
        "message": "Python hooks are file-based; availability is controlled by whether the hook loads successfully.",
    }))
}

async fn dashboard_disable_plugin(Path(plugin): Path<String>) -> impl IntoResponse {
    api_success(json!({
        "plugin": plugin,
        "enabled": false,
        "message": "Runtime disable is not persisted yet. Remove or rename the hook file to disable it in this alpha.",
    }))
}

async fn dashboard_plugin_logs(
    State(state): State<AppState>,
    Path(plugin): Path<String>,
) -> impl IntoResponse {
    match state.storage.recent_logs(100, Some(&plugin)) {
        Ok(logs) => api_success(Value::Array(logs)),
        Err(err) => api_error(StatusCode::INTERNAL_SERVER_ERROR, "storage_error", err),
    }
}

async fn dashboard_workers(State(state): State<AppState>) -> impl IntoResponse {
    let connected = state.executors.read().await.len();
    let configured = state.config.executor.workers.max(1);
    let stats = state.storage.stats().unwrap_or_else(|_| json!({}));
    let workers = (0..configured)
        .map(|index| {
            json!({
                "id": format!("python-executor-{index}"),
                "status": if (index as usize) < connected { "idle" } else { "stopped" },
                "completedJobs": stats["successful_requests"],
                "totalJobs": stats["total_requests"],
                "failedJobs": stats["failed_requests"],
                "uptime": format_uptime(state.started_at.elapsed()),
                "lastActivity": null,
                "performance": {
                    "avgExecutionTime": stats["average_response_time"],
                    "successRate": 100,
                    "errorCount": stats["failed_requests"]
                }
            })
        })
        .collect::<Vec<_>>();

    api_success(Value::Array(workers))
}

async fn dashboard_logs(State(state): State<AppState>) -> impl IntoResponse {
    match state.storage.recent_logs(100, None) {
        Ok(logs) => api_success(Value::Array(logs)),
        Err(err) => api_error(StatusCode::INTERNAL_SERVER_ERROR, "storage_error", err),
    }
}

async fn dashboard_config(State(state): State<AppState>) -> impl IntoResponse {
    api_success(json!([
        {
            "name": "server",
            "description": "Rust HTTP control plane",
            "fields": [
                {"key": "host", "label": "Host", "type": "string", "value": state.config.server.host},
                {"key": "port", "label": "Port", "type": "number", "value": state.config.server.port}
            ]
        },
        {
            "name": "gateway",
            "description": "Unified public webhook gateway",
            "fields": [
                {"key": "enabled", "label": "Enabled", "type": "boolean", "value": state.config.gateway.enabled},
                {"key": "public_path", "label": "Public path", "type": "string", "value": state.config.gateway.public_path},
                {"key": "default_route", "label": "Default route", "type": "string", "value": state.config.gateway.default_route}
            ]
        },
        {
            "name": "python",
            "description": "Rust-managed Python hook executor",
            "fields": [
                {"key": "interpreter", "label": "Interpreter", "type": "string", "value": state.config.python.interpreter},
                {"key": "interpreter_strategy", "label": "Interpreter strategy", "type": "string", "value": state.config.python.interpreter_strategy},
                {"key": "environment_manager", "label": "Environment manager", "type": "string", "value": state.config.python.environment_manager},
                {"key": "uv_project_dir", "label": "uv project", "type": "string", "value": state.config.python.uv_project_dir},
                {"key": "managed_runtime_dir", "label": "Managed runtime", "type": "string", "value": state.config.python.managed_runtime_dir},
                {"key": "auto_start", "label": "Auto start", "type": "boolean", "value": state.config.python.auto_start},
                {"key": "plugin_dirs", "label": "Plugin directories", "type": "textarea", "value": state.config.python.plugin_dirs.join("\n")}
            ]
        },
        {
            "name": "scripts",
            "description": "Local script routes",
            "fields": [
                {"key": "enabled", "label": "Enabled", "type": "boolean", "value": state.config.scripts.enabled},
                {"key": "routes", "label": "Routes", "type": "number", "value": state.config.scripts.routes.len()},
                {"key": "groups", "label": "Groups", "type": "number", "value": state.config.scripts.groups.len()}
            ]
        },
        {
            "name": "forwarding",
            "description": "Webhook request forwarding routes",
            "fields": [
                {"key": "enabled", "label": "Enabled", "type": "boolean", "value": state.config.forwarding.enabled},
                {"key": "routes", "label": "Routes", "type": "number", "value": state.config.forwarding.routes.len()}
            ]
        }
    ]))
}

async fn dashboard_python_env(State(state): State<AppState>) -> impl IntoResponse {
    api_success(json!({
        "interpreter": state.config.python.interpreter,
        "interpreter_strategy": state.config.python.interpreter_strategy,
        "environment_manager": state.config.python.environment_manager,
        "uv_project_dir": state.config.python.uv_project_dir,
        "managed_runtime_dir": state.config.python.managed_runtime_dir,
        "allow_system_fallback": state.config.python.allow_system_fallback,
        "executor_module": state.config.python.executor_module,
        "plugin_dirs": state.config.python.plugin_dirs,
        "auto_start": state.config.python.auto_start,
    }))
}

async fn dashboard_interpreters(State(state): State<AppState>) -> impl IntoResponse {
    api_success(json!([{
        "id": "configured-python",
        "name": "Configured Python",
        "path": state.config.python.interpreter,
        "version": if state.config.python.interpreter_strategy == "managed" { "rust-managed" } else { "managed externally" },
        "status": "available",
        "isDefault": true,
    }]))
}

async fn dashboard_connections(State(state): State<AppState>) -> impl IntoResponse {
    let worker_count = state.executors.read().await.len();
    api_success(json!([
        {
            "name": "Rust HTTP API",
            "type": "api",
            "status": "connected",
            "url": format!("http://{}", state.config.bind_addr()),
        },
        {
            "name": "Python Executor gRPC",
            "type": "service",
            "status": if worker_count > 0 { "connected" } else { "disconnected" },
            "url": state.config.executor_endpoint(),
            "metadata": {"connected_workers": worker_count},
        },
        {
            "name": "Webhook Forwarding",
            "type": "forwarder",
            "status": if state.config.forwarding.enabled { "connected" } else { "disabled" },
            "url": state.config.gateway.public_path,
            "metadata": {
                "routes": state.config.forwarding.routes.len(),
                "script_routes": state.config.scripts.routes.len(),
                "script_groups": state.config.scripts.groups.len(),
                "provider_routes": state.config.gateway.provider_routes,
            },
        }
    ]))
}

async fn execute_gateway(
    State(state): State<AppState>,
    Query(query): Query<HashMap<String, String>>,
    method: Method,
    headers: HeaderMap,
    body: Option<Json<Value>>,
) -> impl IntoResponse {
    if !state.config.gateway.enabled {
        return api_error(
            StatusCode::NOT_FOUND,
            "gateway_disabled",
            "Unified webhook gateway is disabled",
        );
    }

    let provider = detect_provider(&headers, body.as_ref().map(|Json(value)| value));
    let route = query
        .get("route")
        .or_else(|| query.get("plugin"))
        .or_else(|| query.get("hook"))
        .cloned()
        .or_else(|| {
            provider
                .as_ref()
                .and_then(|provider| state.config.gateway.provider_routes.get(provider).cloned())
        })
        .or_else(|| {
            if state.config.gateway.default_route.is_empty() {
                None
            } else {
                Some(state.config.gateway.default_route.clone())
            }
        });

    let Some(route) = route else {
        return api_error(
            StatusCode::BAD_REQUEST,
            "gateway_route_not_found",
            "Could not detect a provider route. Configure gateway.provider_routes or pass ?route=<name>.",
        );
    };

    let mut query = query;
    if let Some(provider) = provider {
        query.entry("provider".to_string()).or_insert(provider);
    }

    process_named_webhook(&state, route, query, method, headers, body).await
}

async fn execute_webhook(
    State(state): State<AppState>,
    Path(plugin): Path<String>,
    Query(query): Query<HashMap<String, String>>,
    method: Method,
    headers: HeaderMap,
    body: Option<Json<Value>>,
) -> impl IntoResponse {
    process_named_webhook(&state, plugin, query, method, headers, body).await
}

async fn process_named_webhook(
    state: &AppState,
    plugin: String,
    query: HashMap<String, String>,
    method: Method,
    headers: HeaderMap,
    body: Option<Json<Value>>,
) -> (StatusCode, Json<Value>) {
    if let Some(group) = find_script_group(state, &plugin) {
        return execute_script_group(state, group, method, headers, query, body).await;
    }

    if let Some(route) = find_script_route(state, &plugin) {
        return execute_script_route(state, route, method, headers, query, body).await;
    }

    if let Some(route) = find_forward_route(state, &plugin) {
        return forward_webhook(state, route, method, headers, query, body).await;
    }

    let Some(client) = ensure_executor(state).await else {
        return api_error(
            StatusCode::SERVICE_UNAVAILABLE,
            "executor_unavailable",
            "Python executor is not connected",
        );
    };

    let mut data = query.clone();
    if let Some(Json(value)) = body {
        flatten_json("", &value, &mut data);
    }

    let header_map = headers
        .iter()
        .filter_map(|(key, value)| {
            value
                .to_str()
                .ok()
                .map(|value| (key.as_str().to_string(), value.to_string()))
        })
        .collect();

    let request_id = uuid::Uuid::new_v4().to_string();
    let started = Instant::now();
    let input_json = serde_json::to_string(&data).unwrap_or_else(|_| "{}".to_string());

    match client
        .execute_plugin(
            plugin.clone(),
            method.to_string(),
            data,
            header_map,
            serde_urlencode_like(&query),
        )
        .await
    {
        Ok(response) => {
            record_execution(
                state,
                &request_id,
                &plugin,
                method.as_str(),
                &input_json,
                &response,
                started.elapsed(),
            );
            plugin_response_with_request_id(response, request_id)
        }
        Err(err) => api_error(
            StatusCode::BAD_GATEWAY,
            "plugin_execution_failed",
            err.to_string(),
        ),
    }
}

async fn execute_script_route(
    state: &AppState,
    route: ScriptRouteConfig,
    method: Method,
    headers: HeaderMap,
    query: HashMap<String, String>,
    body: Option<Json<Value>>,
) -> (StatusCode, Json<Value>) {
    let request_id = uuid::Uuid::new_v4().to_string();
    let started = Instant::now();
    let provider = query.get("provider").cloned();
    let payload = body
        .as_ref()
        .map(|Json(value)| value.clone())
        .unwrap_or_else(|| json!({}));
    let input_json = script_input_json(
        &request_id,
        &route.name,
        None,
        provider,
        method.as_str(),
        &headers,
        &query,
        &payload,
    );

    let result = run_script_route(&route, &input_json);
    match result {
        Ok(output) => {
            let output_json = serde_json::from_str::<Value>(&output)
                .unwrap_or_else(|_| json!({"stdout": output}));
            record_script_execution(
                state,
                &request_id,
                &route.name,
                method.as_str(),
                200,
                true,
                &input_json,
                &output_json,
                "",
                started.elapsed(),
            );
            (
                StatusCode::OK,
                Json(json!({
                    "success": true,
                    "data": output_json,
                    "message": "Executed script route",
                    "error": Value::Null,
                    "execution_time": (started.elapsed().as_secs_f64() * 1000.0).round() as i64,
                    "timestamp": chrono_like_now(),
                    "request_id": request_id,
                })),
            )
        }
        Err(err) => {
            record_script_execution(
                state,
                &request_id,
                &route.name,
                method.as_str(),
                500,
                false,
                &input_json,
                &json!({}),
                &err,
                started.elapsed(),
            );
            api_error(StatusCode::INTERNAL_SERVER_ERROR, "script_failed", err)
        }
    }
}

async fn execute_script_group(
    state: &AppState,
    group: ScriptGroupConfig,
    method: Method,
    headers: HeaderMap,
    query: HashMap<String, String>,
    body: Option<Json<Value>>,
) -> (StatusCode, Json<Value>) {
    let group_request_id = uuid::Uuid::new_v4().to_string();
    let started = Instant::now();
    let provider = query.get("provider").cloned();
    let payload = body
        .as_ref()
        .map(|Json(value)| value.clone())
        .unwrap_or_else(|| json!({}));

    let mut handles = Vec::new();
    for route_name in &group.routes {
        let Some(route) = find_script_route(state, route_name) else {
            continue;
        };
        let request_id = uuid::Uuid::new_v4().to_string();
        let input_json = script_input_json(
            &request_id,
            &route.name,
            Some(&group.name),
            provider.clone(),
            method.as_str(),
            &headers,
            &query,
            &payload,
        );
        let route_for_task = route.clone();
        handles.push(tokio::task::spawn_blocking(move || {
            let result = run_script_route(&route_for_task, &input_json);
            (route_for_task, request_id, input_json, result)
        }));
    }

    let mut results = Vec::new();
    for handle in handles {
        match handle.await {
            Ok((route, request_id, input_json, Ok(output))) => {
                let output_json = serde_json::from_str::<Value>(&output)
                    .unwrap_or_else(|_| json!({"stdout": output}));
                record_script_execution(
                    state,
                    &request_id,
                    &route.name,
                    method.as_str(),
                    200,
                    true,
                    &input_json,
                    &output_json,
                    "",
                    started.elapsed(),
                );
                results.push(json!({
                    "route": route.name,
                    "success": true,
                    "data": output_json,
                }));
            }
            Ok((route, request_id, input_json, Err(err))) => {
                record_script_execution(
                    state,
                    &request_id,
                    &route.name,
                    method.as_str(),
                    500,
                    false,
                    &input_json,
                    &json!({}),
                    &err,
                    started.elapsed(),
                );
                results.push(json!({
                    "route": route.name,
                    "success": false,
                    "error": err,
                }));
            }
            Err(err) => {
                results.push(json!({
                    "route": "unknown",
                    "success": false,
                    "error": err.to_string(),
                }));
            }
        }
    }

    let success = results
        .iter()
        .all(|result| result["success"].as_bool().unwrap_or(false));
    (
        if success {
            StatusCode::OK
        } else {
            StatusCode::INTERNAL_SERVER_ERROR
        },
        Json(json!({
            "success": success,
            "data": {
                "group": group.name,
                "mode": group.mode,
                "results": results,
            },
            "message": "Executed script group",
            "error": if success { Value::Null } else { json!({"code": "script_group_failed", "message": "One or more script routes failed"}) },
            "execution_time": (started.elapsed().as_secs_f64() * 1000.0).round() as i64,
            "timestamp": chrono_like_now(),
            "request_id": group_request_id,
        })),
    )
}

fn script_input_json(
    request_id: &str,
    route: &str,
    group: Option<&str>,
    provider: Option<String>,
    method: &str,
    headers: &HeaderMap,
    query: &HashMap<String, String>,
    payload: &Value,
) -> String {
    let headers_json = headers
        .iter()
        .filter_map(|(key, value)| {
            value
                .to_str()
                .ok()
                .map(|value| (key.as_str().to_string(), json!(value)))
        })
        .collect::<serde_json::Map<String, Value>>();
    let envelope = json!({
        "request_id": request_id,
        "route": route,
        "group": group,
        "provider": provider,
        "method": method,
        "headers": headers_json,
        "query": query,
        "payload": payload,
        "received_at": chrono_like_now(),
    });
    serde_json::to_string(&envelope).unwrap_or_else(|_| "{}".to_string())
}

fn run_script_route(route: &ScriptRouteConfig, input_json: &str) -> Result<String, String> {
    if route.script_path.is_empty() {
        return Err("script_path is empty".to_string());
    }

    let mut command = if route.shell.eq_ignore_ascii_case("powershell") {
        let mut command = std::process::Command::new("powershell");
        command
            .arg("-NoProfile")
            .arg("-ExecutionPolicy")
            .arg("Bypass")
            .arg("-File")
            .arg(&route.script_path);
        command
    } else if route.shell.eq_ignore_ascii_case("pwsh") {
        let mut command = std::process::Command::new("pwsh");
        command
            .arg("-NoProfile")
            .arg("-ExecutionPolicy")
            .arg("Bypass")
            .arg("-File")
            .arg(&route.script_path);
        command
    } else {
        let mut command = std::process::Command::new(&route.shell);
        command.arg(&route.script_path);
        command
    };

    if !route.working_dir.is_empty() {
        command.current_dir(&route.working_dir);
    }
    for (key, value) in &route.env {
        command.env(key, value);
    }

    command
        .env("WEBHOOK_BRIDGE_ROUTE", &route.name)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());

    let mut child = command
        .spawn()
        .map_err(|err| format!("failed to start script {}: {err}", route.script_path))?;

    if let Some(stdin) = child.stdin.as_mut() {
        stdin
            .write_all(input_json.as_bytes())
            .map_err(|err| format!("failed to write webhook payload to script stdin: {err}"))?;
    }

    let output = child
        .wait_with_output()
        .map_err(|err| format!("failed to wait for script {}: {err}", route.script_path))?;
    let stdout = String::from_utf8_lossy(&output.stdout).trim().to_string();
    let stderr = String::from_utf8_lossy(&output.stderr).trim().to_string();

    if output.status.success() {
        Ok(if stdout.is_empty() {
            "{}".to_string()
        } else {
            stdout
        })
    } else {
        Err(if stderr.is_empty() {
            format!("script exited with status {}", output.status)
        } else {
            stderr
        })
    }
}

async fn forward_webhook(
    state: &AppState,
    route: ForwardRouteConfig,
    method: Method,
    headers: HeaderMap,
    query: HashMap<String, String>,
    body: Option<Json<Value>>,
) -> (StatusCode, Json<Value>) {
    let request_id = uuid::Uuid::new_v4().to_string();
    let started = Instant::now();
    let input = body
        .as_ref()
        .map(|Json(value)| value.clone())
        .unwrap_or_else(|| json!({}));
    let input_json = serde_json::to_string(&input).unwrap_or_else(|_| "{}".to_string());

    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(route.timeout_secs.max(1)))
        .build();
    let client = match client {
        Ok(client) => client,
        Err(err) => {
            return api_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                "forwarder_init_failed",
                err,
            )
        }
    };

    let forward_method = route.method.parse::<reqwest::Method>().unwrap_or_else(|_| {
        reqwest::Method::from_bytes(method.as_str().as_bytes()).unwrap_or(reqwest::Method::POST)
    });
    let mut request = client
        .request(forward_method, &route.target_url)
        .query(&query)
        .header("x-webhook-bridge-request-id", &request_id)
        .json(&input);

    for (key, value) in headers.iter() {
        if key.as_str().eq_ignore_ascii_case("host")
            || key.as_str().eq_ignore_ascii_case("content-length")
        {
            continue;
        }
        if let Ok(value) = value.to_str() {
            request = request.header(key.as_str(), value);
        }
    }

    for (key, value) in &route.headers {
        request = request.header(key, value);
    }

    match request.send().await {
        Ok(response) => {
            let status = response.status();
            let response_text = response.text().await.unwrap_or_default();
            let response_json = serde_json::from_str::<Value>(&response_text)
                .unwrap_or_else(|_| json!({"body": response_text}));
            let success = status.is_success();
            record_forward_execution(
                state,
                &request_id,
                &route.name,
                method.as_str(),
                status.as_u16() as i32,
                success,
                &input_json,
                &response_json,
                "",
                started.elapsed(),
            );
            (
                StatusCode::from_u16(status.as_u16()).unwrap_or(StatusCode::OK),
                Json(json!({
                    "success": success,
                    "data": response_json,
                    "message": if success { "Forwarded webhook" } else { "Forward target returned an error" },
                    "error": if success { Value::Null } else { json!({"code": "forward_target_error", "message": status.to_string()}) },
                    "execution_time": (started.elapsed().as_secs_f64() * 1000.0).round() as i64,
                    "timestamp": chrono_like_now(),
                    "request_id": request_id,
                })),
            )
        }
        Err(err) => {
            record_forward_execution(
                state,
                &request_id,
                &route.name,
                method.as_str(),
                502,
                false,
                &input_json,
                &json!({}),
                &err.to_string(),
                started.elapsed(),
            );
            api_error(StatusCode::BAD_GATEWAY, "forward_failed", err)
        }
    }
}

async fn ensure_executor(state: &AppState) -> Option<ExecutorClient> {
    {
        let executors = state.executors.read().await;
        if !executors.is_empty() {
            let idx = state.next_worker.fetch_add(1, Ordering::Relaxed) % executors.len();
            return executors.get(idx).cloned();
        }
    }

    for worker_index in 0..state.config.executor.workers.max(1) {
        match ExecutorClient::connect_worker(&state.config, worker_index).await {
            Ok(client) => {
                state.executors.write().await.push(client.clone());
                return Some(client);
            }
            Err(_) => continue,
        }
    }

    None
}

async fn list_plugins(state: &AppState) -> Result<Vec<PluginInfo>, String> {
    let client = ensure_executor(state)
        .await
        .ok_or_else(|| "Python executor is not connected".to_string())?;
    let response = client
        .list_plugins(String::new())
        .await
        .map_err(|err| err.to_string())?;
    Ok(response.plugins)
}

fn plugin_response_with_request_id(
    response: ExecutePluginResponse,
    request_id: String,
) -> (StatusCode, Json<Value>) {
    let status = StatusCode::from_u16(response.status_code as u16).unwrap_or(StatusCode::OK);
    let body = json!({
        "success": response.error.is_empty(),
        "data": response.data,
        "message": response.message,
        "error": if response.error.is_empty() { Value::Null } else { json!({
            "code": "plugin_error",
            "message": response.error,
        })},
        "execution_time": response.execution_time,
        "timestamp": chrono_like_now(),
        "request_id": request_id,
    });
    (status, Json(body))
}

fn record_execution(
    state: &AppState,
    request_id: &str,
    plugin: &str,
    method: &str,
    input_json: &str,
    response: &ExecutePluginResponse,
    elapsed: Duration,
) {
    let output_json = serde_json::to_string(&response.data).unwrap_or_else(|_| "{}".to_string());
    let success = response.error.is_empty() && response.status_code < 400;
    let execution_time_ms = (elapsed.as_secs_f64() * 1000.0).round() as i64;

    if let Err(err) = state.storage.insert_execution(ExecutionInsert {
        request_id,
        plugin,
        method,
        status_code: response.status_code,
        success,
        execution_time_ms,
        input: input_json,
        output: &output_json,
        error: &response.error,
    }) {
        warn!("failed to record execution: {err}");
    }

    let level = if success { "info" } else { "error" };
    let message = if success {
        format!("Executed plugin {plugin} with {method}")
    } else {
        format!("Plugin {plugin} failed with {method}: {}", response.error)
    };

    if let Err(err) = state.storage.insert_log(
        level,
        "python-executor",
        &message,
        Some(plugin),
        Some(request_id),
        json!({
            "status_code": response.status_code,
            "execution_time_ms": execution_time_ms,
        }),
    ) {
        warn!("failed to record log: {err}");
    }
}

fn record_forward_execution(
    state: &AppState,
    request_id: &str,
    route: &str,
    method: &str,
    status_code: i32,
    success: bool,
    input_json: &str,
    output_json: &Value,
    error: &str,
    elapsed: Duration,
) {
    let output_json = serde_json::to_string(output_json).unwrap_or_else(|_| "{}".to_string());
    let execution_time_ms = (elapsed.as_secs_f64() * 1000.0).round() as i64;

    if let Err(err) = state.storage.insert_execution(ExecutionInsert {
        request_id,
        plugin: route,
        method,
        status_code,
        success,
        execution_time_ms,
        input: input_json,
        output: &output_json,
        error,
    }) {
        warn!("failed to record forwarded execution: {err}");
    }

    let level = if success { "info" } else { "error" };
    let message = if success {
        format!("Forwarded webhook route {route} with {method}")
    } else {
        format!("Webhook forward route {route} failed with {method}: {error}")
    };

    if let Err(err) = state.storage.insert_log(
        level,
        "rust-forwarder",
        &message,
        Some(route),
        Some(request_id),
        json!({
            "status_code": status_code,
            "execution_time_ms": execution_time_ms,
        }),
    ) {
        warn!("failed to record forwarded log: {err}");
    }
}

fn record_script_execution(
    state: &AppState,
    request_id: &str,
    route: &str,
    method: &str,
    status_code: i32,
    success: bool,
    input_json: &str,
    output_json: &Value,
    error: &str,
    elapsed: Duration,
) {
    let output_json = serde_json::to_string(output_json).unwrap_or_else(|_| "{}".to_string());
    let execution_time_ms = (elapsed.as_secs_f64() * 1000.0).round() as i64;

    if let Err(err) = state.storage.insert_execution(ExecutionInsert {
        request_id,
        plugin: route,
        method,
        status_code,
        success,
        execution_time_ms,
        input: input_json,
        output: &output_json,
        error,
    }) {
        warn!("failed to record script execution: {err}");
    }

    let level = if success { "info" } else { "error" };
    let message = if success {
        format!("Executed script route {route} with {method}")
    } else {
        format!("Script route {route} failed with {method}: {error}")
    };

    if let Err(err) = state.storage.insert_log(
        level,
        "script-executor",
        &message,
        Some(route),
        Some(request_id),
        json!({
            "status_code": status_code,
            "execution_time_ms": execution_time_ms,
        }),
    ) {
        warn!("failed to record script log: {err}");
    }
}

fn api_success(data: Value) -> (StatusCode, Json<Value>) {
    (
        StatusCode::OK,
        Json(json!({
            "success": true,
            "data": data,
            "timestamp": chrono_like_now(),
            "request_id": uuid::Uuid::new_v4().to_string(),
        })),
    )
}

fn plugin_to_json(plugin: PluginInfo) -> Value {
    json!({
        "name": plugin.name,
        "id": plugin.name,
        "path": plugin.path,
        "description": plugin.description,
        "supported_methods": plugin.supported_methods,
        "methods": plugin.supported_methods,
        "is_available": plugin.is_available,
        "available": plugin.is_available,
        "status": if plugin.is_available { "active" } else { "error" },
        "enabled": plugin.is_available,
        "last_modified": plugin.last_modified,
        "type": "python",
    })
}

fn route_to_json(route: ForwardRouteConfig) -> Value {
    json!({
        "name": route.name,
        "id": route.name,
        "path": route.target_url,
        "description": "Webhook forwarding route",
        "supported_methods": [route.method],
        "methods": [route.method],
        "is_available": route.enabled,
        "available": route.enabled,
        "status": if route.enabled { "active" } else { "disabled" },
        "enabled": route.enabled,
        "last_modified": null,
        "type": "forward",
        "target_url": route.target_url,
    })
}

fn script_route_to_json(route: ScriptRouteConfig) -> Value {
    json!({
        "name": route.name,
        "id": route.name,
        "path": route.script_path,
        "description": "Local script route",
        "supported_methods": ["POST"],
        "methods": ["POST"],
        "is_available": route.enabled,
        "available": route.enabled,
        "status": if route.enabled { "active" } else { "disabled" },
        "enabled": route.enabled,
        "last_modified": null,
        "type": route.shell,
        "script_path": route.script_path,
    })
}

fn script_group_to_json(group: ScriptGroupConfig) -> Value {
    json!({
        "name": group.name,
        "id": group.name,
        "path": group.routes.join(", "),
        "description": "Parallel script group",
        "supported_methods": ["POST"],
        "methods": ["POST"],
        "is_available": group.enabled,
        "available": group.enabled,
        "status": if group.enabled { "active" } else { "disabled" },
        "enabled": group.enabled,
        "last_modified": null,
        "type": "script-group",
        "routes": group.routes,
        "mode": group.mode,
    })
}

fn find_forward_route(state: &AppState, name: &str) -> Option<ForwardRouteConfig> {
    if !state.config.forwarding.enabled {
        return None;
    }

    state
        .config
        .forwarding
        .routes
        .iter()
        .find(|route| route.enabled && route.name == name && !route.target_url.is_empty())
        .cloned()
}

fn find_script_route(state: &AppState, name: &str) -> Option<ScriptRouteConfig> {
    if !state.config.scripts.enabled {
        return None;
    }

    state
        .config
        .scripts
        .routes
        .iter()
        .find(|route| route.enabled && route.name == name && !route.script_path.is_empty())
        .cloned()
}

fn find_script_group(state: &AppState, name: &str) -> Option<ScriptGroupConfig> {
    if !state.config.scripts.enabled {
        return None;
    }

    state
        .config
        .scripts
        .groups
        .iter()
        .find(|group| group.enabled && group.name == name && !group.routes.is_empty())
        .cloned()
}

fn detect_provider(headers: &HeaderMap, body: Option<&Value>) -> Option<String> {
    if headers.contains_key("x-github-event") {
        return Some("github".to_string());
    }

    if headers.contains_key("x-gitlab-event") || headers.contains_key("x-gitlab-token") {
        return Some("gitlab".to_string());
    }

    if headers.contains_key("x-sentry-hook-resource")
        || headers.contains_key("x-sentry-hook-timestamp")
        || headers.contains_key("sentry-hook-resource")
    {
        return Some("sentry".to_string());
    }

    if let Some(provider) = body
        .and_then(|value| value.get("provider"))
        .and_then(Value::as_str)
    {
        return Some(provider.to_ascii_lowercase());
    }

    if let Some(event) = body
        .and_then(|value| value.get("event"))
        .and_then(Value::as_str)
    {
        let event = event.to_ascii_lowercase();
        if event.contains("github") || event.contains("gitlab") || event.contains("sentry") {
            return event.split('.').next().map(str::to_string);
        }
    }

    None
}

fn api_error(status: StatusCode, code: &str, message: impl ToString) -> (StatusCode, Json<Value>) {
    (
        status,
        Json(json!({
            "success": false,
            "error": {
                "code": code,
                "message": message.to_string(),
            },
            "timestamp": chrono_like_now(),
            "request_id": uuid::Uuid::new_v4().to_string(),
        })),
    )
}

fn flatten_json(prefix: &str, value: &Value, out: &mut HashMap<String, String>) {
    match value {
        Value::Object(map) => {
            for (key, value) in map {
                let next = if prefix.is_empty() {
                    key.to_string()
                } else {
                    format!("{prefix}.{key}")
                };
                flatten_json(&next, value, out);
            }
        }
        Value::Array(_) => {
            out.insert(prefix.to_string(), value.to_string());
        }
        Value::Null => {
            out.insert(prefix.to_string(), String::new());
        }
        Value::Bool(v) => {
            out.insert(prefix.to_string(), v.to_string());
        }
        Value::Number(v) => {
            out.insert(prefix.to_string(), v.to_string());
        }
        Value::String(v) => {
            out.insert(prefix.to_string(), v.to_string());
        }
    }
}

fn serde_urlencode_like(query: &HashMap<String, String>) -> String {
    query
        .iter()
        .map(|(key, value)| format!("{key}={value}"))
        .collect::<Vec<_>>()
        .join("&")
}

fn format_uptime(duration: Duration) -> String {
    let secs = duration.as_secs();
    let hours = secs / 3600;
    let minutes = (secs % 3600) / 60;
    let seconds = secs % 60;
    format!("{hours}h {minutes}m {seconds}s")
}

fn chrono_like_now() -> String {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|duration| format!("{}", duration.as_secs()))
        .unwrap_or_else(|_| "0".to_string())
}
