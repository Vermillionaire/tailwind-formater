package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

var regCSS = `[}(\n\s]\.(?P<class>[\\\/\w:-]*)`
var regClass = `class=".*" +`

func main() {

	input := flag.String("input", "input.css", "tailwind-formatter -input input.css")
	ext := flag.String("extension", ".html", "tailwind-formatter -input input.css -extension .html")

	flag.Parse()

	sufixes := []string{
		":hover",
		":nth-child",
		":after",
		":before",
	}

	file, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile(regCSS)
	matches := r.FindAllStringSubmatch(string(file), -1)
	index := r.SubexpIndex("class")

	if matches == nil {
		fmt.Println("Found no classes in the css.")
		return
	}

	tailwindClasses := make([]string, 0)
	for _, m := range matches {
		class := trimSuffixSlice(m[index], sufixes)
		class = strings.Replace(class, `\`, "", -1)
		tailwindClasses = append(tailwindClasses, class)
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	r = regexp.MustCompile(regClass)

	for _, file := range find(dir, *ext) {

		f, err := os.OpenFile(file, os.O_RDWR, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, f)
		if err != nil {
			panic(err)
		}

		replaced := r.ReplaceAllStringFunc(buf.String(), func(value string) string {
			split := strings.Split(value, `"`)
			classes := strings.Split(strings.Trim(split[1], " "), " ")

			sort.Slice(classes, func(i, j int) bool {

				ci := classes[i]
				cj := classes[j]

				ti := slices.Index(tailwindClasses, ci)
				tj := slices.Index(tailwindClasses, cj)

				return ti < tj

			})

			sorted := strings.Join(classes, " ")

			sufix := ""
			if string(value[len(value)-1:][0]) == " " {
				sufix = " "
			}

			return fmt.Sprintf("class=\"%s\"%s", sorted, sufix)
		})

		_, err = f.Seek(0, 0)
		if err != nil {
			panic(err)
		}

		_, err = f.Write([]byte(replaced))
		if err != nil {
			panic(err)
		}

	}

}

func trimSuffixSlice(val string, suffix []string) string {
	for _, s := range suffix {
		if strings.Contains(val, s) {
			return strings.TrimSuffix(val, s)
		}
	}

	return val
}

func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}
