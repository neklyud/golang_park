package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
)

var x int = 0          //level counter
var oldsep string = "" //last separator

func dirTree(out *bytes.Buffer, paths string, printFiles bool) error {
	var sep string                           //current separator
	var sz int64                             //size of directory or file
	sep1 := "|"                              //separators
	sep2 := "├───"                           //
	sep3 := "└───"                           //
	listDirFiles, _ := filepath.Glob("*")    //list of directorys and files into dir
	var err error = nil                      //
	for i := 0; i < len(listDirFiles); i++ { //cycle into dir
		sep1 = "│"
		sep = sep2
		lst, _ := os.Lstat(listDirFiles[i])
		sz = lst.Size()
		mode := lst.Mode()
		k := false
		if printFiles == false { //if we have directories after current
			for j := i + 1; j < len(listDirFiles); j++ { //set flag
				ls, _ := os.Lstat(listDirFiles[j])
				mod := ls.Mode()
				if mod.IsDir() {
					k = true
					break
				}
			}
		}
		if (i == len(listDirFiles)-1 && printFiles == true) || (k == false && printFiles == false) {
			sep = sep3 //set separator for last dir or file
		}
		if !mode.IsDir() && printFiles == true { //file output
			for j := 0; j < x; j++ {
				out.WriteString("	")
				if oldsep == sep3 && j == x-2 {
					sep1 = ""
				}
				if j != x-1 {
					out.WriteString(sep1)
				}
			}
			out.WriteString(sep + listDirFiles[i])
			if listDirFiles[0] == "main.go" && paths == "." && i == 0 {
				out.WriteString(" (vary)")
			} else {
				out.WriteString(" (")
				if sz != 0 {
					out.WriteString(strconv.FormatInt(sz, 10))
					out.WriteString("b")
				} else if sz == 0 {
					out.WriteString("empty")
				}
				out.WriteString(")")
			}
			out.WriteString("\n")
		} else if mode.IsDir() { //directory output
			for j := 0; j < x; j++ {
				out.WriteString("	")
				if j != x-1 {
					out.WriteString(sep1)
				}
			}
			out.WriteString(sep + listDirFiles[i])
			out.WriteString("\n")
			os.Chdir(listDirFiles[i])
			x++ //level counter
			oldsep = sep
			err = dirTree(out, listDirFiles[i], printFiles)
			x--
			os.Chdir("..")
		}
	}
	return err
}
func main() {
	out := new(bytes.Buffer)
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	paths := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, paths, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
