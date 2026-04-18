package bootstrap

type Module interface {
	Name() string
	Register(*Runtime) error
}

func RegisterAll(runtime *Runtime, modules ...Module) error {
	for _, module := range modules {
		runtime.Logger.Info("register module", "name", module.Name())
		if err := module.Register(runtime); err != nil {
			return err
		}
	}
	return nil
}
