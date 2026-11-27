use crate::types::{Format, HttpVerb};
use clap::Parser;

pub fn validate_args(args: &Args) -> Result<(), String> {
    if let (Some(min), Some(max)) = (args.status_min, args.status_max) {
        if min > max {
            return Err(format!(
                "filter for status-min cannot be greater than status-max: {min} > {max}"
            ));
        }
    }

    Ok(())
}

fn parse_http_status(s: &str) -> Result<usize, String> {
    let n: usize = s.parse().map_err(|_| format!("invalid status: {s}"))?;
    if !(100..=599).contains(&n) {
        return Err(format!("invalid status: {s}"));
    }
    Ok(n)
}

fn parse_min_duration(s: &str) -> Result<f32, String> {
    let n: f32 = s
        .parse()
        .map_err(|_| format!("invalid min-duration: {s}"))?;
    if n < 0.0 {
        return Err(format!("min-duration cannot be negative: {s}"));
    }
    if !n.is_finite() {
        return Err(format!("min-duration must be a finite number: {s}"));
    }
    Ok(n)
}

#[derive(Parser, Debug)]
#[command(about, long_about = None)]
pub struct Args {
    #[arg(
        short,
        long = "status-min",
        value_parser = parse_http_status,
        help = "Keep records with status >= MIN."
    )]
    pub status_min: Option<usize>,

    #[arg(
        short = 'e',
        long = "status-max",
        value_parser = parse_http_status,
        help = "Keep records with status <= MAX."
    )]
    pub status_max: Option<usize>,

    #[arg(
        short,
        long,
        value_enum,
        ignore_case = true,
        help = "Filter by HTTP verb (case-insensitive)."
    )]
    pub method: Option<HttpVerb>,

    #[arg(
        short,
        long = "path-prefix",
        help = "Keep only requests where path starts with the given prefix."
    )]
    pub path_prefix: Option<String>,

    #[arg(
        long = "min-duration",
        value_parser = parse_min_duration,
        help = "Keep records with response time >= duration."
    )]
    pub min_duration: Option<f32>,

    #[arg(
        short = 'n',
        long = "top-routes",
        help = "Number of most requested routes to show in summary.",
        default_value_t = 5
    )]
    pub top_routes: usize,

    #[arg(
        short,
        long,
        value_enum,
        help = "Output format for aggregate results.",
        default_value_t = Format::Text,
    )]
    pub format: Format,

    #[arg(
        long,
        help = "Terminate immediately with non-zero exit code on malformed lines."
    )]
    pub strict: bool,

    #[arg(default_value = "-", help = "Log file to process or - for stdin.")]
    pub file: String,
}

impl Default for Args {
    fn default() -> Self {
        Self {
            status_min: None,
            status_max: None,
            method: None,
            path_prefix: None,
            min_duration: None,
            top_routes: 5,
            format: Format::Text,
            strict: false,
            file: "-".into(),
        }
    }
}

#[cfg(test)]
mod cli_tests {
    use super::{Args, Format, HttpVerb};
    use clap::{CommandFactory, Parser};

    #[test]
    fn verify_cli_definition() {
        Args::command().debug_assert();
    }

    #[test]
    fn all_provided_all_valid() {
        let args = Args::try_parse_from([
            "logriff",
            "-s",
            "100",
            "-e",
            "599",
            "-m",
            "GET",
            "-p",
            "/foo/bar",
            "--min-duration",
            "0.1",
            "-n",
            "3",
            "-f",
            "json",
            "foo.log",
        ])
        .unwrap();
        assert_eq!(args.status_min, Some(100));
        assert_eq!(args.status_max, Some(599));
        assert_eq!(args.method, Some(HttpVerb::Get));
        assert_eq!(args.path_prefix, Some("/foo/bar".to_string()));
        assert_eq!(args.min_duration, Some(0.1));
        assert_eq!(args.top_routes, 3);
        assert_eq!(args.format, Format::Json);
        assert_eq!(args.file, "foo.log".to_string());
    }

    #[test]
    fn non_int_status_min_short() {
        let err = Args::try_parse_from(["logriff", "-s", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn non_int_status_min_long() {
        let err = Args::try_parse_from(["logriff", "--status-min", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn out_of_range_low_status_min() {
        let err = Args::try_parse_from(["logriff", "-s", "99"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn out_of_range_high_status_min() {
        let err = Args::try_parse_from(["logriff", "-s", "600"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn non_int_status_max_short() {
        let err = Args::try_parse_from(["logriff", "-e", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn non_int_status_max_long() {
        let err = Args::try_parse_from(["logriff", "--status-max", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn out_of_range_low_status_max() {
        let err = Args::try_parse_from(["logriff", "-e", "99"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn out_of_range_high_status_max() {
        let err = Args::try_parse_from(["logriff", "-e", "600"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn non_int_top_routes_short() {
        let err = Args::try_parse_from(["logriff", "-n", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn non_int_top_routes_long() {
        let err = Args::try_parse_from(["logriff", "--top-routes", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn negative_top_routes() {
        let err = Args::try_parse_from(["logriff", "-n=-1"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn invalid_method_short() {
        let err = Args::try_parse_from(["logriff", "-m", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::InvalidValue)
    }

    #[test]
    fn invalid_method_long() {
        let err = Args::try_parse_from(["logriff", "--method", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::InvalidValue)
    }

    #[test]
    fn method_case_insensitive() {
        let args = Args::try_parse_from(["logriff", "-m", "GET"]).unwrap();
        assert_eq!(args.method, Some(HttpVerb::Get))
    }

    #[test]
    fn non_numeric_min_duration_long() {
        let err = Args::try_parse_from(["logriff", "--min-duration", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn negative_min_duration() {
        let err = Args::try_parse_from(["logriff", "--min-duration=-1.1"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn infinite_min_duration() {
        let err = Args::try_parse_from(["logriff", "--min-duration", "inf"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn nan_min_duration() {
        let err = Args::try_parse_from(["logriff", "--min-duration", "nan"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::ValueValidation)
    }

    #[test]
    fn invalid_format_short() {
        let err = Args::try_parse_from(["logriff", "-f", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::InvalidValue)
    }

    #[test]
    fn invalid_format_long() {
        let err = Args::try_parse_from(["logriff", "--format", "foo"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::InvalidValue)
    }

    #[test]
    fn strict_flag_handling() {
        let args = Args::try_parse_from(["logriff", "--strict"]).unwrap();
        assert!(args.strict)
    }

    #[test]
    fn short_flag_no_value() {
        let err = Args::try_parse_from(["logriff", "-s"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::InvalidValue)
    }

    #[test]
    fn long_flag_no_value() {
        let err = Args::try_parse_from(["logriff", "--status-min"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::InvalidValue)
    }

    #[test]
    fn multiple_positionals() {
        let err = Args::try_parse_from(["logriff", "pos1", "pos2"]).unwrap_err();
        assert_eq!(err.kind(), clap::error::ErrorKind::UnknownArgument)
    }
}
