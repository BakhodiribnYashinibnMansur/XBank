package domain

import (
	"testing"
)

func TestNewTranslation_Success(t *testing.T) {
	tr, err := NewTranslation("error.not_found", LangEn, "Not found", "errors")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Key != "error.not_found" {
		t.Errorf("Key expected error.not_found, got: %s", tr.Key)
	}
	if tr.Language != LangEn {
		t.Errorf("Language expected en, got: %s", tr.Language)
	}
	if tr.Value != "Not found" {
		t.Errorf("Value expected Not found, got: %s", tr.Value)
	}
	if tr.Group != "errors" {
		t.Errorf("Group expected errors, got: %s", tr.Group)
	}
	if tr.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewTranslation_EmptyKey(t *testing.T) {
	_, err := NewTranslation("", LangUz, "value", "group")
	if err != ErrEmptyKey {
		t.Errorf("expected ErrEmptyKey, got: %v", err)
	}
}

func TestNewTranslation_EmptyValue(t *testing.T) {
	_, err := NewTranslation("key", LangRu, "", "group")
	if err != ErrEmptyValue {
		t.Errorf("expected ErrEmptyValue, got: %v", err)
	}
}

func TestNewTranslation_EmptyLanguage(t *testing.T) {
	_, err := NewTranslation("key", "", "value", "group")
	if err != ErrEmptyLanguage {
		t.Errorf("expected ErrEmptyLanguage, got: %v", err)
	}
}

func TestTranslation_Update(t *testing.T) {
	tr, _ := NewTranslation("key", LangUz, "old value", "group")

	if err := tr.Update("new value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Value != "new value" {
		t.Errorf("Value expected new value, got: %s", tr.Value)
	}
}

func TestTranslation_Update_EmptyValue(t *testing.T) {
	tr, _ := NewTranslation("key", LangEn, "value", "group")

	err := tr.Update("")
	if err != ErrEmptyValue {
		t.Errorf("expected ErrEmptyValue, got: %v", err)
	}
}

func TestLanguageConstants(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LangUz, "uz"},
		{LangRu, "ru"},
		{LangEn, "en"},
	}

	for _, tt := range tests {
		if string(tt.lang) != tt.want {
			t.Errorf("Language expected %s, got: %s", tt.want, tt.lang)
		}
	}
}
