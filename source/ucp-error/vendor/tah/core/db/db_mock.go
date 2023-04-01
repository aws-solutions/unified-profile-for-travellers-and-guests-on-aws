package db

import core "tah/core/core"

type MockDBConfig struct {
}

func (c MockDBConfig) SetTx(tx core.Transaction) error {
	return nil
}
func (c MockDBConfig) Save(prop interface{}) (interface{}, error) {
	return prop, nil
}
func (c MockDBConfig) FindStartingWith(pk string, value string, data interface{}) error {
	return nil
}
func (c MockDBConfig) DeleteMany(data interface{}) error {
	return nil
}
func (c MockDBConfig) SaveMany(data interface{}) error {
	return nil
}
func (c MockDBConfig) Get(pk string, sk string, data interface{}) error {
	return nil
}
