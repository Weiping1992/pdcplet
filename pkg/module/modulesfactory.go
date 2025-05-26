package module

import "strings"

var modulesFactory ModulesFactory

func init() {
	modulesFactory = NewModulesFactory()

	// Register all modules here
	// Example: RegisterConstructor("example", NewExampleModule)
	modulesFactory.Register("vmiproxy", NewVmiProxyModule)
	modulesFactory.Register("vmimetrics", NewVmiMetricsModule)
}

func CreateModule(name string, params map[string]interface{}) (Module, error) {
	return modulesFactory.CreateModule(name, params)
}

type ModulesFactory interface {
	CreateModule(name string, params map[string]interface{}) (Module, error)
	Register(name string, constructor ModuleConstructor)
}

type defaultModulesFactory struct {
	modules map[string]ModuleConstructor
}

func NewModulesFactory() ModulesFactory {
	return &defaultModulesFactory{
		modules: make(map[string]ModuleConstructor),
	}
}

func (f *defaultModulesFactory) Register(name string, constructor ModuleConstructor) {
	f.modules[strings.ToLower(name)] = constructor
}

func (f *defaultModulesFactory) CreateModule(name string, params map[string]interface{}) (Module, error) {
	if constructor, exists := f.modules[strings.ToLower(name)]; exists {
		return constructor(params)
	}
	return nil, nil
}

func (f *defaultModulesFactory) GetAllModules() []string {
	moduleNames := make([]string, 0, len(f.modules))
	for name := range f.modules {
		moduleNames = append(moduleNames, name)
	}
	return moduleNames
}
