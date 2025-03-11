package main

import (
	"errors"
	"fmt"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/urfave/cli/v2"

	sunrise "github.com/sunriselayer/sunrise-op-da-server"
)

const (
	ListenAddrFlagName = "addr"
	PortFlagName       = "port"

	SunriseServerFlagName           = "sunrise.server"
	SunriseDataShardCountFlagName   = "sunrise.data_shard_count"
	SunriseParityShardCountFlagName = "sunrise.parity_shard_count"
)

const EnvVarPrefix = "OP_PLASMA_DA_SERVER"

func prefixEnvVars(name string) []string {
	return opservice.PrefixEnvVar(EnvVarPrefix, name)
}

var (
	ListenAddrFlag = &cli.StringFlag{
		Name:    ListenAddrFlagName,
		Usage:   "server listening address",
		Value:   "127.0.0.1",
		EnvVars: prefixEnvVars("ADDR"),
	}
	PortFlag = &cli.IntFlag{
		Name:    PortFlagName,
		Usage:   "server listening port",
		Value:   3100,
		EnvVars: prefixEnvVars("PORT"),
	}

	SunriseServerFlag = &cli.StringFlag{
		Name:    SunriseServerFlagName,
		Usage:   "sunrise server endpoint",
		Value:   "http://localhost:26658",
		EnvVars: prefixEnvVars("SUNRISE_SERVER"),
	}
	SunriseDataShardCountFlag = &cli.IntFlag{
		Name:    SunriseDataShardCountFlagName,
		Usage:   "sunrise data shard count",
		Value:   10,
		EnvVars: prefixEnvVars("SUNRISE_DATA_SHARD_COUNT"),
	}
	SunriseParityShardCountFlag = &cli.IntFlag{
		Name:    SunriseParityShardCountFlagName,
		Usage:   "sunrise parity shard count",
		Value:   10,
		EnvVars: prefixEnvVars("SUNRISE_PARITY_SHARD_COUNT"),
	}
)

var requiredFlags = []cli.Flag{}

var optionalFlags = []cli.Flag{ListenAddrFlag,
	PortFlag,
	SunriseServerFlag,
	SunriseDataShardCountFlag,
	SunriseParityShardCountFlag}

func init() {
	optionalFlags = append(optionalFlags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

type CLIConfig struct {
	SunriseEndpoint         string
	SunriseDataShardCount   int
	SunriseParityShardCount int
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		SunriseEndpoint:         ctx.String(SunriseServerFlagName),
		SunriseDataShardCount:   ctx.Int(SunriseDataShardCountFlagName),
		SunriseParityShardCount: ctx.Int(SunriseParityShardCountFlagName),
	}
}

func (c CLIConfig) Check() error {
	if c.SunriseEndpoint == "" {
		return errors.New("all Sunrise flags must be set")
	}
	if c.SunriseDataShardCount == 0 || c.SunriseParityShardCount == 0 {
		return errors.New("data and parity shard count must be greater than 0")
	}
	return nil
}

func (c CLIConfig) SunriseConfig() sunrise.SunriseConfig {
	return sunrise.SunriseConfig{
		URL:              c.SunriseEndpoint,
		DataShardCount:   c.SunriseDataShardCount,
		ParityShardCount: c.SunriseParityShardCount,
	}
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	return nil
}
