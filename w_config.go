package wazero

import "github.com/tetratelabs/wazero/api"

type WATERExtendedCompiledModule interface {
	// AllImports returns name and type of all imported functions, tables,
	// memories or globals required for instantiation, per module name.
	AllImports() map[string]map[string]api.ExternType

	// AllExports returns name and type of all exported functions, tables,
	// memories or globals from this module.
	AllExports() map[string]api.ExternType
}

// Implements WATERExtendedCompiledModule.
func (c *compiledModule) AllExports() map[string]api.ExternType {
	var ret = make(map[string]api.ExternType)
	for name, f := range c.module.Exports {
		if f != nil {
			ret[name] = f.Type
		}
	}
	return ret
}

// Implements WATERExtendedCompiledModule.
func (c *compiledModule) AllImports() map[string]map[string]api.ExternType {
	var ret = make(map[string]map[string]api.ExternType)
	for module, imports := range c.module.ImportPerModule {
		if len(imports) == 0 {
			continue
		}

		if _, ok := ret[module]; !ok {
			ret[module] = make(map[string]api.ExternType)
		}
		for _, f := range imports {
			ret[module][f.Name] = f.Type
		}
	}
	return ret
}
