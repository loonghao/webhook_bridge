use std::{
    path::Path,
    sync::{Arc, Mutex},
};

use rusqlite::{params, Connection};
use serde_json::{json, Value};

#[derive(Clone)]
pub struct Storage {
    conn: Arc<Mutex<Connection>>,
}

#[derive(Debug)]
pub struct ExecutionInsert<'a> {
    pub request_id: &'a str,
    pub plugin: &'a str,
    pub method: &'a str,
    pub status_code: i32,
    pub success: bool,
    pub execution_time_ms: i64,
    pub input: &'a str,
    pub output: &'a str,
    pub error: &'a str,
}

impl Storage {
    pub fn open(path: impl AsRef<Path>) -> rusqlite::Result<Self> {
        let path = path.as_ref();
        if let Some(parent) = path.parent() {
            let _ = std::fs::create_dir_all(parent);
        }

        let conn = Connection::open(path)?;
        let storage = Self {
            conn: Arc::new(Mutex::new(conn)),
        };
        storage.migrate()?;
        Ok(storage)
    }

    #[cfg(test)]
    pub fn memory() -> rusqlite::Result<Self> {
        let storage = Self {
            conn: Arc::new(Mutex::new(Connection::open_in_memory()?)),
        };
        storage.migrate()?;
        Ok(storage)
    }

    fn migrate(&self) -> rusqlite::Result<()> {
        let conn = self.conn.lock().expect("sqlite mutex poisoned");
        conn.execute_batch(
            r#"
            CREATE TABLE IF NOT EXISTS executions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                request_id TEXT NOT NULL,
                plugin TEXT NOT NULL,
                method TEXT NOT NULL,
                status_code INTEGER NOT NULL,
                success INTEGER NOT NULL,
                execution_time_ms INTEGER NOT NULL,
                input TEXT NOT NULL,
                output TEXT NOT NULL,
                error TEXT NOT NULL,
                created_at TEXT NOT NULL DEFAULT (datetime('now'))
            );

            CREATE TABLE IF NOT EXISTS logs (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                level TEXT NOT NULL,
                source TEXT NOT NULL,
                message TEXT NOT NULL,
                plugin TEXT,
                request_id TEXT,
                metadata TEXT NOT NULL DEFAULT '{}',
                created_at TEXT NOT NULL DEFAULT (datetime('now'))
            );
            "#,
        )
    }

    pub fn insert_execution(&self, execution: ExecutionInsert<'_>) -> rusqlite::Result<()> {
        let conn = self.conn.lock().expect("sqlite mutex poisoned");
        conn.execute(
            r#"
            INSERT INTO executions
                (request_id, plugin, method, status_code, success, execution_time_ms, input, output, error)
            VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)
            "#,
            params![
                execution.request_id,
                execution.plugin,
                execution.method,
                execution.status_code,
                if execution.success { 1 } else { 0 },
                execution.execution_time_ms,
                execution.input,
                execution.output,
                execution.error,
            ],
        )?;
        Ok(())
    }

    pub fn insert_log(
        &self,
        level: &str,
        source: &str,
        message: &str,
        plugin: Option<&str>,
        request_id: Option<&str>,
        metadata: Value,
    ) -> rusqlite::Result<()> {
        let conn = self.conn.lock().expect("sqlite mutex poisoned");
        conn.execute(
            r#"
            INSERT INTO logs (level, source, message, plugin, request_id, metadata)
            VALUES (?1, ?2, ?3, ?4, ?5, ?6)
            "#,
            params![
                level,
                source,
                message,
                plugin,
                request_id,
                metadata.to_string(),
            ],
        )?;
        Ok(())
    }

    pub fn stats(&self) -> rusqlite::Result<Value> {
        let conn = self.conn.lock().expect("sqlite mutex poisoned");
        let total: i64 = conn.query_row("SELECT COUNT(*) FROM executions", [], |row| row.get(0))?;
        let successful: i64 = conn.query_row(
            "SELECT COUNT(*) FROM executions WHERE success = 1",
            [],
            |row| row.get(0),
        )?;
        let failed = total - successful;
        let average_ms: f64 = conn.query_row(
            "SELECT COALESCE(AVG(execution_time_ms), 0) FROM executions",
            [],
            |row| row.get(0),
        )?;

        Ok(json!({
            "total_requests": total,
            "successful_requests": successful,
            "failed_requests": failed,
            "average_response_time": average_ms,
            "error_rate": if total > 0 { failed as f64 / total as f64 } else { 0.0 },
        }))
    }

    pub fn recent_logs(
        &self,
        limit: i64,
        plugin_filter: Option<&str>,
    ) -> rusqlite::Result<Vec<Value>> {
        let conn = self.conn.lock().expect("sqlite mutex poisoned");
        let mut values = Vec::new();

        if let Some(plugin) = plugin_filter {
            let mut stmt = conn.prepare(
                r#"
                SELECT id, created_at, level, message, source, plugin, request_id, metadata
                FROM logs
                WHERE plugin = ?1
                ORDER BY id DESC
                LIMIT ?2
                "#,
            )?;
            let rows = stmt.query_map(params![plugin, limit], log_row_to_json)?;
            for row in rows {
                values.push(row?);
            }
        } else {
            let mut stmt = conn.prepare(
                r#"
                SELECT id, created_at, level, message, source, plugin, request_id, metadata
                FROM logs
                ORDER BY id DESC
                LIMIT ?1
                "#,
            )?;
            let rows = stmt.query_map(params![limit], log_row_to_json)?;
            for row in rows {
                values.push(row?);
            }
        }

        Ok(values)
    }
}

fn log_row_to_json(row: &rusqlite::Row<'_>) -> rusqlite::Result<Value> {
    let id: i64 = row.get(0)?;
    let timestamp: String = row.get(1)?;
    let level: String = row.get(2)?;
    let message: String = row.get(3)?;
    let source: String = row.get(4)?;
    let plugin: Option<String> = row.get(5)?;
    let request_id: Option<String> = row.get(6)?;
    let metadata: String = row.get(7)?;

    Ok(json!({
        "id": id.to_string(),
        "timestamp": timestamp,
        "level": level,
        "message": message,
        "source": source,
        "plugin": plugin,
        "request_id": request_id,
        "metadata": serde_json::from_str::<Value>(&metadata).unwrap_or_else(|_| json!({})),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn records_execution_and_log() {
        let storage = Storage::memory().unwrap();
        storage
            .insert_execution(ExecutionInsert {
                request_id: "req-1",
                plugin: "github",
                method: "POST",
                status_code: 200,
                success: true,
                execution_time_ms: 12,
                input: "{}",
                output: "{\"ok\":true}",
                error: "",
            })
            .unwrap();
        storage
            .insert_log(
                "info",
                "test",
                "executed",
                Some("github"),
                Some("req-1"),
                json!({"kind": "unit"}),
            )
            .unwrap();

        assert_eq!(storage.stats().unwrap()["total_requests"], 1);
        assert_eq!(storage.recent_logs(10, Some("github")).unwrap().len(), 1);
    }
}
