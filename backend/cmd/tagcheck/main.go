// Tagcheck は struct タグのよくある誤記をコンパイル後・結合前に検出するための小さな静的解析ツールです。
// Gin の binding と GORM の gorm タグだけを対象にし、unknown ルールで exit 1 とします。
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/KENTA0326/run-sync-pro/internal/config"
)

// go-playground/validator でよく使うルール + 本プロジェクトで使っているもの。
var bindingRules = map[string]struct{}{
	"required": {}, "omitempty": {}, "skip": {}, "dive": {}, "keys": {}, "endkeys": {},
	"len": {}, "eq": {}, "ne": {}, "lt": {}, "lte": {}, "gt": {}, "gte": {},
	"min": {}, "max": {}, "oneof": {}, "contains": {}, "excludes": {}, "unique": {},
	"email": {}, "uuid": {}, "url": {}, "uri": {}, "hostname": {}, "ip": {},
	"ipv4": {}, "ipv6": {}, "datetime": {}, "timezone": {}, "json": {}, "jwt": {},
	"lowercase": {}, "uppercase": {}, "number": {}, "numeric": {}, "hexadecimal": {},
	"hexcolor": {}, "rgb": {}, "rgba": {}, "hsl": {}, "hsla": {}, "latitude": {}, "longitude": {},
	"ssn": {}, "cron": {}, "mongodb": {}, "iscolor": {}, "isbn": {}, "issn": {},
}

// GORM のよく使う句（誤記検出用・網羅ではない）。
var gormClauses = map[string]struct{}{
	"primarykey": {}, "primary_key": {}, "autoincrement": {}, "auto_increment": {},
	"column": {}, "type": {}, "size": {}, "precision": {}, "scale": {},
	"default": {}, "not null": {}, "null": {}, "unique": {}, "uniqueindex": {}, "index": {},
	"embedded": {}, "embeddedprefix": {}, "serializer": {}, "constraint": {}, "references": {},
	"foreignkey": {}, "foreign_key": {}, "many2many": {}, "joinforeignkey": {}, "joinreferences": {},
	"associationforeignkey": {}, "comment": {}, "check": {}, "-": {}, "deleted_at": {},
}

func main() {
	rootArg := ""
	if len(os.Args) > 1 {
		rootArg = os.Args[1]
	}
	root := config.ResolveString(rootArg, "TAGCHECK_ROOT", ".")
	var errs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		es := checkFile(path)
		errs = append(errs, es...)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "tagcheck walk: %v\n", err)
		os.Exit(2)
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, e)
	}
	if len(errs) > 0 {
		os.Exit(1)
	}
}

func checkFile(path string) []string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return []string{fmt.Sprintf("%s: parse: %v", path, err)}
	}
	var out []string
	ast.Inspect(node, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, field := range st.Fields.List {
			if field.Tag == nil {
				continue
			}
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			if b := tag.Get("binding"); b != "" {
				for _, msg := range checkBinding(fset, field.Pos(), path, b) {
					out = append(out, msg)
				}
			}
			if g := tag.Get("gorm"); g != "" {
				for _, msg := range checkGorm(fset, field.Pos(), path, g) {
					out = append(out, msg)
				}
			}
		}
		return true
	})
	return out
}

func checkBinding(fset *token.FileSet, pos token.Pos, path, tag string) []string {
	var out []string
	line := fset.Position(pos).Line
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name := part
		if i := strings.Index(part, "="); i >= 0 {
			name = strings.TrimSpace(part[:i])
		}
		nameLower := strings.ToLower(name)
		if _, ok := bindingRules[nameLower]; !ok {
			out = append(out, fmt.Sprintf("%s:%d: unknown binding rule %q (tag=%q)", path, line, name, tag))
		}
	}
	return out
}

func checkGorm(fset *token.FileSet, pos token.Pos, path, tag string) []string {
	var out []string
	line := fset.Position(pos).Line
	for _, seg := range strings.Split(tag, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		key := gormClauseKey(seg)
		keyNorm := strings.ToLower(strings.TrimSpace(key))
		if _, ok := gormClauses[keyNorm]; ok {
			continue
		}
		// column:x / type:varchar(255) などは先頭キーのみ検証
		if colon := strings.Index(seg, ":"); colon >= 0 {
			prefix := strings.ToLower(strings.TrimSpace(seg[:colon]))
			if _, ok := gormClauses[prefix]; ok {
				continue
			}
			out = append(out, fmt.Sprintf("%s:%d: unknown gorm clause prefix %q (segment=%q)", path, line, prefix, seg))
			continue
		}
		out = append(out, fmt.Sprintf("%s:%d: unknown gorm clause %q", path, line, seg))
	}
	return out
}

func gormClauseKey(seg string) string {
	seg = strings.TrimSpace(seg)
	if i := strings.Index(seg, ":"); i >= 0 {
		return strings.TrimSpace(seg[:i])
	}
	return seg
}
