package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

const (
	STATIC_PATH = "views/"
	FILE_EXT    = ".html"
	LAYOUT_DIR  = "views/layouts/"
)

type View struct {
	Template *template.Template
	Layout   string
}

func gatherLayoutFiles() []string {
	files, err := filepath.Glob(LAYOUT_DIR + "*" + FILE_EXT)
	if err != nil {
		panic(err)
	}

	return files
}

func appendFilePathPrefix(files []string) {
	for i, file := range files {
		files[i] = STATIC_PATH + file
	}
}

func appendFileExt(files []string) {
	for i, file := range files {
		files[i] = file + FILE_EXT
	}
}

func NewView(layout string, files ...string) *View {
	appendFilePathPrefix(files)
	appendFileExt(files)

	files = append(files, gatherLayoutFiles()...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, r); err != nil {
		panic(err)
	}
}
