package domain

import "testing"

func BenchmarkMoney_Add(b *testing.B) {
	a, _ := NewMoney(1000, UZS)
	c, _ := NewMoney(2000, UZS)
	for b.Loop() {
		_, _ = a.Add(c)
	}
}

func BenchmarkMoney_Subtract(b *testing.B) {
	a, _ := NewMoney(5000, UZS)
	c, _ := NewMoney(2000, UZS)
	for b.Loop() {
		_, _ = a.Subtract(c)
	}
}

func BenchmarkMoney_String(b *testing.B) {
	m, _ := NewMoney(1500050, UZS)
	for b.Loop() {
		_ = m.String()
	}
}

func BenchmarkNewMoney(b *testing.B) {
	for b.Loop() {
		_, _ = NewMoney(1500050, UZS)
	}
}
