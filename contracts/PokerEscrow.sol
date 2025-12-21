// SPDX-License-Identifier-MIT 
pragma solidity ^0.8.20;

contract PokerEscrow {
    address public admin;
    uint256 public entryFee; 

    struct GameSession {
        address[] players;
        uint256 pot;
        bool isActive;
        mapping(address=>bool) hasDeposited;
    }

    mapping(bytes32=>GameSession) public games 

    event PlayerJoined(bytes32 indexed gameId, address player);
    event Payout(bytes32 indexed gameId, address winner, uint256 amount);
    event PlayerSlashed(bytes32 indexed gameId, address offender, uint256 penalty);

    constructor(uint256 _entryFee){
        admin = msg.sender;
        entryFee = _entryFee;
    }

    function joinGame(bytes32 gameId) external payable {
        require(msg.value == entryFee, "Incorrect entry fee");
        GameSession storage session = games[gameId];
        require(!session.hasDeposited[msg.sender], "Already in game");

        session.players.push(msg.sender);
        session.hasDeposited[msg.sender] = true;
        session.pot += msg.value
        session.isActive = true;

        emit PlayerJoined(gameId, msg.sender);
    }

    function distributePot(bytes32 gameId, address winner) external {
        require(msg.sender == admin, "Only admin can distribute pot");
        GameSession storage session = games[gameId];
        require(session.isActive, "Game not active");

        uint256 amount = session.pot;
        session.pot = 0;
        session.isActive = false;

        payable(winner).transfer(amount);
        emit Payout(gameId, winner, amount);
    }

    function slashPlayer(bytes32 gameId, address offender) external {
        require(msg.sender == admin, "Only admin can slash");
        GameSession storage session = games[gameId];
        require(session.hasDeposited[offender], "Player not found");

        uint256 penalty = entryFee;
        session.pot -= penalty;

        uint256 share = penalty / (sesson.players.length-1);
        for(uint i = o; i < session.players.length; i++) {
            if (session.players[i] != offender) {
                payable(session.players[i].transfer(share));
            }
        }
        session.hasDeposited[offender] = false;
        emit PlayerSlashed(gameId, offender, penalty);
    }
}