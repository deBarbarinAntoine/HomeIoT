package data

const SETUP_MODULE = "setup"

type Setup struct{}

func (s *Setup) GetName() string {
	return SETUP_MODULE
}

func (s *Setup) GetValue() any {
	return nil
}
