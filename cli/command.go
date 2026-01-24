package cli

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	errFlagMustStartWithDash  = errors.New("must start with '-'")
	errRequiredOrDefault      = errors.New("cannot be required and have a default")
	errUnsupportedSliceType   = errors.New("unsupported slice type")
	errUnsupportedType        = errors.New("unsupported type")
	errInvalidInt             = errors.New("expecting integer")
	errArgRequired            = errors.New("argument is required")
	errArgDescriptionRequired = errors.New("argument description missing")
	errHelpTextRequired       = errors.New("help text is missing")
)

type arg struct {
	name        string
	description string
	help        []string
	required    bool
	defaultVal  string
	isFlag      bool
	isParam     bool
	typ         string
}

// Command represents a CLI command configuration.
type Command struct {
	Name        string
	Alias       string
	Help        []string
	description string
	args        []arg
	Handler     Handler

	reflectType  reflect.Type
	structFields map[string]int // argName -> struct field index
}

// cacheReflection builds reflection metadata from struct tags.
// It inspects the c.Handler struct to identify arguments, flags, and parameters.
func (c *Command) cacheReflection() {
	c.structFields = make(map[string]int)
	c.args = nil

	val := reflect.ValueOf(c.Handler)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	c.reflectType = val.Type()

	for i := range c.reflectType.NumField() {
		f := c.reflectType.Field(i)

		argTag := f.Tag.Get("arg")
		if argTag == "" {
			continue
		}

		a := arg{name: argTag}

		// 1) bool => flag
		if f.Type.Kind() == reflect.Bool {
			a.isFlag = true
			a.typ = "bool"
		} else {
			// 2) non-flag + "-" prefix => named param
			if strings.HasPrefix(a.name, "-") {
				a.isParam = true
			}
			a.typ = strings.TrimPrefix(f.Type.String(), "*")
		}

		// 3) required = not pointer + not flag
		if f.Type.Kind() != reflect.Ptr && !a.isFlag {
			a.required = true
		}

		// default from tag, not required anymore
		if def := f.Tag.Get("default"); def != "" {
			a.defaultVal = def
			a.required = false
		}

		c.args = append(c.args, a)
		c.structFields[a.name] = i
	}
}

// prepareHelp processes the help text and extracts argument descriptions.
func (c *Command) prepareHelp() error {
	if len(c.Help) == 0 {
		return fmt.Errorf("command %s: %w", c.Name, errHelpTextRequired)
	}

	if len(c.args) == 0 {
		c.description = strings.Join(c.Help, "\n")
		return nil
	}

	argMap := make(map[string]*arg)
	for i, a := range c.args {
		argMap[a.name] = &c.args[i]
	}

	var commandLines []string
	var curArg *arg

iterateHelp:
	for _, line := range c.Help {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for name, ar := range argMap {
			prefix := name + ":"
			if strings.HasPrefix(line, prefix) {
				curArg = ar
				curArg.description = strings.TrimSpace(strings.TrimPrefix(line, prefix))
				curArg.help = make([]string, 0)
				continue iterateHelp
			}
		}

		if curArg == nil {
			commandLines = append(commandLines, line)
			continue
		}

		curArg.help = append(curArg.help, line)
	}

	for _, a := range c.args {
		if a.description == "" {
			return fmt.Errorf("%s: %w", a.name, errArgDescriptionRequired)
		}
	}

	c.description = strings.Join(commandLines, "\n")
	return nil
}

// prepareHandler binds arguments to the c.Handler struct and returns a Handler instance.
func (c *Command) prepareHandler(rawArgs []string) (Handler, error) {
	posArgs, flags, named := extractArgs(rawArgs, c.args)

	proto := reflect.ValueOf(c.Handler)
	if proto.Kind() == reflect.Ptr {
		proto = proto.Elem()
	}

	val := reflect.New(c.reflectType).Elem()

	val.Set(proto)

	filled := make(map[string]bool)

	for k, v := range named {
		if idx, ok := c.structFields[k]; ok {
			err := setFieldValue(val.Field(idx), v)
			if err != nil {
				return nil, fmt.Errorf("arg %s: %w", k, err)
			}
			filled[k] = true
		}
	}

	for name, present := range flags {
		if idx, ok := c.structFields[name]; ok {
			err := setFieldValue(val.Field(idx), strconv.FormatBool(present))
			if err != nil {
				return nil, fmt.Errorf("arg %s: %w", name, err)
			}
			filled[name] = true
		}
	}

	curr := 0
	for _, a := range c.args {
		if a.isFlag || a.isParam || filled[a.name] {
			continue
		}
		if curr < len(posArgs) {
			idx := c.structFields[a.name]
			err := setFieldValue(val.Field(idx), posArgs[curr])
			if err != nil {
				return nil, fmt.Errorf("arg %s: %w", a.name, err)
			}
			filled[a.name] = true
			curr++
		}
	}

	for _, a := range c.args {
		if !filled[a.name] {
			if a.required {
				return nil, fmt.Errorf("%s: %w. Use '? [command]' for help", a.name, errArgRequired)
			}
			if a.defaultVal != "" {
				idx := c.structFields[a.name]
				err := setFieldValue(val.Field(idx), a.defaultVal)
				if err != nil {
					return nil, fmt.Errorf("arg %s: %w", a.name, err)
				}
			}
		}
	}

	return val.Addr().Interface().(Handler), nil //nolint:forcetypeassert // c.Handler is a Handler by design
}

// extractArgs splits raw args into positional, flags, and named parameters.
func extractArgs(raw []string, args []arg) (pos []string, flags map[string]bool, named map[string]string) {
	flags = make(map[string]bool)
	named = make(map[string]string)

	paramNames := make(map[string]struct{})
	flagNames := make(map[string]struct{})

	for _, a := range args {
		if a.isParam {
			paramNames[a.name] = struct{}{}
		}
		if a.isFlag {
			flagNames[a.name] = struct{}{}
		}
	}

	for _, tok := range raw {
		// Named params: "{argName}=" must match a registered param name.
		if eq := strings.IndexByte(tok, '='); eq >= 0 {
			key := tok[:eq]
			if _, ok := paramNames[key]; ok {
				named[key] = tok[eq+1:]
				continue
			}
		}

		// Flags: token must match a registered flag name exactly.
		if _, ok := flagNames[tok]; ok {
			flags[tok] = true
			continue
		}

		// Otherwise positional value
		pos = append(pos, tok)
	}

	return
}

// setFieldValue assigns a string value to a reflected field, handling type conversion.
func setFieldValue(field reflect.Value, value string) error {
	ft := field.Type()

	if ft.Kind() == reflect.Ptr {
		v := reflect.New(ft.Elem())
		err := setFieldValue(v.Elem(), value)
		if err != nil {
			return err
		}
		field.Set(v)
		return nil
	}

	switch ft.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%w, got '%s'", errInvalidInt, value)
		}
		field.SetInt(v)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Slice:
		if ft.Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			slice := reflect.MakeSlice(ft, len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("%w: %s", errUnsupportedSliceType, ft.Elem().Kind())
		}
	default:
		return fmt.Errorf("%w: %s", errUnsupportedType, ft.Kind())
	}

	return nil
}

// validateCommand checks the command configuration for errors.
func validateCommand(c Command) error {
	for _, a := range c.args {
		if a.required && a.defaultVal != "" {
			return fmt.Errorf("flag %s: %w", a.name, errRequiredOrDefault)
		}

		if a.isFlag && !strings.HasPrefix(a.name, "-") {
			return fmt.Errorf("flag %s: %w", a.name, errFlagMustStartWithDash)
		}
	}

	return nil
}
