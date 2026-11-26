use std::collections::BTreeMap;

use crate::types::{Latency, Output, RouteCount, Stats, Summary, Traffic};

pub fn aggregate(stats: &Stats) -> Result<Output, String> {
    // Calculate latency values
    let Some(min) = stats.response_times.iter().copied().reduce(f32::min) else {
        return Err("could not get min latency".into());
    };
    let Some(max) = stats.response_times.iter().copied().reduce(f32::max) else {
        return Err("could not get max latency".into());
    };
    let Some(mean) = (if stats.response_times.is_empty() {
        None
    } else {
        let sum: f64 = stats.response_times.iter().map(|&x| x as f64).sum();
        Some((sum / stats.response_times.len() as f64) as f32)
    }) else {
        return Err("could not calculate mean latency".into());
    };
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

    Ok(Output {
        summary: Summary {
            total_lines: stats.lines,
            matched_lines: stats.matched,
            error_lines: stats.error,
        },
        traffic: Traffic {
            total_bytes: stats.bytes_volume,
        },
        latency: Latency {
            min,
            mean,
            p95,
            max,
        },
        status_codes,
        top_routes: Vec::new(),
    })
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
