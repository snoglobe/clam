package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var library = map[string]interface{}{}

func load(a string) interface{} {
	file, err := os.Open(a)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	var b interface{}
	err = json.Unmarshal([]byte(a), &b)
	if err != nil {
		return nil
	}
	return b
}

func save(a string, b interface{}) {
	file, _ := os.Create(a)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	err := file.Truncate(0)
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(data)
}

func buildFile(file *os.File) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"_file": file,
		"read": func() interface{} {
			var b []byte
			_, err := file.Read(b)
			if err != nil {
				return err
			}
			return b
		},
		"read_b": func(a int) interface{} {
			var b = make([]byte, a)
			_, err := file.Read(b)
			if err != nil {
				return err
			}
			return b
		},
		"seek": func(a int64, b int) interface{} {
			_, err := file.Seek(a, b)
			if err != nil {
				return err
			}
			return nil
		},
		"write": func(a []byte) interface{} {
			_, err := file.Write(a)
			if err != nil {
				return err
			}
			return nil
		},
		"write_str": func(a string) interface{} {
			_, err := file.WriteString(a)
			if err != nil {
				return err
			}
			return nil
		},
		"read_str": func() interface{} {
			var b []byte
			_, err := file.Read(b)
			if err != nil {
				return err
			}
			return string(b)
		},
		"close": func() interface{} {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func timeToMap(a time.Time) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"year":     a.Year(),
		"month":    a.Month(),
		"day":      a.Day(),
		"hour":     a.Hour(),
		"minute":   a.Minute(),
		"second":   a.Second(),
		"nsec":     a.Nanosecond(),
		"location": a.Location().String(),
	}
}

func convInt(a interface{}) int {
	switch a.(type) {
	case int:
		return a.(int)
	case int8:
		return int(a.(int8))
	case int16:
		return int(a.(int16))
	case int32:
		return int(a.(int32))
	case int64:
		return int(a.(int64))
	case float64:
		return int(a.(float64))
	default:
		return 0
	}
}

func convFloat(a interface{}) float64 {
	switch a.(type) {
	case int:
		return float64(a.(int))
	case float64:
		return a.(float64)
	default:
		return 0
	}
}

func connToMap(a net.Conn) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"_conn": a,
		"read": func(n int) interface{} {
			var b = make([]byte, n)
			_, err := a.Read(b)
			if err != nil {
				return err
			}
			return b
		},
		"read_all": func() interface{} {
			var b []byte
			_, err := io.ReadAll(a)
			if err != nil {
				return err
			}
			return b
		},
		"read_str": func() interface{} {
			var b []byte
			_, err := a.Read(b)
			if err != nil {
				return err
			}
			return string(b)
		},
		"write": func(b []byte) interface{} {
			_, err := a.Write(b)
			if err != nil {
				return err
			}
			return nil
		},
		"write_str": func(b string) interface{} {
			_, err := a.Write([]byte(b))
			if err != nil {
				return err
			}
			return nil
		},
		"close": func() interface{} {
			err := a.Close()
			if err != nil {
				return err
			}
			return nil
		},
		"local":  a.LocalAddr().String(),
		"remote": a.RemoteAddr().String(),
	}
}

func respToMap(a *http.Response) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"_response": a,
		"status":    a.Status,
		"header": func() map[interface{}]interface{} {
			var m = map[interface{}]interface{}{}
			for k, v := range a.Header {
				m[k] = v
			}
			return m
		}(),
		"body": func() interface{} {
			b, err := io.ReadAll(a.Body)
			if err != nil {
				return err
			}
			return string(b)
		}(),
		"close": func() interface{} {
			err := a.Body.Close()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func requestToMap(a *http.Request) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"_request": a,
		"method":   a.Method,
		"url":      a.URL.String(),
		"path":     a.URL.Path,
		"query":    a.URL.RawQuery,
		"fragment": a.URL.Fragment,
		"proto":    a.Proto,
		"host":     a.RemoteAddr,
		"form": func() map[interface{}]interface{} {
			err := a.ParseForm()
			if err != nil {
				return nil
			}
			var m = map[interface{}]interface{}{}
			for k, v := range a.Form {
				m[k] = v
			}
			return m
		},
		"header": func() map[interface{}]interface{} {
			var m = map[interface{}]interface{}{}
			for k, v := range a.Header {
				m[k] = v
			}
			return m
		}(),
		"body": func() interface{} {
			var b []byte
			_, err := a.Body.Read(b)
			if err != nil {
				return err
			}
			return string(b)
		}(),
		"close": func() interface{} {
			err := a.Body.Close()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func init() {
	library["print"] = fmt.Print
	doc_fn("print", ManyArgs("args"), "Prints the arguments to the standard output.", "nil")
	library["println"] = fmt.Println
	doc_fn("println", ManyArgs("args"), "Prints the arguments to the standard output, followed by a newline.", "nil")
	library["printf"] = fmt.Printf
	doc_fn("printf", ManyArgs("format", "args"), "Prints the formatted arguments to the standard output.", "nil")
	library["sprintf"] = fmt.Sprintf
	doc_fn("sprintf", ManyArgs("format", "args"), "Returns the formatted arguments as a string.", "string")
	library["len"] = func(a interface{}) int {
		switch a.(type) {
		case string:
			return len(a.(string))
		case []interface{}:
			return len(a.([]interface{}))
		case map[interface{}]interface{}:
			return len(a.(map[interface{}]interface{}))
		default:
			return 0
		}
	}
	doc_fn("len", ArgsOf("a"), "Returns the length of a string, array, or map.", "int")
	library["push"] = func(a []interface{}, b interface{}) []interface{} {
		return append(a, b)
	}
	doc_fn("push", ArgsOf("a", "b"), "Appends b to the end of a.", "array")
	library["pop"] = func(a []interface{}) (interface{}, []interface{}) {
		return a[len(a)-1], a[:len(a)-1]
	}
	doc_fn("pop", ArgsOf("a"), "Removes and returns the last element of a.", "value, array")
	library["shift"] = func(a []interface{}) (interface{}, []interface{}) {
		return a[0], a[1:]
	}
	doc_fn("shift", ArgsOf("a"), "Removes and returns the first element of a.", "value, array")
	library["unshift"] = func(a []interface{}, b interface{}) []interface{} {
		return append([]interface{}{b}, a...)
	}
	doc_fn("unshift", ArgsOf("a", "b"), "Prepends b to the beginning of a.", "array")
	library["join"] = func(a []interface{}, b string) string {
		var s string
		for i, v := range a {
			if i > 0 {
				s += b
			}
			s += fmt.Sprint(v)
		}
		return s
	}
	doc_fn("join", ArgsOf("a", "b"), "Concatenates the elements of a to create a single string, separated by b.", "string")
	library["split"] = func(a string, b string) []string {
		return strings.Split(a, b)
	}
	doc_fn("split", ArgsOf("a", "b"), "Splits a into substrings separated by b.", "array")
	library["index"] = func(a []interface{}, b interface{}) int {
		for i, v := range a {
			if v == b {
				return i
			}
		}
		return -1
	}
	doc_fn("index", ArgsOf("a", "b"), "Returns the index of b in a, or -1 if b is not present.", "int")
	library["slice"] = func(a []interface{}, b interface{}, c interface{}) []interface{} {
		return a[convInt(b):convInt(c)]
	}
	doc_fn("slice", ArgsOf("a", "b", "c"), "Returns a slice of a from b to c.", "array")
	library["map"] = func(a []interface{}, b func(interface{}) interface{}) []interface{} {
		var s []interface{}
		for _, v := range a {
			s = append(s, b(v))
		}
		return s
	}
	doc_fn("map", ArgsOf("a", "b"), "Applies b to each element of a and returns the results.", "array")
	library["filter"] = func(a []interface{}, b func(interface{}) interface{}) []interface{} {
		var s []interface{}
		for _, v := range a {
			if truthy(b(v)) {
				s = append(s, v)
			}
		}
		return s
	}
	doc_fn("filter", ArgsOf("a", "b"), "Returns a new array containing the elements of a for which b returns true.", "array")
	library["reduce"] = func(a []interface{}, b func(interface{}, interface{}) interface{}) interface{} {
		var s = a[0]
		for _, v := range a[1:] {
			s = b(s, v)
		}
		return s
	}
	doc_fn("reduce", ArgsOf("a", "b"), "Applies b to each element of a, accumulating the results.", "value")
	library["reverse"] = func(a []interface{}) []interface{} {
		var s []interface{}
		for i := len(a) - 1; i >= 0; i-- {
			s = append(s, a[i])
		}
		return s
	}
	doc_fn("reverse", ArgsOf("a"), "Returns a new array containing the elements of a in reverse order.", "array")

	library["file"] = map[interface{}]interface{}{
		"persist": func(a string, b interface{}) interface{} {
			// if a exists, load it and deserialize it, then return it
			// if a does not exist, save b to a and return b
			if _, err := os.Stat(a); err == nil {
				// exists
				return load(a)
			}
			// does not exist
			save(a, b)
			return b
		},
		"open": func(a string) interface{} {
			var file, err = os.Open(a)
			if err != nil {
				return err
			}
			return buildFile(file)
		},
		"create": func(a string) interface{} {
			var file, err = os.Create(a)
			if err != nil {
				return err
			}
			return buildFile(file)
		},
		"remove": func(a string) interface{} {
			err := os.Remove(a)
			if err != nil {
				return err
			}
			return nil
		},
		"rename": func(a string, b string) interface{} {
			err := os.Rename(a, b)
			if err != nil {
				return err
			}
			return nil
		},
		"stat": func(a string) interface{} {
			var info, err = os.Stat(a)
			if err != nil {
				return err
			}
			return map[interface{}]interface{}{
				"name": info.Name(),
				"size": info.Size(),
				"mode": map[interface{}]interface{}{
					"isdir":     info.Mode().IsDir(),
					"isregular": info.Mode().IsRegular(),
					"perm":      info.Mode().Perm(),
				},
				"modtime": timeToMap(info.ModTime()),
				"isdir":   info.IsDir(),
			}
		},
	}
	doc_obj("file_info",
		"File info and operations.",
		doc_fn_str("read", ArgsOf(), "Reads the file as an array of bytes.", "value"),
		doc_fn_str("read_b", ArgsOf("n"), "Reads n bytes from the file.", "value"),
		doc_fn_str("seek", ArgsOf("offset", "whence"), "Sets the position for the next read or write.", "nil"),
		doc_fn_str("write", ArgsOf("data"), "Writes data to the file.", "nil"),
		doc_fn_str("write_str", ArgsOf("data"), "Writes data to the file.", "nil"),
		doc_fn_str("read_str", ArgsOf(), "Reads the file as a string.", "value"),
		doc_fn_str("close", ArgsOf(), "Closes the file.", "nil"),
	)
	doc_obj("stat_info",
		"File info.",
		doc_fn_str("name", ArgsOf(), "The name of the file.", "string"),
		doc_fn_str("size", ArgsOf(), "The size of the file.", "int"),
		doc_obj_str("mode",
			"The mode of the file.",
			doc_str("isdir", "True if the file is a directory."),
			doc_str("isregular", "True if the file is a regular file."),
			doc_str("perm", "The permission bits of the file."),
		),
		doc_str("modtime", "The modification time of the file."),
		doc_str("isdir", "True if the file is a directory."),
	)
	doc_obj("file",
		"File I/O operations.",
		doc_fn_str("persist", ArgsOf("path", "data"), "Persists data to path if it does not exist, or loads and returns the data at path if it does.", "value"),
		doc_fn_str("open", ArgsOf("path"), "Opens the file at path for reading.", "file_info"),
		doc_fn_str("create", ArgsOf("path"), "Creates the file at path for writing.", "file_info"),
		doc_fn_str("remove", ArgsOf("path"), "Removes the file at path.", "nil"),
		doc_fn_str("rename", ArgsOf("old", "new"), "Renames the file at old to new.", "nil"),
		doc_fn_str("stat", ArgsOf("path"), "Returns information about the file at path.", "stat_info"),
	)

	library["time"] = map[interface{}]interface{}{
		"now": func() map[interface{}]interface{} {
			return timeToMap(time.Now())
		},
		"parse": func(a string, b string) map[interface{}]interface{} {
			var t, err = time.Parse(a, b)
			if err != nil {
				return nil
			}
			return timeToMap(t)
		},
		"format": func(a string, b map[interface{}]interface{}) string {
			var t = time.Date(
				convInt(b["year"]),
				time.Month(convInt(b["month"])),
				convInt(b["day"]),
				convInt(b["hour"]),
				convInt(b["minute"]),
				convInt(b["second"]),
				convInt(b["nsec"]),
				time.FixedZone("", 0),
			)
			return t.Format(a)
		},
		"str": func(a map[interface{}]interface{}) string {
			var t = time.Date(
				convInt(a["year"]),
				time.Month(convInt(a["month"])),
				convInt(a["day"]),
				convInt(a["hour"]),
				convInt(a["minute"]),
				convInt(a["second"]),
				convInt(a["nsec"]),
				time.FixedZone("", 0),
			)
			return t.String()
		},
		"from_unix": func(a interface{}) map[interface{}]interface{} {
			return timeToMap(time.Unix(int64(convInt(a)), 0))
		},
		"from": func(args ...interface{}) interface{} {
			switch len(args) {
			case 1:
				return timeToMap(time.Date(
					convInt(args[0]),
					time.January,
					1,
					0,
					0,
					0,
					0,
					time.FixedZone("", 0),
				))
			case 2:
				return timeToMap(time.Date(
					convInt(args[0]),
					time.Month(convInt(args[1])),
					1,
					0,
					0,
					0,
					0,
					time.FixedZone("", 0),
				))
			case 3:
				return timeToMap(time.Date(
					convInt(args[0]),
					time.Month(convInt(args[1])),
					convInt(args[2]),
					0,
					0,
					0,
					0,
					time.FixedZone("", 0),
				))
			case 6:
				return timeToMap(time.Date(
					convInt(args[0]),
					time.Month(convInt(args[1])),
					convInt(args[2]),
					convInt(args[3]),
					convInt(args[4]),
					convInt(args[5]),
					0,
					time.FixedZone("", 0),
				))
			case 7:
				return timeToMap(time.Date(
					convInt(args[0]),
					time.Month(convInt(args[1])),
					convInt(args[2]),
					convInt(args[3]),
					convInt(args[4]),
					convInt(args[5]),
					convInt(args[6]),
					time.FixedZone("", 0),
				))
			default:
				return fmt.Errorf("time.from(): invalid arguments")
			}
		},
		"diff": func(a map[interface{}]interface{}, b map[interface{}]interface{}) map[interface{}]interface{} {
			var t1 = time.Date(
				convInt(a["year"]),
				time.Month(convInt(a["month"])),
				convInt(a["day"]),
				convInt(a["hour"]),
				convInt(a["minute"]),
				convInt(a["second"]),
				convInt(a["nsec"]),
				time.FixedZone("", 0),
			)
			var t2 = time.Date(
				convInt(b["year"]),
				time.Month(convInt(b["month"])),
				convInt(b["day"]),
				convInt(b["hour"]),
				convInt(b["minute"]),
				convInt(b["second"]),
				convInt(b["nsec"]),
				time.FixedZone("", 0),
			)
			var d = t2.Sub(t1)
			return map[interface{}]interface{}{
				"hours":   int(d.Hours()),
				"minutes": int(d.Minutes()),
				"seconds": int(d.Seconds()),
				"mills":   int(d.Milliseconds()),
				"nsec":    int(d.Nanoseconds()),
			}
		},
		"January":   time.January,
		"February":  time.February,
		"March":     time.March,
		"April":     time.April,
		"May":       time.May,
		"June":      time.June,
		"July":      time.July,
		"August":    time.August,
		"September": time.September,
		"October":   time.October,
		"November":  time.November,
		"December":  time.December,
	}
	doc_obj("time_info",
		"Time info.",
		doc_str("year", "The year."),
		doc_str("month", "The month."),
		doc_str("day", "The day."),
		doc_str("hour", "The hour."),
		doc_str("minute", "The minute."),
		doc_str("second", "The second."),
		doc_str("nsec", "The nanosecond."),
		doc_str("location", "The location."),
	)
	doc_obj("time_diff",
		"Time difference.",
		doc_str("hours", "The difference in hours."),
		doc_str("minutes", "The difference in minutes."),
		doc_str("seconds", "The difference in seconds."),
		doc_str("mills", "The difference in milliseconds."),
		doc_str("nsec", "The difference in nanoseconds."),
	)
	doc_obj("time",
		"Time info and operations.",
		doc_fn_str("now", ArgsOf(), "Returns the current time.", "time_info"),
		doc_fn_str("parse", ArgsOf("layout", "value"), "Parses value using layout and returns the time.", "time_info"),
		doc_fn_str("format", ArgsOf("layout", "time"), "Formats time using layout.", "string"),
		doc_fn_str("str", ArgsOf("time"), "Returns a string representation of time.", "string"),
		doc_fn_str("from_unix", ArgsOf("unix"), "Returns the time from a Unix timestamp.", "time_info"),
		doc_fn_str("from", ManyArgs("year", "month", "day", "hour", "minute", "second", "nsec"), "Returns the time from the given year, month, day, hour, minute, second, and nanosecond. Granularity can be selected by omitting args.", "time_info"),
		doc_fn_str("diff", ArgsOf("a", "b"), "Returns the difference between a and b.", "time_diff"),
		doc_str("January", "The month of January."),
		doc_str("February", "The month of February."),
		doc_str("March", "The month of March."),
		doc_str("April", "The month of April."),
		doc_str("May", "The month of May."),
		doc_str("June", "The month of June."),
		doc_str("July", "The month of July."),
		doc_str("August", "The month of August."),
		doc_str("September", "The month of September."),
		doc_str("October", "The month of October."),
		doc_str("November", "The month of November."),
		doc_str("December", "The month of December."),
	)

	library["math"] = map[interface{}]interface{}{
		"abs":       math.Abs,
		"acos":      math.Acos,
		"acosh":     math.Acosh,
		"asin":      math.Asin,
		"asinh":     math.Asinh,
		"atan":      math.Atan,
		"atan2":     math.Atan2,
		"atanh":     math.Atanh,
		"cbrt":      math.Cbrt,
		"ceil":      math.Ceil,
		"copysign":  math.Copysign,
		"cos":       math.Cos,
		"cosh":      math.Cosh,
		"exp":       math.Exp,
		"exp2":      math.Exp2,
		"floor":     math.Floor,
		"gamma":     math.Gamma,
		"hypot":     math.Hypot,
		"inf":       math.Inf,
		"log":       math.Log,
		"log10":     math.Log10,
		"log2":      math.Log2,
		"max":       math.Max,
		"min":       math.Min,
		"mod":       math.Mod,
		"nan":       math.NaN,
		"pow":       math.Pow,
		"pow10":     math.Pow10,
		"remainder": math.Remainder,
		"round":     math.Round,
		"signbit":   math.Signbit,
		"sin":       math.Sin,
		"sinh":      math.Sinh,
		"sqrt":      math.Sqrt,
		"tan":       math.Tan,
		"tanh":      math.Tanh,
		"trunc":     math.Trunc,
		"epsilon":   math.SmallestNonzeroFloat64,
		"pi":        math.Pi,
		"e":         math.E,
	}
	doc_obj("math",
		"Mathematical functions and constants.",
		doc_fn_str("abs", ArgsOf("x"), "Returns the absolute value of x.", "float"),
		doc_fn_str("acos", ArgsOf("x"), "Returns the arccosine of x.", "float"),
		doc_fn_str("acosh", ArgsOf("x"), "Returns the hyperbolic arccosine of x.", "float"),
		doc_fn_str("asin", ArgsOf("x"), "Returns the arcsine of x.", "float"),
		doc_fn_str("asinh", ArgsOf("x"), "Returns the hyperbolic arcsine of x.", "float"),
		doc_fn_str("atan", ArgsOf("x"), "Returns the arctangent of x.", "float"),
		doc_fn_str("atan2", ArgsOf("y", "x"), "Returns the arctangent of y/x.", "float"),
		doc_fn_str("atanh", ArgsOf("x"), "Returns the hyperbolic arctangent of x.", "float"),
		doc_fn_str("cbrt", ArgsOf("x"), "Returns the cube root of x.", "float"),
		doc_fn_str("ceil", ArgsOf("x"), "Returns the smallest integer value greater than or equal to x.", "float"),
		doc_fn_str("copysign", ArgsOf("x", "y"), "Returns x with the sign of y.", "float"),
		doc_fn_str("cos", ArgsOf("x"), "Returns the cosine of x.", "float"),
		doc_fn_str("cosh", ArgsOf("x"), "Returns the hyperbolic cosine of x.", "float"),
		doc_fn_str("exp", ArgsOf("x"), "Returns e**x.", "float"),
		doc_fn_str("exp2", ArgsOf("x"), "Returns 2**x.", "float"),
		doc_fn_str("floor", ArgsOf("x"), "Returns the largest integer value less than or equal to x.", "float"),
		doc_fn_str("gamma", ArgsOf("x"), "Returns the gamma function of x.", "float"),
		doc_fn_str("hypot", ArgsOf("x", "y"), "Returns the square root of x**2 + y**2.", "float"),
		doc_fn_str("inf", ArgsOf("sign"), "Returns positive or negative infinity.", "float"),
		doc_fn_str("log", ArgsOf("x"), "Returns the natural logarithm of x.", "float"),
		doc_fn_str("log10", ArgsOf("x"), "Returns the base 10 logarithm of x.", "float"),
		doc_fn_str("log2", ArgsOf("x"), "Returns the base 2 logarithm of x.", "float"),
		doc_fn_str("max", ArgsOf("x", "y"), "Returns the larger of x or y.", "float"),
		doc_fn_str("min", ArgsOf("x", "y"), "Returns the smaller of x or y.", "float"),
		doc_fn_str("mod", ArgsOf("x", "y"), "Returns the floating-point remainder of x/y.", "float"),
		doc_fn_str("nan", ArgsOf("sign"), "Returns a quiet NaN.", "float"),
		doc_fn_str("pow", ArgsOf("x", "y"), "Returns x**y.", "float"),
		doc_fn_str("pow10", ArgsOf("n"), "Returns 10**n.", "float"),
		doc_fn_str("remainder", ArgsOf("x", "y"), "Returns the IEEE 754 floating-point remainder of x/y.", "float"),
		doc_fn_str("round", ArgsOf("x"), "Returns the nearest integer, rounding half away from zero.", "float"),
		doc_fn_str("signbit", ArgsOf("x"), "Reports whether x is negative.", "bool"),
		doc_fn_str("sin", ArgsOf("x"), "Returns the sine of x.", "float"),
		doc_fn_str("sinh", ArgsOf("x"), "Returns the hyperbolic sine of x.", "float"),
		doc_fn_str("sqrt", ArgsOf("x"), "Returns the square root of x.", "float"),
		doc_fn_str("tan", ArgsOf("x"), "Returns the tangent of x.", "float"),
		doc_fn_str("tanh", ArgsOf("x"), "Returns the hyperbolic tangent of x.", "float"),
		doc_fn_str("trunc", ArgsOf("x"), "Returns the integer value of x.", "float"),
		doc_str("epsilon", "The smallest positive number that can be represented as a float64."),
		doc_str("pi", "The ratio of the circumference of a circle to its diameter."),
		doc_str("e", "The base of the natural logarithm."),
	)
	library["net"] = map[interface{}]interface{}{
		"resolve": func(a string) interface{} {
			var ips, err = net.LookupIP(a)
			if err != nil {
				return err
			}
			var s []string
			for _, ip := range ips {
				s = append([]string{ip.String()}, s...)
			}
			return s
		},
		"lookup": func(a string) interface{} {
			var ips, err = net.LookupAddr(a)
			if err != nil {
				return err
			}
			return ips
		},
		"dial_tcp": func(a string, b string) interface{} {
			var conn, err = net.Dial("tcp", a+":"+b)
			if err != nil {
				return err
			}
			return connToMap(conn)
		},
		"dial_udp": func(a string, b string) interface{} {
			var conn, err = net.Dial("udp", a+":"+b)
			if err != nil {
				return err
			}
			return connToMap(conn)
		},
		"listen_tcp": func(a string) interface{} {
			var listener, err = net.Listen("tcp", a)
			if err != nil {
				return err
			}
			return map[interface{}]interface{}{
				"_listener": listener,
				"accept": func() interface{} {
					var conn, err = listener.Accept()
					if err != nil {
						return err
					}
					return connToMap(conn)
				},
				"close": func() interface{} {
					err := listener.Close()
					if err != nil {
						return err
					}
					return nil
				},
			}
		},
		"listen_udp": func(a string) interface{} {
			var listener, err = net.ListenPacket("udp", a)
			if err != nil {
				return err
			}
			return map[interface{}]interface{}{
				"_listener": listener,
				"read": func(b []byte) interface{} {
					n, addr, err := listener.ReadFrom(b)
					if err != nil {
						return err
					}
					return map[interface{}]interface{}{
						"n":    n,
						"addr": addr.String(),
					}
				},
				"write": func(b []byte) interface{} {
					_, err := listener.WriteTo(b, listener.LocalAddr())
					if err != nil {
						return err
					}
					return nil
				},
				"write_str": func(b string) interface{} {
					_, err := listener.WriteTo([]byte(b), listener.LocalAddr())
					if err != nil {
						return err
					}
					return nil
				},
				"close": func() interface{} {
					err := listener.Close()
					if err != nil {
						return err
					}
					return nil
				},
			}
		},
	}
	doc_obj("conn",
		"Network connection.",
		doc_fn_str("read", ArgsOf("n"), "Reads n bytes from the connection.", "nil"),
		doc_fn_str("read_all", ArgsOf(), "Reads all available bytes from the connection.", "byte array"),
		doc_fn_str("read_str", ArgsOf(), "Reads all available bytes from the connection as a string.", "string"),
		doc_fn_str("write", ArgsOf("data"), "Writes the byte array data to the connection.", "nil"),
		doc_fn_str("write_str", ArgsOf("data"), "Writes the string data to the connection.", "nil"),
		doc_fn_str("close", ArgsOf(), "Closes the connection.", "nil"),
		doc_str("local", "The local address of the connection."),
		doc_str("remote", "The remote address of the connection."),
	)
	doc_obj("tcp_listener",
		"TCP network listener.",
		doc_fn_str("accept", ArgsOf(), "Accepts a connection.", "conn"),
		doc_fn_str("close", ArgsOf(), "Closes the listener.", "nil"),
	)
	doc_obj("udp_listener",
		"UDP network listener.",
		doc_fn_str("read", ArgsOf("n"), "Reads n bytes from the port.", "value"),
		doc_fn_str("write", ArgsOf("data"), "Writes data to the port.", "nil"),
		doc_fn_str("write_str", ArgsOf("data"), "Writes data to the port.", "nil"),
		doc_fn_str("close", ArgsOf(), "Closes the listener.", "nil"),
	)
	doc_obj("net",
		"Network operations.",
		doc_fn_str("resolve", ArgsOf("host"), "Resolves the IP addresses of a host.", "array"),
		doc_fn_str("lookup", ArgsOf("ip"), "Looks up the hostnames of an IP address.", "array"),
		doc_fn_str("dial_tcp", ArgsOf("host", "port"), "Dials a TCP connection to a host and port.", "conn"),
		doc_fn_str("dial_udp", ArgsOf("host", "port"), "Dials a UDP connection to a host and port.", "conn"),
		doc_fn_str("listen_tcp", ArgsOf("port"), "Listens for TCP connections on a port.", "tcp_listener"),
		doc_fn_str("listen_udp", ArgsOf("port"), "Listens for UDP connections on a port.", "udp_listener"),
	)
	library["http"] = map[interface{}]interface{}{
		"get": func(a string) interface{} {
			var client = &http.Client{}
			var resp, err = client.Get(a)
			if err != nil {
				return err
			}
			return respToMap(resp)
		},
		"post": func(a string, b string, c string) interface{} {
			var client = &http.Client{}
			var resp, err = client.Post(a, b, strings.NewReader(c))
			if err != nil {
				return err
			}
			return respToMap(resp)
		},
		"head": func(a string) interface{} {
			var client = &http.Client{}
			var resp, err = client.Head(a)
			if err != nil {
				return err
			}
			return respToMap(resp)
		},
		"new_request": func(a string, b string, c string) interface{} {
			var req, err = http.NewRequest(a, b, strings.NewReader(c))
			if err != nil {
				return err
			}
			return map[interface{}]interface{}{
				"_request": req,
				"header": func() map[interface{}]interface{} {
					var m = map[interface{}]interface{}{}
					for k, v := range req.Header {
						m[k] = v
					}
					return m
				}(),
				"body": func() interface{} {
					var b []byte
					_, err := req.Body.Read(b)
					if err != nil {
						return err
					}
					return string(b)
				}(),
				"close": func() interface{} {
					err := req.Body.Close()
					if err != nil {
						return err
					}
					return nil
				},
			}
		},
		"do": func(a map[interface{}]interface{}) interface{} {
			var client = &http.Client{}
			var req = a["_request"].(*http.Request)
			var resp, err = client.Do(req)
			if err != nil {
				return err
			}
			return respToMap(resp)
		},
		"server": func(a string, b func(...interface{}) interface{}) interface{} {
			return http.ListenAndServe(a, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var m = requestToMap(r)
				var resp = b(m).(map[interface{}]interface{})
				w.WriteHeader(convInt(resp["status"]))
				for k, v := range resp["header"].(map[interface{}]interface{}) {
					w.Header().Add(k.(string), v.(string))
				}
				_, _ = w.Write([]byte(resp["body"].(string)))
			}))
		},
		"response": func(a int, b map[interface{}]interface{}, c string) map[interface{}]interface{} {
			return map[interface{}]interface{}{
				"status": a,
				"header": b,
				"body":   c,
			}
		},
	}
	doc_obj("http_request",
		"HTTP request.",
		doc_str("header", "The request header."),
		doc_str("body", "The request body."),
		doc_fn_str("close", ArgsOf(), "Closes the request body.", "nil"),
	)
	doc_obj("http_response",
		"HTTP response.",
		doc_str("status", "The response status code."),
		doc_str("header", "The response header."),
		doc_str("body", "The response body."),
	)
	doc_obj("http",
		"HTTP operations.",
		doc_fn_str("get", ArgsOf("url"), "Performs an HTTP GET request.", "http_response"),
		doc_fn_str("post", ArgsOf("url", "type", "data"), "Performs an HTTP POST request.", "http_response"),
		doc_fn_str("head", ArgsOf("url"), "Performs an HTTP HEAD request.", "http_response"),
		doc_fn_str("new_request", ArgsOf("method", "url", "data"), "Creates a new HTTP request.", "http_request"),
		doc_fn_str("do", ArgsOf("request"), "Performs an HTTP request.", "http_response"),
		doc_fn_str("server", ArgsOf("addr", "handler"), "Starts an HTTP server.", "nil"),
		doc_fn_str("response", ArgsOf("status", "header", "body"), "Creates an HTTP response.", "http_response"),
	)
	library["json"] = map[interface{}]interface{}{
		"from": func(a string) interface{} {
			var b interface{}
			err := json.Unmarshal([]byte(a), &b)
			if err != nil {
				return err
			}
			return b
		},
		"to": func(a interface{}) string {
			b, err := json.Marshal(a)
			if err != nil {
				return ""
			}
			return string(b)
		},
		"valid": func(a string) bool {
			return json.Valid([]byte(a))
		},
	}
	doc_obj("json",
		"JSON operations.",
		doc_fn_str("from", ArgsOf("string"), "Parses a JSON string.", "value"),
		doc_fn_str("to", ArgsOf("value"), "Serializes a value to a JSON string.", "string"),
		doc_fn_str("valid", ArgsOf("string"), "Reports whether a string is a valid JSON.", "bool"),
	)
	library["os"] = map[interface{}]interface{}{
		"args": func() []string {
			return os.Args
		},
		"env": func(a string) string {
			return os.Getenv(a)
		},
		"setenv": func(a string, b string) {
			err := os.Setenv(a, b)
			if err != nil {
				return
			}
		},
		"unsetenv": func(a string) {
			err := os.Unsetenv(a)
			if err != nil {
				return
			}
		},
		"getwd": func() string {
			s, err := os.Getwd()
			if err != nil {
				return ""
			}
			return s
		},
		"chdir": func(a string) {
			err := os.Chdir(a)
			if err != nil {
				return
			}
		},
		"mkdir": func(a string) {
			err := os.Mkdir(a, 0755)
			if err != nil {
				return
			}
		},
		"mkdir_all": func(a string) {
			err := os.MkdirAll(a, 0755)
			if err != nil {
				return
			}
		},
		"cp": func(a string, b string) {
			src, err := os.Open(a)
			if err != nil {
				return
			}
			defer func(src *os.File) {
				err := src.Close()
				if err != nil {
					return
				}
			}(src)
			dst, err := os.Create(b)
			if err != nil {
				return
			}
			defer func(dst *os.File) {
				err := dst.Close()
				if err != nil {
					return
				}
			}(dst)
			_, err = io.Copy(dst, src)
			if err != nil {
				return
			}
		},
		"mv": func(a string, b string) {
			err := os.Rename(a, b)
			if err != nil {
				return
			}
		},
		"system": func(a string) interface{} {
			err := exec.Command(strings.Split(a, " ")[0], strings.Split(a, " ")[1:]...).Run()
			if err != nil {
				return err
			}
			return nil
		},
		"os": runtime.GOOS,
		"exit": func(a int) {
			os.Exit(a)
		},
	}
	doc_obj("os",
		"Operating system operations.",
		doc_fn_str("args", ArgsOf(), "Returns the command-line arguments.", "array"),
		doc_fn_str("env", ArgsOf("key"), "Returns the value of an environment variable.", "string"),
		doc_fn_str("setenv", ArgsOf("key", "value"), "Sets the value of an environment variable.", "nil"),
		doc_fn_str("unsetenv", ArgsOf("key"), "Unsets an environment variable.", "nil"),
		doc_fn_str("getwd", ArgsOf(), "Returns the current working directory.", "string"),
		doc_fn_str("chdir", ArgsOf("dir"), "Changes the current working directory.", "nil"),
		doc_fn_str("mkdir", ArgsOf("dir"), "Creates a directory.", "nil"),
		doc_fn_str("mkdir_all", ArgsOf("dir"), "Creates a directory and any necessary parents.", "nil"),
		doc_fn_str("cp", ArgsOf("src", "dst"), "Copies a file.", "nil"),
		doc_fn_str("mv", ArgsOf("src", "dst"), "Moves a file.", "nil"),
		doc_fn_str("system", ArgsOf("command"), "Executes a system command.", "nil"),
		doc_str("os", "The operating system."),
		doc_fn_str("exit", ArgsOf("code"), "Exits the program with a status code.", "nil"),
	)
	library["exec"] = func(a string, b ...string) interface{} {
		out, err := exec.Command(a, b...).Output()
		if err != nil {
			return err
		}
		return string(out)
	}
	doc_fn("exec", ArgsOf("command", "args"), "Executes a system command.", "value")
}
