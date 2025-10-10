package directives

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

/*
Summary: Initializes a directive within the local configuration scope by examining the provided configFiles, loading and merging
configuration data, and persisting the updated directive into the project configuration. Use when you need to ensure a given
directive is established in the current local scope based on existing configuration files.

Signature: func InitDirectiveInLocalScope(directive types.Directive, toFocus []string, configFiles []string) error

Parameters:
- directive: types.Directive — the directive to initialize in the local scope.
- toFocus: []string — additional focus items to attach to the directive during processing.
- configFiles: []string — ordered list of configuration file paths to consult for initialization.

Returns:
- error — non-nil on failure; nil on success.

Errors/Exceptions:
- Returns an error if the directive cannot be loaded from any of the configFiles.
- Propagates errors from config operations (PushLoadedConfig, LoadConfigFile, PopLoadedConfig, SaveConfigFile) as encountered.

Side Effects:
- Mutates global configuration state via config.PushLoadedConfig, config.LoadConfigFile, and config.SaveConfigFile.
- Updates config.Settings.Directives and may modify config.Settings.Files and directive.Scope.
- May write updated project configuration to disk.

Edge Cases & Assumptions:
- Assumes config.Settings.Directives uses directive.Name as the key and supports updates/merges via directive operations.
- If the directive already exists in the current settings, the function may exit early with no changes.
- Relies on the presence and behavior of config.ProjectConfigFile for project-scope persistence.

*/
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

/*
Summary: Exports a custom directive by normalizing its fields and persisting it across the provided config files. It resets certain fields on the directive copy, then for each config file it pushes the current Settings onto the ConfigStack, loads the file into Settings, stores the updated directive under Settings.Directives[directive.Name], and saves the updated Settings to the file.
Signature: func ExportCustomDirective(directive types.Directive, configFiles []string) error
Parameters:
- directive: types.Directive - the directive to export; its Name identifies the entry in Settings.Directives to update.
- configFiles: []string - paths to config files to which the updated Settings will be written.
Returns:
- error: non-nil on failure; errors from SaveConfigFile are wrapped as "Failed to create new directive: %v".
Errors/Exceptions:
- Non-nil if config.SaveConfigFile returns an error (wrapped as described above).
Side Effects:
- Mutates global config.Settings (including config.Settings.Directives) and per-file directive scope.
- Calls config.PushLoadedConfig(), config.LoadConfigFile(string), and config.SaveConfigFile(string, Config).
Edge Cases & Assumptions:
- Assumes directive.Name exists in config.Settings.Directives; the directive is read, then its Focus, Ignore, Model, ApiKey, and Scope are reset.
- If configFiles is empty, no files are written and the function returns nil.
- Assumes the Config type can be serialized to YAML by SaveConfigFile and that updating Settings.Directives persists as intended.

*/
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

/*
Summary: SetField sets the field named by name on the struct pointed to by obj to value, using reflection. It supports both direct and pointer fields and will assign or convert values when possible. Use it to set fields by name at runtime.
Signature: func SetField(obj any, name string, value any) error
Parameters:
- obj: any; must be a non-nil pointer to a struct
- name: string; name of the exported field to set
- value: any; new value to assign; may be of the field's type, a convertible type, or for pointer fields, the element type or a *T matching the field type
Returns: error; nil on success, non-nil if the field cannot be set or inputs are invalid
Errors/Exceptions:
- "obj must be a non-nil pointer to a struct"
- "obj must point to a struct"
- "no such field: %s"
- "cannot set field %s (unexported?)"
- "cannot assign %s to field %s of type %s"
Side Effects: Mutates the field of the provided obj via reflection; may allocate a new pointer for pointer fields.
Edge Cases & Assumptions:
- obj must be a non-nil pointer to a struct; otherwise an error is returned
- The target field must be exported and settable
- For pointer fields, value may be a T (assigned by constructing a new pointer) or a *T; otherwise assignment may fail
- If value is invalid, not assignable, or not convertible to the field's type, an error is returned
- If the field is found but cannot be set, an error is returned

*/
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

/*
Summary:
Sets the value of an exported array or slice field (by name) on a struct pointed to by obj. If the field does not already contain the provided value (as determined by DeepEqual), the value is appended. Use to grow a collection field on a struct via reflection.

Signature:
func SetArray(obj any, name string, value any) error

Parameters:
- obj: any; a non-nil pointer to a struct. The function requires obj to be a pointer to a struct; otherwise an error is returned.
- name: string; the name of the field on the struct to modify. The field must be exportable and addressable.
- value: any; the value to append to the field if it is not already present. The field must be of kind array or slice.

Returns:
- error: non-nil on failure; nil on success (when the value is appended or already present).

Errors/Exceptions:
- "obj must be a non-nil pointer to a struct" if obj is not a non-nil pointer to a struct.
- "obj must point to a struct" if obj points to a non-struct value.
- "no such field: %s" if the field named by name does not exist.
- "cannot set field %s (unexported?)" if the field cannot be set (likely unexported).
- "Member variable %v is not an array; is a: %v" if the field is not an array or slice.
- "cannot assign %s to field %s of type %s" if the value cannot be assigned to the field element type.

Side Effects:
- Mutates the target struct's field by potentially appending value via reflection.
- May allocate memory when appending to the field.

Edge Cases & Assumptions:
- The function assumes the target field is either an array or a slice and attempts to append using reflect.Append; for arrays, this may be semantically inconsistent with the fixed length of arrays (runtime behavior depends on reflection). The behavior is determined by the code path that appends when the field is array or slice.
- Existing elements are detected with DeepEqual; the value is only appended if an equal element is not already present.
- The code includes the comment: "Handle pointer fields by accepting both T and *T" which indicates an intended flexibility, but the actual handling is determined by reflect on the field's element type.

*/
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

/*
Summary: Updates a field on a Directive and persists the change to each config file specified in configFiles. It uses reflection to set the field, validates the Directive, and then saves the updated configuration back to disk for each file.

Signature: func UpdateDirectiveFieldInConfigs(directive types.Directive, field string, arg string, configFiles []string) error

Parameters:
- directive: types.Directive; the directive to update (identified by directive.Name)
- field: string; the exported field name on the Directive to set
- arg: string; the new value for the field; converted to types.DirectiveType(strings.ToLower(arg))
- configFiles: []string; list of config file paths to update and persist

Returns:
- error; nil on success; non-nil if an update fails or a config file cannot be saved

Errors/Exceptions:
- "can't update directive %v: %v" when CheckForSave() fails after updating the field
- "Failed to set field %v on directive %v: %v" when SetField() cannot assign the value
- "Failed to create new directive: %v" if SaveConfigFile() fails for any config file

Side Effects:
- Mutates config.Settings.Directives[directive.Name]
- For each configFiles entry: pushes the current config onto the stack (config.PushLoadedConfig()), loads the file (config.LoadConfigFile), updates the directive in memory, and saves the updated config (config.SaveConfigFile)
- Logs debug information about saving and updating config files

Edge Cases & Assumptions:
- If configFiles is empty, the function updates the in-memory directive and may perform no file I/O
- SetField requires a valid, exported field name that is settable; the field is updated via reflection
- arg is lowercased and converted to types.DirectiveType for the field value
- LoadConfigFile may be a no-op if the file does not exist
- The function assumes config.Settings and directive mappings are initialized and available

*/
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

/*
Summary:
UpdateDirectiveArrayInConfigs updates the value of an exported array or slice field (by name) on a Directive
and persists these changes across the provided configFiles. It validates the directive for saving once,
then, for each config file, reloads the file, appends the given args to the specified field (without duplicates),
and saves the updated configuration back to disk. The in-memory Settings are restored after processing all files.

Signature:
func UpdateDirectiveArrayInConfigs(directive types.Directive, field string, args []string, configFiles []string) error

Parameters:
- directive: types.Directive; the directive to update. Used to locate the directive within config.Settings.Directives.
- field: string; the name of the exported array or slice field on the directive to modify.
- args: []string; values to append to the field if not already present.
- configFiles: []string; paths to config files to update.

Returns:
- error: non-nil on failure (wrapped with context). Nil if all updates and saves succeed.

Errors/Exceptions:
- Returns an error wrapping the result of d.CheckForSave() as "can't update directive: %v" if validation fails.
- Returns an error wrapping failures from SetArray when updating the array field within the directive.
- Returns an error wrapping failures from config.SaveConfigFile for any configFile as "Failed to create new directive: %v".

Side Effects:
- Mutates the in-memory config.Settings.Directives[directive.Name] and writes updated configurations to each configFile.
- May allocate memory when appending to the field during SetArray.
- Resets config.Settings per file during processing and restores it afterwards.

Edge Cases & Assumptions:
- The function validates the directive before applying changes to files; if validation fails, no changes are written.
- For each configFile, the function resets in-memory configuration (config.Settings = config.NewConfig()) and loads the file before applying updates.
- Existing elements are detected with DeepEqual; a value is appended only if it is not already present.
- Assumes SetArray correctly handles pointer/addressable field access and works with array or slice fields.

*/
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
