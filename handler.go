package faasv

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

var (
	ErrHandlerNotFound                    = errors.New("faasv Handler not found")
	ErrHandlerMustBeFunc                  = errors.New("faasv Handler must be a func")
	ErrHandlerMustBeError                 = errors.New("faasv Handler must be a func with an error return type")
	ErrHandlerMustBeCtx                   = errors.New("faasv Handler must be a func with a context.Context argument")
	ErrHandlerMustBeAtLeastOneArgument    = errors.New("faasv Handler must be a func with at least one argument")
	ErrHandlerMustBeAtLeastOneReturnValue = errors.New("faasv Handler must be a func with at least one return value")
)

// must match
// func(ctx.context, args...) (any, error)
// func(ctx.context) (error)
// arg any
type Handler any

// isValidHandler check handler should match the following:
// func(ctx.context, args...) (any, error)
// func(ctx.context) (error)
// arg any
func isValidHandler(handler interface{}) error {
	cbType := reflect.TypeOf(handler)
	if cbType.Kind() != reflect.Func {
		return ErrHandlerMustBeFunc
	}
	numArgs := cbType.NumIn()
	if numArgs < 1 {
		return ErrHandlerMustBeAtLeastOneArgument
	}

	numOuts := cbType.NumOut()
	if numOuts < 1 {
		return ErrHandlerMustBeAtLeastOneReturnValue
	}

	lastTypeOutPut := cbType.Out(cbType.NumOut() - 1)
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	if !lastTypeOutPut.Implements(errorInterface) {
		return ErrHandlerMustBeError
	}

	firstTypeInput := cbType.In(0)
	ctxInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !firstTypeInput.Implements(ctxInterface) {
		return ErrHandlerMustBeCtx
	}

	return nil
}

func callHandlerWithByteArgs(handler Handler, ctx context.Context, args []byte) (any, error) {
	err := isValidHandler(handler)
	if err != nil {
		return nil, err
	}
	var paramsArgs interface{}
	err = json.Unmarshal(args, &paramsArgs)
	if err != nil {
		return nil, err
	}
	return callHandler(handler, ctx, paramsArgs)
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

type user struct {
	Name string `json:"name"`
}

func callHandler(handler Handler, ctx context.Context, arg any) (any, error) {
	method := reflect.ValueOf(handler)
	inputs := []reflect.Value{
		reflect.ValueOf(ctx),
	}

	if arg != nil {
		methodInType := method.Type().In(1)
		inputNew := reflect.New(methodInType).Elem().Interface()
		err := mapstructure.Decode(arg, &inputNew)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, reflect.ValueOf(inputNew))
	}

	response := method.Call(inputs)
	resultValue := response[0]
	errorValue := response[1]
	if errorValue.Interface() != nil {
		return nil, errorValue.Interface().(error)
	}
	return resultValue.Interface(), nil
}
