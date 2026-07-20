package botapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type GoInventory struct {
	Types     map[string]*GoType
	Functions []GoFunction
}

type GoType struct {
	Names      []string
	Struct     bool
	Interface  bool
	CustomJSON bool
	Alias      string
	Embeds     []string
	JSONFields map[string]GoField
}

type GoFunction struct {
	Parameters []GoParameter
	Strings    []string
}

type GoParameter struct {
	Name string
	Type string
}

type GoField struct {
	Name      string
	OmitEmpty bool
	Tagged    bool
	Type      string
}

type ScanOptions struct {
	ExcludeFiles map[string]bool
	Shallow      bool
}

type AuditReport struct {
	MissingMethodBindings       []string         `json:"missing_method_bindings"`
	MissingMethodParams         []MissingGap     `json:"missing_method_parameters"`
	MismatchedMethodOptionality []OptionalityGap `json:"mismatched_method_optionality"`
	MismatchedMethodTypes       []TypeGap        `json:"mismatched_method_types"`
	NonNilableMethodOptionals   []TypeGap        `json:"non_nilable_method_optionals"`
	MissingObjectTypes          []string         `json:"missing_object_types"`
	MissingObjectFields         []MissingGap     `json:"missing_object_fields"`
	MismatchedObjectOptionality []OptionalityGap `json:"mismatched_object_optionality"`
	MismatchedObjectTypes       []TypeGap        `json:"mismatched_object_types"`
	NonNilableObjectOptionals   []TypeGap        `json:"non_nilable_object_optionals"`
	MissingUnionTypes           []string         `json:"missing_union_types"`
	MissingUnionVariants        []MissingGap     `json:"missing_union_variants"`
}

type MissingGap struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type OptionalityGap struct {
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Required     bool   `json:"required"`
	HasOmitEmpty bool   `json:"has_omitempty"`
}

type TypeGap struct {
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

func ScanGo(root string, directories ...string) (GoInventory, error) {
	return ScanGoWithOptions(root, ScanOptions{}, directories...)
}

func ScanGoWithOptions(root string, options ScanOptions, directories ...string) (GoInventory, error) {
	result := GoInventory{Types: make(map[string]*GoType)}
	fileset := token.NewFileSet()
	for _, directory := range directories {
		base := filepath.Join(root, directory)
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if options.Shallow && path != base {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || options.ExcludeFiles[entry.Name()] {
				return nil
			}
			file, err := parser.ParseFile(fileset, path, nil, 0)
			if err != nil {
				return err
			}
			for _, declaration := range file.Decls {
				if function, ok := declaration.(*ast.FuncDecl); ok {
					result.addFunction(function)
					continue
				}
				general, ok := declaration.(*ast.GenDecl)
				if !ok || general.Tok != token.TYPE {
					continue
				}
				for _, item := range general.Specs {
					typeSpec, ok := item.(*ast.TypeSpec)
					if !ok {
						continue
					}
					key := NormalizeName(typeSpec.Name.Name)
					entry := result.Types[key]
					if entry == nil {
						entry = &GoType{JSONFields: make(map[string]GoField)}
						result.Types[key] = entry
					}
					entry.Names = appendUnique(entry.Names, typeSpec.Name.Name)
					switch value := typeSpec.Type.(type) {
					case *ast.StructType:
						entry.Struct = true
						for _, field := range value.Fields.List {
							if len(field.Names) == 0 {
								if embedded := expressionTypeName(field.Type); embedded != "" {
									entry.Embeds = appendUnique(entry.Embeds, NormalizeName(embedded))
								}
							}
							name, omitEmpty, ok := jsonField(field)
							if ok {
								entry.JSONFields[name] = GoField{
									Name: goFieldName(field), OmitEmpty: omitEmpty, Tagged: true, Type: expressionTypeShape(field.Type),
								}
							}
						}
					case *ast.InterfaceType:
						entry.Interface = true
					default:
						if alias := expressionTypeName(typeSpec.Type); alias != "" {
							entry.Alias = NormalizeName(alias)
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			return GoInventory{}, fmt.Errorf("scan %s: %w", base, err)
		}
	}
	for _, entry := range result.Types {
		sort.Strings(entry.Names)
		sort.Strings(entry.Embeds)
	}
	return result, nil
}

func (inventory *GoInventory) addFunction(function *ast.FuncDecl) {
	if function.Name.Name == "MarshalJSON" && function.Recv != nil && len(function.Recv.List) != 0 {
		if receiver := expressionTypeName(function.Recv.List[0].Type); receiver != "" {
			key := NormalizeName(receiver)
			entry := inventory.Types[key]
			if entry == nil {
				entry = &GoType{JSONFields: make(map[string]GoField)}
				inventory.Types[key] = entry
			}
			entry.CustomJSON = true
		}
	}
	if function.Body == nil {
		return
	}
	entry := GoFunction{}
	if function.Type.Params != nil {
		for _, field := range function.Type.Params.List {
			typeName := expressionTypeName(field.Type)
			if len(field.Names) == 0 {
				entry.Parameters = append(entry.Parameters, GoParameter{Type: typeName})
				continue
			}
			for _, name := range field.Names {
				entry.Parameters = append(entry.Parameters, GoParameter{Name: name.Name, Type: typeName})
			}
		}
	}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err == nil {
			entry.Strings = appendUnique(entry.Strings, value)
		}
		return true
	})
	if len(entry.Strings) != 0 {
		inventory.Functions = append(inventory.Functions, entry)
	}
}

func (inventory GoInventory) addTypeFields(result map[string]struct{}, key string, visiting map[string]bool) {
	fields := make(map[string]GoField)
	inventory.addTypeFieldInfo(fields, key, visiting)
	for name := range fields {
		result[name] = struct{}{}
	}
}

func (inventory GoInventory) addTypeFieldInfo(result map[string]GoField, key string, visiting map[string]bool) {
	if key == "" || visiting[key] {
		return
	}
	entry := inventory.Types[key]
	if entry == nil {
		return
	}
	visiting[key] = true
	for name, field := range entry.JSONFields {
		result[name] = field
	}
	inventory.addTypeFieldInfo(result, entry.Alias, visiting)
	for _, embedded := range entry.Embeds {
		inventory.addTypeFieldInfo(result, embedded, visiting)
	}
	delete(visiting, key)
}

func Audit(schema Manifest, inventory GoInventory) AuditReport {
	var report AuditReport
	for _, method := range schema.Methods {
		covered := make(map[string]GoField)
		inventory.addTypeFieldInfo(covered, NormalizeName(exportedName(method.Name)+"Params"), make(map[string]bool))
		bound := false
		for _, function := range inventory.Functions {
			if !contains(function.Strings, method.Name) {
				continue
			}
			bound = true
			for _, parameter := range function.Parameters {
				if parameter.Type != "" {
					inventory.addTypeFieldInfo(covered, NormalizeName(parameter.Type), make(map[string]bool))
				}
				if parameter.Name != "" {
					covered[snakeCase(parameter.Name)] = GoField{}
				}
			}
			for _, literal := range function.Strings {
				covered[literal] = GoField{}
			}
		}
		if !bound {
			report.MissingMethodBindings = append(report.MissingMethodBindings, method.Name)
		}
		for _, parameter := range method.Parameters {
			field, exists := covered[parameter.Name]
			if !exists {
				report.MissingMethodParams = append(report.MissingMethodParams, MissingGap{Owner: method.Name, Name: parameter.Name})
			} else if field.Tagged && parameter.Required == field.OmitEmpty {
				report.MismatchedMethodOptionality = append(report.MismatchedMethodOptionality, OptionalityGap{
					Owner: method.Name, Name: parameter.Name, Required: parameter.Required, HasOmitEmpty: field.OmitEmpty,
				})
			}
			if exists && field.Tagged && !fieldTypeMatches(parameter.Type, field.Type) {
				report.MismatchedMethodTypes = append(report.MismatchedMethodTypes, TypeGap{
					Owner: method.Name, Name: parameter.Name, Expected: parameter.Type, Actual: field.Type,
				})
			}
			if exists && field.Tagged && !parameter.Required && optionalNeedsNilable(parameter.Type) && !goTypeNilable(field.Type, inventory) {
				report.NonNilableMethodOptionals = append(report.NonNilableMethodOptionals, TypeGap{
					Owner: method.Name, Name: parameter.Name, Expected: "nilable " + parameter.Type, Actual: field.Type,
				})
			}
		}
	}
	for _, object := range schema.Objects {
		entry := inventory.Types[NormalizeName(object.Name)]
		if entry == nil || (!entry.Struct && len(object.Fields) != 0) {
			report.MissingObjectTypes = append(report.MissingObjectTypes, object.Name)
			continue
		}
		covered := make(map[string]GoField)
		inventory.addTypeFieldInfo(covered, NormalizeName(object.Name), make(map[string]bool))
		for _, field := range object.Fields {
			goField, exists := covered[field.Name]
			if !exists && (field.Name == "type" || field.Name == "source") && entry.CustomJSON {
				exists = true
			}
			if !exists {
				report.MissingObjectFields = append(report.MissingObjectFields, MissingGap{Owner: object.Name, Name: field.Name})
			} else if !(entry.CustomJSON && (field.Name == "type" || field.Name == "source")) && field.Required == goField.OmitEmpty {
				report.MismatchedObjectOptionality = append(report.MismatchedObjectOptionality, OptionalityGap{
					Owner: object.Name, Name: field.Name, Required: field.Required, HasOmitEmpty: goField.OmitEmpty,
				})
			}
			if exists && !(entry.CustomJSON && (field.Name == "type" || field.Name == "source")) && !fieldTypeMatches(field.Type, goField.Type) {
				report.MismatchedObjectTypes = append(report.MismatchedObjectTypes, TypeGap{
					Owner: object.Name, Name: field.Name, Expected: field.Type, Actual: goField.Type,
				})
			}
			if exists && !field.Required && optionalNeedsNilable(field.Type) && !goTypeNilable(goField.Type, inventory) {
				report.NonNilableObjectOptionals = append(report.NonNilableObjectOptionals, TypeGap{
					Owner: object.Name, Name: field.Name, Expected: "nilable " + field.Type, Actual: goField.Type,
				})
			}
		}
	}
	for _, union := range schema.Unions {
		if _, exists := inventory.Types[NormalizeName(union.Name)]; !exists {
			report.MissingUnionTypes = append(report.MissingUnionTypes, union.Name)
		}
		for _, variant := range union.Variants {
			if _, exists := inventory.Types[NormalizeName(variant)]; !exists {
				report.MissingUnionVariants = append(report.MissingUnionVariants, MissingGap{Owner: union.Name, Name: variant})
			}
		}
	}
	report.sort()
	return report
}

func (report AuditReport) GapCount() int {
	return len(report.MissingMethodBindings) + len(report.MissingMethodParams) + len(report.MismatchedMethodOptionality) + len(report.MismatchedMethodTypes) + len(report.NonNilableMethodOptionals) +
		len(report.MissingObjectTypes) + len(report.MissingObjectFields) + len(report.MismatchedObjectOptionality) + len(report.MismatchedObjectTypes) + len(report.NonNilableObjectOptionals) +
		len(report.MissingUnionTypes) + len(report.MissingUnionVariants)
}

func NormalizeName(name string) string {
	var result strings.Builder
	result.Grow(len(name))
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}

func exportedName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func expressionTypeName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return expressionTypeName(value.X)
	case *ast.Ellipsis:
		return expressionTypeName(value.Elt)
	case *ast.IndexExpr:
		return expressionTypeName(value.X)
	case *ast.IndexListExpr:
		return expressionTypeName(value.X)
	case *ast.SelectorExpr:
		return value.Sel.Name
	default:
		return ""
	}
}

func expressionTypeShape(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return "*" + expressionTypeShape(value.X)
	case *ast.ArrayType:
		return "[]" + expressionTypeShape(value.Elt)
	case *ast.SelectorExpr:
		return value.Sel.Name
	case *ast.InterfaceType:
		if value.Methods == nil || len(value.Methods.List) == 0 {
			return "any"
		}
		return "interface"
	case *ast.MapType:
		return "map[" + expressionTypeShape(value.Key) + "]" + expressionTypeShape(value.Value)
	default:
		return ""
	}
}

func fieldTypeMatches(expected, actual string) bool {
	actual = strings.TrimPrefix(actual, "*")
	const array = "Array of "
	if strings.HasPrefix(expected, array) {
		return strings.HasPrefix(actual, "[]") && fieldTypeMatches(strings.TrimPrefix(expected, array), strings.TrimPrefix(actual, "[]"))
	}
	if strings.HasPrefix(actual, "[]") {
		return false
	}
	switch expected {
	case "Boolean", "True":
		return actual == "bool"
	case "Float", "Float number":
		return actual == "float64"
	case "Integer":
		return actual == "int" || actual == "int64"
	case "String":
		return actual == "string"
	case "Integer or String":
		return actual == "any" || NormalizeName(actual) == "chatid"
	case "InputMediaAnimation or InputMediaAudio or InputMediaPhoto or InputMediaVideo or InputMediaVoiceNote":
		return NormalizeName(actual) == "richmessagemedia"
	case "InlineKeyboardMarkup or ReplyKeyboardMarkup or ReplyKeyboardRemove or ForceReply":
		return NormalizeName(actual) == "replymarkup"
	default:
		return NormalizeName(expected) == NormalizeName(actual)
	}
}

func optionalNeedsNilable(expected string) bool {
	if strings.HasPrefix(expected, "Array of ") || strings.Contains(expected, " or ") {
		return false
	}
	switch expected {
	case "Boolean", "True", "Float", "Float number", "Integer", "String":
		return false
	default:
		return true
	}
}

func goTypeNilable(actual string, inventory GoInventory) bool {
	if strings.HasPrefix(actual, "*") || strings.HasPrefix(actual, "[]") || strings.HasPrefix(actual, "map[") || actual == "any" || actual == "interface" {
		return true
	}
	entry := inventory.Types[NormalizeName(actual)]
	return entry != nil && (entry.Interface || entry.Alias == "any" || entry.Alias == "interface")
}

func snakeCase(name string) string {
	runes := []rune(name)
	var result strings.Builder
	result.Grow(len(name) + 4)
	for index, r := range runes {
		if unicode.IsUpper(r) && index > 0 {
			previousLower := unicode.IsLower(runes[index-1]) || unicode.IsDigit(runes[index-1])
			nextLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			if previousLower || nextLower {
				result.WriteByte('_')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func contains(values []string, value string) bool {
	for _, current := range values {
		if current == value {
			return true
		}
	}
	return false
}

func jsonField(field *ast.Field) (string, bool, bool) {
	if field.Tag == nil {
		return "", false, false
	}
	tag, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return "", false, false
	}
	value := reflect.StructTag(tag).Get("json")
	if value == "" {
		return "", false, false
	}
	parts := strings.Split(value, ",")
	if parts[0] == "" || parts[0] == "-" {
		return "", false, false
	}
	omitEmpty := false
	for _, option := range parts[1:] {
		omitEmpty = omitEmpty || option == "omitempty"
	}
	return parts[0], omitEmpty, true
}

func goFieldName(field *ast.Field) string {
	if len(field.Names) == 0 {
		return ""
	}
	return field.Names[0].Name
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func (report *AuditReport) sort() {
	sort.Strings(report.MissingMethodBindings)
	sort.Strings(report.MissingObjectTypes)
	sort.Strings(report.MissingUnionTypes)
	sort.Slice(report.MismatchedMethodOptionality, func(i, j int) bool {
		if report.MismatchedMethodOptionality[i].Owner == report.MismatchedMethodOptionality[j].Owner {
			return report.MismatchedMethodOptionality[i].Name < report.MismatchedMethodOptionality[j].Name
		}
		return report.MismatchedMethodOptionality[i].Owner < report.MismatchedMethodOptionality[j].Owner
	})
	sort.Slice(report.MismatchedMethodTypes, func(i, j int) bool {
		if report.MismatchedMethodTypes[i].Owner == report.MismatchedMethodTypes[j].Owner {
			return report.MismatchedMethodTypes[i].Name < report.MismatchedMethodTypes[j].Name
		}
		return report.MismatchedMethodTypes[i].Owner < report.MismatchedMethodTypes[j].Owner
	})
	sort.Slice(report.NonNilableMethodOptionals, func(i, j int) bool {
		if report.NonNilableMethodOptionals[i].Owner == report.NonNilableMethodOptionals[j].Owner {
			return report.NonNilableMethodOptionals[i].Name < report.NonNilableMethodOptionals[j].Name
		}
		return report.NonNilableMethodOptionals[i].Owner < report.NonNilableMethodOptionals[j].Owner
	})
	sort.Slice(report.MismatchedObjectOptionality, func(i, j int) bool {
		if report.MismatchedObjectOptionality[i].Owner == report.MismatchedObjectOptionality[j].Owner {
			return report.MismatchedObjectOptionality[i].Name < report.MismatchedObjectOptionality[j].Name
		}
		return report.MismatchedObjectOptionality[i].Owner < report.MismatchedObjectOptionality[j].Owner
	})
	sort.Slice(report.MismatchedObjectTypes, func(i, j int) bool {
		if report.MismatchedObjectTypes[i].Owner == report.MismatchedObjectTypes[j].Owner {
			return report.MismatchedObjectTypes[i].Name < report.MismatchedObjectTypes[j].Name
		}
		return report.MismatchedObjectTypes[i].Owner < report.MismatchedObjectTypes[j].Owner
	})
	sort.Slice(report.NonNilableObjectOptionals, func(i, j int) bool {
		if report.NonNilableObjectOptionals[i].Owner == report.NonNilableObjectOptionals[j].Owner {
			return report.NonNilableObjectOptionals[i].Name < report.NonNilableObjectOptionals[j].Name
		}
		return report.NonNilableObjectOptionals[i].Owner < report.NonNilableObjectOptionals[j].Owner
	})
	sort.Slice(report.MissingMethodParams, func(i, j int) bool {
		if report.MissingMethodParams[i].Owner == report.MissingMethodParams[j].Owner {
			return report.MissingMethodParams[i].Name < report.MissingMethodParams[j].Name
		}
		return report.MissingMethodParams[i].Owner < report.MissingMethodParams[j].Owner
	})
	sort.Slice(report.MissingObjectFields, func(i, j int) bool {
		if report.MissingObjectFields[i].Owner == report.MissingObjectFields[j].Owner {
			return report.MissingObjectFields[i].Name < report.MissingObjectFields[j].Name
		}
		return report.MissingObjectFields[i].Owner < report.MissingObjectFields[j].Owner
	})
	sort.Slice(report.MissingUnionVariants, func(i, j int) bool {
		if report.MissingUnionVariants[i].Owner == report.MissingUnionVariants[j].Owner {
			return report.MissingUnionVariants[i].Name < report.MissingUnionVariants[j].Name
		}
		return report.MissingUnionVariants[i].Owner < report.MissingUnionVariants[j].Owner
	})
}
