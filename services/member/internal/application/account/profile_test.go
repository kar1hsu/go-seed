package account

import "testing"

func TestNicknameRejectsInvisibleFormatting(t *testing.T) {
	for _, v := range []string{"name\u202e", "name\u200b", "bad\nname"} {
		if ValidNickname(v) {
			t.Fatalf("unsafe nickname accepted: %q", v)
		}
	}
	if !ValidNickname("正常名字") {
		t.Fatal("normal nickname rejected")
	}
}
