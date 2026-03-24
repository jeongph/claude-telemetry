package i18n

import "testing"

func TestGetLabel(t *testing.T) {
	l := New("ko")
	if l.Get("context") != "컨텍스트" {
		t.Errorf("ko context = %q", l.Get("context"))
	}
}

func TestGetLabelEnglish(t *testing.T) {
	l := New("en")
	if l.Get("context") != "Context" {
		t.Errorf("en context = %q", l.Get("context"))
	}
}

func TestGetLabelJapanese(t *testing.T) {
	l := New("ja")
	if l.Get("context") != "コンテキスト" {
		t.Errorf("ja context = %q", l.Get("context"))
	}
}

func TestGetLabelChinese(t *testing.T) {
	l := New("zh")
	if l.Get("context") != "上下文" {
		t.Errorf("zh context = %q", l.Get("context"))
	}
}

func TestGetLabelFallback(t *testing.T) {
	l := New("fr")
	if l.Get("context") != "Context" {
		t.Errorf("fallback should return English")
	}
}

func TestGetLabelUnknownKey(t *testing.T) {
	l := New("en")
	if l.Get("nonexistent") != "nonexistent" {
		t.Errorf("unknown key should return key itself")
	}
}

func TestLang(t *testing.T) {
	l := New("ko")
	if l.Lang() != "ko" {
		t.Errorf("Lang() = %q, want ko", l.Lang())
	}
}
