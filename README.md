# TrustDraw

A protocol for dealing and playing with a shuffled deck of cards in the open, using RSA, AES and Ed25519 encryption.

This protocol does not currently support returning cards to the deck, but will be extended to support it. It will require a further call to the dealer.

## Try it out

```sh
# Install the TrustDraw CLI
$ go install github.com/jphastings/trustdraw@latest
go: downloading github.com/jphastings/trustdraw v1.0.0

# Deal a deck to play with
$ trustdraw deal standard52-fr test_data/dealer.pem test_data/player1.pub.pem test_data/player2.pub.pem > example.deal

# Verify that the deck was created by the dealer to prevent cheating
$ trustdraw verify example.deal test_data/dealer.pub.pem
✅ example.deal is a valid deck of 52 cards for 2 players

# As Player 2, get an allowKey, to allow Player 1 to draw a card
$ trustdraw allow-draw example.deal test_data/player2.pem 1
Creating example-game.player2.state to hold game state…
Your allowKey: BABFpJBzhiVJwMonZIDVDjk4

# As Player 1, use the allowKey given by Player 2 to draw a card
$ trustdraw draw example.deal test_data/player1.pem BABFpJBzhiVJwMonZIDVDjk4
Creating example-game.player1.state to hold game state…
You drew: 3♦️
Prove with: AACH+oA5nhR+JoulasCyHrmv

# As Player 2, when Player 1 plays 🃓, verify that they really drew that card
$ trustdraw verify-draw example.deal test_data/player2.pem 🃓 AACH+oA5nhR+JoulasCyHrmv
✅ This was a valid draw

# Demonstrate that a cheating draw is detectable
$ trustdraw verify-draw example.deal test_data/player2.pem 🂱 AACH+oA5nhR+JoulasCyHrmv
❌ This was not a valid draw
```

## Protocol

Below is a walk-through of the deal and a draw of a two player card game using this protocol. This also works for more players.

To **deal the cards**:

1. Both players send their public RSA keys to the dealer.
2. Dealer generates 52 AES keys for player 1, and 52 for player 2.
3. Dealer pairs off the keys for player 1 and 2, and XORs them to make 52 combined keys.
4. Dealer pairs off each of the (shuffled) cards ("K♥", "2♣️", "A♦", …N) with each of the combined keys, and symmetrically encrypts the card with the key — this is the "shuffled deck". _(`AES-128-ECB`)_
5. Dealer encrypts all Player 1's keys (in order, the "key stack"), for Player 1's eyes only, using Player 1's public RSA key. _(`AES-128-CTR` preceeded by `RSA(key)`)_
6. …and does the same for Player 2.
7. Dealer publishes the shuffled deck and these two encrypted blocks, all signed with a dealer's key (`Ed25519`), to demonstrate authenticity, as the "deal file".

To **draw a card**:

1. Player 2 retrieves the deal file and decrypts their key stack (with their private key)
2. Player 2 finds the top-most unused AES key from their key stack (recording it as "dealt to player 1") and shares it with Player 1.
3. Player 1 retrieves the deal file and decrypts their own key stack, and finds their top-most unused AES key from it in the same way, recording it as used by themselves
4. Player 1 XORs their key and the one received from Player 2 to make the combined key.
5. Player 1 uses this combined key to decrypt the top-most unused card from the "shuffled deck" and now has drawn a card!
6. During play, when this card is played, Player 1 declares their part of the AES key used to decrypt the card, so Player 2 can verify they actually drew it.
