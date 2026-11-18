package translate

import (
	"strings"
	"testing"
)

func TestTranslatorChineseToEnglish(t *testing.T) {
	trans := NewTranslator()

	tests := []struct {
		cnQuery  string
		expected string
	}{
		{"赛博朋克2077", "Cyberpunk 2077"},
		{"2077", "Cyberpunk 2077"},
		{"艾尔登法环", "ELDEN RING"},
		{"老头环", "ELDEN RING"},
		{"荒野大镖客2", "Red Dead Redemption 2"},
		{"大镖客2", "Red Dead Redemption 2"},
		{"黑神话悟空", "Black Myth: Wukong"},
		{"黑猴", "Black Myth: Wukong"},
		{"博德之门3", "Baldur's Gate 3"},
		{"生化危机4", "Resident Evil 4"},
		{"怪猎世界", "Monster Hunter: World"},
	}

	for _, tt := range tests {
		t.Run(tt.cnQuery, func(t *testing.T) {
			res := trans.TranslateChineseToEnglish(tt.cnQuery)
			found := false
			for _, item := range res {
				if item == tt.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("TranslateChineseToEnglish(%s) expected %s in %v", tt.cnQuery, tt.expected, res)
			}
		})
	}
}

func TestTranslatorEnglishToChinese(t *testing.T) {
	trans := NewTranslator()

	tests := []struct {
		enQuery  string
		expected string
	}{
		{"Cyberpunk 2077", "赛博朋克2077"},
		{"ELDEN RING", "艾尔登法环"},
		{"Red Dead Redemption 2", "荒野大镖客：救赎2"},
		{"Black Myth: Wukong", "黑神话：悟空"},
		{"Baldur's Gate 3", "博德之门3"},
	}

	for _, tt := range tests {
		t.Run(tt.enQuery, func(t *testing.T) {
			res := trans.TranslateEnglishToChinese(tt.enQuery)
			found := false
			for _, item := range res {
				if item == tt.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("TranslateEnglishToChinese(%s) expected %s in %v", tt.enQuery, tt.expected, res)
			}
		})
	}
}

func TestPinyinConversion(t *testing.T) {
	trans := NewTranslator()

	// 1. 全拼转换
	py := trans.ToPinyin("赛博朋克2077")
	if !strings.HasPrefix(py, "saibopengke") {
		t.Errorf("expected pinyin prefix 'saibopengke', got '%s'", py)
	}

	// 2. 首字母转换
	initials := trans.ToPinyinInitials("赛博朋克2077")
	if !strings.HasPrefix(initials, "sbpk") {
		t.Errorf("expected initials prefix 'sbpk', got '%s'", initials)
	}

	// 3. 艾尔登法环 -> aedfh
	ringInitials := trans.ToPinyinInitials("艾尔登法环")
	if ringInitials != "aedfh" {
		t.Errorf("expected initials 'aedfh', got '%s'", ringInitials)
	}
}

func TestExpandQueryTerms(t *testing.T) {
	trans := NewTranslator()

	// 1. 中文扩展包含英文与拼音
	termsCN := trans.ExpandQueryTerms("赛博朋克2077")
	containsEN := false
	containsPinyin := false
	for _, term := range termsCN {
		if strings.Contains(term, "Cyberpunk") {
			containsEN = true
		}
		if strings.Contains(term, "saibo") {
			containsPinyin = true
		}
	}
	if !containsEN || !containsPinyin {
		t.Errorf("expected expanded terms to contain English and Pinyin, got %v", termsCN)
	}

	// 2. 拼音首字母匹配 (sbpk -> Cyberpunk 2077)
	termsInit := trans.ExpandQueryTerms("sbpk")
	containsCyberpunk := false
	for _, term := range termsInit {
		if strings.Contains(term, "Cyberpunk") {
			containsCyberpunk = true
		}
	}
	if !containsCyberpunk {
		t.Errorf("expected initials 'sbpk' to expand to Cyberpunk, got %v", termsInit)
	}
}

