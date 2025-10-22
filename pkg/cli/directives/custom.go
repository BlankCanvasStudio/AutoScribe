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
Summary: Initialize a directive within the local configuration scope by loading, merging, and
persisting directive data from a sequence of config files. The function searches the provided
configFiles for a directive with the given name, augments its Focus with toFocus, and saves the
updated directive to config.ProjectConfigFile. Use this to constrain a directive to a set of
configurations while preserving global settings.

Signature: func InitDirectiveInLocalScope(directive types.Directive, toFocus []string, configFiles []string) error

Parameters:
- directive: types.Directive — the directive to initialize and scope locally.
- toFocus: []string — additional focus items to append to the directive's Focus when saving.
- configFiles: []string — ordered config file paths to search for the directive.

Returns:
- error: nil on success; non-nil if initialization fails (e.g., directive not found in provided configs,
  or saving the updated configuration fails).

Errors/Exceptions:
- Non-nil errors returned if pushing/popping loaded configs or saving the config fail.
- If the directive cannot be found in any of the configFiles, returns an error indicating the failure.

Side Effects:
- Mutates package-level ConfigStack and Settings via PushLoadedConfig, LoadConfigFile, and PopLoadedConfig.
- Writes the updated Settings to config.ProjectConfigFile via SaveConfigFile.
- Logs progress and results to the logger (e.g., Debug/Info messages).

Edge Cases & Assumptions:
- If the directiveName is already initialized in the initial ProjectConfigFile, the function returns nil
  without changes.
- Assumes ConfigStack and Settings are package-global identifiers; config.ProjectConfigFile is the target
  persistence file.
- If a configFile does not define the directive, the function continues with the next file; if none do, it
  returns an error.

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
Summary: Exports a given Directive into a set of configuration files by loading each file into the global Settings, normalizing the directive, and persisting the updated directive to each file. For each configFile, it pushes the currently loaded config, loads the file, then stores the updated directive under toAdd.Name and saves the full Settings to the file.

Signature: func ExportCustomDirective(directive types.Directive, configFiles []string) error

Parameters:
- directive: types.Directive — the directive to export; its corresponding entry in config.Settings.Directives is used as the base.
- configFiles: []string — list of config file paths to which the directive should be exported.

Returns:
- error: non-nil if persisting to any configFile fails; nil if all files are updated successfully.

Errors/Exceptions:
- If config.SaveConfigFile(configFile, config.Settings) fails for any configFile, returns an error of the form "Failed to create new directive: %v".

Side Effects:
- Mutates package-level state via config.PushLoadedConfig and config.LoadConfigFile (per file) and config.Settings.Directives.
- Writes YAML configuration to disk for each configFile.

Edge Cases & Assumptions:
- If a config file does not exist, LoadConfigFile behaves as a no-op, allowing iteration to continue.
- If directive.Name is not present in config.Settings.Directives, toAdd will be the zero-value Directive; the code stores it using toAdd.Name as the key.
- toAdd.Focus, toAdd.Ignore, toAdd.Model, toAdd.ApiKey, and toAdd.Scope are reset before export.

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
Summary: SetField uses reflection to assign a value to the field named by name on the struct pointed to by obj.
Use when you need to set a struct field by name at runtime, with type-safety checks.
Signature: func SetField(obj any, name string, value any) error
Parameters:
  obj: any - a non-nil pointer to a struct to modify.
  name: string - the field name to set (must be exported and valid for the struct).
  value: any - the value to assign to the field; for pointer fields, a value of the element type or a *element-type is accepted.
Returns: error - non-nil if the field cannot be found, is unexported, or cannot be assigned to.
Errors/Exceptions:
  - "obj must be a non-nil pointer to a struct" if obj is not a non-nil pointer to a struct.
  - "obj must point to a struct" if obj does not reference a struct.
  - "no such field: %s" if the field name does not exist.
  - "cannot set field %s (unexported?)" if the field is not settable.
  - "cannot assign %s to field %s of type %s" if value cannot be assigned or converted to the field type.
Side Effects: Mutates the field on obj.
Edge Cases & Assumptions:
  - For pointer fields, accepts both T and *T for value.
  - If value is not valid or cannot be assigned/converted, an error is returned.
  - Only exported fields can be set via reflection; unexported fields produce an error.

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
Summary: Sets or appends a value to a struct field by name when the field is an array or slice. If the value is not already present, it appends it. Use when you need to mutate a slice/array field of a struct via reflection.

Signature:
func SetArray(obj any, name string, value any) error

Parameters:
- obj: any — a non-nil pointer to a struct whose field will be modified.
- name: string — the name of the field to modify.
- value: any — the value to append to the field if not already present.

Returns:
- error — non-nil on failure; nil on success.

Errors/Exceptions:
- "obj must be a non-nil pointer to a struct" if obj is not a non-nil pointer to a struct.
- "obj must point to a struct" if obj does not point to a struct.
- "no such field: %s" if the struct has no field named name.
- "cannot set field %s (unexported?)" if the field cannot be set.
- "Member variable %v is not an array; is a: %v" if the field is not an array or slice.
- "cannot assign %s to field %s of type %s" if the value cannot be assigned to the field’s element type.

Side Effects:
- Modifies the target field of obj by appending value (when not already present).

Edge Cases & Assumptions:
- Works for both array and slice fields; uses DeepEqual to detect existing membership.
- If the value already exists in the field, nothing is changed.
- Assumes type compatibility between value and the field element type; mismatches may cause runtime errors.

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
Summary:
UpdateDirectiveFieldInConfigs updates a field on a Directive (identified by directive.Name) and persists the change to each file in configFiles. The field is set by name using reflection via SetField, with the value derived from arg as a lower-cased types.DirectiveType. The directive is validated with CheckForSave before saving, and each config file is updated by loading its content, applying the change, and saving it.

Signature:
func UpdateDirectiveFieldInConfigs(directive types.Directive, field string, arg string, configFiles []string) error

Parameters:
- directive: types.Directive — target directive; uses directive.Name to locate the entry in config.Settings.Directives.
- field: string — the field on the Directive to set (must be exported).
- arg: string — value to assign to the field; converted to a DirectiveType via strings.ToLower.
- configFiles: []string — list of config file paths to update with the new field value.

Returns:
- error — non-nil if setting the field or persisting any config fails; nil on success.

Errors/Exceptions:
- "Failed to set field %v on directive %v: %v" if SetField fails.
- "can't update directive %v: %v" if d.CheckForSave() returns an error.
- "Failed to create new directive: %v" if saving any per-file config fails.

Side Effects:
- Mutates config.Settings.Directives[directive.Name].
- Writes updated configuration to each configFile via SaveConfigFile, using a transient per-file load/save context (PushLoadedConfig/LoadConfigFile/SaveConfigFile/PopLoadedConfig).

Edge Cases & Assumptions:
- Operates on config.Settings.Directives[directive.Name]; behavior is defined for the target directive in in-memory settings and per-file configs.
- arg is lower-cased and cast to types.DirectiveType for field assignment.
- Each configFile is processed independently; an error in one aborts the operation.
- If configFiles is empty, only the in-memory directive is updated.

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
Summary: Updates the array/slice field named field on the provided directive by appending each value from args (if not already present), and persists the updated directive to each YAML config file in configFiles. The directive is identified in config.Settings.Directives by directive.Name. Validation is performed with d.CheckForSave() before persisting. If configFiles is non-empty, each file is loaded, updated, and saved; the in-memory global config is restored after processing. If configFiles is empty, the in-memory update remains only in config.Settings.
Signature: func UpdateDirectiveArrayInConfigs(directive types.Directive, field string, args []string, configFiles []string) error
Parameters:
- directive: types.Directive — the directive whose data is updated; uses directive.Name to locate the directive in config.Settings.Directives.
- field: string — the name of the array/slice field on the directive to mutate.
- args: []string — values to append to the field; each value is processed in order.
- configFiles: []string — config file paths to persist changes to; each file is loaded, updated, and saved.
Returns: error — non-nil on failure; nil on success. Errors may indicate inability to update an array field, validation failure, or failures writing config files.
Errors/Exceptions:
- non-nil when d.CheckForSave() fails (wrapped as "can't update directive: %v").
- non-nil when SetArray(&d, field, arg) fails for any arg (wrapped as "failed to update %v array with %v: %v").
- non-nil when saving to a config file fails (wrapped as "Failed to create new directive: %v").
Side Effects:
- Mutates the in-memory directive (config.Settings.Directives[directive.Name]) and, for each configFile, writes updated directives to disk via LoadConfigFile/SaveConfigFile.
- Logs debug information about saving to config files.
Edge Cases & Assumptions:
- Works for array and slice fields; avoids duplicates by membership check during SetArray (DeepEqual-based).
- If configFiles is empty, the in-memory update to config.Settings.Directives is applied but nothing is written to disk.
- Assumes type compatibility between Arg values and the field's element type; mismatches may cause runtime errors during SetArray.

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
