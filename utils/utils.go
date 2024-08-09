package utils

import (
	"crypto/sha256"

	native_mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

func ByteSlicesToDoubleHashes(inputData [][]byte) [][]byte {
	var convertedData [][]byte
	for _, data := range inputData {
		convertedData = append(convertedData, DoubleHashMimc(data))
	}
	return convertedData
}

func HashMimc(data []byte) []byte {
	m := native_mimc.NewMiMC()
	m.Write(data)
	return m.Sum(nil)
}

func DoubleHashMimc(data []byte) []byte {
	hashData := HashMimc(data)
	return HashMimc(hashData)
}

func HashSha256(data []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
