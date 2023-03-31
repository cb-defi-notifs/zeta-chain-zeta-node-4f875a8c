//go:build PRIVNET
// +build PRIVNET

package common

func GoerliLocalNetChain() Chain {
	return Chain{
		ChainName: ChainName_goerli_localnet,
		ChainId:   1337,
	}
}

func ZetaChainPrivateNet() Chain {
	return Chain{
		ChainName: ChainName_zeta_mainnet,
		ChainId:   101,
	}
}

func BtcRegtestChain() Chain {
	return Chain{
		ChainName: ChainName_btc_regtest,
		ChainId:   18444,
	}
}

func DefaultChainsList() []*Chain {
	chains := []Chain{
		BtcRegtestChain(),
		GoerliLocalNetChain(),
		ZetaChainPrivateNet(),
	}
	var c []*Chain
	for i := 0; i < len(chains); i++ {
		c = append(c, &chains[i])
	}
	return c
}
