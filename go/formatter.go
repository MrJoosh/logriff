package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const bytesPerGiB = 1 << 30

func (o output) toJSON(n int) (string, error) {
	if len(o.TopRoutes) > n {
		o.TopRoutes = o.TopRoutes[:n]
	}
	s, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(s), nil
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

	sBuf := make([]string, 0, len(o.StatusCodes))
	for _, s := range o.StatusCodes {
		pct := 0.0
		if o.Summary.Matched != 0 {
			pct = ((float64(s.Count) / float64(o.Summary.Matched)) * 100.0)
		}
		sBuf = append(sBuf, fmt.Sprintf(`%v : %7s (%5.2f%%)`, s.Status, comma(s.Count), pct))
	}
	sc := strings.Join(sBuf, "\n")

	rBuf := make([]string, 0, min(n, len(o.TopRoutes)))
	for i, r := range o.TopRoutes {
		if i == n {
			break
		}
		rBuf = append(rBuf, fmt.Sprintf(`%v. %-33s : %s`, i+1, r.Path, comma(r.Count)))
	}

	tr := strings.Join(rBuf, "\n")

	return fmt.Sprintf(
		strings.Join(textTemplate, "\n"),
		comma(o.Summary.Total),
		comma(o.Summary.Matched),
		comma(o.Summary.Error),
		float64(o.Traffic.TotalBytes)/bytesPerGiB,
		comma(o.Traffic.TotalBytes),
		o.Latency.Min,
		o.Latency.Mean,
		o.Latency.P95,
		o.Latency.Max,
		sc,
		n,
		tr,
	)
}

func comma(i int) string {
	s := strconv.Itoa(i)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func writeOutput(out StdoutProvider, format string, output output, numRoutes int) error {
	switch format {
	case "json":
		o, err := output.toJSON(numRoutes)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out.Stdout(), o)
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
