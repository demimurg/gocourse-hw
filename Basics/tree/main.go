package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var interfaceElements = map[string]string{
	"Т": "├───",
	"Г": "└───",
	"-": "│",
}

func dirTree(out io.Writer, path string, needFiles bool) error {
	var crawler func(path, indent string) error

	crawler = func(path, indent string) error {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Have problem with path, error: %v", err)
		}

		folderContent, _ := file.Readdir(0)

		type data struct {
			name  string
			size  int
			isDir bool
		}
		var filteredContent []data

		for _, item := range folderContent {
			hidden := strings.HasPrefix(item.Name(), ".")
			if !item.IsDir() && !needFiles || hidden {
				continue
			}

			filteredContent = append(filteredContent, data{
				item.Name(),
				int(item.Size()),
				item.IsDir(),
			})
		}

		sort.Slice(filteredContent, func(i, j int) bool {
			return filteredContent[i].name < filteredContent[j].name
		})

		var row string
		for i, item := range filteredContent {

			row = indent
			if i == len(filteredContent)-1 {
				row += interfaceElements["Г"]
			} else {
				row += interfaceElements["Т"]
			}
			row += item.name
			if !item.isDir {
				if item.size == 0 {
					row += " (empty)"
				} else {
					row += fmt.Sprintf(" (%vb)", item.size)
				}
			}
			fmt.Fprintf(out, "%v", row+"\n")

			if item.isDir {
				if i < len(filteredContent)-1 {
					crawler(path+"/"+item.name+"/",
						indent+interfaceElements["-"]+"\t")
				} else {
					crawler(path+"/"+item.name+"/", indent+"\t")
				}
			}
		}

		return nil
	}

	return crawler(path, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
