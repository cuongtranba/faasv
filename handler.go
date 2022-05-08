package faasv

import (
	"errors"
	"reflect"
)

type Handler any

func isValidHandler(handler interface{}) error {
	cbType := reflect.TypeOf(handler)
	if cbType.Kind() != reflect.Func {
		return errors.New("faasv Handler needs to be a func")
	}
	numArgs := cbType.NumIn()
	if numArgs < 1 {
		return errors.New("faasv Handler needs to be a func with at least one argument")
	}

	lastTypeOutPut := cbType.Out(cbType.NumOut() - 1)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	if !lastTypeOutPut.Implements(errorInterface) {
		return errors.New("faasv Handler needs to be a func with an error return type")
	}

	return nil
}

func getArgs(handler Handler) []reflect.Value {
	method := reflect.ValueOf(handler)
	in := make([]reflect.Value, method.Type().NumIn())
	objects := make(map[reflect.Type]interface{})

	for i := 0; i < method.Type().NumIn(); i++ {
		t := method.Type().In(i)
		object := objects[t]
		in[i] = reflect.ValueOf(object)
	}
	return in
}

func callHandler(handler Handler, args ...interface{}) (any, error) {
	err := isValidHandler(handler)
	if err != nil {
		return nil, err
	}
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	method := reflect.ValueOf(handler)
	response := method.Call(inputs)
	if len(response) == 0 {
		return nil, errors.New("faasv Handler needs to be a func with an error return type")
	}
	resultValue := response[0]
	errorValue := response[1]
	if errorValue.Interface() != nil {
		return nil, errorValue.Interface().(error)
	}
	return resultValue.Interface(), nil
}
