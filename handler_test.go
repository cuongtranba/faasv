package faasv

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

type dummyStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type dummyStructResult struct {
	UserType []string
}

func Test_callFunc(t *testing.T) {

	result, err := callHandler(func(ctx context.Context, in dummyStruct) (string, error) {
		return in.Name, nil
	}, context.TODO(), map[string]interface{}{
		"Name": "test",
	})

	require.Nil(t, err)
	require.Equal(t, "test", result)

	dummyHandlerStringResult := func(ctx context.Context, userId string) (string, error) {
		return userId, nil
	}

	resultTypeValues, err := callHandler(dummyHandlerStringResult, context.TODO(), "test")
	require.Nil(t, err)
	require.Equal(t, "test", resultTypeValues)

	dummyHandlerStructResult := func(ctx context.Context, req dummyStruct) (*dummyStructResult, error) {
		return &dummyStructResult{
			UserType: []string{"asdasd"},
		}, nil
	}
	resultTypeValues, err = callHandler(dummyHandlerStructResult, context.TODO(), dummyStruct{Name: "test", Age: 1})
	require.Nil(t, err)
	require.Equal(t, "asdasd", resultTypeValues.(*dummyStructResult).UserType[0])

	dummyHandlerWithErrorReturn := func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("test error")
	}

	resultTypeValues, err = callHandler(dummyHandlerWithErrorReturn, context.TODO(), nil)
	require.NotNil(t, err)
	require.Nil(t, resultTypeValues)
	require.Equal(t, "test error", err.Error())
}

type userType struct {
	Name string
}

type result struct {
	Name string
}

func Test_isValidHandler(t *testing.T) {
	dummyHandlerStringResultFunc := func(ctx context.Context, userId string) (string, error) {
		return userId, nil
	}
	err := isValidHandler(dummyHandlerStringResultFunc)
	require.Nil(t, err)
	dummyNoArgFunc := func(ctx context.Context) error {
		return nil
	}
	require.Nil(t, isValidHandler(dummyNoArgFunc))

	require.Nil(t, isValidHandler(func(ctx context.Context, user userType) (*result, error) {
		return &result{
			Name: user.Name,
		}, nil
	}))

	dummyNoCtxFunc := func() error {
		return nil
	}
	require.NotNil(t, isValidHandler(dummyNoCtxFunc))

	dummyNoReturnFunc := func(ctx context.Context) {
	}
	require.NotNil(t, isValidHandler(dummyNoReturnFunc))

	dummyNoCtxArgsReturnFunc := func(userId string) string {
		return userId
	}
	require.NotNil(t, isValidHandler(dummyNoCtxArgsReturnFunc))
}

type YourT2 struct{}

func (y YourT2) MethodFoo(u user) string {
	//do something
	return u.Name
}

func example() {
	value := map[string]interface{}{
		"Name": "test",
	}

	method := reflect.ValueOf(YourT2{}).MethodByName("MethodFoo")
	methodInType := method.Type().In(0)
	inputNew := reflect.New(methodInType).Elem().Interface()
	err := mapstructure.Decode(value, &inputNew)
	if err != nil {
		panic(err)
	}
	method.Call([]reflect.Value{reflect.ValueOf(inputNew)})
}

func Test_example(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			example()
		})
	}
}
