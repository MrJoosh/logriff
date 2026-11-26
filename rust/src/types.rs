use std::{collections::HashMap, fmt, io, str::FromStr};

use clap::ValueEnum;

#[derive(Copy, Clone, Debug, PartialEq, Eq, ValueEnum)]
pub enum Format {
    Text,
    Json,
}

#[derive(Copy, Clone, Debug, PartialEq, Eq, ValueEnum)]
pub enum HttpVerb {
    Connect,
    Delete,
    Get,
    Head,
    Options,
    Patch,
    Post,
    Put,
    Trace,
}

impl FromStr for HttpVerb {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_ascii_uppercase().as_str() {
            "CONNECT" => Ok(Self::Connect),
            "DELETE" => Ok(Self::Delete),
            "GET" => Ok(Self::Get),
            "HEAD" => Ok(Self::Head),
            "OPTIONS" => Ok(Self::Options),
            "PATCH" => Ok(Self::Patch),
            "POST" => Ok(Self::Post),
            "PUT" => Ok(Self::Put),
            "TRACE" => Ok(Self::Trace),
            _ => Err(format!("unknown HTTP verb: '{s}'")),
        }
    }
}

#[derive(Debug)]
pub struct ParsedLine {
    pub remote_address: String,
    pub timestamp: String,
    pub method: HttpVerb,
    pub path: String,
    pub protocol: String,
    pub status: u16,
    pub bytes_sent: u64,
    pub response_time_ms: f32,
    pub user_agent: String,
}

#[derive(Debug)]
pub struct RawParsedLine {
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

#[derive(Debug)]
pub enum ParseError {
    Io(io::Error),
    Strict { line: usize, message: String },
}

impl fmt::Display for ParseError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ParseError::Io(err) => write!(f, "I/O error: {err}"),
            ParseError::Strict { line, message } => {
                write!(f, "malformed input in line {line}: {message}")
            }
        }
    }
}

impl std::error::Error for ParseError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            ParseError::Io(err) => Some(err),
            ParseError::Strict { .. } => None,
        }
    }
}

impl From<io::Error> for ParseError {
    fn from(err: io::Error) -> Self {
        ParseError::Io(err)
    }
}

#[derive(Default, Debug)]
pub struct Stats {
    pub lines: usize,
    pub matched: usize,
    pub error: usize,
    pub bytes_volume: u64,
    pub response_times: Vec<f32>,
    pub status_codes: HashMap<u16, u16>,
    pub route_distribution: HashMap<String, u32>,
}

impl Stats {
    pub fn new() -> Self {
        Default::default()
    }
}
