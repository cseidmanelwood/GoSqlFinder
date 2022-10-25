package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/*
Represents the source code file in the string with a pointer to the line the
system is currently reading
*/
type code struct {
	code string
	line int
}

type visitor struct {
	fileSet     *token.FileSet
	occurrences []code
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch d := n.(type) {
	case *ast.AssignStmt:
		{
			//fmt.Printf("Assignment %s at %v\n", d.Tok.String(), v.fset.Position(d.Pos()))
		}
	case *ast.GoStmt:
		{
			//fmt.Printf("Statement at %v\n", v.fset.Position(d.Pos()))
		}
	case *ast.Ident:
		{
			//fmt.Printf("Ident %s at %v\n", d.String(), v.fset.Position(d.Pos()))
		}
	case *ast.BasicLit:
		{
			val := d.Value
			comp := strings.ToLower(val)
			if strings.Contains(comp, "select") ||
				strings.Contains(comp, "insert") ||
				strings.Contains(comp, "delete") ||
				strings.Contains(comp, "update") {

				v.occurrences = append(v.occurrences, code{val, v.fileSet.Position(d.Pos()).Line})
			}
		}

	default:
		//fmt.Printf("Other %v at %d\n", d, d.Pos())
	}

	return v
}

func getSqlCode(srcFile string, v visitor, outfile os.File) {

	v.fileSet = token.NewFileSet()

	f, err := parser.ParseFile(v.fileSet, srcFile, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast.Walk(&v, f)

	if len(v.occurrences) > 0 {
		_, _ = outfile.WriteString(fmt.Sprintf("### File: %s\n\n", srcFile))
		_, _ = outfile.WriteString(fmt.Sprintf("**Found %d occurrence(s) of SQL**\n\n", len(v.occurrences)))
		_, _ = outfile.WriteString(fmt.Sprintf("---\n"))
		for _, o := range v.occurrences {
			_, _ = outfile.WriteString(fmt.Sprintf("Line %d :\n", o.line))
			_, _ = outfile.WriteString(fmt.Sprintf("```SQL\n"))
			_, _ = outfile.WriteString(fmt.Sprintf(o.code))
			_, _ = outfile.WriteString(fmt.Sprintf("\n```\n\n"))
		}
	}

}

func main() {

	inputPath := flag.String("input", ".", "Directory containing go files")
	outputPath := flag.String("output", "output.md", "Name and path of output file")

	flag.Parse()

	if _, err := os.Stat(*inputPath); os.IsNotExist(err) {
		fmt.Printf("Path '%s' does not exist\n", *inputPath)
		return
	}

	var v visitor
	v.fileSet = token.NewFileSet()

	outFile, createFileErr := os.Create(*outputPath)
	if createFileErr != nil {
		fmt.Printf("Unable to create file %s\n", *outputPath)
		return
	}

	err := filepath.Walk(*inputPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".go" {
				getSqlCode(path, v, *outFile)
			}

			return nil
		})

	_ = outFile.Close()
	if err != nil {
		return
	}

	if err != nil {
		log.Println(err)
	}
}
