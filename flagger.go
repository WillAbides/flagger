package flagger

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"os"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
	"io"
)

type FlagCfg struct {
	Name        string
	Description string `hcl:"desc"`
	ShortFlag   string `hcl:"short"`
	Type        string
	Default     string
	EnvVar      string `hcl:"env"`
	Required    bool
}

type Arg struct {
	Name        string
	Description string `hcl:"desc"`
	Type        string
	Default     string
	Required    bool
}

type FlaggerConfig struct {
	Name        string
	Description string `hcl:"desc"`
	Flags       map[string]*FlagCfg
}

type Flagger struct {
	stringVars map[string]*string
	intVars    map[string]*int
	app        *kingpin.Application
	cfg        *FlaggerConfig
}

func ReadConfig(b []byte) (*FlaggerConfig, error) {
	cfg := new(FlaggerConfig)
	err := hcl.Unmarshal(b, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed reading config")
	}
	for k, v := range cfg.Flags {
		if v.Name == "" {
			v.Name = k
		}
		if len(v.ShortFlag) > 1 {
			return nil, errors.Errorf("short flags must be a single character: %q is longer", v.ShortFlag)
		}
	}
	return cfg, nil
}

func New(cfg *FlaggerConfig) *Flagger {
	app := kingpin.New(cfg.Name, cfg.Description)
	app.HelpFlag.PreAction(func(context *kingpin.ParseContext) error {
		app.Usage([]string{})
		os.Exit(1)
		return nil
	})
	return &Flagger{
		stringVars: make(map[string]*string),
		intVars:    make(map[string]*int),
		app:        app,
		cfg:        cfg,
	}
}

func (f *Flagger) EchoVars() {
	f.WriteEnvOutput(os.Stdout)
}


func (f *Flagger) WriteEnvOutput(w io.Writer) {
	for name, val := range f.stringVars {
		fmt.Fprintf(w, "%s=%q\n", name, *val)
	}
	for name, val := range f.intVars {
		fmt.Fprintf(w, "%s=%d\n", name, *val)
	}
}

func (f *Flagger) AddFlags() error {
	config := f.cfg
	flags := config.Flags
	for _, cfg := range flags {
		err := f.AddFlag(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfg *FlagCfg)AddToApp(app *kingpin.Application) *kingpin.FlagClause{
	flag := app.Flag(cfg.Name, cfg.Description)
	if len(cfg.ShortFlag) == 1 {
		flag.Short([]rune(cfg.ShortFlag)[0])
	}
	if cfg.Required {
		flag.Required()
	}
	if cfg.Default != "" {
		flag.Default(cfg.Default)
	}
	return flag
}

func (f *Flagger)AddFlag(cfg *FlagCfg) error {
	var err error
	flag := cfg.AddToApp(f.app)
	env := cfg.EnvVar
	if env == "" {
		env = strings.ToUpper(cfg.Name)
	}
	flagType := cfg.Type
	if flagType == "" {
		flagType = "string"
	}
	switch flagType {
	case "string":
		f.stringVars[env] = flag.String()
	case "int":
		f.intVars[env] = flag.Int()
	default:
		err = errors.Errorf("The flag %q has an unknown type: %q", cfg.Name, flagType)
	}
	return err
}

func (f *Flagger) Parse() {
	_, err := f.app.Parse(os.Args[1:])

	if err != nil {
		f.app.FatalUsage("%v", err)
	}
}

func GetConfigFromStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", errors.New("You need to pipe config to stdin")
	}
	bytes, err := ioutil.ReadAll(os.Stdin)
	return string(bytes), err
}
