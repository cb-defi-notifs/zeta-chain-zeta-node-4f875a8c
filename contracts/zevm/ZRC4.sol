// SPDX-License-Identifier: MIT
pragma solidity 0.8.7;
import "./Interfaces.sol";

interface ZRC4Errors {
    error CallerIsNotFungibleModule();

    error InvalidSender();
}

contract ZRC4 is Context, IZRC4, IZRC4Metadata, ZRC4Errors {
    address public constant FUNGIBLE_MODULE_ADDRESS = 0x735b14BB79463307AAcBED86DAf3322B1e6226aB;
    address public SYSTEM_CONTRACT_ADDRESS;
    uint256 public CHAIN_ID;
    CoinType public COIN_TYPE;
    uint256 public GAS_LIMIT;

    mapping(address => uint256) private _balances;

    mapping(address => mapping(address => uint256)) private _allowances;

    uint256 private _totalSupply;

    string private _name;
    string private _symbol;
    uint8 private _decimals;

    constructor(string memory name_, string memory symbol_, uint8 decimals_, uint256 chainid_, CoinType coinType_, uint256 gasLimit_, address systemContractAddress_) {
        if(msg.sender != FUNGIBLE_MODULE_ADDRESS) revert CallerIsNotFungibleModule();
        _name = name_;
        _symbol = symbol_;
        _decimals = decimals_;
        CHAIN_ID = chainid_;
        COIN_TYPE = coinType_;
        GAS_LIMIT = gasLimit_;
        SYSTEM_CONTRACT_ADDRESS = systemContractAddress_;
    }

    function name() public view virtual override returns (string memory) {
        return _name;
    }

    function symbol() public view virtual override returns (string memory) {
        return _symbol;
    }

    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }

    function totalSupply() public view virtual override returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) public view virtual override returns (uint256) {
        return _balances[account];
    }

    function transfer(address recipient, uint256 amount) public virtual override returns (bool) {
        _transfer(_msgSender(), recipient, amount);
        return true;
    }

    function allowance(address owner, address spender) public view virtual override returns (uint256) {
        return _allowances[owner][spender];
    }

    function approve(address spender, uint256 amount) public virtual override returns (bool) {
        _approve(_msgSender(), spender, amount);
        return true;
    }

    function transferFrom(address sender,address recipient,uint256 amount) public virtual override returns (bool) {
        _transfer(sender, recipient, amount);

        uint256 currentAllowance = _allowances[sender][_msgSender()];
        require(currentAllowance >= amount, "ERC20: transfer amount exceeds allowance");

        _approve(sender, _msgSender(), currentAllowance - amount);

        return true;
    }

    function _transfer(address sender, address recipient, uint256 amount) internal virtual {
        require(sender != address(0), "ERC20: transfer from the zero address");
        require(recipient != address(0), "ERC20: transfer to the zero address");

        uint256 senderBalance = _balances[sender];
        require(senderBalance >= amount, "ERC20: transfer amount exceeds balance");

        _balances[sender] = senderBalance - amount;
        _balances[recipient] += amount;

        emit Transfer(sender, recipient, amount);
    }

    function _mint(address account, uint256 amount) internal virtual {
        require(account != address(0), "ERC20: mint to the zero address");

        _totalSupply += amount;
        _balances[account] += amount;
        emit Transfer(address(0), account, amount);
    }

    function _burn(address account, uint256 amount) internal virtual {
        require(account != address(0), "ERC20: burn from the zero address");

        uint256 accountBalance = _balances[account];
        require(accountBalance >= amount, "ERC20: burn amount exceeds balance");

        _balances[account] = accountBalance - amount;
        _totalSupply -= amount;

        emit Transfer(account, address(0), amount);
    }

    function _approve(address owner, address spender, uint256 amount) internal virtual {
        require(owner != address(0), "ERC20: approve from the zero address");
        require(spender != address(0), "ERC20: approve to the zero address");

        _allowances[owner][spender] = amount;
        emit Approval(owner, spender, amount);
    }

    function deposit(address to, uint256 amount) external override returns (bool) {
        if(msg.sender != FUNGIBLE_MODULE_ADDRESS && msg.sender != SYSTEM_CONTRACT_ADDRESS) revert InvalidSender();
        _mint(to, amount);
        emit Deposit(abi.encodePacked(FUNGIBLE_MODULE_ADDRESS), to, amount);
        return true;
    }

    // returns the ZRC4 address for gas on the same chain of this ZRC4,
    // and calculate the gas fee for withdraw()
    function withdrawGasFee() public override view returns (address,uint256) {
        address gasZRC4 = ISystem(SYSTEM_CONTRACT_ADDRESS).gasCoinZRC4(CHAIN_ID);
        require(gasZRC4 != address(0), "gas coin not set");
        uint256 gasPrice = ISystem(SYSTEM_CONTRACT_ADDRESS).gasPrice(CHAIN_ID);
        require(gasPrice > 0, "gas price not set");
        uint256 gasFee = gasPrice * GAS_LIMIT;
        return (gasZRC4, gasFee);
    }

    // this function causes cctx module to send out outbound tx to the outbound chain
    // this contract should be given enough allowance of the gas ZRC4 to pay for outbound tx gas fee
    function withdraw(bytes memory to, uint256 amount) external override returns (bool) {
        (address gasZRC4, uint256 gasFee)= withdrawGasFee();
        require(IZRC4(gasZRC4).transferFrom(msg.sender, (FUNGIBLE_MODULE_ADDRESS), gasFee), "transfer gas fee failed");

        _burn(msg.sender, amount);
        emit Withdrawal(msg.sender, to, amount, gasFee);
        return true;
    }

    function updateSystemContractAddress(address addr) external {
        require(msg.sender == FUNGIBLE_MODULE_ADDRESS, "permission error");
        SYSTEM_CONTRACT_ADDRESS = addr;
    }

    function updateGasLimit(uint256 gasLimit) external {
        require(msg.sender == FUNGIBLE_MODULE_ADDRESS, "permission error");
        GAS_LIMIT = gasLimit;
    }
}