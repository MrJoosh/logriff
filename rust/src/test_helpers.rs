use crate::types::{HttpVerb, ParsedLine, RawParsedLine};

#[derive(Debug, Clone)]
pub struct LogLine {
    pub remote_address: String,
    pub timestamp: String,
    pub method: String,
    pub path: String,
    pub protocol: String,
    pub status: String,
    pub bytes_sent: String,
    pub response_time_ms: String,
    pub user_agent: String,
}

impl Default for LogLine {
    fn default() -> Self {
        Self {
            remote_address: "192.168.1.10".into(),
            timestamp: "06/Nov/2026:08:12:01 +0000".into(),
            method: "GET".into(),
            path: "/api/v1/health".into(),
            protocol: "HTTP/1.1".into(),
            status: "200".into(),
            bytes_sent: "42".into(),
            response_time_ms: "1.2".into(),
            user_agent: "Go-http-client/1.1".into(),
        }
    }
}

impl LogLine {
    pub fn as_string(&self) -> String {
        format!(
            r#"{} - - [{}] "{} {} {}" {} {} {} "{}""#,
            self.remote_address,
            self.timestamp,
            self.method,
            self.path,
            self.protocol,
            self.status,
            self.bytes_sent,
            self.response_time_ms,
            self.user_agent
        )
    }
    pub fn as_raw_parsed(&self) -> RawParsedLine {
        RawParsedLine {
            remote_address: self.remote_address.clone(),
            timestamp: self.timestamp.clone(),
            method: self.method.clone(),
            path: self.path.clone(),
            protocol: self.protocol.clone(),
            status: self.status.clone(),
            bytes_sent: self.bytes_sent.clone(),
            response_time_ms: self.response_time_ms.clone(),
            user_agent: self.user_agent.clone(),
        }
    }
    pub fn as_parsed(&self) -> ParsedLine {
        ParsedLine {
            remote_address: self.remote_address.clone(),
            timestamp: self.timestamp.clone(),
            method: self.method.parse::<HttpVerb>().unwrap(),
            path: self.path.clone(),
            protocol: self.protocol.clone(),
            status: self.status.clone().parse().unwrap(),
            bytes_sent: self.bytes_sent.clone().parse().unwrap(),
            response_time_ms: self.response_time_ms.clone().parse().unwrap(),
            user_agent: self.user_agent.clone(),
        }
    }
}
