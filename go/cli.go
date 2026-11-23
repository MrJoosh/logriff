package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

var errHelp = errors.New("help requested")

const usageText = `usage: logriff [-h] [-s STATUS_MIN] [-e STATUS_MAX] [-m {CONNECT,DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT,TRACE}] [-p PATH_PREFIX] [--min-duration MIN_DURATION] [-n TOP_ROUTES] [-f {text,json}] [--strict] [file]

Filter and aggregate logs.

positional arguments:
  file                  input file, or - for stdin

options:
  -h, --help            show this help message and exit
  -s, --status-min STATUS_MIN
                        keep records with status >= MIN
  -e, --status-max STATUS_MAX
                        keep records with status <= MAX
  -m, --method {CONNECT,DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT,TRACE}
                        filter by HTTP verb (case-insensitive)
  -p, --path-prefix PATH_PREFIX
                        keep paths starting with this prefix
  --min-duration MIN_DURATION
                        keep records with response_time_ms >= duration
  -n, --top-routes TOP_ROUTES
                        number of top routes to show
  -f, --format {text,json}
                        output format
  --strict              exit non-zero on malformed lines
`

func usage(w io.Writer) {
	fmt.Fprint(w, usageText)
}

func validateArgs(a args) error {
	if a.statusMin != nil && (*a.statusMin < 100 || *a.statusMin > 599) {
		return fmt.Errorf("invalid status-min: %v", *a.statusMin)
	}
	if a.statusMax != nil && (*a.statusMax < 100 || *a.statusMax > 599) {
		return fmt.Errorf("invalid status-max: %v", *a.statusMax)
	}
	if a.statusMin != nil && a.statusMax != nil && *a.statusMin > *a.statusMax {
		return fmt.Errorf("filter for status-min cannot be greater than status-max: %v > %v", *a.statusMin, *a.statusMax)
	}
	if _, ok := httpMethods[a.method]; a.method != "" && !ok {
		return fmt.Errorf("invalid http method: %s", a.method)
	}
	if a.minDuration != nil && *a.minDuration < 0 {
		return fmt.Errorf("filter for min-duration cannot be less than 0: %v", *a.minDuration)
	}
	if a.minDuration != nil && (math.IsNaN(*a.minDuration) || math.IsInf(*a.minDuration, 1)) {
		return fmt.Errorf("invalid min-duration: %v", *a.minDuration)
	}
	if a.topRoutes != nil && *a.topRoutes < 0 {
		return fmt.Errorf("top-routes cannot be negative: %v", *a.topRoutes)
	}
	if a.format != "" && a.format != "text" && a.format != "json" {
		return fmt.Errorf("invalid output format: %s", a.format)
	}
	return nil
}

func parseArgs(argv []string) (args, error) {
	var positional []string
	out := args{}

	for i := 0; i < len(argv); i++ {
		token := argv[i]
		name, val, hasEq := strings.Cut(token, "=")

		switch name {
		case "-h", "--help":
			return args{}, errHelp
		case "-s", "--status-min":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", name, v)
			}
			out.statusMin = &n
		case "-e", "--status-max":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", name, v)
			}
			out.statusMax = &n
		case "-m", "--method":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			method := strings.ToUpper(v)
			out.method = method
		case "-p", "--path-prefix":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			out.path = v
		case "--min-duration":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", name, v)
			}
			out.minDuration = &n
		case "-n", "--top-routes":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", name, v)
			}
			out.topRoutes = &n
		case "-f", "--format":
			v, err := optValue(argv, &i, name, val, hasEq)
			if err != nil {
				return args{}, err
			}
			out.format = v
		case "--strict":
			if hasEq {
				return args{}, fmt.Errorf("%s does not take a value", name)
			}
			out.strict = true
		case "-":
			positional = append(positional, name)
		default:
			if strings.HasPrefix(token, "-") {
				return args{}, fmt.Errorf("unknown option: %s", name)
			}
			positional = append(positional, token)
		}
	}

	if len(positional) > 1 {
		return args{}, fmt.Errorf("unexpected arguments: %v", positional[1:])
	}
	if len(positional) == 0 {
		out.file = "-"
	} else {
		out.file = positional[0]
	}

	if err := validateArgs(out); err != nil {
		return args{}, err
	}

	// Apply defaults
	if out.statusMin == nil {
		out.statusMin = new(0)
	}
	if out.statusMax == nil {
		out.statusMax = new(0)
	}
	if out.minDuration == nil {
		out.minDuration = new(0.0)
	}
	if out.topRoutes == nil {
		out.topRoutes = new(5)
	}
	if out.format == "" {
		out.format = "text"
	}

	return out, nil
}

func optValue(argv []string, i *int, name, val string, hasEq bool) (string, error) {
	if hasEq {
		return val, nil
	}
	if *i+1 >= len(argv) {
		return "", fmt.Errorf("%s requires a value", name)
	}
	*i++
	return argv[*i], nil
}
