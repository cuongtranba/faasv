package faasv

import "encoding/json"

type Msg interface {
	Body() []byte
}

type natMsg struct {
	body []byte // can be string or struct
}

// Body implements Msg
func (n *natMsg) Body() []byte {
	return n.body
}

func NewNatMsg(payload any) Msg {
	switch payloadType := payload.(type) {
	case []byte:
		return &natMsg{
			body: payloadType,
		}
	case error:
		return &natMsg{
			body: []byte(payloadType.Error()),
		}
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		return &natMsg{
			body: b,
		}
	}
}
