use std::{
    fs,
    io::{Read, Write},
    net::TcpListener,
    path::{Path, PathBuf},
    process::{Child, Command, Stdio},
    sync::mpsc,
    thread,
    time::{Duration, Instant},
};

use reqwest::Client;
use serde_json::{json, Value};
use tempfile::TempDir;

fn free_port() -> u16 {
    TcpListener::bind("127.0.0.1:0")
        .expect("bind ephemeral port")
        .local_addr()
        .expect("local addr")
        .port()
}

fn repo_root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .and_then(Path::parent)
        .expect("repo root")
        .to_path_buf()
}

fn write_config(dir: &TempDir, server_port: u16, executor_port: u16, workers: u16) -> PathBuf {
    let root = repo_root();
    let db = dir.path().join("e2e.db");
    let config = dir.path().join("config.yaml");
    let plugin_dir = root.join("example_plugins");
    let config_text = format!(
        r#"
server:
  host: "127.0.0.1"
  port: {server_port}
python:
  interpreter: "python"
  environment_manager: "system"
  executor_module: "python_executor.main"
  auto_start: true
  plugin_dirs:
    - "{}"
executor:
  host: "127.0.0.1"
  port: {executor_port}
  workers: {workers}
  connect_timeout_secs: 5
gateway:
  enabled: true
  public_path: "/gateway"
  provider_routes:
    github: "github"
    gitlab: "gitlab"
    sentry: "sentry"
storage:
  sqlite_path: "{}"
logging:
  level: "warn"
  format: "text"
dashboard:
  enabled: true
  api_prefix: "/api/dashboard"
"#,
        plugin_dir.display().to_string().replace('\\', "/"),
        db.display().to_string().replace('\\', "/"),
    );
    fs::write(&config, config_text).expect("write config");
    config
}

fn write_forward_config(
    dir: &TempDir,
    server_port: u16,
    executor_port: u16,
    target_port: u16,
) -> PathBuf {
    let db = dir.path().join("forward.db");
    let config = dir.path().join("forward-config.yaml");
    let config_text = format!(
        r#"
server:
  host: "127.0.0.1"
  port: {server_port}
python:
  interpreter: "python"
  interpreter_strategy: "managed"
  environment_manager: "system"
  allow_system_fallback: true
  executor_module: "python_executor.main"
  auto_start: false
executor:
  host: "127.0.0.1"
  port: {executor_port}
  workers: 1
  connect_timeout_secs: 1
forwarding:
  enabled: true
  routes:
    - name: "github-forward"
      target_url: "http://127.0.0.1:{target_port}/github"
      method: "POST"
      enabled: true
      timeout_secs: 5
      headers:
        x-forwarded-by: "webhook-bridge"
storage:
  sqlite_path: "{}"
logging:
  level: "warn"
  format: "text"
dashboard:
  enabled: true
  api_prefix: "/api/dashboard"
"#,
        db.display().to_string().replace('\\', "/"),
    );
    fs::write(&config, config_text).expect("write config");
    config
}

fn spawn_server(config: &Path) -> Child {
    Command::new(env!("CARGO_BIN_EXE_webhook-bridge"))
        .arg("run")
        .arg("--config")
        .arg(config)
        .current_dir(repo_root())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .expect("spawn server")
}

fn spawn_mock_receiver() -> (u16, mpsc::Receiver<String>, thread::JoinHandle<()>) {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind mock receiver");
    let port = listener.local_addr().expect("mock addr").port();
    let (tx, rx) = mpsc::channel();
    let handle = thread::spawn(move || {
        let (mut stream, _) = listener.accept().expect("accept request");
        let mut buffer = vec![0; 8192];
        let read = stream.read(&mut buffer).expect("read request");
        let request = String::from_utf8_lossy(&buffer[..read]).to_string();
        tx.send(request).expect("send request");
        let body = r#"{"accepted":true,"target":"mock-github"}"#;
        let response = format!(
            "HTTP/1.1 202 Accepted\r\ncontent-type: application/json\r\ncontent-length: {}\r\nconnection: close\r\n\r\n{}",
            body.len(),
            body
        );
        stream
            .write_all(response.as_bytes())
            .expect("write response");
    });
    (port, rx, handle)
}

async fn wait_for_health(port: u16) {
    let client = Client::new();
    let deadline = Instant::now() + Duration::from_secs(20);
    loop {
        if Instant::now() > deadline {
            panic!("server did not become healthy");
        }

        if let Ok(response) = client
            .get(format!("http://127.0.0.1:{port}/health"))
            .send()
            .await
        {
            if response.status().is_success() {
                return;
            }
        }

        thread::sleep(Duration::from_millis(250));
    }
}

#[tokio::test]
async fn records_github_webhook_execution_and_logs() {
    let temp = TempDir::new().unwrap();
    let server_port = free_port();
    let executor_port = free_port();
    let config = write_config(&temp, server_port, executor_port, 1);
    let mut server = spawn_server(&config);

    let result = async {
        wait_for_health(server_port).await;
        let client = Client::new();
        let payload = json!({
            "ref": "refs/heads/main",
            "repository": {"full_name": "loonghao/webhook_bridge"},
            "sender": {"login": "octocat"}
        });

        let webhook: Value = client
            .post(format!("http://127.0.0.1:{server_port}/api/webhook/github"))
            .header("X-GitHub-Event", "push")
            .json(&payload)
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();

        assert_eq!(webhook["success"], true);
        assert_eq!(webhook["data"]["provider"], "github");
        assert_eq!(webhook["data"]["repository"], "loonghao/webhook_bridge");

        let stats: Value = client
            .get(format!(
                "http://127.0.0.1:{server_port}/api/dashboard/stats"
            ))
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();
        assert_eq!(stats["data"]["total_requests"], 1);

        let logs: Value = client
            .get(format!("http://127.0.0.1:{server_port}/api/dashboard/logs"))
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();
        assert!(logs["data"].as_array().unwrap().iter().any(|entry| {
            entry["plugin"] == "github" && entry["message"].as_str().unwrap().contains("Executed")
        }));
    }
    .await;

    let _ = server.kill();
    result
}

#[tokio::test]
async fn starts_multiple_workers_and_reports_them() {
    let temp = TempDir::new().unwrap();
    let server_port = free_port();
    let executor_port = free_port();
    let config = write_config(&temp, server_port, executor_port, 2);
    let mut server = spawn_server(&config);

    let result = async {
        wait_for_health(server_port).await;
        let client = Client::new();
        let workers: Value = client
            .get(format!(
                "http://127.0.0.1:{server_port}/api/dashboard/workers"
            ))
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();

        assert_eq!(workers["data"].as_array().unwrap().len(), 2);
    }
    .await;

    let _ = server.kill();
    result
}

#[tokio::test]
async fn unified_gateway_detects_github_and_runs_local_hook() {
    let temp = TempDir::new().unwrap();
    let server_port = free_port();
    let executor_port = free_port();
    let config = write_config(&temp, server_port, executor_port, 1);
    let mut server = spawn_server(&config);

    let result = async {
        wait_for_health(server_port).await;
        let client = Client::new();
        let payload = json!({
            "ref": "refs/heads/main",
            "repository": {"full_name": "loonghao/webhook_bridge"},
            "sender": {"login": "octocat"}
        });

        let webhook: Value = client
            .post(format!("http://127.0.0.1:{server_port}/gateway"))
            .header("X-GitHub-Event", "push")
            .json(&payload)
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();

        assert_eq!(webhook["success"], true);
        assert_eq!(webhook["data"]["provider"], "github");
        assert_eq!(webhook["data"]["repository"], "loonghao/webhook_bridge");

        let logs: Value = client
            .get(format!("http://127.0.0.1:{server_port}/api/dashboard/logs"))
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();
        assert!(logs["data"].as_array().unwrap().iter().any(|entry| {
            entry["plugin"] == "github" && entry["message"].as_str().unwrap().contains("Executed")
        }));
    }
    .await;

    let _ = server.kill();
    result
}

#[tokio::test]
async fn forwards_webhook_route_and_records_execution() {
    let temp = TempDir::new().unwrap();
    let server_port = free_port();
    let executor_port = free_port();
    let (target_port, received, receiver_thread) = spawn_mock_receiver();
    let config = write_forward_config(&temp, server_port, executor_port, target_port);
    let mut server = spawn_server(&config);

    let result = async {
        wait_for_health(server_port).await;
        let client = Client::new();
        let payload = json!({
            "repository": {"full_name": "loonghao/webhook_bridge"},
            "sender": {"login": "octocat"}
        });

        let forwarded: Value = client
            .post(format!(
                "http://127.0.0.1:{server_port}/api/webhook/github-forward?source=github"
            ))
            .header("X-GitHub-Event", "push")
            .json(&payload)
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();

        assert_eq!(forwarded["success"], true);
        assert_eq!(forwarded["data"]["accepted"], true);

        let raw_request = received
            .recv_timeout(Duration::from_secs(5))
            .expect("mock receiver request");
        assert!(raw_request.starts_with("POST /github?source=github"));
        assert!(raw_request.contains("x-forwarded-by: webhook-bridge"));
        assert!(raw_request.contains("loonghao/webhook_bridge"));

        let logs: Value = client
            .get(format!("http://127.0.0.1:{server_port}/api/dashboard/logs"))
            .send()
            .await
            .unwrap()
            .json()
            .await
            .unwrap();
        assert!(logs["data"].as_array().unwrap().iter().any(|entry| {
            entry["plugin"] == "github-forward"
                && entry["source"] == "rust-forwarder"
                && entry["message"].as_str().unwrap().contains("Forwarded")
        }));
    }
    .await;

    let _ = server.kill();
    let _ = receiver_thread.join();
    result
}
