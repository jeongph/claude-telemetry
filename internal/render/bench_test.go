package render

import "testing"

func BenchmarkDisplayWidth(b *testing.B) {
	s := "Opus 💭high │ ◷ 경과 12m 34s │ myproject:main ↑1 +5/-2"
	for i := 0; i < b.N; i++ {
		DisplayWidth(s)
	}
}

func BenchmarkDisplayWidthCJK(b *testing.B) {
	s := "◆ 컨텍스트 ▰▰▰▱▱ 72% (200k) │ 2h 12m/5h ▰▰▰▰▱ 88%"
	for i := 0; i < b.N; i++ {
		DisplayWidth(s)
	}
}

func BenchmarkProgressBar(b *testing.B) {
	c := NewColors(true)
	for i := 0; i < b.N; i++ {
		ProgressBarRemaining(72, 5, c, 50, 20)
	}
}

func BenchmarkStripANSI(b *testing.B) {
	s := "\033[1;36mOpus\033[0m \033[2;37m💭high\033[0m │ \033[32m+156\033[0m"
	for i := 0; i < b.N; i++ {
		StripANSI(s)
	}
}
