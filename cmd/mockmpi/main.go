package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zeta-chain/zetacore/common"
)

var ALL_CHAINS = []*ChainETHish{
	{
		name:                         common.Chain("ETH"),
		MPI_CONTRACT:                 "0xDe8802902Ff3136bdACe5FFC9a2423B1d37F6833",
		DEFAULT_DESTINATION_CONTRACT: "0xFf6B270ac3790589A1Fe90d0303e9D4d9A54FD1A",
	},
	{
		name:                         common.Chain("BSC"),
		MPI_CONTRACT:                 "0xCC3e1C9460B7803d4d79F32342b2b27543362536",
		DEFAULT_DESTINATION_CONTRACT: "0xF47bd84B86d1667e7621c38c72C6905Ca8710b0d",
	},
	{
		name:                         common.Chain("POLYGON"),
		MPI_CONTRACT:                 "0x692E8A48634B530b4BFF1e621FC18C82F471892c",
		DEFAULT_DESTINATION_CONTRACT: "0x22696Bef41E49FEf5beac1D4765a5b7B1E0Dcb01",
	},
}

func startAllChainListeners() {
	for _, chain := range ALL_CHAINS {
		chain.Start()
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	debug := flag.Bool("debug", false, "sets log level to debug")
	onlyChain := flag.String("only-chain", "all", "Uppercase name of a supported chain")
	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *onlyChain == "all" {
		log.Info().Msg("Starting all chains")
		startAllChainListeners()
	} else {
		log.Info().Msg(fmt.Sprintf("Running 1 chain only: %s", *onlyChain))
		chain, err := FindChainByName(*onlyChain)
		if err != nil {
			log.Fatal().Err(err).Msg("Chain not found")
			os.Exit(1)
		}
		chain.Start()
	}

	ch3 := make(chan os.Signal, 1)
	signal.Notify(ch3, syscall.SIGINT, syscall.SIGTERM)
	<-ch3
	log.Info().Msg("stop signal received")
}
