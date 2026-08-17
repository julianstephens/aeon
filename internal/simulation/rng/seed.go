package rng

import "crypto/sha256"

// DeriveSeed derives a new seed from the given base seed and derivation path.
func DeriveSeed(baseSeed [32]byte, derivationPath string) [32]byte {
	h := sha256.New()
	h.Write(baseSeed[:])
	h.Write([]byte(derivationPath))

	var newSeed [32]byte
	copy(newSeed[:], h.Sum(nil))
	return newSeed
}
