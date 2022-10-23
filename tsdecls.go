package tsdecls

import (
	_ "embed"
	"fmt"
	"go/types"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/bobg/go-generics/maps"
	"github.com/bobg/go-generics/set"
	"github.com/fatih/camelcase"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

//go:embed tmpl
var tmplstr string

var tmpl = template.Must(template.New("").Parse(tmplstr))

func Write(w io.Writer, dir, typename string) error {
	data := tsDecls{
		ClassName: typename,
	}

	config := &packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(config, dir)
	if err != nil {
		return errors.Wrapf(err, "loading %s", dir)
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("want 1 package in %s, got %d", dir, len(pkgs))
	}
	if pkgs[0].Types == nil {
		return fmt.Errorf("pkgs[0].Types == nil")
	}
	scope := pkgs[0].Types.Scope()
	if scope == nil {
		return fmt.Errorf("scope == nil")
	}
	obj := scope.Lookup(typename)
	if obj == nil {
		return fmt.Errorf("obj == nil")
	}

	methods := make(map[string]methodInfo)
	printed := set.New[string]()

	for _, typ := range []types.Type{obj.Type(), types.NewPointer(obj.Type())} {
		methodSet := types.NewMethodSet(typ)
		for i := 0; i < methodSet.Len(); i++ {
			method := methodSet.At(i)
			fn, ok := method.Obj().(*types.Func)
			if !ok {
				continue
			}
			if !fn.Exported() {
				continue
			}
			fname := fn.Name()

			ftype := fn.Type()
			sig, ok := ftype.(*types.Signature)
			if !ok {
				return fmt.Errorf("type of function %s is %T", fname, ftype)
			}

			var param, result *types.Var

			params := sig.Params()
			switch params.Len() {
			case 0:
				// do nothing
			case 1:
				p := params.At(0)
				if types.TypeString(p.Type(), nil) != "context.Context" {
					param = p
				}
			case 2:
				if types.TypeString(params.At(0).Type(), nil) != "context.Context" {
					continue // xxx skip this method
				}
				param = params.At(1)
			default:
				continue // xxx skip this method
			}

			if param != nil {
				types, err := printNamedTypes(param.Type(), printed, nil)
				if err != nil {
					return err
				}
				data.Interfaces = append(data.Interfaces, types...)
			}

			results := sig.Results()
			switch results.Len() {
			case 0:
				// do nothing
			case 1:
				r := results.At(0)
				if types.TypeString(r.Type(), nil) != "error" {
					result = r
				}
			case 2:
				if types.TypeString(results.At(1).Type(), nil) != "error" {
					continue // xxx skip this method
				}
				result = results.At(0)
			default:
				continue // xxx skip this method
			}

			if result != nil {
				types, err := printNamedTypes(result.Type(), printed, nil)
				if err != nil {
					return err
				}
				data.Interfaces = append(data.Interfaces, types...)
			}

			methods[fname] = methodInfo{param: param, result: result}
		}
	}

	names := maps.Keys(methods)
	sort.Strings(names)
	for _, name := range names {
		info := methods[name]
		m := tsMethod{SnakeName: toSnake(name)}
		if info.param != nil {
			m.ParamName = info.param.Name()
			m.ReqType = printType(info.param.Type())
		}
		if info.result != nil {
			m.RespType = printType(info.result.Type())
		}
		data.Methods = append(data.Methods, m)
	}

	return tmpl.Execute(w, data)
}

type methodInfo struct {
	param, result *types.Var
}

type (
	tsInterface struct {
		Name   string
		Fields []tsField
	}
	tsField struct {
		Name, Type string
	}
	tsMethod struct {
		SnakeName, ParamName, ReqType, RespType string
	}
	tsDecls struct {
		ClassName  string
		Interfaces []tsInterface
		Methods    []tsMethod
	}
)

// This traverses typ recursively,
// printing TypeScript declarations for named types it finds.
// Depended-on types are printed earlier than the types depending on them.
func printNamedTypes(typ types.Type, printed set.Of[string], namer func(*types.Named) string) ([]tsInterface, error) {
	key := types.TypeString(typ, nil)
	switch key {
	case "error", "context.Context":
		return nil, nil
	}

	if printed.Has(key) {
		return nil, nil
	}

	if namer == nil {
		namer = defaultNamedName
	}

	switch typ := typ.(type) {
	case *types.Array:
		return printNamedTypes(typ.Elem(), printed, namer)

	case *types.Basic:
		return nil, nil

	case *types.Chan:
		return printNamedTypes(typ.Elem(), printed, namer)

		// case *types.Interface:

	case *types.Map:
		types1, err := printNamedTypes(typ.Key(), printed, namer)
		if err != nil {
			return nil, err
		}
		types2, err := printNamedTypes(typ.Elem(), printed, namer)
		if err != nil {
			return nil, err
		}
		return append(types1, types2...), nil

	case *types.Named:
		printed.Add(key)
		u := typ.Underlying()
		result, err := printNamedTypes(u, printed, namer)
		if err != nil {
			return nil, err
		}
		if ptr, ok := u.(*types.Pointer); ok {
			u = ptr.Elem()
		}
		switch u := u.(type) {
		case *types.Struct:
			intf := tsInterface{
				Name: namer(typ),
			}
			for i := 0; i < u.NumFields(); i++ {
				intf.Fields = append(intf.Fields, printStructField(u, i))
			}
			result = append(result, intf)
			return result, nil

		default:
			return nil, fmt.Errorf("unexpected underlying type for %s: %T", key, u)
		}

	case *types.Pointer:
		return printNamedTypes(typ.Elem(), printed, namer)

	case *types.Signature:
		var result []tsInterface
		for i := 0; i < typ.Params().Len(); i++ {
			v := typ.Params().At(i)
			types, err := printNamedTypes(v.Type(), printed, namer)
			if err != nil {
				return nil, err
			}
			result = append(result, types...)
		}
		for i := 0; i < typ.Results().Len(); i++ {
			v := typ.Results().At(i)
			types, err := printNamedTypes(v.Type(), printed, namer)
			if err != nil {
				return nil, err
			}
			result = append(result, types...)
		}
		return result, nil

	case *types.Slice:
		return printNamedTypes(typ.Elem(), printed, namer)

	case *types.Struct:
		var result []tsInterface
		for i := 0; i < typ.NumFields(); i++ {
			v := typ.Field(i)
			types, err := printNamedTypes(v.Type(), printed, namer)
			if err != nil {
				return nil, err
			}
			result = append(result, types...)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unexpected type %T", typ)
	}
}

func printStructField(str *types.Struct, idx int) tsField {
	var (
		field = str.Field(idx)
		name  = field.Name()
		tag   = reflect.StructTag(str.Tag(idx))
	)
	if jsonTag, ok := tag.Lookup("json"); ok && jsonTag != "-" {
		parts := strings.Split(jsonTag, ",")
		name = parts[0]
	}
	return tsField{Name: name, Type: printType(field.Type())}
}

func defaultNamedName(n *types.Named) string {
	return n.Obj().Name()
}

func printType(typ types.Type) string {
	switch typ := typ.(type) {
	case *types.Array:
		return printType(typ.Elem()) + "[]"

	case *types.Basic:
		switch typ.Kind() {
		case types.Bool:
			return "boolean"
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
			return "number"
		case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			return "number"
		case types.Float32, types.Float64:
			return "number"
		case types.String:
			return "string"
		}

	case *types.Map:
		return fmt.Sprintf("{[x: %s]: %s}", printType(typ.Key()), printType(typ.Elem()))

	case *types.Named:
		return typ.Obj().Name() // xxx

	case *types.Pointer:
		return printType(typ.Elem())

	case *types.Slice:
		el := typ.Elem()
		if b, ok := el.Underlying().(*types.Basic); ok && b.Kind() == types.Byte {
			return "string"
		}
		return printType(typ.Elem()) + "[]"

		// xxx case *types.Struct:
	}

	return ""
}

func toSnake(inp string) string {
	parts := camelcase.Split(inp)
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.ToLower(parts[i])
	}
	return strings.Join(parts, "_")
}
