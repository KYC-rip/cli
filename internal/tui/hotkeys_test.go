package tui

import "testing"

// TestPickersDisableGlobalHotkeys is a regression for a real user-reported
// bug: typing 'LTC' in the From picker jumped to the Track tab because
// 't' was a global hotkey and the picker state wasn't flagged as a
// typing state. Same for 'TRX', 'TRC20', 'TON', etc.
func TestPickersDisableGlobalHotkeys(t *testing.T) {
	cases := []struct {
		name string
		tab  tab
		st   swapState
		want bool
	}{
		{"PickFrom typing", tabSwap, stPickFrom, true},
		{"PickTo typing", tabSwap, stPickTo, true},
		{"Amount typing", tabSwap, stAmount, true},
		{"Address typing", tabSwap, stAddress, true},
		{"Memo typing", tabSwap, stMemo, true},

		// Non-typing states — global hotkeys should still work.
		{"Quoting non-typing", tabSwap, stQuoting, false},
		{"Quoted non-typing", tabSwap, stQuoted, false},
		{"Ordered non-typing", tabSwap, stOrdered, false},
		{"Error non-typing", tabSwap, stError, false},
	}
	for _, tc := range cases {
		m := Model{tab: tc.tab, state: tc.st}
		if got := m.isTypingState(); got != tc.want {
			t.Errorf("%s: isTypingState=%v, want %v", tc.name, got, tc.want)
		}
	}
}
