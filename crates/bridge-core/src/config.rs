use std::{fs, path::Path};

use serde::{Deserialize, Serialize};

#[derive(Debug, thiserror::Error)]
pub enum ConfigError {
    #[error("failed to read config file {path}: {source}")]
    Read {
        path: String,
        #[source]
        source: std::io::Error,
    },
    #[error("failed to parse config file {path}: {source}")]
    Parse {
        path: String,
        #[source]
        source: serde_yaml::Error,
    },
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct BridgeConfig {
    pub server: ServerConfig,
    pub python: PythonConfig,
    pub executor: ExecutorConfig,
    pub gateway: GatewayConfig,
    pub scripts: ScriptConfig,
    pub forwarding: ForwardingConfig,
    pub storage: StorageConfig,
    pub logging: LoggingConfig,
    pub dashboard: DashboardConfig,
}

impl Default for BridgeConfig {
    fn default() -> Self {
        Self {
            server: ServerConfig::default(),
            python: PythonConfig::default(),
            executor: ExecutorConfig::default(),
            gateway: GatewayConfig::default(),
            scripts: ScriptConfig::default(),
            forwarding: ForwardingConfig::default(),
            storage: StorageConfig::default(),
            logging: LoggingConfig::default(),
            dashboard: DashboardConfig::default(),
        }
    }
}

impl BridgeConfig {
    pub fn from_path(path: impl AsRef<Path>) -> Result<Self, ConfigError> {
        let path_ref = path.as_ref();
        let raw = fs::read_to_string(path_ref).map_err(|source| ConfigError::Read {
            path: path_ref.display().to_string(),
            source,
        })?;
        serde_yaml::from_str(&raw).map_err(|source| ConfigError::Parse {
            path: path_ref.display().to_string(),
            source,
        })
    }

    pub fn bind_addr(&self) -> String {
        format!("{}:{}", self.server.host, self.server.port)
    }

    pub fn executor_endpoint(&self) -> String {
        self.executor_endpoint_for(0)
    }

    pub fn executor_endpoint_for(&self, worker_index: u16) -> String {
        format!(
            "http://{}:{}",
            self.executor.host,
            self.executor.port + worker_index
        )
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ServerConfig {
    pub host: String,
    pub port: u16,
    pub request_timeout_secs: u64,
    pub cors_allowed_origins: Vec<String>,
}

impl Default for ServerConfig {
    fn default() -> Self {
        Self {
            host: "0.0.0.0".to_string(),
            port: 8080,
            request_timeout_secs: 30,
            cors_allowed_origins: vec![
                "http://localhost:3000".to_string(),
                "http://localhost:3002".to_string(),
                "http://127.0.0.1:3000".to_string(),
                "http://127.0.0.1:3002".to_string(),
            ],
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct PythonConfig {
    pub interpreter: String,
    pub interpreter_strategy: String,
    pub environment_manager: String,
    pub uv_project_dir: String,
    pub managed_runtime_dir: String,
    pub allow_system_fallback: bool,
    pub executor_module: String,
    pub plugin_dirs: Vec<String>,
    pub auto_start: bool,
    pub embedded_runtime: bool,
}

impl Default for PythonConfig {
    fn default() -> Self {
        Self {
            interpreter: "python".to_string(),
            interpreter_strategy: "managed".to_string(),
            environment_manager: "uv".to_string(),
            uv_project_dir: ".".to_string(),
            managed_runtime_dir: "data/python".to_string(),
            allow_system_fallback: true,
            executor_module: "python_executor.main".to_string(),
            plugin_dirs: vec!["example_plugins".to_string(), "plugins".to_string()],
            auto_start: true,
            embedded_runtime: true,
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct GatewayConfig {
    pub enabled: bool,
    pub public_path: String,
    pub default_route: String,
    pub provider_routes: std::collections::HashMap<String, String>,
}

impl Default for GatewayConfig {
    fn default() -> Self {
        let mut provider_routes = std::collections::HashMap::new();
        provider_routes.insert("github".to_string(), "github".to_string());
        provider_routes.insert("gitlab".to_string(), "gitlab".to_string());
        provider_routes.insert("sentry".to_string(), "sentry".to_string());

        Self {
            enabled: true,
            public_path: "/gateway".to_string(),
            default_route: String::new(),
            provider_routes,
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ScriptConfig {
    pub enabled: bool,
    pub routes: Vec<ScriptRouteConfig>,
    pub groups: Vec<ScriptGroupConfig>,
}

impl Default for ScriptConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            routes: Vec::new(),
            groups: Vec::new(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ScriptGroupConfig {
    pub name: String,
    pub enabled: bool,
    pub mode: String,
    pub routes: Vec<String>,
}

impl Default for ScriptGroupConfig {
    fn default() -> Self {
        Self {
            name: String::new(),
            enabled: true,
            mode: "parallel".to_string(),
            routes: Vec::new(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ScriptRouteConfig {
    pub name: String,
    pub shell: String,
    pub script_path: String,
    pub working_dir: String,
    pub enabled: bool,
    pub timeout_secs: u64,
    pub env: std::collections::HashMap<String, String>,
}

impl Default for ScriptRouteConfig {
    fn default() -> Self {
        Self {
            name: String::new(),
            shell: "powershell".to_string(),
            script_path: String::new(),
            working_dir: ".".to_string(),
            enabled: true,
            timeout_secs: 30,
            env: std::collections::HashMap::new(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ForwardingConfig {
    pub enabled: bool,
    pub routes: Vec<ForwardRouteConfig>,
}

impl Default for ForwardingConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            routes: Vec::new(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ForwardRouteConfig {
    pub name: String,
    pub target_url: String,
    pub method: String,
    pub enabled: bool,
    pub timeout_secs: u64,
    pub headers: std::collections::HashMap<String, String>,
}

impl Default for ForwardRouteConfig {
    fn default() -> Self {
        Self {
            name: String::new(),
            target_url: String::new(),
            method: "POST".to_string(),
            enabled: true,
            timeout_secs: 30,
            headers: std::collections::HashMap::new(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct ExecutorConfig {
    pub host: String,
    pub port: u16,
    pub workers: u16,
    pub connect_timeout_secs: u64,
}

impl Default for ExecutorConfig {
    fn default() -> Self {
        Self {
            host: "127.0.0.1".to_string(),
            port: 50051,
            workers: 1,
            connect_timeout_secs: 5,
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct StorageConfig {
    pub sqlite_path: String,
}

impl Default for StorageConfig {
    fn default() -> Self {
        Self {
            sqlite_path: "data/webhook-bridge.db".to_string(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct LoggingConfig {
    pub level: String,
    pub format: String,
}

impl Default for LoggingConfig {
    fn default() -> Self {
        Self {
            level: "info".to_string(),
            format: "text".to_string(),
        }
    }
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(default)]
pub struct DashboardConfig {
    pub enabled: bool,
    pub api_prefix: String,
}

impl Default for DashboardConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            api_prefix: "/api/dashboard".to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn defaults_keep_python_executor_enabled() {
        let cfg = BridgeConfig::default();

        assert_eq!(cfg.server.port, 8080);
        assert_eq!(cfg.executor.port, 50051);
        assert_eq!(cfg.executor.workers, 1);
        assert!(cfg.python.auto_start);
        assert_eq!(cfg.python.interpreter_strategy, "managed");
        assert_eq!(cfg.python.environment_manager, "uv");
        assert!(cfg.gateway.enabled);
        assert_eq!(cfg.gateway.provider_routes["github"], "github");
        assert!(cfg.scripts.enabled);
        assert!(cfg.forwarding.enabled);
        assert!(cfg
            .python
            .plugin_dirs
            .contains(&"example_plugins".to_string()));
    }
}
