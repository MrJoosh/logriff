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
        output.latency_ms.min,
        output.latency_ms.mean,
        output.latency_ms.p95,
        output.latency_ms.max,
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
mod test_format_output {
    use pretty_assertions::assert_eq;
    use std::collections::BTreeMap;

    use crate::types::{Format, Latency, Output, RouteCount, Summary, Traffic};

    use super::format_output;

    #[test]
    fn test_json_format_sample() {
        let o: Output = Output {
            summary: Summary {
                total_lines: 5,
                matched_lines: 5,
                error_lines: 0,
            },
            traffic: Traffic { total_bytes: 9330 },
            latency_ms: Latency {
                min: 0.8,
                mean: 62.56,
                p95: 185.56,
                max: 210.6,
            },
            status_codes: BTreeMap::from([
                ("200".into(), 2),
                ("401".into(), 1),
                ("304".into(), 1),
                ("500".into(), 1),
            ]),
            top_routes: vec![
                RouteCount {
                    path: "/api/v1/auth/login".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/api/v1/health".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/api/v1/items/4912".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/api/v1/sessions/current".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/assets/main.css".into(),
                    count: 1,
                },
            ],
        };
        let r = format_output(o, Format::Json).unwrap();
        assert_eq!(r, r#"{"summary":{"total_lines":5,"matched_lines":5,"error_lines":0},"traffic":{"total_bytes":9330},"latency_ms":{"min":0.8,"mean":62.56,"p95":185.56,"max":210.6},"status_codes":{"200":2,"304":1,"401":1,"500":1},"top_routes":[{"path":"/api/v1/auth/login","count":1},{"path":"/api/v1/health","count":1},{"path":"/api/v1/items/4912","count":1},{"path":"/api/v1/sessions/current","count":1},{"path":"/assets/main.css","count":1}]}"#.to_string())
    }

    #[test]
    fn test_json_format_empty() {
        let o: Output = Output {
            summary: Summary {
                total_lines: 0,
                matched_lines: 0,
                error_lines: 0,
            },
            traffic: Traffic { total_bytes: 0 },
            latency_ms: Latency {
                min: 0.0,
                mean: 0.0,
                p95: 0.0,
                max: 0.0,
            },
            status_codes: BTreeMap::from([]),
            top_routes: vec![],
        };
        let r = format_output(o, Format::Json).unwrap();
        assert_eq!(r, r#"{"summary":{"total_lines":0,"matched_lines":0,"error_lines":0},"traffic":{"total_bytes":0},"latency_ms":{"min":0.0,"mean":0.0,"p95":0.0,"max":0.0},"status_codes":{},"top_routes":[]}"#.to_string())
    }

    #[test]
    fn test_text_format_sample() {
        let o: Output = Output {
            summary: Summary {
                total_lines: 5,
                matched_lines: 5,
                error_lines: 0,
            },
            traffic: Traffic { total_bytes: 9330 },
            latency_ms: Latency {
                min: 0.8,
                mean: 62.56,
                p95: 185.56,
                max: 210.6,
            },
            status_codes: BTreeMap::from([
                ("200".into(), 2),
                ("401".into(), 1),
                ("304".into(), 1),
                ("500".into(), 1),
            ]),
            top_routes: vec![
                RouteCount {
                    path: "/api/v1/auth/login".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/api/v1/health".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/api/v1/items/4912".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/api/v1/sessions/current".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/assets/main.css".into(),
                    count: 1,
                },
            ],
        };
        let r = format_output(o, Format::Text).unwrap();
        let expected = r#"=== LogRiff Summary ===
Lines Processed : 5
Matched Records : 5
Skipped (Errors): 0

--- Traffic & Bandwidth ---
Total Volume    : 0.00 GB (9,330 bytes)

--- Latency (ms) ---
Min             : 0.80 ms
Mean            : 62.56 ms
p95             : 185.56 ms
Max             : 210.60 ms

--- Status Codes ---
200 :       2 (40.00%)
304 :       1 (20.00%)
401 :       1 (20.00%)
500 :       1 (20.00%)

--- Top Routes (Top 5) ---
1. /api/v1/auth/login                : 1
2. /api/v1/health                    : 1
3. /api/v1/items/4912                : 1
4. /api/v1/sessions/current          : 1
5. /assets/main.css                  : 1
========================"#
            .to_string();
        assert_eq!(r, expected)
    }

    #[test]
    fn test_text_format_empty() {
        let o: Output = Output {
            summary: Summary {
                total_lines: 0,
                matched_lines: 0,
                error_lines: 0,
            },
            traffic: Traffic { total_bytes: 0 },
            latency_ms: Latency {
                min: 0.0,
                mean: 0.0,
                p95: 0.0,
                max: 0.0,
            },
            status_codes: BTreeMap::from([]),
            top_routes: vec![],
        };
        let r = format_output(o, Format::Text).unwrap();
        let expected = r#"=== LogRiff Summary ===
Lines Processed : 0
Matched Records : 0
Skipped (Errors): 0

--- Traffic & Bandwidth ---
Total Volume    : 0.00 GB (0 bytes)

--- Latency (ms) ---
Min             : 0.00 ms
Mean            : 0.00 ms
p95             : 0.00 ms
Max             : 0.00 ms

--- Status Codes ---


--- Top Routes (Top 0) ---

========================"#
            .to_string();
        assert_eq!(r, expected)
    }
}

#[cfg(test)]
mod test_text_format {}
