package g8

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"e8vm.io/e8vm/build8"
	"e8vm.io/e8vm/dagvis"
	"e8vm.io/e8vm/g8/ast"
	"e8vm.io/e8vm/g8/ir"
	"e8vm.io/e8vm/g8/parse"
	"e8vm.io/e8vm/lex8"
)

type lang struct {
	golike bool
}

// Lang returns the G language builder for the building system
func Lang() build8.Lang { return new(lang) }

// LangGolike returns the G lauguage that uses a subset of golang AST.
func LangGolike() build8.Lang {
	return &lang{golike: true}
}

func (l *lang) IsSrc(filename string) bool {
	return strings.HasSuffix(filename, ".g")
}

func (l *lang) Prepare(
	src map[string]*build8.File, importer build8.Importer,
) []*lex8.Error {
	importer.Import("$", "asm/builtin", nil)
	if f := build8.OnlyFile(src); f != nil {
		return listImport(f.Path, f, importer, l.golike)
	}

	f := src["import.g"]
	if f == nil {
		return nil
	}
	return listImport(f.Path, f, importer, l.golike)
}

func makeBuilder(pinfo *build8.PkgInfo, golike bool) *builder {
	b := newBuilder(pinfo.Path, golike)
	initBuilder(b, pinfo.Import)
	return b
}

func initBuilder(b *builder, imp map[string]*build8.Import) {
	b.exprFunc = buildExpr
	b.stmtFunc = buildStmt
	b.typeFunc = buildType
	b.constFunc = buildConstExpr

	builtin, ok := imp["$"]
	if !ok {
		b.Errorf(nil, "builtin import missing for %q", b.path)
		return
	}

	declareBuiltin(b, builtin.Compiled.Lib())
}

// parse all files
func (l *lang) parsePkg(pinfo *build8.PkgInfo) (
	map[string]*ast.File, []*lex8.Error,
) {
	var parseErrs []*lex8.Error
	asts := make(map[string]*ast.File)
	for name, src := range pinfo.Src {
		if filepath.Base(src.Path) != name {
			panic("basename in path is different from the file name")
		}

		f, es := parse.File(src.Path, src, l.golike)
		if es != nil {
			parseErrs = append(parseErrs, es...)
		}

		if err := src.Close(); err != nil {
			parseErrs = append(parseErrs, &lex8.Error{Err: err})
		}

		asts[name] = f
	}
	if len(parseErrs) > 0 {
		return nil, parseErrs
	}

	return asts, nil
}

func logIr(pinfo *build8.PkgInfo, b *builder) error {
	w := pinfo.CreateLog("ir")
	ir.PrintPkg(w, b.p)
	return w.Close()
}

func logDeps(pinfo *build8.PkgInfo, g *dagvis.Graph) error {
	bs, err := json.MarshalIndent(g.Nodes, "", "    ")
	if err != nil {
		panic(err)
	}

	w := pinfo.CreateLog("deps")
	if _, err := w.Write(bs); err != nil {
		return err
	}
	return w.Close()
}

func logDepMap(pinfo *build8.PkgInfo, deps []byte) error {
	w := pinfo.CreateLog("depmap")
	if _, err := w.Write(deps); err != nil {
		return err
	}
	return w.Close()
}

func (l *lang) Compile(pinfo *build8.PkgInfo) (
	compiled build8.Linkable, es []*lex8.Error,
) {
	// parsing
	asts, es := l.parsePkg(pinfo)
	if es != nil {
		return nil, es
	}

	// building
	b := makeBuilder(pinfo, l.golike)
	if es = b.Errs(); es != nil {
		return nil, es
	}

	p := newPkg(asts)
	b.deps = newDeps(asts) // dependency map init

	p.build(b, pinfo)
	if es = b.Errs(); es != nil {
		return nil, es
	}

	// circular dep check
	g := b.deps.graph()
	if err := logDeps(pinfo, g); err != nil {
		return nil, lex8.SingleErr(err)
	}

	bs, err := dagvis.LayoutJSON(g.Reverse())
	if err != nil {
		return nil, lex8.SingleErr(err)
	}
	if err := logDepMap(pinfo, bs); err != nil {
		return nil, lex8.SingleErr(err)
	}

	// codegen
	lib := ir.BuildPkg(b.p)

	// IR logging
	if err := logIr(pinfo, b); err != nil {
		return nil, lex8.SingleErr(err)
	}

	return &builtPkg{p: p, lib: lib}, nil
}
