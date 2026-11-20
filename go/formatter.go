package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func (o output) toJSON(n int) string {
	jsonTemplate := []string{
		`{`,
		`"summary": {`,
		`"total_lines": %v,`,
		`"matched_lines": %v,`,
		`"error_lines": %v`,
		`},`,
		`"traffic": {`,
		`"total_bytes": %v`,
		`},`,
		`"latency_ms": {`,
		`"min": %.2f,`,
		`"mean": %.2f,`,
		`"p95": %.2f,`,
		`"max": %.2f`,
		`},`,
		`"status_codes": {`,
		`%s`,
		`},`,
		`"top_routes": [`,
		`%s`,
		`]`,
		`}`,
	}

	var b strings.Builder

	for _, s := range jsonTemplate {
		b.WriteString(s)
	}

	sBuf := make([]string, 0, len(o.statusCodes))
	for _, s := range o.statusCodes {
		sBuf = append(sBuf, fmt.Sprintf(`"%v": %v`, s.status, s.count))
	}
	sc := strings.Join(sBuf, ",")

	rBuf := make([]string, 0, min(n, len(o.topRoutes)))
	for i, r := range o.topRoutes {
		if i == n {
			break
		}
		rBuf = append(rBuf, fmt.Sprintf(`{"path": "%s", "count": %v}`, r.path, r.count))
	}

	tr := strings.Join(rBuf, ",")

	return fmt.Sprintf(
		b.String(),
		o.summary.total,
		o.summary.matched,
		o.summary.error,
		o.traffic,
		o.latency.min,
		o.latency.mean,
		o.latency.p95,
		o.latency.max,
		sc,
		tr,
	)
}

func (o output) toText(n int) string {
	textTemplate := []string{
		`=== LogRiff Summary ===`,
		`Lines Processed : %s`,
		`Matched Records : %s`,
		`Skipped (Errors): %s`,
		``,
		`--- Traffic & Bandwidth ---`,
		`Total Volume    : %.2f GB (%s bytes)`,
		``,
		`--- Latency (ms) ---`,
		`Min             : %.2f ms`,
		`Mean            : %.2f ms`,
		`p95             : %.2f ms`,
		`Max             : %.2f ms`,
		``,
		`--- Status Codes ---`,
		`%s`,
		``,
		`--- Top Routes (Top %v) ---`,
		`%s`,
		`========================`,
	}

	sBuf := make([]string, 0, len(o.statusCodes))
	for _, s := range o.statusCodes {
		sBuf = append(sBuf, fmt.Sprintf(`%v : %7s (%5.2f%%)`, s.status, comma(s.count), ((float64(s.count)/float64(o.summary.matched))*100.0)))
	}
	sc := strings.Join(sBuf, "\n")

	rBuf := make([]string, 0, min(n, len(o.topRoutes)))
	for i, r := range o.topRoutes {
		if i == n {
			break
		}
		rBuf = append(rBuf, fmt.Sprintf(`%v. %-33s : %s`, i+1, r.path, comma(r.count)))
	}

	tr := strings.Join(rBuf, "\n")

	return fmt.Sprintf(
		strings.Join(textTemplate, "\n"),
		comma(o.summary.total),
		comma(o.summary.matched),
		comma(o.summary.error),
		float64(o.traffic)/math.Pow(2, 30),
		comma(o.traffic),
		o.latency.min,
		o.latency.mean,
		o.latency.p95,
		o.latency.max,
		sc,
		n,
		tr,
	)
}

func comma(i int) string {
	s := strconv.Itoa(i)
	o := make([]string, 0)
	for x := 0; x < len(s); x++ {
		if x != 0 && x%3 == 0 {
			c := []string{","}
			o = append(c, o...)
		}
		n := []string{s[len(s)-x-1 : len(s)-x]}
		o = append(n, o...)
	}
	return strings.Join(o, "")
}

func writeOutput(out StdoutProvider, format string, output output, numRoutes int) error {
	switch format {
	case "json":
		_, err := fmt.Fprintln(out.Stdout(), output.toJSON(numRoutes))
		if err != nil {
			return err
		}
	case "text":
		_, err := fmt.Fprintln(out.Stdout(), output.toText(numRoutes))
		if err != nil {
			return err
		}
	}
	return nil
}
