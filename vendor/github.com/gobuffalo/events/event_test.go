package events

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gobuffalo/mapi"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_Event_Validate(t *testing.T) {
	r := require.New(t)

	e := Event{}
	r.Error(e.Validate())

	e.Kind = "foo"
	r.NoError(e.Validate())
}

func Test_Event_MarshalJSON(t *testing.T) {

	table := []struct {
		in  Event
		out map[string]interface{}
	}{
		{
			in: Event{
				Kind:    "K",
				Message: "M",
				Payload: mapi.Mapi{
					"p": "P",
				},
				Error: errors.New("E"),
			},
			out: map[string]interface{}{
				"kind":    "K",
				"message": "M",
				"payload": map[string]string{"p": "P"},
				"error":   "E",
			},
		},
		{
			in: Event{
				Kind: "K",
				Payload: mapi.Mapi{
					"func": func() {},
				},
			},
			out: map[string]interface{}{
				"kind":    "K",
				"payload": Payload{},
			},
		},
		{
			in: Event{
				Kind: "K",
				Payload: map[string]interface{}{
					"f":   func() {},
					"a":   "A",
					"two": 2,
				},
			},
			out: map[string]interface{}{
				"kind": "K",
				"payload": map[string]interface{}{
					"a":   "A",
					"two": 2,
				},
			},
		},
	}

	for i, tt := range table {
		t.Run(fmt.Sprintf("%d:%s", i, tt.in), func(st *testing.T) {
			r := require.New(st)
			act, err := json.Marshal(tt.in)
			r.NoError(err)

			exp, err := json.Marshal(tt.out)
			r.NoError(err)

			r.Equal(string(exp), string(act))

			e := Event{}
			r.NoError(json.Unmarshal(act, &e))
			r.Equal(e.Kind, tt.in.Kind)
			r.Equal(e.Message, tt.in.Message)
		})
	}
}
