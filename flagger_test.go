package flagger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func strPtr(str string) *string {
	return &str
}

func intPtr(i int) *int {
	return &i
}

func TestReadConfigErrs(t *testing.T) {
	t.Run("long short flag", func(t *testing.T) {
		in := []byte(`flags = { abc = { short = "hi" } }`)
		exErr := `short flags must be a single character: "hi" is longer`
		out, err := ReadConfig(in)
		assert.Nil(t, out)
		assert.EqualError(t, err, exErr)
	})

	t.Run("invalid hcl", func(t *testing.T) {
		in := []byte(`asdf`)
		exErr := "Failed reading config"
		out, err := ReadConfig(in)
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), exErr)
	})
}

func TestReadConfigValid(t *testing.T) {
	hh :=
		`
name = "foo"
desc = "bar"
flags = {
 flagfoo = {
   desc = "flagdesc" short = "h" type = "string" default = "blah"
   required = true
 },
 bar = {},
}`
	ex := FlaggerConfig{
		Name:        "foo",
		Description: "bar",
		Flags: map[string]*Flag{
			"flagfoo": {
				Name:        "flagfoo",
				Description: "flagdesc",
				ShortFlag:   "h",
				Type:        "string",
				Default:     "blah",
				Required:    true,
			},
			"bar": {
				Name: "bar",
			},
		},
	}
	flaggerConfig, err := ReadConfig([]byte(hh))
	assert.Nil(t, err)
	assert.Equal(t, ex, *flaggerConfig)
}

func TestFlagger_WriteEnvOutput(t *testing.T) {
	f := &Flagger{
		stringVars: map[string]*string{
			"V1": strPtr("v1"),
			"V2": strPtr("v2"),
		},
		intVars: map[string]*int{
			"I1": intPtr(1),
			"I2": intPtr(2),
		},
	}
	b := new(bytes.Buffer)
	f.WriteEnvOutput(b)
	ex := `
V1="v1"
V2="v2"
I1=1
I2=2`
	exLines := strings.Split(ex, "\n")
	lines := b.String()
	for _, e := range exLines {
		assert.Contains(t, lines, e)
	}
	assert.Len(t, strings.Split(lines, "\n"), len(exLines))
}
