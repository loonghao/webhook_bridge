use std::{collections::HashMap, path::PathBuf, process::Stdio, time::Duration};

use tokio::process::{Child, Command};
use tonic::{transport::Channel, Request};

use crate::{
    config::BridgeConfig,
    proto::webhook::{
        webhook_executor_client::WebhookExecutorClient, ExecutePluginRequest,
        ExecutePluginResponse, GetPluginInfoRequest, GetPluginInfoResponse, HealthCheckRequest,
        HealthCheckResponse, ListPluginsRequest, ListPluginsResponse,
    },
};

#[derive(Debug, thiserror::Error)]
pub enum ExecutorConnectError {
    #[error("invalid executor endpoint: {0}")]
    InvalidEndpoint(#[from] tonic::codegen::http::uri::InvalidUri),
    #[error("failed to connect to executor: {0}")]
    Transport(#[from] tonic::transport::Error),
}

#[derive(Clone)]
pub struct ExecutorClient {
    inner: WebhookExecutorClient<Channel>,
}

impl ExecutorClient {
    pub async fn connect(cfg: &BridgeConfig) -> Result<Self, ExecutorConnectError> {
        Self::connect_endpoint(cfg.executor_endpoint(), cfg.executor.connect_timeout_secs).await
    }

    pub async fn connect_worker(
        cfg: &BridgeConfig,
        worker_index: u16,
    ) -> Result<Self, ExecutorConnectError> {
        Self::connect_endpoint(
            cfg.executor_endpoint_for(worker_index),
            cfg.executor.connect_timeout_secs,
        )
        .await
    }

    async fn connect_endpoint(
        endpoint: String,
        connect_timeout_secs: u64,
    ) -> Result<Self, ExecutorConnectError> {
        let channel = Channel::from_shared(endpoint)?
            .connect_timeout(Duration::from_secs(connect_timeout_secs))
            .connect()
            .await?;

        Ok(Self {
            inner: WebhookExecutorClient::new(channel),
        })
    }

    pub async fn execute_plugin(
        &self,
        plugin_name: String,
        http_method: String,
        data: HashMap<String, String>,
        headers: HashMap<String, String>,
        query_string: String,
    ) -> Result<ExecutePluginResponse, tonic::Status> {
        let mut client = self.inner.clone();
        let request = ExecutePluginRequest {
            plugin_name,
            http_method,
            data,
            headers,
            query_string,
        };

        Ok(client
            .execute_plugin(Request::new(request))
            .await?
            .into_inner())
    }

    pub async fn list_plugins(&self, filter: String) -> Result<ListPluginsResponse, tonic::Status> {
        let mut client = self.inner.clone();
        Ok(client
            .list_plugins(Request::new(ListPluginsRequest { filter }))
            .await?
            .into_inner())
    }

    pub async fn get_plugin_info(
        &self,
        plugin_name: String,
    ) -> Result<GetPluginInfoResponse, tonic::Status> {
        let mut client = self.inner.clone();
        Ok(client
            .get_plugin_info(Request::new(GetPluginInfoRequest { plugin_name }))
            .await?
            .into_inner())
    }

    pub async fn health_check(&self) -> Result<HealthCheckResponse, tonic::Status> {
        let mut client = self.inner.clone();
        Ok(client
            .health_check(Request::new(HealthCheckRequest {
                service: "python-executor".to_string(),
            }))
            .await?
            .into_inner())
    }
}

pub fn spawn_python_executor(
    cfg: &BridgeConfig,
    config_path: Option<&str>,
) -> std::io::Result<Child> {
    spawn_python_executor_on_port(cfg, cfg.executor.port, config_path)
}

pub fn spawn_python_executor_on_port(
    cfg: &BridgeConfig,
    port: u16,
    config_path: Option<&str>,
) -> std::io::Result<Child> {
    let runtime_root = if cfg.python.embedded_runtime {
        Some(crate::runtime::materialize_embedded_runtime(cfg)?)
    } else {
        None
    };

    let interpreter = crate::runtime::resolve_python_interpreter(cfg);
    let mut command = python_command(cfg, interpreter);
    command
        .arg("-m")
        .arg(&cfg.python.executor_module)
        .arg("--host")
        .arg(&cfg.executor.host)
        .arg("--port")
        .arg(port.to_string())
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit());

    if let Some(root) = &runtime_root {
        let proto_path = root.join("api").join("proto");
        let python_path = format!("{};{}", root.display(), proto_path.display());
        command.env("PYTHONPATH", python_path);
    }

    if let Some(path) = config_path {
        let config_path = std::fs::canonicalize(path).unwrap_or_else(|_| path.into());
        command.arg("--config").arg(config_path);
    }

    if config_path.is_none() && !cfg.python.plugin_dirs.is_empty() {
        command.arg("--plugin-dirs");
        command.args(&cfg.python.plugin_dirs);
    }

    command.spawn()
}

fn python_command(cfg: &BridgeConfig, interpreter: PathBuf) -> Command {
    if cfg.python.environment_manager.eq_ignore_ascii_case("uv") && command_exists("uv") {
        let mut command = Command::new("uv");
        command.arg("run");
        if !cfg.python.uv_project_dir.is_empty() {
            command.arg("--project").arg(&cfg.python.uv_project_dir);
        }
        command.arg(interpreter);
        return command;
    }

    if cfg.python.environment_manager.eq_ignore_ascii_case("uvx") && command_exists("uvx") {
        let mut command = Command::new("uvx");
        command.arg(interpreter);
        return command;
    }

    Command::new(interpreter)
}

fn command_exists(name: &str) -> bool {
    let Some(paths) = std::env::var_os("PATH") else {
        return false;
    };

    let candidates = if cfg!(windows) {
        vec![
            format!("{name}.exe"),
            format!("{name}.cmd"),
            name.to_string(),
        ]
    } else {
        vec![name.to_string()]
    };

    std::env::split_paths(&paths).any(|dir| {
        candidates
            .iter()
            .any(|candidate| dir.join(candidate).is_file())
    })
}
