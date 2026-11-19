package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func usage(l *log.Logger) {
	l.Println("usage: logriff [-h] [-s STATUS_MIN] [-e STATUS_MAX] [-m {CONNECT,DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT,TRACE}] [-p PATH_PREFIX] [--min-duration MIN_DURATION] [-n TOP_ROUTES] [-f {text,json}] [--strict] [file]")
	l.Println("")
	l.Println("Filter and aggregate logs.")
	l.Println("")
	l.Println("positional arguments:")
	l.Println("  file                  input file, or - for stdin")
	l.Println("")
	l.Println("options:")
	l.Println("  -h, --help            show this help message and exit")
	l.Println("  -s, --status-min STATUS_MIN")
	l.Println("                        keep records with status >= MIN")
	l.Println("  -e, --status-max STATUS_MAX")
	l.Println("                        keep records with status <= MAX")
	l.Println("  -m, --method {CONNECT,DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT,TRACE}")
	l.Println("                        filter by HTTP verb (case-insensitive)")
	l.Println("  -p, --path-prefix PATH_PREFIX")
	l.Println("                        keep paths starting with this prefix")
	l.Println("  --min-duration MIN_DURATION")
	l.Println("                        keep records with response_time_ms >= duration")
	l.Println("  -n, --top-routes TOP_ROUTES")
	l.Println("                        number of top routes to show")
	l.Println("  -f, --format {text,json}")
	l.Println("                        output format")
	l.Println("  --strict              exit non-zero on malformed lines")
}

func validateArgs(a args) (args, error) {
	if a.statusMin != 0 && (a.statusMin < 100 || a.statusMin > 599) {
		return a, fmt.Errorf("Invalid HTTP Status in status-min: %v", a.statusMin)
	}
	if a.statusMax != 0 && (a.statusMax < 100 || a.statusMax > 599) {
		return a, fmt.Errorf("Invalid HTTP Status in status-max: %v", a.statusMax)
	}
	if a.statusMin != 0 && a.statusMax != 0 && a.statusMin > a.statusMax {
		return a, fmt.Errorf("Filter for status-max cannot be greater than status-min: %v > %v", a.statusMin, a.statusMax)
	}
	if a.method != "" && !httpMethods[a.method] {
		return a, fmt.Errorf("Invalid HTTP Method in method filter: %s", a.method)
	}
	if a.minDuration < 0 {
		return a, fmt.Errorf("Filter for min-duration cannot be less than 0: %v", a.minDuration)
	}
	if a.topRoutes < 0 {
		return a, fmt.Errorf("Number of top-routes must be a positive non-zero integer: %v", a.topRoutes)
	}
	if a.format != "" && a.format != "text" && a.format != "json" {
		return a, fmt.Errorf("Invalid output format: %s", a.format)
	}
	return a, nil
}

func parseArgs(argv []string) (args, error, bool) {
	var positional []string
	out := args{
		topRoutes: -1,
	}

	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "-h", "--help":
			return args{}, nil, true
		case "-s", "--status-min":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			v, err := strconv.Atoi(argv[i])
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", argv[i-1], argv[i]), false
			}
			out.statusMin = v
		case "-e", "--status-max":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			v, err := strconv.Atoi(argv[i])
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", argv[i-1], argv[i]), false
			}
			out.statusMax = v
		case "-m", "--method":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			out.method = strings.ToUpper(argv[i])
		case "-p", "--path-prefix":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			out.path = argv[i]
		case "--min-duration":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			v, err := strconv.ParseFloat(argv[i], 64)
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", argv[i-1], argv[i]), false
			}
			out.minDuration = v
		case "-n", "--top-routes":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			v, err := strconv.Atoi(argv[i])
			if err != nil {
				return args{}, fmt.Errorf("invalid %s: %q", argv[i-1], argv[i]), false
			}
			out.topRoutes = v
		case "-f", "--format":
			if i+1 >= len(argv) {
				return args{}, fmt.Errorf("%s requires a value", argv[i]), false
			}
			i++
			out.format = argv[i]
		case "--strict":
			out.strict = true
		case "-", "":
			positional = append(positional, argv[i])
		default:
			if strings.HasPrefix(argv[i], "-") {
				return args{}, fmt.Errorf("unknown option: %s", argv[i]), false
			}
			positional = append(positional, argv[i])
		}
	}

	if len(positional) > 1 {
		return args{}, fmt.Errorf("unexpected arguments: %v", positional[1:]), false
	} else if len(positional) == 0 {
		out.file = "-"
	} else {
		out.file = positional[0]
	}

	// Apply defaults
	if out.topRoutes == -1 {
		out.topRoutes = 5
	}
	if out.format == "" {
		out.format = "text"
	}

	out, err := validateArgs(out)
	if err != nil {
		return out, err, false
	}

	return out, nil, false
}
