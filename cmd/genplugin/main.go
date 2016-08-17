package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	w     = new(bytes.Buffer)
	tmpls = make([]string, 0, 4)
)

func main() {
	inp := flag.String("inp", "dev", "Input directory containing .tmpl files")
	outp := flag.String("outp", "plugins/gorums", "Output directory for templates formated as Go consts")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// start writing to the in-mem buffer
	w.WriteString("package gorums\n\n")
	err := filepath.Walk(*inp, visit)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteString("var tmpls = map[string]string{\n")
	for _, s := range tmpls {
		w.WriteString("\t")
		w.WriteString("\"")
		w.WriteString(s)
		w.WriteString("\":\t")
		w.WriteString(s)
		w.WriteString(",\n")
	}
	w.WriteString("}\n")

	// do the actual writing of the in-mem buffer to the file
	oname := filepath.Join(*outp, "templates.go")
	if err = ioutil.WriteFile(oname, w.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

// visit generates a _gen.go file if the supplied path is an .tmpl file.
// Note that the function rely on data in global variable pkgData.
func visit(fpath string, f os.FileInfo, err error) error {
	if strings.HasSuffix(fpath, ".tmpl") {
		_, fname := path.Split(fpath)
		name := strings.TrimSuffix(fname, ".tmpl")
		log.Println("Processing: " + fname)
		b, err := ioutil.ReadFile(fpath)
		if err != nil {
			return err
		}
		tmplName := name + "_tmpl"
		tmpls = append(tmpls, tmplName)

		// for each new template file, write to the in-mem buffer
		w.WriteString("const " + tmplName + " = `")
		_, err = w.Write(b)
		if err != nil {
			return err
		}
		w.WriteString("`\n\n")
	}
	return nil
}