use std::{
    fs,
    path::{Path, PathBuf},
};

use crate::config::BridgeConfig;

const EMBEDDED_FILES: &[(&str, &str)] = &[
    (
        "python_executor/__init__.py",
        include_str!("../../../python_executor/__init__.py"),
    ),
    (
        "python_executor/main.py",
        include_str!("../../../python_executor/main.py"),
    ),
    (
        "python_executor/server.py",
        include_str!("../../../python_executor/server.py"),
    ),
    (
        "python_executor/utils.py",
        include_str!("../../../python_executor/utils.py"),
    ),
    (
        "webhook_bridge/__init__.py",
        include_str!("../../../webhook_bridge/__init__.py"),
    ),
    (
        "webhook_bridge/plugin.py",
        include_str!("../../../webhook_bridge/plugin.py"),
    ),
    (
        "webhook_bridge/filesystem.py",
        include_str!("../../../webhook_bridge/filesystem.py"),
    ),
    ("api/__init__.py", include_str!("../../../api/__init__.py")),
    (
        "api/proto/__init__.py",
        include_str!("../../../api/proto/__init__.py"),
    ),
    (
        "api/proto/webhook_pb2.py",
        include_str!("../../../api/proto/webhook_pb2.py"),
    ),
    (
        "api/proto/webhook_pb2_grpc.py",
        include_str!("../../../api/proto/webhook_pb2_grpc.py"),
    ),
];

pub fn materialize_embedded_runtime(cfg: &BridgeConfig) -> std::io::Result<PathBuf> {
    let root = PathBuf::from(&cfg.storage.sqlite_path)
        .parent()
        .unwrap_or_else(|| Path::new("data"))
        .join("runtime")
        .join(env!("CARGO_PKG_VERSION"));

    for (relative, contents) in EMBEDDED_FILES {
        let path = root.join(relative);
        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }
        fs::write(path, contents)?;
    }

    Ok(root)
}

pub fn resolve_python_interpreter(cfg: &BridgeConfig) -> PathBuf {
    if cfg
        .python
        .interpreter_strategy
        .eq_ignore_ascii_case("system")
    {
        return PathBuf::from(&cfg.python.interpreter);
    }

    let managed_root = PathBuf::from(&cfg.python.managed_runtime_dir);
    let managed_python = if cfg!(windows) {
        managed_root.join("python.exe")
    } else {
        managed_root.join("bin").join("python")
    };

    if managed_python.exists() {
        return managed_python;
    }

    if cfg.python.allow_system_fallback {
        return PathBuf::from(&cfg.python.interpreter);
    }

    managed_python
}
