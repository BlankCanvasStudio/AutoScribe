package directives;

import (
    // "os"
    "fmt"
    "strings"
    "reflect"
    // "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    // "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/calls"
    // "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
    // "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/formatting"
)


func InitDirectiveInLocalScope(directive types.Directive, toFocus []string, configFiles []string) error {
    log.Debugf("Looking for directive: %v", directive.Name)
    log.Debugf("Looking at configs: %v", configFiles)

    config.PushLoadedConfig()

    {

        config.LoadConfigFile(config.ProjectConfigFile)

        _, exists := config.Settings.Directives[directive.Name]
        if exists {
            log.Infof("Directive %v already initialized", directive.Name)
            return nil
        }

    }

    err := config.PopLoadedConfig()
    if err != nil {
        return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
    }


    // Iterate to move "up" in scope
    for i := 0; i < len(configFiles); i++ {
        config.PushLoadedConfig()

        {

            // Load those settings cofigs
            config.LoadConfigFile(configFiles[i])

            // Verify that directive exists
            toLoad, exists := config.Settings.Directives[directive.Name]
            if !exists {
                log.Debugf("Directive %v not defined in %v", directive.Name, configFiles[i])
                err := config.PopLoadedConfig()
                if err != nil {
                    return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
                }

                continue
            }

            config.PushLoadedConfig()

            {

                // Load the config to append to
                config.LoadConfigFile(config.ProjectConfigFile)

                // load things to focus for ease of use
                for _, arg := range toFocus {
                    toLoad.Focus = append(toLoad.Focus, arg)
                }

                config.Settings.Directives[directive.Name] = toLoad

                err = config.SaveConfigFile(config.ProjectConfigFile, config.Settings)
                if err != nil {
                    return fmt.Errorf("Failed to save to config %v: %v", configFiles[i], err)
                }

            }

            err := config.PopLoadedConfig()
            if err != nil {
                return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
            }

        }

        err := config.PopLoadedConfig()
        if err != nil {
            return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
        }

        log.Infof("Initialized %v from %v", directive.Name, configFiles[i])

        return nil
    }

    return fmt.Errorf("Couldn't load directive %v not found in configs: %v", directive, configFiles)
}


func ExportCustomDirective(directive types.Directive, configFiles []string) error {

    toAdd := config.Settings.Directives[directive.Name]

    // Idk if keeping the focus & ignore is the play
    toAdd.Focus = nil
    toAdd.Ignore = nil
    toAdd.Model = types.NoModel
    toAdd.ApiKey = ""
    toAdd.Scope = ""

    // Iterate to move "up" in scope
    for _, configFile := range configFiles {
        config.PushLoadedConfig()
        {

            config.LoadConfigFile(configFile)

            config.Settings.Directives[toAdd.Name] = toAdd

            err := config.SaveConfigFile(configFile, config.Settings)
            if err != nil {
                return fmt.Errorf("Failed to create new directive: %v", err)
            }
        }
    }

    return nil
}


func SetField(obj any, name string, value any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("obj must be a non-nil pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("obj must point to a struct")
	}

	f := v.FieldByName(name)
	if !f.IsValid() {
		return fmt.Errorf("no such field: %s", name)
	}
	if !f.CanSet() {
		return fmt.Errorf("cannot set field %s (unexported?)", name)
	}

	val := reflect.ValueOf(value)

	// Handle pointer fields by accepting both T and *T
	if f.Kind() == reflect.Pointer {
		ft := f.Type().Elem()
		if val.IsValid() && val.Type().AssignableTo(ft) {
			ptr := reflect.New(ft)
			ptr.Elem().Set(val)
			f.Set(ptr)
			return nil
		}
		if val.IsValid() && val.Type().AssignableTo(f.Type()) {
			f.Set(val)
			return nil
		}
	}

	// Direct assign or convertible types (e.g., int->int64, float64->int, etc.)
	if val.IsValid() && val.Type().AssignableTo(f.Type()) {
		f.Set(val)
		return nil
	}
	if val.IsValid() && val.Type().ConvertibleTo(f.Type()) {
		f.Set(val.Convert(f.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign %s to field %s of type %s", val.Type(), name, f.Type())
}


func SetArray(obj any, name string, value any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("obj must be a non-nil pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("obj must point to a struct")
	}

	f := v.FieldByName(name)
	if !f.IsValid() {
		return fmt.Errorf("no such field: %s", name)
	}
	if !f.CanSet() {
		return fmt.Errorf("cannot set field %s (unexported?)", name)
	}

	val := reflect.ValueOf(value)

        if f.Kind() != reflect.Array && f.Kind() != reflect.Slice {
            return fmt.Errorf("Member variable %v is not an array; is a: %v", name, f.String())
        }

	// Handle pointer fields by accepting both T and *T
        if f.Kind() == reflect.Array || f.Kind() == reflect.Slice {
            exists := false
            for i := 0; i < f.Len(); i++ {
                if reflect.DeepEqual(f.Index(i).Interface(), val.Interface()) {
                    exists = true
                    break
                }
            }
            if !exists {
                f.Set(reflect.Append(f, val))
            }
            return nil
        }

	return fmt.Errorf("cannot assign %s to field %s of type %s", val.Type(), name, f.Type())
}


func UpdateDirectiveFieldInConfigs(directive types.Directive, field string, arg string, configFiles []string) error {
    log.Debugf("Saving to config files: %v", configFiles)

    d := config.Settings.Directives[directive.Name]

    err := SetField(&d, field, types.DirectiveType(strings.ToLower(arg)))
    if err != nil {
        return fmt.Errorf("Failed to set field %v on directive %v: %v", field, directive.Name, err)
    }

    err = d.CheckForSave()
    if err != nil {
        return fmt.Errorf("can't update directive %v: %v", field, err)
    }

    config.Settings.Directives[directive.Name] = d

    for _, configFile := range configFiles {
        log.Debugf("Updating settings in: %v", configFile)
        config.PushLoadedConfig()
        {

            config.LoadConfigFile(configFile)

            d := config.Settings.Directives[directive.Name]

            err := SetField(&d, field, types.DirectiveType(strings.ToLower(arg)))
            if err != nil {
                return fmt.Errorf("Failed to set field %v on directive %v: %v", field, directive.Name, err)
            }

            config.Settings.Directives[directive.Name] = d

            err = config.SaveConfigFile(configFile, config.Settings)
            if err != nil {
                return fmt.Errorf("Failed to create new directive: %v", err)
            }
        }
        config.PopLoadedConfig()
    }

    return nil
}


func UpdateDirectiveArrayInConfigs(directive types.Directive, field string, args []string, configFiles []string) error {
    log.Debugf("Saving to config files: %v", configFiles)

    d := config.Settings.Directives[directive.Name]

    for _, arg := range args {
        err := SetArray(&d, field, arg)
        if err != nil {
            return fmt.Errorf("failed to update %v array with %v: %v", field, arg, err)
        }
    }

    err := d.CheckForSave()
    if err != nil {
        return fmt.Errorf("can't update directive: %v", err)
    }

    config.Settings.Directives[directive.Name] = d

    savedConfig := config.Settings

    for _, configFile := range configFiles {
        log.Debugf("Updating settings in: %v", configFile)
        config.Settings = config.NewConfig()

        config.LoadConfigFile(configFile)

        d := config.Settings.Directives[directive.Name]

        for _, arg := range args {
            err := SetArray(&d, field, arg)
            if err != nil {
                return fmt.Errorf("failed to update %v array with %v: %v", field, arg, err)
            }
        }

        config.Settings.Directives[directive.Name] = d

        err := config.SaveConfigFile(configFile, config.Settings)
        if err != nil {
            return fmt.Errorf("Failed to create new directive: %v", err)
        }
    }

    config.Settings = savedConfig

    return nil
}

