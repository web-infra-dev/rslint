package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"log"
	"maps"
	"os"
	"path"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
)

const tsgoInternalPrefix = "github.com/microsoft/TypeScript/tsc/internal/"

type ExtraShim struct {
	ExtraFunctions []string            `json:"ExtraFunctions"`
	ExtraMethods   map[string][]string `json:"ExtraMethods"`
	ExtraFields    map[string][]string `json:"ExtraFields"`
	// Field pointers preserve the identity of checker state and embedded stores.
	// A dotted selector names a field inside a mirrored private struct.
	ExtraFieldPointers map[string][]string `json:"ExtraFieldPointers"`
	IgnoreFunctions    []string            `json:"IgnoreFunctions"`
}

func emitFieldAccessor(out *strings.Builder, receiver *types.Named, selector string, pointer bool, qualifier types.Qualifier) {
	var fieldType types.Type = receiver
	for part := range strings.SplitSeq(selector, ".") {
		member, _, _ := types.LookupFieldOrMethod(fieldType, true, receiver.Obj().Pkg(), part)
		field, ok := member.(*types.Var)
		if !ok || !field.IsField() {
			log.Fatalf("expected %s to contain field %s", receiver.Obj().Name(), selector)
		}
		fieldType = field.Type()
	}
	name := receiver.Obj().Name()
	accessor := name + "_" + strings.ReplaceAll(selector, ".", "_")
	address := ""
	if pointer {
		accessor += "Ptr"
		address = "&"
		fieldType = types.NewPointer(fieldType)
	}
	fmt.Fprintf(out, "func %s(v *%s) %s {\n  return %s((*extra_%s)(unsafe.Pointer(v))).%s\n}\n",
		accessor, types.TypeString(receiver, qualifier), types.TypeString(fieldType, qualifier), address, name, selector)
}

// check whether signature can be exported
func canBeExported(t types.Signature) bool {
	if params := t.Params(); params != nil {
		for i := range params.Len() {
			ty := params.At(i).Type()

			if ptrType, ok := ty.(*types.Pointer); ok {
				ty = ptrType.Elem()
			}
			if named, ok := ty.(*types.Named); ok {
				if !named.Obj().Exported() {
					return false
				}
			}
		}
	}
	return true
}

func main() {
	generateShims(func(name string, source []byte) error {
		if err := os.MkdirAll(path.Dir(name), 0o755); err != nil {
			return err
		}
		return os.WriteFile(name, source, 0o644)
	})
}

func generateShims(writeFile func(string, []byte) error) {
	packagesToShim := []string{
		"ast",
		"binder",
		"bundled",
		"checker",
		"compiler",
		"core",
		"diagnostics",
		"evaluator",
		"jsnum",
		"jsonrpc",
		"parser",
		"scanner",
		"semver",
		"tsoptions",
		"tspath",
		"transformers/jsxtransforms",
		"vfs",
		"vfs/cachedvfs",
		"vfs/iovfs",
		"vfs/osvfs",
		"vfs/trackingvfs",
		"vfs/vfsmatch",
		"collections",
		"lsp/lsproto",
		"ls",
		"locale",
		"module",
		"nodebuilder",
		"printer",
		"project",
		"project/logging",
		"api",
		"api/encoder",
	}

	packagesToShimFullNames := make([]string, len(packagesToShim))
	for i, pkg := range packagesToShim {
		packagesToShimFullNames[i] = tsgoInternalPrefix + pkg
	}

	loadedPackages, err := packages.Load(&packages.Config{
		// TODO: path relative to repo root
		Dir:  "./shim/compiler",
		Mode: packages.LoadSyntax,
	}, packagesToShimFullNames...)
	if err != nil {
		log.Fatalf("Error loading package: %v", err)
	}
	if packages.PrintErrors(loadedPackages) > 0 {
		log.Fatal("Cannot generate shims: fix the pinned compiler's package errors first")
	}

	var shimHeaderBuilder strings.Builder
	var shimBuilder strings.Builder
	var tempBuffer bytes.Buffer

	for _, pkg := range loadedPackages {
		shimDirPath := path.Join("./shim/", strings.TrimPrefix(pkg.PkgPath, tsgoInternalPrefix))
		var extraShim ExtraShim
		extraShimFilePath := path.Join(shimDirPath, "extra-shim.json")
		if data, err := os.ReadFile(extraShimFilePath); err == nil {
			if err := json.Unmarshal(data, &extraShim); err != nil {
				log.Fatalf("error parsing %s: %v", extraShimFilePath, err)
			}
		} else if !os.IsNotExist(err) {
			log.Fatalf("error reading %s: %v", extraShimFilePath, err)
		}

		// true if directly used, false otherwise
		importedPackages := map[string]bool{}

		importPackage := func(pkg string, directly bool) {
			if directly {
				importedPackages[pkg] = true
			} else if _, ok := importedPackages[pkg]; !ok {
				importedPackages[pkg] = false
			}
		}

		var qualifierOnlyPackageName types.Qualifier = func(p *types.Package) string {
			importPackage(p.Path(), true)
			return p.Name()
		}
		var qualifierEmptyPackageName types.Qualifier = func(p *types.Package) string {
			return ""
		}

		emitGoLinknameDirective := func(localName string, fn *types.Func) {
			// //go:linkname only allowed in Go files that import "unsafe"
			importPackage("unsafe", false)
			importPackage(pkg.Types.Path(), false)
			shimBuilder.WriteString("//go:linkname ")
			shimBuilder.WriteString(localName)
			shimBuilder.WriteByte(' ')
			shimBuilder.WriteString(fn.Pkg().Path())
			shimBuilder.WriteByte('.')
			if recv := fn.Signature().Recv(); recv != nil {
				shimBuilder.WriteByte('(')
				shimBuilder.WriteString(types.TypeString(recv.Type(), qualifierEmptyPackageName))
				shimBuilder.WriteByte(')')
				shimBuilder.WriteByte('.')
			}
			shimBuilder.WriteString(fn.Name())
			shimBuilder.WriteByte('\n')
		}

		emitLinkedFunction := func(fn *types.Func) bool {
			if fn.Signature().TypeParams() != nil {
				// https://github.com/golang/go/issues/60425
				// linking to functions with generics is not supported in go:linkname
				return false
			}
			if !canBeExported(*fn.Signature()) {
				log.Printf("Skip unexported %s.%s", fn.Pkg().Name(), fn.Name())
				return false
			}
			name := cases.Title(language.English, cases.NoLower).String(fn.Name())
			emitGoLinknameDirective(name, fn)
			shimBuilder.WriteString("func ")
			shimBuilder.WriteString(name)
			types.WriteSignature(&tempBuffer, fn.Signature(), qualifierOnlyPackageName)
			shimBuilder.Write(tempBuffer.Bytes())
			tempBuffer.Reset()
			shimBuilder.WriteString("\n")
			return true
		}

		emitGenericWrapper := func(fn *types.Func) bool {
			signature := fn.Signature()
			if signature.TypeParams() == nil {
				return false
			}
			for param := range signature.Params().Variables() {
				if param.Name() == "" {
					return false
				}
			}

			importPackage(pkg.Types.Path(), true)
			shimBuilder.WriteString("func ")
			shimBuilder.WriteString(fn.Name())
			types.WriteSignature(&tempBuffer, signature, qualifierOnlyPackageName)
			shimBuilder.Write(tempBuffer.Bytes())
			tempBuffer.Reset()
			shimBuilder.WriteString(" {\n\t")
			if signature.Results().Len() > 0 {
				shimBuilder.WriteString("return ")
			}
			shimBuilder.WriteString(pkg.Name)
			shimBuilder.WriteByte('.')
			shimBuilder.WriteString(fn.Name())
			shimBuilder.WriteByte('[')
			for index := range signature.TypeParams().Len() {
				if index > 0 {
					shimBuilder.WriteString(", ")
				}
				shimBuilder.WriteString(signature.TypeParams().At(index).Obj().Name())
			}
			shimBuilder.WriteString("](")
			for index := range signature.Params().Len() {
				if index > 0 {
					shimBuilder.WriteString(", ")
				}
				shimBuilder.WriteString(signature.Params().At(index).Name())
				if signature.Variadic() && index == signature.Params().Len()-1 {
					shimBuilder.WriteString("...")
				}
			}
			shimBuilder.WriteString(")\n}\n")
			return true
		}

		matchedExtraFunctions := make(map[string]bool, len(extraShim.ExtraFunctions))
		for _, name := range extraShim.ExtraFunctions {
			matchedExtraFunctions[name] = false
		}
		matchedExtraMethods := make(map[string](map[string]bool), len(extraShim.ExtraMethods))
		for name, methods := range extraShim.ExtraMethods {
			matchedExtraMethods[name] = make(map[string]bool, len(methods))
			for _, method := range methods {
				matchedExtraMethods[name][method] = false
			}
		}
		matchedExtraFields := make(map[string]bool, len(extraShim.ExtraFields))
		for name := range extraShim.ExtraFields {
			matchedExtraFields[name] = false
		}
		for name := range extraShim.ExtraFieldPointers {
			matchedExtraFields[name] = false
		}

		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			object := scope.Lookup(name)
			if !object.Exported() {
				fn, isFunc := object.(*types.Func)
				if _, exists := matchedExtraFunctions[name]; isFunc && exists {
					if emitLinkedFunction(fn) {
						matchedExtraFunctions[name] = true
					}
				}
				continue
			}

			printReexport := func(kind string) {
				importPackage(pkg.Types.Path(), true)
				shimBuilder.WriteString(kind)
				shimBuilder.WriteString(" ")
				shimBuilder.WriteString(name)
				shimBuilder.WriteString(" = ")
				shimBuilder.WriteString(pkg.Name)
				shimBuilder.WriteString(".")
				shimBuilder.WriteString(name)
				shimBuilder.WriteString("\n")
			}

			switch object := object.(type) {
			case *types.TypeName:
				t := object.Type()
				named, isNamed := t.(*types.Named)
				if isNamed {
					_, nameWithTypeParams, _ := strings.Cut(types.TypeString(named, qualifierOnlyPackageName), ".")
					importPackage(pkg.Types.Path(), true)
					shimBuilder.WriteString("type ")
					shimBuilder.WriteString(nameWithTypeParams)
					shimBuilder.WriteString(" = ")
					shimBuilder.WriteString(pkg.Name)
					shimBuilder.WriteString(".")
					shimBuilder.WriteString(name)

					typeParams := slices.Collect(named.TypeParams().TypeParams())
					if len(typeParams) > 0 {
						// (*typeWriter)typeList
						shimBuilder.WriteByte('[')
						for i, typ := range typeParams {
							if i > 0 {
								shimBuilder.WriteByte(',')
							}
							shimBuilder.WriteString(typ.String())
						}
						shimBuilder.WriteByte(']')
					}

					shimBuilder.WriteString("\n")
				} else {
					printReexport("type")
				}

				if extraMethods, ok := matchedExtraMethods[name]; isNamed && ok {
					for method := range named.Methods() {
						methodName := method.Name()
						if _, exists := extraMethods[methodName]; !exists {
							continue
						}
						extraMethods[methodName] = true
						prefix := name + "_"
						emitGoLinknameDirective(prefix+methodName, method)
						funcDeclStr := types.ObjectString(method, qualifierOnlyPackageName)
						recvStart := 0
						recvEnd := 0
						paramsStart := 0
						for i, s := range funcDeclStr {
							if s == '(' {
								if recvStart == 0 {
									recvStart = i + 1
								}
								if recvEnd != 0 {
									paramsStart = i + 1
									break
								}
							}
							if s == ')' && recvEnd == 0 {
								recvEnd = i
							}
						}
						shimBuilder.WriteString("func ")
						shimBuilder.WriteString(prefix)
						shimBuilder.WriteString(funcDeclStr[recvEnd+2 : paramsStart])
						shimBuilder.WriteString("recv ")
						shimBuilder.WriteString(funcDeclStr[recvStart:recvEnd])
						if method.Signature().Params() != nil {
							shimBuilder.WriteString(", ")
						}
						shimBuilder.WriteString(funcDeclStr[paramsStart:])
						shimBuilder.WriteString("\n")
					}
				}

				if _, ok := matchedExtraFields[name]; isNamed && ok {
					importPackage("unsafe", true)

					matchedExtraFields[name] = true
					var emitExtraStruct func(name string, s *types.Struct)
					emitExtraStruct = func(name string, s *types.Struct) {
						shimBuilder.WriteString("type extra_")
						shimBuilder.WriteString(name)
						shimBuilder.WriteString(" struct {")

						dependencies := [](struct {
							string
							*types.Struct
						}){}
						for field := range s.Fields() {
							shimBuilder.WriteString("\n  ")
							if !field.Embedded() {
								shimBuilder.WriteString(field.Name())
								shimBuilder.WriteByte(' ')
							}

							// Private struct values (including instantiated generic stores)
							// need an anonymous mirror with the same layout. Their names
							// cannot be referenced from this shim's package.
							if named, ok := field.Type().(*types.Named); ok && !named.Obj().Exported() {
								if structType, ok := named.Underlying().(*types.Struct); ok {
									shimBuilder.WriteString(types.TypeString(structType, qualifierOnlyPackageName))
									continue
								}
							}

							ptrType, ok := field.Type().(*types.Pointer)
							if ok {
								named, ok := ptrType.Elem().(*types.Named)
								if ok && !named.Obj().Exported() {
									structType, ok := named.Underlying().(*types.Struct)
									if ok {
										n := named.Obj().Name()
										dependencies = append(dependencies, struct {
											string
											*types.Struct
										}{n, structType})
										shimBuilder.WriteString("extra_")
										shimBuilder.WriteString(n)
										continue
									}
								}
							}

							// TODO: move to extra-shim.json
							typeStr := types.TypeString(field.Type(), qualifierOnlyPackageName)
							typeStr = strings.ReplaceAll(typeStr, "checker.thisAssignmentDeclarationKind", "int32")
							typeStr = strings.ReplaceAll(typeStr, "checker.symbolTableID", "uint64")
							shimBuilder.WriteString(typeStr)
						}
						shimBuilder.WriteString("\n}\n")

						for _, dep := range dependencies {
							emitExtraStruct(dep.string, dep.Struct)
						}
					}

					structType, ok := named.Underlying().(*types.Struct)
					if !ok {
						log.Fatalf("expected %v to be struct", name)
					}

					emitExtraStruct(name, structType)

					for _, field := range extraShim.ExtraFields[name] {
						emitFieldAccessor(&shimBuilder, named, field, false, qualifierOnlyPackageName)
					}
					for _, field := range extraShim.ExtraFieldPointers[name] {
						emitFieldAccessor(&shimBuilder, named, field, true, qualifierOnlyPackageName)
					}
				}
			case *types.Const:
				printReexport("const")
			case *types.Var:
				printReexport("var")
			case *types.Func:
				if !slices.Contains(extraShim.IgnoreFunctions, name) {
					if object.Signature().TypeParams() != nil {
						if _, exists := matchedExtraFunctions[name]; exists {
							matchedExtraFunctions[name] = emitGenericWrapper(object)
						}
					} else {
						emitLinkedFunction(object)
					}
				}
			}
		}

		exit := false
		for name, found := range matchedExtraFields {
			if !found {
				log.Fatalf("couldn't find struct %s.%s for field access", pkg.Name, name)
			}
		}
		for fnName, found := range matchedExtraFunctions {
			if found {
				continue
			}
			fmt.Printf("ERROR: couldn't find %v function\n", fnName)
			exit = true
		}
		for name, methods := range matchedExtraMethods {
			for methodName, found := range methods {
				if found {
					continue
				}
				fmt.Printf("ERROR: couldn't find %v.%v method\n", name, methodName)
				exit = true
			}
		}
		if exit {
			os.Exit(1)
		}

		// https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
		shimHeaderBuilder.WriteString("\n// Code generated by tools/gen_shims. DO NOT EDIT.\n\n")
		shimHeaderBuilder.WriteString("package ")
		shimHeaderBuilder.WriteString(pkg.Name)
		shimHeaderBuilder.WriteString("\n\n")
		importsList := slices.Collect(maps.Keys(importedPackages))
		slices.Sort(importsList)
		for _, imported := range importsList {
			shimHeaderBuilder.WriteString("import ")
			if !importedPackages[imported] {
				shimHeaderBuilder.WriteString("_ ")
			}
			shimHeaderBuilder.WriteString("\"")
			shimHeaderBuilder.WriteString(imported)
			shimHeaderBuilder.WriteString("\"\n")
		}
		shimHeaderBuilder.WriteString("\n")

		shimGoPath := path.Join(shimDirPath, "shim.go")
		if err := writeFile(shimGoPath, []byte(shimHeaderBuilder.String()+shimBuilder.String())); err != nil {
			log.Fatalf("error writing %s: %v", shimGoPath, err)
		}

		shimHeaderBuilder.Reset()
		shimBuilder.Reset()
	}
}
