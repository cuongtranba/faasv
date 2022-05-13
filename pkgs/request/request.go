// for testing
package request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
)

func Post[T, V any](url string, req T) (*V, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var result V
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func WS[T, V any](url string, req T) (*V, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	defer ws.Close()
	err = ws.WriteJSON(req)
	if err != nil {
		return nil, err
	}
	var res V
	err = ws.ReadJSON(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
