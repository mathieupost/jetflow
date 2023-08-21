package generate

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
)

//go:embed templates/types.gotmpl
var typesTemplate string

//go:embed templates/proxy.gotmpl templates/handler.gotmpl
var templatesFS embed.FS

type Writer struct {
	state *State
	dir   string
}

func NewWriter(state *State, dir string) *Writer {
	return &Writer{
		state: state,
		dir:   dir,
	}
}

func (w *Writer) Write() error {
	tmplAll, err := template.New("writer").
		Funcs(template.FuncMap{
			"toCamel": strcase.ToCamel,
		}).
		ParseFS(templatesFS, "templates/*.gotmpl")
	if err != nil {
		return errors.Wrap(err, "parsing templates")
	}

	for _, tmpl := range tmplAll.Templates() {
		for _, t := range w.state.Types {
			typename := strcase.ToSnake(t.Name)
			filename := fmt.Sprintf("%s_%s", typename, tmpl.ParseName)
			filename = strings.TrimSuffix(filename, "tmpl")
			err := w.executeTemplate(tmpl, filename, t)
			if err != nil {
				return errors.Wrap(err, "executing template")
			}
		}
	}

	typesTmpl, err := template.New("types").Parse(typesTemplate)
	err = w.executeTemplate(typesTmpl, "types.go", nil)
	if err != nil {
		return errors.Wrap(err, "executing types template")
	}

	return nil
}

func (w *Writer) executeTemplate(tmpl *template.Template, filename string, t *Type) error {
	dir := w.dir
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "creating directory")
	}

	filename = path.Join(dir, filename)
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return errors.Wrap(err, "creating file")
	}

	var buf bytes.Buffer
	err = tmpl.
		Execute(&buf, map[string]interface{}{
			"Package": w.state.Package,
			"Types":   w.state.Types,
			"Type":    t,
		})
	if err != nil {
		return errors.Wrap(err, "executing template")
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "formatting source")
	}

	println(filename, string(out))

	_, err = file.Write(out)
	if err != nil {
		return errors.Wrap(err, "writing to file")
	}

	println("filename", filename)
	return nil
}
