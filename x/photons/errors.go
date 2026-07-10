package photons

import "cosmossdk.io/errors"

var ErrBadRecipient = errors.Register(ModuleName, 2, "storer_reward_recipient must be a valid address or empty")
