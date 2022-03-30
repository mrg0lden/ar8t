package ar8t

func Correct(block []byte, blockInfo BlockInfo) ([]byte, error)

func CorrectWithErrorCount(block []byte, blockInfo BlockInfo) ([]byte, int, error)

func calculateSyndromes(block []byte, blockInfo BlockInfo) bool
