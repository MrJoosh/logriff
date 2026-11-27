use std::collections::BTreeMap;

use crate::types::{Latency, Output, RouteCount, Stats, Summary, Traffic};

pub fn aggregate(stats: &Stats, num_routes: &usize) -> Output {
    // Calculate latency values
    let min = stats
        .response_times
        .iter()
        .copied()
        .reduce(f32::min)
        .unwrap_or(0.0);
    let max = stats
        .response_times
        .iter()
        .copied()
        .reduce(f32::max)
        .unwrap_or(0.0);
    let mean = mean(&stats.response_times);
    let p95 = p95(&stats.response_times);

    // Sort status and routes objects
    let status_codes: BTreeMap<String, usize> = stats
        .status_codes
        .iter()
        .map(|(k, v)| (k.to_string(), *v))
        .collect();

    let mut top_routes: Vec<RouteCount> = stats
        .route_distribution
        .iter()
        .map(|(p, &c)| RouteCount {
            path: p.into(),
            count: c,
        })
        .collect();
    top_routes.sort_unstable_by(|a, b| {
        if a.count == b.count {
            a.path.cmp(&b.path)
        } else {
            b.count.cmp(&a.count)
        }
    });
    top_routes.truncate(*num_routes);

    Output {
        summary: Summary {
            total_lines: stats.lines,
            matched_lines: stats.matched,
            error_lines: stats.error,
        },
        traffic: Traffic {
            total_bytes: stats.bytes_volume,
        },
        latency_ms: Latency {
            min,
            mean,
            p95,
            max,
        },
        status_codes,
        top_routes,
    }
}

fn p95(v: &[f32]) -> f32 {
    if v.is_empty() {
        return 0.0;
    }
    if v.iter().any(|x| x.is_nan()) {
        return f32::NAN;
    }
    if v.len() == 1 {
        return v[0];
    }

    let r = 0.95 * (v.len() - 1) as f64;
    let k = r as usize;
    let f = (r - k as f64) as f32;

    let mut nums = v.to_vec();
    nums.sort_unstable_by(f32::total_cmp);

    if f == 0.0 || k + 1 >= nums.len() || nums[k] == nums[k + 1] {
        return nums[k];
    }

    let delta = nums[k + 1] - nums[k];
    if delta.is_infinite() {
        (1.0 - f) * nums[k] + f * nums[k + 1]
    } else {
        nums[k] + f * delta
    }
}

fn mean(v: &[f32]) -> f32 {
    if v.is_empty() {
        return 0.0;
    }
    if v.len() == 1 {
        return v[0];
    }
    let sum: f64 = v.iter().map(|&x| x as f64).sum();
    (sum / v.len() as f64) as f32
}

#[cfg(test)]
mod test_aggregate {
    use std::collections::{BTreeMap, HashMap};

    use crate::types::{Latency, Output, RouteCount, Stats, Summary, Traffic};

    use super::aggregate;

    #[test]
    fn valid_input_aggregates() {
        let stats = Stats {
            lines: 5,
            matched: 5,
            error: 0,
            bytes_volume: 250,
            response_times: vec![1.2, 1.6, 1.4, 1.5, 1.3],
            status_codes: HashMap::from([(200, 3), (401, 1), (500, 1)]),
            route_distribution: HashMap::from([
                ("/api/health".into(), 3),
                ("/assets/foo.css".into(), 1),
                ("/index.html".into(), 1),
            ]),
        };
        let want = Output {
            summary: Summary {
                total_lines: stats.lines,
                matched_lines: stats.matched,
                error_lines: stats.error,
            },
            traffic: Traffic {
                total_bytes: stats.bytes_volume,
            },
            latency_ms: Latency {
                min: 1.2,
                mean: 1.4,
                p95: 1.58,
                max: 1.6,
            },
            status_codes: BTreeMap::from([("200".into(), 3), ("401".into(), 1), ("500".into(), 1)]),
            top_routes: vec![
                RouteCount {
                    path: "/api/health".into(),
                    count: 3,
                },
                RouteCount {
                    path: "/assets/foo.css".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/index.html".into(),
                    count: 1,
                },
            ],
        };
        let got = aggregate(&stats, &5);
        assert_eq!(got, want)
    }

    #[test]
    fn empty_input_aggregates() {
        let stats = Stats {
            lines: 0,
            matched: 0,
            error: 0,
            bytes_volume: 0,
            response_times: vec![],
            status_codes: HashMap::from([]),
            route_distribution: HashMap::from([]),
        };
        let want = Output {
            summary: Summary {
                total_lines: stats.lines,
                matched_lines: stats.matched,
                error_lines: stats.error,
            },
            traffic: Traffic {
                total_bytes: stats.bytes_volume,
            },
            latency_ms: Latency {
                min: 0.0,
                mean: 0.0,
                p95: 0.0,
                max: 0.0,
            },
            status_codes: BTreeMap::from([]),
            top_routes: vec![],
        };
        let got = aggregate(&stats, &5);
        assert_eq!(got, want)
    }

    #[test]
    fn status_sorting() {
        let stats = Stats {
            lines: 5,
            matched: 5,
            error: 0,
            bytes_volume: 250,
            response_times: vec![1.2, 1.6, 1.4, 1.5, 1.3],
            status_codes: HashMap::from([(500, 1), (401, 1), (200, 3)]),
            route_distribution: HashMap::from([
                ("/api/health".into(), 3),
                ("/assets/foo.css".into(), 1),
                ("/index.html".into(), 1),
            ]),
        };
        let want = Output {
            summary: Summary {
                total_lines: stats.lines,
                matched_lines: stats.matched,
                error_lines: stats.error,
            },
            traffic: Traffic {
                total_bytes: stats.bytes_volume,
            },
            latency_ms: Latency {
                min: 1.2,
                mean: 1.4,
                p95: 1.58,
                max: 1.6,
            },
            status_codes: BTreeMap::from([("200".into(), 3), ("401".into(), 1), ("500".into(), 1)]),
            top_routes: vec![
                RouteCount {
                    path: "/api/health".into(),
                    count: 3,
                },
                RouteCount {
                    path: "/assets/foo.css".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/index.html".into(),
                    count: 1,
                },
            ],
        };
        let got = aggregate(&stats, &5);
        assert_eq!(got, want)
    }

    #[test]
    fn routes_sorting() {
        let stats = Stats {
            lines: 5,
            matched: 5,
            error: 0,
            bytes_volume: 250,
            response_times: vec![1.2, 1.6, 1.4, 1.5, 1.3],
            status_codes: HashMap::from([(200, 3), (401, 1), (500, 1)]),
            route_distribution: HashMap::from([
                ("/index.html".into(), 1),
                ("/assets/foo.css".into(), 1),
                ("/api/health".into(), 3),
            ]),
        };
        let want = Output {
            summary: Summary {
                total_lines: stats.lines,
                matched_lines: stats.matched,
                error_lines: stats.error,
            },
            traffic: Traffic {
                total_bytes: stats.bytes_volume,
            },
            latency_ms: Latency {
                min: 1.2,
                mean: 1.4,
                p95: 1.58,
                max: 1.6,
            },
            status_codes: BTreeMap::from([("200".into(), 3), ("401".into(), 1), ("500".into(), 1)]),
            top_routes: vec![
                RouteCount {
                    path: "/api/health".into(),
                    count: 3,
                },
                RouteCount {
                    path: "/assets/foo.css".into(),
                    count: 1,
                },
                RouteCount {
                    path: "/index.html".into(),
                    count: 1,
                },
            ],
        };
        let got = aggregate(&stats, &5);
        assert_eq!(got, want)
    }

    #[test]
    fn routes_limiting() {
        let stats = Stats {
            lines: 5,
            matched: 5,
            error: 0,
            bytes_volume: 250,
            response_times: vec![1.2, 1.6, 1.4, 1.5, 1.3],
            status_codes: HashMap::from([(200, 3), (401, 1), (500, 1)]),
            route_distribution: HashMap::from([
                ("/index.html".into(), 1),
                ("/assets/foo.css".into(), 1),
                ("/api/health".into(), 3),
            ]),
        };
        let want = Output {
            summary: Summary {
                total_lines: stats.lines,
                matched_lines: stats.matched,
                error_lines: stats.error,
            },
            traffic: Traffic {
                total_bytes: stats.bytes_volume,
            },
            latency_ms: Latency {
                min: 1.2,
                mean: 1.4,
                p95: 1.58,
                max: 1.6,
            },
            status_codes: BTreeMap::from([("200".into(), 3), ("401".into(), 1), ("500".into(), 1)]),
            top_routes: vec![RouteCount {
                path: "/api/health".into(),
                count: 3,
            }],
        };
        let got = aggregate(&stats, &1);
        assert_eq!(got, want)
    }
}

#[cfg(test)]
mod test_p95 {
    use super::p95;

    #[test]
    fn empty_input() {
        let v: Vec<f32> = vec![];
        let r = p95(&v);
        assert_eq!(r, 0.0)
    }

    #[test]
    fn single_element() {
        let v: Vec<f32> = vec![42.5];
        let r = p95(&v);
        assert_eq!(r, 42.5)
    }

    #[test]
    fn two_elements() {
        let v: Vec<f32> = vec![10.0, 20.0];
        let r = p95(&v);
        assert_eq!(r, 19.5)
    }

    #[test]
    fn exact_index() {
        let v: Vec<f32> = (0..=20).map(|x| x as f32).collect();
        let r = p95(&v);
        assert_eq!(r, 19.0)
    }

    #[test]
    fn interpolated_index() {
        let v: Vec<f32> = (1..=20).map(|x| x as f32).collect();
        let r = p95(&v);
        assert_eq!(r, 19.05)
    }

    #[test]
    fn dupes_at_boundary() {
        let v: Vec<f32> = vec![1.0, 2.0, 3.0, 50.0, 50.0];
        let r = p95(&v);
        assert_eq!(r, 50.0)
    }

    #[test]
    fn extreme_values() {
        let v: Vec<f32> = vec![f32::MIN, f32::MAX];
        let r = p95(&v);
        assert_eq!(r, 3.062541e38)
    }

    #[test]
    fn infinite_values() {
        let v: Vec<f32> = vec![1.0, f32::INFINITY];
        let r = p95(&v);
        assert_eq!(r, f32::INFINITY)
    }

    #[test]
    fn nan_values() {
        let v: Vec<f32> = vec![1.0, f32::NAN];
        let r = p95(&v);
        assert!(r.is_nan())
    }

    #[test]
    fn negative_values() {
        let v: Vec<f32> = vec![-1.0, -1.0];
        let r = p95(&v);
        assert_eq!(r, -1.0)
    }
}

#[cfg(test)]
mod test_mean {
    use super::mean;

    #[test]
    fn empty_input() {
        let v: Vec<f32> = vec![];
        let r = mean(&v);
        assert_eq!(r, 0.0);
    }

    #[test]
    fn single_element() {
        let v: Vec<f32> = vec![1.0];
        let r = mean(&v);
        assert_eq!(r, 1.0)
    }

    #[test]
    fn positive_floats() {
        let v: Vec<f32> = vec![10.0, 20.0, 30.0];
        let r = mean(&v);
        assert_eq!(r, 20.0)
    }

    #[test]
    fn negative_and_zero_values() {
        let v: Vec<f32> = vec![-10.0, 0.0, 10.0];
        let r = mean(&v);
        assert_eq!(r, 0.0)
    }

    #[test]
    fn non_terminating_decimal() {
        let v: Vec<f32> = vec![1.0, 2.0, 4.0];
        let r = mean(&v);
        assert_eq!(r, 7.0 / 3.0)
    }
}
