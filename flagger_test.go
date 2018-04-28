package flagger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
)

func strPtr(str string) *string {
	return &str
}

func intPtr(i int) *int {
	return &i
}

func TestFlagCfg_AddToApp(t *testing.T) {
	app := kingpin.New("", "")
	flagCfg := &FlagCfg{
		Name: "foo", Help: "bar", Short: "f", Default: "hi", Required: true,
	}
	ex := &kingpin.FlagModel{
		Name: "foo",
		Help: "bar",
		Short: rune("f"[0]),
		Default: []string{"hi"},
		Required: true,
	}
	clause := flagCfg.AddToApp(app)
	assert.Equal(t, ex, clause.Model())
}

func TestArgCfg_AddToApp(t *testing.T) {
	app := kingpin.New("", "")
	argCfg := &ArgCfg{
		Name: "foo", Help: "bar", Default: "hi", Required: true,
	}
	ex := &kingpin.ArgModel{
		Name: "foo",
		Help: "bar",
		Default: []string{"hi"},
		Required: true,
	}
	clause := argCfg.AddToApp(app)
	assert.Equal(t, ex, clause.Model())
}

func TestFlagger_AddFlag(t *testing.T) {
	t.Run("string var", func(t *testing.T) {
		app := kingpin.New("", "")
		flagger := &Flagger{
			app: app,
			stringVars: make(map[string]*string),
			intVars:    make(map[string]*int),
		}
		flagCfg := &FlagCfg{
			Name: "foo", Help: "bar", Short: "f", Default: "hi", Required: true,
		}
		err := flagger.AddFlag(flagCfg)
		assert.Nil(t, err)
		assert.Len(t, flagger.stringVars, 1)
		_, ok := flagger.stringVars["FOO"]
		assert.True(t, ok)
	})

	t.Run("int var", func(t *testing.T) {
		app := kingpin.New("", "")
		flagger := &Flagger{
			app: app,
			stringVars: make(map[string]*string),
			intVars:    make(map[string]*int),
		}
		flagCfg := &FlagCfg{
			Name: "foo", Default: "2", Type: "int",
		}
		err := flagger.AddFlag(flagCfg)
		assert.Nil(t, err)
		assert.Len(t, flagger.intVars, 1)
		_, ok := flagger.intVars["FOO"]
		assert.True(t, ok)
	})

}

func TestFlagger_AddArg(t *testing.T) {
	t.Run("string var", func(t *testing.T) {
		app := kingpin.New("", "")
		flagger := &Flagger{
			app: app,
			stringVars: make(map[string]*string),
			intVars:    make(map[string]*int),
		}
		argCfg := &ArgCfg{
			Name: "foo", Help: "bar", Default: "hi", Required: true,
		}
		err := flagger.AddArg(argCfg)
		assert.Nil(t, err)
		assert.Len(t, flagger.stringVars, 1)
		_, ok := flagger.stringVars["FOO"]
		assert.True(t, ok)
	})

	t.Run("int var", func(t *testing.T) {
		app := kingpin.New("", "")
		flagger := &Flagger{
			app: app,
			stringVars: make(map[string]*string),
			intVars:    make(map[string]*int),
		}
		argCfg := &ArgCfg{
			Name: "foo", Help: "bar", Default: "hi", Required: true, Type: "int",
		}
		err := flagger.AddArg(argCfg)
		assert.Nil(t, err)
		assert.Len(t, flagger.intVars, 1)
		_, ok := flagger.intVars["FOO"]
		assert.True(t, ok)
	})
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
help = "bar"
args = [
  { name = "myarg" help = "myarghelp" type = "int"required = true },
]
flags = {
  flagfoo = {
    help = "flaghelp" short = "h" type = "string" default = "blah"
    required = true
  },
  bar = {},
}`
	ex := FlaggerConfig{
		Name: "foo",
		Help: "bar",
		Args: []*ArgCfg{
			{
				Name: "myarg", Help: "myarghelp", Type: "int", Required: true,
			},
		},
		Flags: map[string]*FlagCfg{
			"flagfoo": {
				Name:     "flagfoo",
				Help:     "flaghelp",
				Short:    "h",
				Type:     "string",
				Default:  "blah",
				Required: true,
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
