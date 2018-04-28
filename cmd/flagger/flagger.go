package main

import "github.com/WillAbides/flagger"

func main() {
	h, err := flagger.GetConfigFromStdin()
	if err != nil {
		panic(err)
	}
	cfg, err := flagger.ReadConfig([]byte(h))
	if err != nil {
		panic(err)
	}
	flaggy := flagger.New(cfg)
	err = flaggy.AddArgs()
	if err != nil {
		panic(err)
	}
	err = flaggy.AddFlags()
	if err != nil {
		panic(err)
	}
	flaggy.Parse()
	flaggy.EchoVars()
}
