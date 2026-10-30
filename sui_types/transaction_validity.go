// Copyright (c) Mysten Labs, Inc.
// SPDX-License-Identifier: Apache-2.0
package sui_types

import (
	"fmt"
	"io"

	"github.com/fardream/go-bcs/bcs"
)

// ValidityExpiration is TransactionExpiration variant 3. Field order matches
// Sui's wire format; proposer restrictions must survive transaction re-encoding.
type ValidityExpiration struct {
	MinEpoch         *EpochId `bcs:"optional"`
	MaxEpoch         *EpochId `bcs:"optional"`
	MinTimestamp     *uint64  `bcs:"optional"`
	MaxTimestamp     *uint64  `bcs:"optional"`
	Chain            CheckpointDigest
	Nonce            uint32
	AllowedProposers *AllowedProposers `bcs:"optional"`
}

type AllowedProposers struct {
	Epoch     EpochId
	Proposers []uint32
}

// Sui serializes NonEmpty<u32> as Vec<u32> but rejects empty lists on decode.
// Ordering and committee membership are transaction validation, not BCS rules.
type allowedProposersWire AllowedProposers

func (p AllowedProposers) MarshalBCS() ([]byte, error) {
	if len(p.Proposers) == 0 {
		return nil, fmt.Errorf("allowed proposers must be nonempty")
	}
	return bcs.Marshal(allowedProposersWire(p))
}

func (p *AllowedProposers) UnmarshalBCS(r io.Reader) (int, error) {
	var wire allowedProposersWire
	n, err := bcs.NewDecoder(r).Decode(&wire)
	if err != nil {
		return n, err
	}
	if len(wire.Proposers) == 0 {
		return n, fmt.Errorf("allowed proposers must be nonempty")
	}
	*p = AllowedProposers(wire)
	return n, nil
}
