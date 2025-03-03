package utils

import (
	"io"
	"os"
)

func ComputeXORDelta(oldPath, newPath string) ([]byte, error) {
	oldData, err := os.ReadFile(oldPath)
	if err != nil {
		return nil, err
	}
	newData, err := os.ReadFile(newPath)
	if err != nil {
		return nil, err
	}

	// Take the shorter length
	minLen := len(oldData)
	if len(newData) < minLen {
		minLen = len(newData)
	}

	delta := make([]byte, minLen)
	for i := 0; i < minLen; i++ {
		delta[i] = oldData[i] ^ newData[i]
	}

	return delta, nil
}

// ApplyXORDelta reconstructs a file using an old version and XOR delta
func ApplyXORDelta(oldPath string, delta []byte, outputPath string) error {
	oldData, err := os.ReadFile(oldPath)
	if err != nil {
		return err
	}

	// Apply XOR delta
	minLen := len(oldData)
	if len(delta) < minLen {
		minLen = len(delta)
	}

	reconstructed := make([]byte, minLen)
	for i := 0; i < minLen; i++ {
		reconstructed[i] = oldData[i] ^ delta[i]
	}

	// Write reconstructed data to output file
	return os.WriteFile(outputPath, reconstructed, 0644)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
