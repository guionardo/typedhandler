package banchmarks

import "testing"

func BenchmarkStringEmpty(b *testing.B) {
	s := "Some text"

	b.Run("compare_to_empty_string", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			if s == "" {
				_ = true
			}
		}
	})
	b.Run("compare_to_empty_string_variable", func(b *testing.B) {
		empty := ""

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			if s != empty {
				_ = false
			}
		}
	})
	b.Run("len_equals_zero", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			if len(s) == 0 { //nolint
				_ = true
			}
		}
	})
}
