use crate::matcher::match_line;
use crate::parser::{parse_line, validate_line};
use crate::types::{ParseError, Stats};
use std::fs::File;
use std::io::stdin;
use std::io::BufRead;
use std::io::BufReader;

pub fn generate_stats(path: &str, strict: &bool) -> Result<Stats, ParseError> {
    let mut reader: Box<dyn BufRead> = match path {
        "-" => Box::new(BufReader::new(stdin())),
        file_path => Box::new(BufReader::new(File::open(file_path)?)),
    };

    let mut stats: Stats = Stats::new();
    let mut line: String = String::new();
    loop {
        line.clear();
        let bytes_read = reader.read_line(&mut line)?;
        if bytes_read == 0 {
            break;
        }
        stats.lines += 1;
        line = line.trim().to_string();
        let parsed_raw = match parse_line(&stats.lines, &line) {
            Ok(parsed_raw) => parsed_raw,
            Err(err) => {
                if *strict {
                    return Err(ParseError::Strict {
                        line: stats.lines,
                        message: format!("{err}"),
                    });
                }
                stats.error += 1;
                continue;
            }
        };
        let parsed = match validate_line(&parsed_raw) {
            Ok(parsed) => parsed,
            Err(err) => {
                if *strict {
                    return Err(ParseError::Strict {
                        line: stats.lines,
                        message: err.to_string(),
                    });
                }
                stats.error += 1;
                continue;
            }
        };
        if !match_line(&parsed) {
            continue;
        }
        stats.matched += 1;
        stats.bytes_volume += parsed.bytes_sent;
        stats.response_times.push(parsed.response_time_ms);
        *stats
            .status_codes
            .entry(parsed.status.to_owned())
            .or_default() += 1;
        *stats
            .route_distribution
            .entry(parsed.path.to_owned())
            .or_default() += 1
    }

    Ok(stats)
}
