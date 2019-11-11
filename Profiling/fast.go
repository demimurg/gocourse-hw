package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	jlexer "github.com/mailru/easyjson/jlexer"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

// UnmarshalJSON supports json.Unmarshaler interface
func (out *User) UnmarshalJSON(data []byte) error {
	in := jlexer.Lexer{Data: data}

	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return in.Error()
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = string(in.String())
		case "name":
			out.Name = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}

	return in.Error()
}

// FastSearch - like SlowSearch, but better
func FastSearch(out io.Writer) {
	seenBrowsers := map[string]bool{}
	uniqueBrowsers := 0

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	buf := bufio.NewReader(file)
	i := -1

	fmt.Fprintln(out, "found users:")

	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}

		user := &User{}
		err = user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}
		i++

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {

			androidBrowser := strings.Contains(browser, "Android")
			if androidBrowser {
				isAndroid = true
			}

			msieBrowser := strings.Contains(browser, "MSIE")
			if msieBrowser {
				isMSIE = true
			}

			if (androidBrowser || msieBrowser) && !seenBrowsers[browser] {
				seenBrowsers[browser] = true
				uniqueBrowsers++
			}
		}

		if isAndroid && isMSIE {
			email := strings.Replace(user.Email, "@", " [at] ", 1)
			fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
		}
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
