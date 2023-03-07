//go:build PRIVNET
// +build PRIVNET

package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func (sm *SmokeTest) TestCrosschainSwap() {
	startTime := time.Now()
	defer func() {
		fmt.Printf("test finishes in %s\n", time.Since(startTime))
	}()
	LoudPrintf("Testing Bitcoin ERC20 crosschain swap...\n")
	// Firstly, deposit 1.15 BTC into Zeta for liquidity
	//sm.DepositBTC()
	// Secondly, deposit 1000.0 USDT into Zeta for liquidity
	txhash := sm.DepositERC20(big.NewInt(1e9), []byte{})
	WaitCctxMinedByInTxHash(txhash.Hex(), sm.cctxClient)

	sm.zevmAuth.GasLimit = 10000000
	tx, err := sm.UniswapV2Factory.CreatePair(sm.zevmAuth, sm.USDTZRC20Addr, sm.BTCZRC20Addr)
	if err != nil {
		panic(err)
	}
	receipt := MustWaitForTxReceipt(sm.zevmClient, tx)
	usdtBtcPair, err := sm.UniswapV2Factory.GetPair(&bind.CallOpts{}, sm.USDTZRC20Addr, sm.BTCZRC20Addr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("USDT-BTC pair receipt txhash %s status %d pair addr %s\n", receipt.TxHash, receipt.Status, usdtBtcPair.Hex())

	tx, err = sm.USDTZRC20.Approve(sm.zevmAuth, sm.UniswapV2RouterAddr, big.NewInt(1e18))
	if err != nil {
		panic(err)
	}
	receipt = MustWaitForTxReceipt(sm.zevmClient, tx)
	fmt.Printf("USDT ZRC20 approval receipt txhash %s status %d\n", receipt.TxHash, receipt.Status)

	tx, err = sm.BTCZRC20.Approve(sm.zevmAuth, sm.UniswapV2RouterAddr, big.NewInt(1e18))
	if err != nil {
		panic(err)
	}
	receipt = MustWaitForTxReceipt(sm.zevmClient, tx)
	fmt.Printf("BTC ZRC20 approval receipt txhash %s status %d\n", receipt.TxHash, receipt.Status)

	// Add 100 USDT liq and 0.001 BTC
	bal, err := sm.BTCZRC20.BalanceOf(&bind.CallOpts{}, DeployerAddress)
	if err != nil {
		panic(err)
	}
	fmt.Printf("balance of deployer on BTC ZRC20: %d\n", bal)
	bal, err = sm.USDTZRC20.BalanceOf(&bind.CallOpts{}, DeployerAddress)
	if err != nil {
		panic(err)
	}
	fmt.Printf("balance of deployer on USDT ZRC20: %d\n", bal)
	tx, err = sm.UniswapV2Router.AddLiquidity(sm.zevmAuth, sm.USDTZRC20Addr, sm.BTCZRC20Addr, big.NewInt(1e8), big.NewInt(1e8), big.NewInt(1e8), big.NewInt(1e5), DeployerAddress, big.NewInt(time.Now().Add(10*time.Minute).Unix()))
	if err != nil {
		fmt.Printf("Error liq %s", err.Error())
		panic(err)
	}
	receipt = MustWaitForTxReceipt(sm.zevmClient, tx)
	fmt.Printf("Add liquidity receipt txhash %s status %d\n", receipt.TxHash, receipt.Status)

	btcMinOutAmount := big.NewInt(0)
	msg := []byte{}
	for i := 0; i < 20-len(HexToAddress(ZEVMSwapAppAddr).Bytes()); i++ {
		msg = append(msg, 0)
	}
	msg = append(msg, HexToAddress(ZEVMSwapAppAddr).Bytes()...)
	for i := 0; i < 32-len(sm.BTCZRC20Addr.Bytes()); i++ {
		msg = append(msg, 0)
	}
	msg = append(msg, sm.BTCZRC20Addr.Bytes()...)
	for i := 0; i < 32-len(DeployerAddress.Bytes()); i++ {
		msg = append(msg, 0)
	}
	msg = append(msg, DeployerAddress.Bytes()...)
	for i := 0; i < 32-len(btcMinOutAmount.Bytes()); i++ {
		msg = append(msg, 0)
	}
	msg = append(msg, btcMinOutAmount.Bytes()...)

	for {
		// Should deposit USDT for swap, swap for BTC and withdraw BTC
		txhash = sm.DepositERC20(big.NewInt(4e7), msg)
		cctx1 := WaitCctxMinedByInTxHash(txhash.Hex(), sm.cctxClient)

		_, err = sm.btcRPCClient.GenerateToAddress(10, BTCDeployerAddress, nil)
		if err != nil {
			panic(err)
		}
		// cctx1 index acts like the inTxHash for the second cctx (the one that withdraws BTC)
		cctx2 := WaitCctxMinedByInTxHash(cctx1.Index, sm.cctxClient)
		_ = cctx2
	}
}
