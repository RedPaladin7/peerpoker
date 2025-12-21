// SPDX-License-Identifier-MIT 
pragma solidity ^0.8.20;

import "./PokerEscrow.sol";

contract PokerFactoy {
    struct GameInfo {
        address gameAddress;
        bytes32 gameId;
        address[] participants;
    }

    GameInfo[] public allGames;

    event GameCreated(address indexed gameAddress, bytes32 indexed gameId);

    function createGame(
        bytes32 gameId,
        uint256 entryFee,
        address[] calldata participants
    ) external returns (address) {
        require(participants.length >= 2, "Need at least 2 players");

        PokerEscrow newGame = new PokerEscrow(gameId, entryFee, participants);

        allGames.push(GameInfo({
            gameAddress: address(newGame),
            gameId: gameId,
            participants: participants
        }));

        emit GameCreated(address(newGame), gameId);

        return address(newGame);
    }

    function getGameCount() external view returns (uint256) {
        return allGames.length;
    }
}