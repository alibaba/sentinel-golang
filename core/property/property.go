package property

//PropertyListener
type PropertyListener interface {
	ConfigUpdate(value interface{})

	ConfigLoad(value interface{})
}
