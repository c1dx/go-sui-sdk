// Copyright (c) Mysten Labs, Inc.
// SPDX-License-Identifier: Apache-2.0
package sui_types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/fardream/go-bcs/bcs"
	"os"
	"testing"
)

// Vectors are independently encoded from sui-types transaction.rs at e847565c:
// enum index, four Option<u64>, byte-vector chain digest, u32 nonce, and for
// Validity only, Option<(u64, nonempty Vec<u32>)>. Old indices stay unchanged.
func expirationWireVectors(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("testdata/transaction_expiration_wire.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string]string
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func TestTransactionExpirationWireVectors(t *testing.T) {
	for name, h := range expirationWireVectors(t) {
		t.Run(name, func(t *testing.T) {
			raw, err := hex.DecodeString(h)
			if err != nil {
				t.Fatal(err)
			}
			var got TransactionExpiration
			n, err := bcs.Unmarshal(raw, &got)
			if err != nil || n != len(raw) {
				t.Fatalf("decode consumed %d/%d: %v", n, len(raw), err)
			}
			back, err := bcs.Marshal(got)
			if err != nil || !bytes.Equal(raw, back) {
				t.Fatalf("wire changed: %x -> %x: %v", raw, back, err)
			}
		})
	}
}

func TestTransactionValidityRejectsMalformedWire(t *testing.T) {
	raw, _ := hex.DecodeString(expirationWireVectors(t)["validity_proposers"])
	for n := 0; n < len(raw); n++ {
		var got TransactionExpiration
		if _, err := bcs.Unmarshal(raw[:n], &got); err == nil {
			t.Fatalf("accepted truncation at %d", n)
		}
	}
	// Replace Vec length with zero and remove its three u32 entries.
	empty := append([]byte(nil), raw[:len(raw)-12]...)
	empty[len(empty)-1] = 0
	for _, bad := range [][]byte{empty, {4}, {255, 1}} {
		var got TransactionExpiration
		if _, err := bcs.Unmarshal(bad, &got); err == nil {
			t.Fatalf("accepted malformed wire %x", bad)
		}
	}
}

func TestTransactionValidityFieldsAndNonemptyEncoding(t *testing.T) {
	raw, _ := hex.DecodeString(expirationWireVectors(t)["validity_proposers"])
	var got TransactionExpiration
	if _, err := bcs.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	v := got.Validity
	if v == nil || v.MinEpoch == nil || v.MaxEpoch == nil || v.MinTimestamp == nil || v.MaxTimestamp == nil {
		t.Fatal("missing validity fields")
	}
	if *v.MinEpoch != 1296 || *v.MaxEpoch != 1297 || *v.MinTimestamp != 1000 || *v.MaxTimestamp != 2000 || v.Nonce != 0x12345678 {
		t.Fatal("wrong validity values")
	}
	if v.AllowedProposers == nil || v.AllowedProposers.Epoch != 1296 || len(v.AllowedProposers.Proposers) != 3 || v.AllowedProposers.Proposers[1] != 31 {
		t.Fatal("wrong proposer restriction")
	}
	v.AllowedProposers.Proposers = nil
	if _, err := bcs.Marshal(got); err == nil {
		t.Fatal("encoded empty proposer restriction")
	}
}
