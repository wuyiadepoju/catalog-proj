package domain

import "math/big"

type Money *big.Rat

func NewMoney(amount int64) Money {
	return big.NewRat(amount, 100)
}

// NewMoneyFromFraction creates Money from numerator and denominator
func NewMoneyFromFraction(numerator, denominator int64) Money {
	return big.NewRat(numerator, denominator)
}

func Add(m, other Money) Money {
	result := new(big.Rat).Add(m, other)
	return result
}

func Subtract(m, other Money) Money {
	result := new(big.Rat).Sub(m, other)
	return result
}

func Multiply(m, other Money) Money {
	result := new(big.Rat).Mul(m, other)
	return result
}
