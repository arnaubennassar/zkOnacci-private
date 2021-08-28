pragma solidity ^0.8.6;

import * as _verifier from "./verifier.sol";
import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";

contract ZKOnacci is ERC721 {
    // Circuit
    uint256 public root;
    // NFT metadata
    string constant public baseURI = "ipfs://";
    uint256 public tokenCounter;
    uint8 constant public nTiers = 4;
    uint16[nTiers] public tokenTiers = [2, 4, 8, 16];
    string[nTiers] public tokenURIs = [
        "1", // 3 copies (0,1,2)
        "2", // 2 copies (3,4)
        "3", // 4 copies (5,6,7,8)
        "4"  // 8 copies (9,10,...,16)
    ];
    _verifier.Verifier private verifier;

    constructor(address verifierAddr) public ERC721 ("zkOnacci", "ZKO"){
        // Set the first two numbers of the sequence [0, 1]
        tokenCounter = 0;
        root = 19733998167332688543494136895553318319796515049857122158390636597337826955912;
        verifier = _verifier.Verifier(verifierAddr);
    }

    function captureTheFlag (
            uint[2] memory proofA,
            uint[2][2] memory proofB,
            uint[2] memory proofC
    ) public returns (uint256) {
        // Check if all tokens have been minted
        require(
            tokenCounter <= tokenTiers[nTiers-1],
            "ZKOnacci::captureTheFlag: ALL_TOKENS_MINTED"
        );
        // Verify proof
        require(
            verifier.verifyProof(
                proofA, proofB, proofC,
                [
                    uint256(uint160(msg.sender)),
                    root
                    // nextRoot
                ]
            ) == true,
            "ZKOnacci::captureTheFlag: INVALID_ZK_PROOF"
        );
        // Mint NFT
        return mintNFT();
    }

    function mintNFT() private returns (uint256) {
        uint256 newItemId = tokenCounter;
        _safeMint(msg.sender, newItemId);
        tokenCounter++;
        return newItemId;
    }

    function tokenURI(uint256 tokenId) public view virtual override returns (string memory) {
        require(_exists(tokenId), "ERC721Metadata: URI query for nonexistent token");

        // NFTs have different URIs according to how many of them had been minted when they where created.
        uint8 tokenTierIndex = 0;
        while (tokenTierIndex < tokenTiers.length - 1 && tokenId > tokenTiers[tokenTierIndex]) {
            tokenTierIndex++;
        }
        return string(abi.encodePacked(baseURI, tokenURIs[tokenTierIndex]));
    }
}