use crate::types::{HttpVerb, ParseError, ParsedLine, RawParsedLine};
use regex::Regex;
use std::{net::IpAddr, sync::LazyLock};

static LINE_RE: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(concat!(
        // start of line
        r#"^"#,
        // ip
        r#"(?P<remote_addr>[A-Za-z0-9.:%]+)"#,
        // extra
        r#" - - \["#,
        // timestamp
        r#"(?P<timestamp>[\w/: +TZ.-]+)"#,
        // extra
        r#"\] ""#,
        // method
        r#"(?P<method>[A-Z]{3,7})"#,
        // extra
        r#" "#,
        // path
        r#"(?P<path>[\w\p{L}/.?=%&@:#\[\]@!$'\(\)*+,;~\|-]+)"#,
        // extra
        r#" "#,
        // protocol
        r#"(?P<protocol>[A-Z0-9./]+)"#,
        // extra
        r#"" "#,
        // status
        r#"(?P<status>\d{3})"#,
        // extra
        r#" "#,
        // bytes
        r#"(?P<bytes_sent>\d+)"#,
        // extra
        r#" "#,
        // response time
        r#"(?P<response_time_ms>[\d.e+-]+)"#,
        // extra
        r#" ""#,
        // user agent
        r#"(?P<user_agent>.*)"#,
        // extra
        r#"""#,
        // end of line
        r#"$"#,
    ))
    .expect("invalid log line regex")
});

pub fn parse_line(n: &usize, line: &str) -> Result<RawParsedLine, ParseError> {
    if let Some(caps) = LINE_RE.captures(line) {
        return Ok(RawParsedLine {
            remote_address: String::from(&caps["remote_addr"]),
            timestamp: String::from(&caps["timestamp"]),
            method: String::from(&caps["method"]),
            path: String::from(&caps["path"]),
            protocol: String::from(&caps["protocol"]),
            status: String::from(&caps["status"]),
            bytes_sent: String::from(&caps["bytes_sent"]),
            response_time_ms: String::from(&caps["response_time_ms"]),
            user_agent: String::from(&caps["user_agent"]),
        });
    }
    Err(ParseError::Strict {
        line: *n,
        message: "failed to parse line".to_string(),
    })
}

pub fn validate_line(l: &RawParsedLine) -> Result<ParsedLine, String> {
    // Validate remote address
    if l.remote_address.parse::<IpAddr>().is_err() {
        return Err(format!("invalid remote address: {}", l.remote_address));
    };

    // Validate method
    let method = l
        .method
        .parse::<HttpVerb>()
        .map_err(|_| format!("invalid method: {}", l.method))?;

    // Cut query string from path
    let path_parts = &l.path.split_once('?');
    let path = if path_parts.is_none() {
        &l.path
    } else {
        path_parts.unwrap().0
    };

    // Validate status
    let status = l
        .status
        .parse()
        .map_err(|_| format!("invalid status: {}", l.status))?;

    if !(100..=599).contains(&status) {
        return Err(format!("invalid status: {status}"));
    }

    // Validate bytes sent
    let bytes_sent = l
        .bytes_sent
        .parse::<usize>()
        .map_err(|_| format!("invalid bytes sent: {}", l.bytes_sent))?;

    // Validate response time
    let response_time = l
        .response_time_ms
        .parse::<f32>()
        .map_err(|_| format!("invalid response time: {}", l.response_time_ms))?;

    if response_time < 0.0 || !response_time.is_finite() {
        return Err(format!("invalid response time: {}", l.response_time_ms));
    }

    Ok(ParsedLine {
        remote_address: l.remote_address.clone(),
        timestamp: l.timestamp.clone(),
        method,
        path: path.to_string(),
        protocol: l.protocol.clone(),
        status,
        bytes_sent,
        response_time_ms: response_time,
        user_agent: l.user_agent.clone(),
    })
}

#[cfg(test)]
mod parse_line_tests {
    use super::parse_line;
    use crate::{test_helpers::LogLine, types::ParseError};

    #[test]
    fn extracts_all_fields() {
        let line = LogLine::default();
        let r = parse_line(&1, &line.as_string()).unwrap();

        assert_eq!(r.remote_address, line.remote_address);
        assert_eq!(r.timestamp, line.timestamp);
        assert_eq!(r.method, line.method);
        assert_eq!(r.path, line.path);
        assert_eq!(r.protocol, line.protocol);
        assert_eq!(r.status, line.status);
        assert_eq!(r.bytes_sent, line.bytes_sent);
        assert_eq!(r.response_time_ms, line.response_time_ms);
        assert_eq!(r.user_agent, line.user_agent);
    }

    #[test]
    fn preserves_spaces_in_user_agent() {
        let line = LogLine {
            user_agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.user_agent, line.user_agent)
    }

    #[test]
    fn accepts_zero_bytes() {
        let line = LogLine {
            bytes_sent: "0".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.bytes_sent, line.bytes_sent)
    }

    #[test]
    fn accepts_large_byte_counts() {
        let line = LogLine {
            bytes_sent: "1000000000000".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.bytes_sent, line.bytes_sent);
    }

    #[test]
    fn accepts_http_3() {
        let line = LogLine {
            protocol: "HTTP/3".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.protocol, line.protocol)
    }

    #[test]
    fn reject_obvious_malform_blank() {
        let err = parse_line(&1, "").unwrap_err();
        assert!(matches!(err, ParseError::Strict { line: 1, .. }))
    }

    #[test]
    fn reject_obvious_malform_text() {
        let err = parse_line(&1, "this is not a log line").unwrap_err();
        assert!(matches!(err, ParseError::Strict { line: 1, .. }))
    }

    #[test]
    fn reject_obvious_malform_truncated() {
        let err = parse_line(&1, "10.0.0.1 - - [06/Nov/2026:08:12:01 +0000]").unwrap_err();
        assert!(matches!(err, ParseError::Strict { line: 1, .. }))
    }

    #[test]
    fn reject_obvious_malform_truncated_user_agent() {
        let err = parse_line(
            &1,
            r#"10.0.0.1 - - [06/Nov/2026:08:12:01 +0000] "GET / HTTP/1.1" 200 42 1.2 "Go"#,
        )
        .unwrap_err();
        assert!(matches!(err, ParseError::Strict { line: 1, .. }))
    }

    #[test]
    fn accepts_ipv6() {
        let line = LogLine {
            remote_address: "2001:db8::1".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.remote_address, line.remote_address)
    }

    #[test]
    fn accepts_ipv6_ipv4_mapped() {
        let line = LogLine {
            remote_address: "::ffff:192.168.1.1".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.remote_address, line.remote_address)
    }

    #[test]
    fn accepts_ipv6_zone_scoped() {
        let line = LogLine {
            remote_address: "fe80::1%eth0".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.remote_address, line.remote_address)
    }

    #[test]
    fn accepts_simple_query_string() {
        let line = LogLine {
            path: "/api/v1/items?page=2".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.path, line.path)
    }

    #[test]
    fn accepts_complex_query_string() {
        let line = LogLine {
            path: "/api/v1/items?page=2&foo=bar&user=@me&~group=|bar".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.path, line.path)
    }

    #[test]
    fn accepts_common_url_chars_dash() {
        let line = LogLine {
            path: "/files/hello-world".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.path, line.path)
    }

    #[test]
    fn accepts_common_url_chars_url_encode() {
        let line = LogLine {
            path: "/files/a%20b.txt".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.path, line.path)
    }

    #[test]
    fn accepts_iso_like_timestamps() {
        let line = LogLine {
            timestamp: "2026-11-15T08:08:19.12345Z".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.timestamp, line.timestamp)
    }

    #[test]
    fn accepts_empty_quote_user_aqgent() {
        let line = LogLine {
            user_agent: "".into(),
            ..Default::default()
        };
        let r = parse_line(&1, &line.as_string()).unwrap();
        assert_eq!(r.user_agent, line.user_agent)
    }
}

#[cfg(test)]
mod validate_line_tests {
    use crate::{parser::validate_line, test_helpers::LogLine};

    #[test]
    fn accepts_valid_line() {
        let line = LogLine::default();
        let r = validate_line(&line.as_raw_parsed()).unwrap();
        let want = line.as_parsed();

        assert_eq!(r.remote_address, want.remote_address);
        assert_eq!(r.timestamp, want.timestamp);
        assert_eq!(r.method, want.method);
        assert_eq!(r.path, want.path);
        assert_eq!(r.protocol, want.protocol);
        assert_eq!(r.status, want.status);
        assert_eq!(r.bytes_sent, want.bytes_sent);
        assert_eq!(r.response_time_ms, want.response_time_ms);
        assert_eq!(r.user_agent, want.user_agent);
    }

    #[test]
    fn invalid_ipv4_all_octets() {
        let line = LogLine {
            remote_address: "999.999.999.999".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid remote address"))
    }

    #[test]
    fn invalid_ipv4_one_octet() {
        let line = LogLine {
            remote_address: "256.1.1.1".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid remote address"))
    }

    #[test]
    fn invalid_status_not_int() {
        let line = LogLine {
            status: "20x".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid status"))
    }

    #[test]
    fn invalid_status_too_small() {
        let line = LogLine {
            status: "99".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid status"))
    }

    #[test]
    fn invalid_status_too_big() {
        let line = LogLine {
            status: "999".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid status"))
    }

    #[test]
    fn invalid_bytes_sent_negative() {
        let line = LogLine {
            bytes_sent: "-1".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid bytes sent"))
    }

    #[test]
    fn invalid_response_time_word() {
        let line = LogLine {
            response_time_ms: "slow".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid response time"))
    }

    #[test]
    fn invalid_response_time_semver() {
        let line = LogLine {
            response_time_ms: "1.2.3".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid response time"))
    }

    #[test]
    fn invalid_response_time_negative() {
        let line = LogLine {
            response_time_ms: "-1.1".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid response time"))
    }

    #[test]
    fn invalid_method() {
        let line = LogLine {
            method: "FLY".into(),
            ..Default::default()
        };
        let err = validate_line(&line.as_raw_parsed()).unwrap_err();

        println!("{}", err);
        assert!(err.contains("invalid method"))
    }
}
