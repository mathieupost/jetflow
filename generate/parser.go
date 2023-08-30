package generate

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Parser struct {
	state *State
	queue []func()
}

func ParsePackage(path string) {
	state := &State{
		Package: determineModuleImportPath(path),
		Types:   map[string]*Type{},
	}
	parser := Parser{
		state: state,
		queue: []func(){},
	}

	// List all files in the directory.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	// Parse each file.
	for _, f := range files {
		if !f.IsDir() {
			parser.ParseFile(path + "/" + f.Name())
		}
	}

	for len(parser.queue) > 0 {
		parser.queue[0]()
		parser.queue = parser.queue[1:]
	}

	// fmt.Println(state)

	w := NewWriter(state, filepath.Join(path, "gen"))
	err = w.Write()
	if err != nil {
		panic(err.Error())
	}
}

// ParseFile parses a single Go file.
func (p *Parser) ParseFile(path string) {
	// Open the file.
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new token file set and parse the source file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, file, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	state := p.state

	for _, decl := range node.Decls {
		if g, ok := decl.(*ast.GenDecl); ok {
			fmt.Printf("GenDecl: %T, %v\n", g, g)
			for _, spec := range g.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					fmt.Printf(" TypeSpec: %T, %v\n", s.Type, s)
					switch t := s.Type.(type) {
					case *ast.InterfaceType:
						name := s.Name.Name
						typ := &Type{
							Name:    name,
							Methods: []*Method{},
						}
						state.Types[name] = typ

						f := func() {
							fmt.Printf("  InterfaceType: %T, %v\n", t, t)
							for _, m := range t.Methods.List {
								fmt.Printf("   Method: %T, %v\n", m.Type, m)
								switch t := m.Type.(type) {
								case *ast.SelectorExpr:
									fmt.Printf("    SelectorExpr: %T, %v\n", t, t)
								case *ast.FuncType:
									fmt.Printf("    FuncType: %T, %v\n", t, t)
									name := m.Names[0].Name
									method := &Method{
										Name:       name,
										Parameters: []*Parameter{},
										Results:    []*Parameter{},
									}
									typ.Methods = append(typ.Methods, method)
									for _, r := range t.Results.List {
										result := &Parameter{
											Type: &Type{},
										}
										switch t := r.Type.(type) {
										case *ast.SelectorExpr:
											fmt.Printf("     SelectorExpr: %T, %v\n", t, t)
										case *ast.Ident:
											fmt.Printf("     Ident: %T, %v\n", t, t)
											if typ, ok := state.Types[t.Name]; ok {
												result.Type = typ
											} else {
												result.Type = &Type{Name: t.Name}
											}
										}
										method.Results = append(method.Results, result)
									}
									retlen := len(method.Results) - 1
									if method.Results[retlen].Type.Name == "error" {
										method.Results = method.Results[:retlen]
									}
									for _, p := range t.Params.List {
										name := p.Names[0].Name
										if name == "ctx" {
											continue
										}
										parameter := &Parameter{
											Name: name,
											Type: &Type{},
										}
										switch t := p.Type.(type) {
										case *ast.SelectorExpr:
											fmt.Printf("     SelectorExpr: %T, %v\n", t, t)
										case *ast.Ident:
											fmt.Printf("     Ident: %T, %v\n", t, t)
											if typ, ok := state.Types[t.Name]; ok {
												parameter.Type = typ
											} else {
												parameter.Type = &Type{Name: t.Name}
											}
										}
										method.Parameters = append(method.Parameters, parameter)
									}
								}
							}
						}
						p.queue = append(p.queue, f)
					}
				}
			}
		}
	}
}

func determineModuleImportPath(dirPath string) string {
	// Look for the go.mod file in the current directory and parent directories
	modFilePath := filepath.Join(dirPath, "go.mod")
	if _, err := os.Stat(modFilePath); err == nil {
		// Read the module path from go.mod
		modulePath, err := readModulePath(modFilePath)
		if err != nil {
			err = errors.Wrap(err, "reading module path")
			log.Fatal(err)
		}
		return modulePath
	}

	dirName := filepath.Base(dirPath)
	parentDir := filepath.Dir(dirPath)
	parentPackage := determineModuleImportPath(parentDir)
	return filepath.Join(parentPackage, dirName)
}

func readModulePath(modFilePath string) (string, error) {
	file, err := os.Open(modFilePath)
	if err != nil {
		return "", errors.Wrap(err, "opening module file")
	}
	defer file.Close()

	modulePath := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "scanning module file")
	}

	if modulePath == "" {
		return "", fmt.Errorf("module path not found in %s", modFilePath)
	}

	return modulePath, nil
}
