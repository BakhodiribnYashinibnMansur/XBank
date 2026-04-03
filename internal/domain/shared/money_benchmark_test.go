package shared

import "testing"

func BenchmarkMoney_Add(b *testing.B) {
	m1 := Money{Amount: 1000000, Currency: UZS}
	m2 := Money{Amount: 500000, Currency: UZS}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m1.Add(m2)
	}
}

func BenchmarkMoney_Subtract(b *testing.B) {
	m1 := Money{Amount: 1000000, Currency: UZS}
	m2 := Money{Amount: 500000, Currency: UZS}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m1.Subtract(m2)
	}
}

func BenchmarkMoney_String(b *testing.B) {
	m := Money{Amount: 1500050, Currency: UZS}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.String()
	}
}

func BenchmarkNewMoney(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewMoney(int64(i), USD)
	}
}
