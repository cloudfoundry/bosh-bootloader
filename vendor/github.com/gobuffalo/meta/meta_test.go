package meta

import (
	"strings"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/stretchr/testify/require"
)

func Test_Named(t *testing.T) {
	envy.Set(envy.GO111MODULE, "off")
	r := require.New(t)

	app := Named("coke", ".")
	r.Equal("coke", app.Name.String())
	r.True(strings.HasSuffix(app.PackagePkg, "/coke"))
	r.True(strings.HasSuffix(app.ModelsPkg, "/coke/models"))
}
