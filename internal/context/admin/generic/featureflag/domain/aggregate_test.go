package domain

import (
	"testing"
)

func TestNewFeatureFlag_Success(t *testing.T) {
	f, err := NewFeatureFlag("dark_mode", "Enable dark mode", FlagTypeBool, "false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Key != "dark_mode" {
		t.Errorf("Key expected dark_mode, got: %s", f.Key)
	}
	if f.FlagType != FlagTypeBool {
		t.Errorf("FlagType expected bool, got: %s", f.FlagType)
	}
	if f.Active {
		t.Error("new flag should be inactive")
	}
	if f.RolloutPct != 0 {
		t.Errorf("RolloutPct expected 0, got: %d", f.RolloutPct)
	}
}

func TestNewFeatureFlag_EmptyKey(t *testing.T) {
	_, err := NewFeatureFlag("", "desc", FlagTypeBool, "false")
	if err != ErrEmptyKey {
		t.Errorf("expected ErrEmptyKey, got: %v", err)
	}
}

func TestNewFeatureFlag_DefaultType(t *testing.T) {
	f, err := NewFeatureFlag("test", "test", "", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.FlagType != FlagTypeBool {
		t.Errorf("empty FlagType should default to bool, got: %s", f.FlagType)
	}
}

func TestFeatureFlag_Activate_Deactivate(t *testing.T) {
	f, _ := NewFeatureFlag("test", "test", FlagTypeBool, "false")

	f.Activate()
	if !f.Active {
		t.Error("flag should be active after Activate()")
	}

	f.Deactivate()
	if f.Active {
		t.Error("flag should be inactive after Deactivate()")
	}
}

func TestFeatureFlag_SetRollout(t *testing.T) {
	f, _ := NewFeatureFlag("test", "test", FlagTypeBool, "false")

	if err := f.SetRollout(50); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.RolloutPct != 50 {
		t.Errorf("RolloutPct expected 50, got: %d", f.RolloutPct)
	}
}

func TestFeatureFlag_SetRollout_Invalid(t *testing.T) {
	f, _ := NewFeatureFlag("test", "test", FlagTypeBool, "false")

	tests := []struct {
		name string
		pct  int
	}{
		{"negative", -1},
		{"over 100", 101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := f.SetRollout(tt.pct); err != ErrInvalidRolloutPct {
				t.Errorf("expected ErrInvalidRolloutPct, got: %v", err)
			}
		})
	}
}

func TestFeatureFlag_IsEnabledForUser_InactiveFlag(t *testing.T) {
	f, _ := NewFeatureFlag("test", "test", FlagTypeBool, "false")
	// Flag is inactive by default
	if f.IsEnabledForUser("user-1") {
		t.Error("inactive flag should not be enabled for any user")
	}
}

func TestFeatureFlag_IsEnabledForUser_FullRollout(t *testing.T) {
	f, _ := NewFeatureFlag("test", "test", FlagTypeBool, "false")
	f.Activate()
	f.SetRollout(100)

	if !f.IsEnabledForUser("user-1") {
		t.Error("100% rollout should enable for all users")
	}
	if !f.IsEnabledForUser("user-2") {
		t.Error("100% rollout should enable for all users")
	}
}

func TestFeatureFlag_IsEnabledForUser_ZeroRollout(t *testing.T) {
	f, _ := NewFeatureFlag("test", "test", FlagTypeBool, "false")
	f.Activate()
	f.SetRollout(0)

	if f.IsEnabledForUser("user-1") {
		t.Error("0% rollout should not enable for any user")
	}
}

func TestFeatureFlag_IsEnabledForUser_Deterministic(t *testing.T) {
	f, _ := NewFeatureFlag("feature_x", "test", FlagTypeBool, "false")
	f.Activate()
	f.SetRollout(50)

	// Same user should always get the same result
	result1 := f.IsEnabledForUser("user-stable")
	result2 := f.IsEnabledForUser("user-stable")
	if result1 != result2 {
		t.Error("IsEnabledForUser should be deterministic for the same user")
	}
}

func TestFeatureFlag_Update(t *testing.T) {
	f, _ := NewFeatureFlag("test", "old desc", FlagTypeBool, "false")

	newDesc := "new desc"
	newVal := "true"
	active := true
	rollout := 75

	err := f.Update(&newDesc, &newVal, &active, &rollout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Description != "new desc" {
		t.Errorf("Description expected new desc, got: %s", f.Description)
	}
	if f.DefaultValue != "true" {
		t.Errorf("DefaultValue expected true, got: %s", f.DefaultValue)
	}
	if !f.Active {
		t.Error("Active should be true")
	}
	if f.RolloutPct != 75 {
		t.Errorf("RolloutPct expected 75, got: %d", f.RolloutPct)
	}
}

func TestFeatureFlag_Update_InvalidRollout(t *testing.T) {
	f, _ := NewFeatureFlag("test", "desc", FlagTypeBool, "false")
	rollout := 150
	err := f.Update(nil, nil, nil, &rollout)
	if err != ErrInvalidRolloutPct {
		t.Errorf("expected ErrInvalidRolloutPct, got: %v", err)
	}
}
