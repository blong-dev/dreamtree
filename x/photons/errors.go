package photons

import "cosmossdk.io/errors"

var ErrBadRecipient = errors.Register(ModuleName, 2, "storer_reward_recipient must be a valid address or empty")

// ErrMintCeilingExceeded is returned when a batch's mint would push the current
// block's total minted photons past MaxMintPerBlock (trust-layer W3). It fails
// the whole MsgCommitBatch, so no seeds are registered and nothing is minted.
var ErrMintCeilingExceeded = errors.Register(ModuleName, 3, "per-block photon mint ceiling exceeded")
