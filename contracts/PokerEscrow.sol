// SPDX-License-Identifier-MIT 
pragma solidity ^0.8.20;

contract PokerEscrow {
    bytes32 public gameId;
    uint256 public entryFee;
    address[] public players;

    enum GameState {Joining, Active, Resolved}
    GameState public state;

    mapping(address=>bool) public isPlayer;
    mapping(address=>uint256) public securityDeposits;
    mapping(address=>uint256) public playerStacks;
    mapping(address=>address) public votes;
    mapping(address=>uint256) public voteCount;
    mapping(address=>uint256) public challengeDeadline;

    uint256 public constant CHALLENGE_WINDOW = 30 minutes;
    bool private locked;

    event DepositMade(address indexed player, uint256 deposit, uint256 stack);
    event StackTopUp(address indexed player, uint256 addedAmount, uint256 newTotal);
    event WinnerVoted(address indexed voter, address indexed candidate);
    event HandResolved(address indexed winner, uint256 potAmount);
    event PlayerSlashed(address indexed offender, uint256 redistributedAmount);
    event FairExit(address indexed player, uint256 stackReturned, uint256 depositReturned);
    event ChallengeStarted(address indexed challenger, address indexed offender);

    modifier onlyPlayers() {
        require(isPlayer[msg.sender], "Not a participant");
        _;
    }

    modifier noReentrant() {
        require(!locked, "No reentrancy");
        locked = true;
        _;
        locked = false;
    }

    constructor(bytes32 _gameId, uint256 _entryFee, address[] memory _players) {
        gameId = _gameId;
        entryFee = _entryFee;
        players = _players;
        state = GameState.Joining;
        for(uint i = 0; i < _players.length; i++) {
            isPlayer[_players[i]] = true;
        }
    }

    function depositStake() external payable onlyPlayers {
        require(state == GameState.Joining, "Not in joining phase");
        require(msg.value > entryFee, "Must provide entry fee + stack");
        require(securityDeposits[msg.sender] == 0, "Already deposited");

        securityDeposits[msg.sender] = entryFee;
        playerStacks[msg.sender] = msg.value - entryFee;

        emit DepositMade(msg.sender, entryFee, playerStacks[msg.sender]);

        bool allJoined = true;
        for(uint i=0; i<players.length; i++){
            if(securityDeposits[players[i]] == 0){
                allJoined = false;
                break;
            }
        }
        if (allJoined) {
            state = GameState.Active;
        }
    }

    function topUpStack() external payable onlyPlayers {
        require(state == GameState.Active, "Game must be active to top up");
        require(msg.value > 0, "Must send ETH");

        playerStacks[msg.sender] += msg.value;
        emit StackTopUp(msg.sender, msg.value, playerStacks[msg.sender]);
    }

    function submitResult(address winner, uint256 potAmount) external onlyPlayers {
        require(state == GameState.Active, "Game not active");
        require(isPlayer[winner], "Invalid winner candidate");
        require(votes[msg.sender] == address(0), "Already voted");

        votes[msg.sender] = winner;
        voteCount[winner]++;

        emit WinnerVoted(msg.sender, winner);
        if (voteCount[winner] > players.length / 2) {
            _resolveHand(winner, potAmount);
        }
    }

    function _resolveHand(address winner, uint256 potAmount) internal {
        playerStacks[winner] += potAmount;
        for(uint i = 0; i < players.length; i++) {
            delete votes[players[i]];
            delete voteCount[players[i]];
        }
        emit HandResolved(winner, potAmount);
    }

    function exitGameFairly() external onlyPlayers noReentrant {
        require(securityDeposits[msg.sender] > 0, "No funds in contract");
        uint256 totalReturn  = securityDeposits[msg.sender] + playerStacks[msg.sender];

        uint256 sLog = playerStacks[msg.sender];
        uint256 dLog = securityDeposits[msg.sender];

        securityDeposits[msg.sender] = 0;
        playerStacks[msg.sender] = 0;
        isPlayer[msg.sender] = false;

        (bool success, ) = payable(msg.sender).call{value:totalReturn}("");
        require(success, "Withdrawal failed");

        emit FairExit(msg.sender, sLog, dLog);
    }

    function challengePlayer(address offender) external onlyPlayers {
        require(isPlayer[offender], "Not an active player");
        require(challengeDeadline[offender] == 0, "Challenge already active");

        challengeDeadline[offender] = block.timestamp + CHALLENGE_WINDOW;
        emit ChallengeStarted(msg.sender, offender);
    }

    function respondToChallenge() external onlyPlayers {
        challengeDeadline[msg.sender] = 0;
    }

    function performSlash(address offender) external noReentrant {
        require(challengeDeadline[offender] != 0, "No challenge exists");
        require(block.timestamp > challengeDeadline[offender], "Challenge window still open");

        uint256 slashAmount = securityDeposits[offender];
        uint256 stackToReturn = playerStacks[offender];

        securityDeposits[offender] = 0;
        playerStacks[offender] = 0;
        isPlayer[offender] = false;
        challengeDeadline[offender] = 0;

        if (stackToReturn > 0) {
            (bool success, ) = payable(offender).call{value: stackToReturn}("");
            require(success, "Stack withdrawal failed");
        }

        uint256 activeCount = 0;
        for (uint i = 0; i < players.length; i++) {
            if (isPlayer[players[i]]) activeCount++;
        }

        if (activeCount > 0) {
            uint256 share = slashAmount / activeCount;
            for (uint i = 0; i < players.length; i++) {
                if (isPlayer[players[i]]) {
                    (bool success, ) = payable(players[i]).call{value: share}("");
                    require(success, "Failed to redeem slashed amount");
                }
            }
        }
        emit PlayerSlashed(offender, slashAmount);
    }

    function getPlayerStacks(address player) external view returns (uint256) {
        return playerStacks[player];
    }
}