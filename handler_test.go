package faasv

import (
	"context"
	"fmt"
	"testing"

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
