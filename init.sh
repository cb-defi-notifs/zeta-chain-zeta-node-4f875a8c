#!/bin/bash

KEY1="alice"
KEH2="bob"
CHAINID="athens_9000-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# to trace evm
TRACE="--trace"
#TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# remove existing daemon and client
rm -rf ~/.zetacore*

make install

zetacored config keyring-backend $KEYRING
zetacored config chain-id $CHAINID

# if $KEY exists it should be deleted
#zetacored keys add $KEY1 --keyring-backend $KEYRING --algo $KEYALGO
echo "Generating deterministic account - alice"
echo "race draft rival universe maid cheese steel logic crowd fork comic easy truth drift tomorrow eye buddy head time cash swing swift midnight borrow" | zetacored keys add alice --recover --keyring-backend $KEYRING

echo "Generating deterministic account - bob"
echo "hand inmate canvas head lunar naive increase recycle dog ecology inhale december wide bubble hockey dice worth gravity ketchup feed balance parent secret orchard" | zetacored keys add bob --recover --keyring-backend $KEYRING


# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
zetacored init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to aphoton
cat $HOME/.zetacore/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="azeta"' > $HOME/.zetacore/config/tmp_genesis.json && mv $HOME/.zetacore/config/tmp_genesis.json $HOME/.zetacore/config/genesis.json
cat $HOME/.zetacore/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="azeta"' > $HOME/.zetacore/config/tmp_genesis.json && mv $HOME/.zetacore/config/tmp_genesis.json $HOME/.zetacore/config/genesis.json
cat $HOME/.zetacore/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="azeta"' > $HOME/.zetacore/config/tmp_genesis.json && mv $HOME/.zetacore/config/tmp_genesis.json $HOME/.zetacore/config/genesis.json
cat $HOME/.zetacore/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="azeta"' > $HOME/.zetacore/config/tmp_genesis.json && mv $HOME/.zetacore/config/tmp_genesis.json $HOME/.zetacore/config/genesis.json
cat $HOME/.zetacore/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="azeta"' > $HOME/.zetacore/config/tmp_genesis.json && mv $HOME/.zetacore/config/tmp_genesis.json $HOME/.zetacore/config/genesis.json


# Set gas limit in genesis
cat $HOME/.zetacore/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="10000000"' > $HOME/.zetacore/config/tmp_genesis.json && mv $HOME/.zetacore/config/tmp_genesis.json $HOME/.zetacore/config/genesis.json

# disable produce empty block
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.zetacore/config/config.toml
  else
    sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.zetacore/config/config.toml
fi

if [[ $1 == "pending" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.zetacore/config/config.toml
      sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.zetacore/config/config.toml
  else
      sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.zetacore/config/config.toml
      sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.zetacore/config/config.toml
  fi
fi

# Allocate genesis accounts (cosmos formatted addresses)
zetacored add-genesis-account $KEY1 100000000000000000000000000azeta --keyring-backend $KEYRING
zetacored add-genesis-account $KEY2 1000000000000000000000azeta --keyring-backend $KEYRING


# Sign genesis transaction
zetacored gentx $KEY1 1000000000000000000000azeta --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
zetacored collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
zetacored validate-genesis

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
zetacored start --pruning=nothing --evm.tracer=json $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0001azeta --json-rpc.api eth,txpool,personal,net,debug,web3,miner --api.enable
