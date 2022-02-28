package main

import "github.com/jessevdk/go-flags"

type InterchainGet struct {
	Args struct {
		TargetChainId  string `positional-arg-name:"tid" required:"true" description:"target pier id"`
		CCid string `positional-arg-name:"ccid" required:"true" description:"target contract's address"`
		Key string `positional-arg-name:"key" required:"true" description: value's key"`
	} `positional-args:"true"`
	Url     string `long:"url" description:"Specify URL of REST API"`
	Keyfile string `long:"keyfile" description:"Identify file containing user's private key"`
	Wait    uint   `long:"wait" description:"Set time, in seconds, to wait for transaction to commit"`
}

func (args *InterchainGet) Name() string {
	return "interchainGet"
}

func (args *InterchainGet) KeyfilePassed() string {
	return args.Keyfile
}

func (args *InterchainGet) UrlPassed() string {
	return args.Url
}

func (args *InterchainGet) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "init meta value", "Sends an dsswapper transaction to set <name> to <value>.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *InterchainGet) Run() error {
	// Construct client
	//name := args.Args.Name
	//value := args.Args.Value
	wait := args.Wait

	dsClient, err := GetClient(args, true)
	if err != nil {
		return err
	}
	_, err = dsClient.Init(wait)
	return err
}