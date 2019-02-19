package events

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Filter(t *testing.T) {
	r := require.New(t)

	var log []string
	of := func(e Event) {
		log = append(log, e.Kind)
	}
	ff := Filter("^buffalo:.+", of)
	df, err := Listen(ff)
	r.NoError(err)
	defer df()

	msgs := []string{
		"buffalo:foo",
		"foo:buffalo",
		"buffalo:foo:bar",
		"bar:buffalo:foo",
		"buffalofoo",
		"buffalo:",
	}

	for _, m := range msgs {
		ff(Event{Kind: m})
	}

	r.Len(log, 2)
}
