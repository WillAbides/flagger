package flagger

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"os"
	"gopkg.in/alecthomas/kingpin.v3-unstable"
	"github.com/pkg/errors"
	"io/ioutil"
	"strings"
	"io"
)

type FlagCfg struct {
	Name     string
	Help     string
	Short    string
	Type     string
	Default  string
	Env      string
	Required bool
}

type ArgCfg struct {
	Name     string
	Help     string
	Type     string
	Default  string
	Env      string
	Required bool
}

type FlaggerConfig struct {
	Name  string
	Help  string
	Flags map[string]*FlagCfg
	Args  []*ArgCfg
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
		if len(v.Short) > 1 {
			return nil, errors.Errorf("short flags must be a single character: %q is longer", v.Short)
		}
	}
	return cfg, nil
}

func New(cfg *FlaggerConfig) *Flagger {
	return &Flagger{
		stringVars: make(map[string]*string),
		intVars:    make(map[string]*int),
		app:        kingpin.New(cfg.Name, cfg.Help).Writers(os.Stderr, os.Stderr),
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
	for _, cfg := range f.cfg.Flags {
		err := f.AddFlag(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Flagger) AddArgs() error {
	for _, cfg := range f.cfg.Args {
		err := f.AddArg(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfg *ArgCfg) AddToApp(app *kingpin.Application) *kingpin.Clause{
	arg := app.Arg(cfg.Name, cfg.Help)
	if cfg.Required {
		arg.Required()
	}
	if cfg.Default != "" {
		arg.Default(cfg.Default)
	}
	return arg
}

func (cfg *FlagCfg)AddToApp(app *kingpin.Application) *kingpin.Clause{
	flag := app.Flag(cfg.Name, cfg.Help)
	if len(cfg.Short) == 1 {
		flag.Short([]rune(cfg.Short)[0])
	}
	if cfg.Required {
		flag.Required()
	}
	if cfg.Default != "" {
		flag.Default(cfg.Default)
	}
	return flag
}

func (f *Flagger)AddArg(cfg *ArgCfg) error {
	arg := cfg.AddToApp(f.app)
	env := cfg.Env
	if env == "" {
		env = strings.ToUpper(cfg.Name)
	}
	argType := cfg.Type
	if argType == "" {
		argType = "string"
	}
	switch argType {
	case "string":
		f.stringVars[env] = arg.String()
	case "int":
		f.intVars[env] = arg.Int()
	default:
		return errors.Errorf("The arg %q has an unknown type: %q", cfg.Name, argType)
	}
	return nil
}

func (f *Flagger)AddFlag(cfg *FlagCfg) error {
	flag := cfg.AddToApp(f.app)
	env := cfg.Env
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
		return errors.Errorf("The flag %q has an unknown type: %q", cfg.Name, flagType)
	}
	return nil
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
