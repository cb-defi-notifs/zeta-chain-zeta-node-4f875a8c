package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/zeta-chain/zetacore/common"
	contracts "github.com/zeta-chain/zetacore/contracts/zevm"
	"github.com/zeta-chain/zetacore/x/fungible/types"
	clientconfig "github.com/zeta-chain/zetacore/zetaclient/config"
	"math/big"
)

// FIXME: This is for testnet only
func (k Keeper) BlockOneDeploySystemContracts(goCtx context.Context) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// setup uniswap v2 factory
	uniswapV2Factory, err := k.DeployUniswapV2Factory(ctx)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployUniswapV2Factory")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("UniswapV2Factory", uniswapV2Factory.String()),
		),
	)

	// setup WZETA contract
	wzeta, err := k.DeployWZETA(ctx)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployWZetaContract")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("DeployWZetaContract", wzeta.String()),
		),
	)

	router, err := k.DeployUniswapV2Router02(ctx, uniswapV2Factory, wzeta)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to DeployUniswapV2Router02")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("DeployUniswapV2Router02", router.String()),
		),
	)

	SystemContractAddress, err := k.DeploySystemContract(ctx, wzeta, uniswapV2Factory, router)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to SystemContractAddress")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("SystemContractAddress", SystemContractAddress.String()),
		),
	)

	// set the system contract
	system, _ := k.GetSystemContract(ctx)
	system.SystemContract = SystemContractAddress.String()
	k.SetSystemContract(ctx, system)

	_, err = k.setupChainGasCoinAndPool(ctx, "GOERLI", "ETH", "gETH", 18)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}
	_, err = k.setupChainGasCoinAndPool(ctx, "BSCTESTNET", "BNB", "tBNB", 18)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}
	_, err = k.setupChainGasCoinAndPool(ctx, "MUMBAI", "MATIC", "tMATIC", 18)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}
	_, err = k.setupChainGasCoinAndPool(ctx, "BTCTESTNET", "BTC", "tBTC", 8)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to setupChainGasCoinAndPool")
	}

	return nil
}

// setup gas ZRC20, and ZETA/gas pool for a chain
// add 0.1gas/0.1wzeta to the pool
func (k Keeper) setupChainGasCoinAndPool(ctx sdk.Context, chain string, gasAssetName string, symbol string, decimals uint8) (ethcommon.Address, error) {
	name := fmt.Sprintf("%s-%s", gasAssetName, chain)
	transferGasLimit := big.NewInt(21_000)
	zrc20Addr, err := k.DeployZRC20Contract(ctx, name, symbol, decimals, chain, common.CoinType_Gas, "", transferGasLimit)
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to DeployZRC20Contract")
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(name, zrc20Addr.String()),
		),
	)
	chainid := clientconfig.Chains[chain].ChainID
	err = k.SetGasCoin(ctx, chainid, zrc20Addr)
	if err != nil {
		return ethcommon.Address{}, err
	}
	amount := big.NewInt(10)
	amount.Exp(amount, big.NewInt(int64(decimals-1)), nil)
	amountAZeta := big.NewInt(1e17)

	_, err = k.DepositZRC20(ctx, zrc20Addr, types.ModuleAddressEVM, amount)
	if err != nil {
		return ethcommon.Address{}, err
	}
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("azeta", sdk.NewIntFromBigInt(amountAZeta))))
	if err != nil {
		return ethcommon.Address{}, err
	}

	systemContractAddress, err := k.GetSystemContractAddress(ctx)
	if err != nil || systemContractAddress == (ethcommon.Address{}) {
		return ethcommon.Address{}, sdkerrors.Wrapf(types.ErrContractNotFound, "system contract address invalid: %s", systemContractAddress)
	}

	systemABI, err := contracts.SystemContractMetaData.GetAbi()
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to get system contract abi")
	}
	_, err = k.CallEVM(ctx, *systemABI, types.ModuleAddressEVM, systemContractAddress, BigIntZero, nil, true, "setGasZetaPool", chainid, zrc20Addr)
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to CallEVM method setGasZetaPool(%d, %s)", chainid, zrc20Addr.String())
	}

	// setup uniswap v2 pools gas/zeta
	routerAddress, err := k.GetUniswapV2Router02Address(ctx)
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to GetUniswapV2Router02Address")
	}
	routerABI, err := contracts.UniswapV2Router02MetaData.GetAbi()
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to get uniswap router abi")
	}
	zrc4ABI, err := contracts.ZRC20MetaData.GetAbi()
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to GetAbi zrc20")
	}
	_, err = k.CallEVM(ctx, *zrc4ABI, types.ModuleAddressEVM, zrc20Addr, BigIntZero, nil, true, "approve", routerAddress, amount)
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to CallEVM method approve(%s, %d)", routerAddress.String(), amount)
	}
	//function addLiquidityETH(
	//	address token,
	//	uint amountTokenDesired,
	//	uint amountTokenMin,
	//	uint amountETHMin,
	//	address to,
	//	uint deadline
	//) external payable returns (uint amountToken, uint amountETH, uint liquidity);
	res, err := k.CallEVM(ctx, *routerABI, types.ModuleAddressEVM, routerAddress, amount, big.NewInt(20_000_000), true,
		"addLiquidityETH", zrc20Addr, amount, BigIntZero, BigIntZero, types.ModuleAddressEVM, amountAZeta)
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to CallEVM method addLiquidityETH(%s, %s)", zrc20Addr.String(), amountAZeta.String())
	}
	AmountToken := new(*big.Int)
	AmountETH := new(*big.Int)
	Liquidity := new(*big.Int)
	err = routerABI.UnpackIntoInterface(&[]interface{}{AmountToken, AmountETH, Liquidity}, "addLiquidityETH", res.Ret)
	if err != nil {
		return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to UnpackIntoInterface addLiquidityETH")

	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute("function", "addLiquidityETH"),
			sdk.NewAttribute("amountToken", (*AmountToken).String()),
			sdk.NewAttribute("amountETH", (*AmountETH).String()),
			sdk.NewAttribute("liquidity", (*Liquidity).String()),
		),
	)

	//k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("azeta", sdk.NewIntFromBigInt(amount))))
	//amounts, err := k.CallUniswapv2RouterSwapExactETHForToken(ctx, types.ModuleAddressEVM, types.ModuleAddressEVM, big.NewInt(1e16), zrc20Addr)
	//if err != nil {
	//	return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to CallUniswapv2RouterSwapExactETHForToken")
	//}
	//ctx.EventManager().EmitEvent(
	//	sdk.NewEvent(sdk.EventTypeMessage,
	//		sdk.NewAttribute("function", "swapExactETHForTokens"),
	//		sdk.NewAttribute("amounts", fmt.Sprintf("%v", amounts)),
	//	),
	//)
	//
	//k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("azeta", sdk.NewInt(1e18))))
	//
	//amounts, err = k.QueryUniswapv2RouterGetAmountsIn(ctx, big.NewInt(1e16), zrc20Addr)
	//if err != nil {
	//	return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to QueryUniswapv2RouterGetAmountsIn")
	//}
	//ctx.EventManager().EmitEvent(
	//	sdk.NewEvent(sdk.EventTypeMessage,
	//		sdk.NewAttribute("function", "GetAmountsIn"),
	//		sdk.NewAttribute("amounts[0]", amounts[0].String()),
	//		sdk.NewAttribute("amounts[1]", amounts[1].String()),
	//	),
	//)
	//
	//amounts, err = k.CallUniswapv2RouterSwapEthForExactToken(ctx, types.ModuleAddressEVM, types.ModuleAddressEVM, amounts[0], amounts[1], zrc20Addr)
	//if err != nil {
	//	return ethcommon.Address{}, sdkerrors.Wrapf(err, "failed to CallUniswapv2RouterSwapEthForExactToken")
	//}
	//ctx.EventManager().EmitEvent(
	//	sdk.NewEvent(sdk.EventTypeMessage,
	//		sdk.NewAttribute("function", "SwapEthForExactToken"),
	//		sdk.NewAttribute("amounts", fmt.Sprintf("%v", amounts)),
	//	),
	//)

	return zrc20Addr, nil
}