package module

import "strings"

func init() {
	modulesFactory := NewModulesFactory()

	// Register all modules here
	// Example: RegisterConstructor("example", NewExampleModule)
	modulesFactory.Register("vmiproxy", NewVmiProxyModule)
	modulesFactory.Register("vmimetrics", NewDefaultVmiMetricsModule)
}

func CreateModule(name string, params ...interface{}) (Module, error) {
	return modulesFactory.CreateModule(name, params...)
}

type ModulesFactory interface {
	CreateModule(name string, params ...interface{}) (Module, error)
	Register(name string, constructor ModuleConstructor)
}

type modulesFactory struct {
	modules map[string]ModuleConstructor
}

func NewModulesFactory() ModulesFactory {
	return &modulesFactory{
		modules: make(map[string]ModuleConstructor),
	}
}

func (f *modulesFactory) Register(name string, constructor ModuleConstructor) {
	f.modules[strings.ToLower(name)] = constructor
}

func (f *modulesFactory) CreateModule(name string, params ...interface{}) (Module, error) {
	if constructor, exists := f.modules[strings.ToLower(name)]; exists {
		return constructor(params...)
	}
	return nil, nil
}

func (f *modulesFactory) GetAllModules() []string {
	moduleNames := make([]string, 0, len(f.modules))
	for name := range f.modules {
		moduleNames = append(moduleNames, name)
	}
	return moduleNames
}
