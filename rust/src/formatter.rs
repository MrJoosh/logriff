use std::collections::BTreeMap;

use crate::types::{Format, Output, RouteCount};
use serde_json::Result;

pub fn format_output(output: Output, format: Format) -> Result<String> {
    match format {
        Format::Json => serde_json::to_string(&output),
        Format::Text => Ok(get_text_output(&output)),
    }
}

fn comma(i: usize) -> String {
    let s = i.to_string();
    if s.len() <= 3 {
        return s;
    }
    let mut pre = s.len() % 3;
    if pre == 0 {
        pre = 3;
    }
    let mut out = s[..pre].to_string();
    let mut i = pre;
    while i < s.len() {
        out += ",";
        out += &s[i..i + 3];
        i += 3
    }
    out
}

fn get_text_output(output: &Output) -> String {
    format!(
        r#"=== LogRiff Summary ===
Lines Processed : {}
Matched Records : {}
Skipped (Errors): {}

--- Traffic & Bandwidth ---
Total Volume    : {:.2} GB ({} bytes)

--- Latency (ms) ---
Min             : {:.2} ms
Mean            : {:.2} ms
p95             : {:.2} ms
Max             : {:.2} ms

--- Status Codes ---
{}

--- Top Routes (Top {}) ---
{}
========================"#,
        comma(output.summary.total_lines),
        comma(output.summary.matched_lines),
        comma(output.summary.error_lines),
        output.traffic.total_bytes as f64 / (1 << 30) as f64,
        comma(output.traffic.total_bytes),
        output.latency.min,
        output.latency.mean,
        output.latency.p95,
        output.latency.max,
        format_status_codes_text(&output.status_codes, &output.summary.total_lines),
        output.top_routes.len(),
        format_top_routes_text(&output.top_routes)
    )
}

fn format_status_codes_text(status_codes: &BTreeMap<String, usize>, total: &usize) -> String {
    if status_codes.is_empty() {
        return "".into();
    }
    let mut out: Vec<String> = vec![];
    for (status_code, count) in status_codes {
        out.push(format!(
            "{} : {:7} ({:5.2}%)",
            status_code,
            count,
            (*count as f64 / *total as f64) * 100.0
        ))
    }
    out.join("\n")
}

fn format_top_routes_text(top_routes: &[RouteCount]) -> String {
    if top_routes.is_empty() {
        return "".to_string();
    }
    let mut out: Vec<String> = vec![];
    for (i, top_route) in top_routes.iter().enumerate() {
        out.push(format!(
            "{}. {:33} : {}",
            i + 1,
            top_route.path,
            comma(top_route.count),
        ))
    }
    out.join("\n")
}

#[cfg(test)]
mod test_comma {
    use super::comma;

    #[test]
    fn test_1m() {
        let r = comma(1000000);
        assert_eq!(r, "1,000,000".to_string())
    }

    #[test]
    fn test_6_digits() {
        let r = comma(100000);
        assert_eq!(r, "100,000".to_string())
    }

    #[test]
    fn test_no_commas() {
        let r = comma(100);
        assert_eq!(r, "100".to_string())
    }
}

#[cfg(test)]
mod test_json_format {}

#[cfg(test)]
mod test_text_format {}
