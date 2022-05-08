package faasv

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type dummyStruct struct {
	Name string
	Age  int
}

type dummyStructResult struct {
	UserType []string
}

func Test_callFunc(t *testing.T) {
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

	resultTypeValues, err = callHandler(dummyHandlerWithErrorReturn, context.TODO())
	require.NotNil(t, err)
	require.Nil(t, resultTypeValues)
	require.Equal(t, "test error", err.Error())
}
