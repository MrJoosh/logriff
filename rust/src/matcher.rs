use crate::{cli::Args, types::ParsedLine};

pub fn match_line(l: &ParsedLine, args: &Args) -> bool {
    if let Some(status_min) = args.status_min {
        if l.status < status_min {
            return false;
        }
    }
    if let Some(status_max) = args.status_max {
        if l.status > status_max {
            return false;
        }
    }
    if !args.method.is_none() && l.method != args.method.unwrap() {
        return false;
    }
    if let Some(path_prefix) = args.path_prefix.as_deref() {
        if !l.path.starts_with(path_prefix) {
            return false;
        }
    }
    if let Some(min_duration) = args.min_duration {
        if l.response_time_ms < min_duration {
            return false;
        }
    }
    true
}

#[cfg(test)]
mod match_line_tests {
    use super::match_line;
    use crate::{cli::Args, test_helpers::LogLine, types::HttpVerb};

    #[test]
    fn match_multi_filter() {
        let line = LogLine::default();
        let args = Args {
            status_min: Some(200),
            status_max: Some(201),
            method: Some(HttpVerb::Get),
            path_prefix: Some("/api".into()),
            min_duration: Some(1.1),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args));
    }

    #[test]
    fn no_match_multi_filter() {
        let line = LogLine::default();
        let args = Args {
            status_min: Some(200),
            status_max: Some(201),
            method: Some(HttpVerb::Get),
            path_prefix: Some("/api".into()),
            min_duration: Some(1.5),
            ..Default::default()
        };
        assert!(!match_line(&line.as_parsed(), &args));
    }

    #[test]
    fn match_status_min() {
        let line = LogLine::default();
        let args = Args {
            status_min: Some(200),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn no_match_status_min() {
        let line = LogLine::default();
        let args = Args {
            status_min: Some(201),
            ..Default::default()
        };
        assert!(!match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn match_status_max() {
        let line = LogLine::default();
        let args = Args {
            status_max: Some(201),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn no_match_status_max() {
        let line = LogLine::default();
        let args = Args {
            status_max: Some(199),
            ..Default::default()
        };
        assert!(!match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn match_method() {
        let line = LogLine::default();
        let args = Args {
            method: Some(HttpVerb::Get),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn no_match_method() {
        let line = LogLine::default();
        let args = Args {
            method: Some(HttpVerb::Post),
            ..Default::default()
        };
        assert!(!match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn match_path_prefix() {
        let line = LogLine::default();
        let args = Args {
            path_prefix: Some("/api".into()),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn match_path_prefix_full() {
        let line = LogLine::default();
        let args = Args {
            path_prefix: Some("/api/v1/health".into()),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn no_match_path_prefix() {
        let line = LogLine::default();
        let args = Args {
            path_prefix: Some("/assets".into()),
            ..Default::default()
        };
        assert!(!match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn match_min_duration() {
        let line = LogLine::default();
        let args = Args {
            min_duration: Some(1.2),
            ..Default::default()
        };
        assert!(match_line(&line.as_parsed(), &args))
    }

    #[test]
    fn no_match_min_duration() {
        let line = LogLine::default();
        let args = Args {
            min_duration: Some(1.3),
            ..Default::default()
        };
        assert!(!match_line(&line.as_parsed(), &args))
    }
}
