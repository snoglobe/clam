package main

var docs = map[string]string{}

func doc(name string, desc string) {
	docs[name] = "# " + name + "\n" + desc
	docs[name] += "\n\n<br>\n\n"
}

func doc_str(name string, desc string) string {
	return "## " + name + "\n" + desc
}

type Args struct {
	Args []string
	Many bool
}

func (a Args) String() string {
	if len(a.Args) == 0 {
		return ""
	}
	var result = ""
	for _, arg := range a.Args {
		result += arg + ", "
	}
	result = result[:len(result)-2]
	if a.Many {
		result += "..."
	}
	return result
}

func ArgsOf(args ...string) Args {
	return Args{Args: args}
}

func ManyArgs(args ...string) Args {
	return Args{Args: args, Many: true}
}

func doc_fn(name string, args Args, desc string, returns string) {
	docs[name] = "# " + name + "\n`" + name + "(" + args.String() + ") -> " + returns + "` : " + desc
	docs[name] += "\n\n<br>\n\n"
}

func doc_fn_str(name string, args Args, desc string, returns string) string {
	return "## " + name + "\n`" + name + "(" + args.String() + ") -> " + returns + "` : " + desc
}

func doc_obj(name string, desc string, fields ...string) {
	docs[name] = "# " + name + "\n" + desc
	for _, field := range fields {
		docs[name] += "\n\n" + field
	}
	docs[name] += "\n\n<br>\n\n"
}

func doc_obj_str(name string, desc string, fields ...string) string {
	var result = "## " + name + "\n" + desc
	for _, field := range fields {
		result += "\n\n" + field
	}
	return result
}
