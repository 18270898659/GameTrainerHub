package translate

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

type Translator struct {
	items []GameDictionaryItem
}

func NewTranslator() *Translator {
	return &Translator{
		items: BuiltinDictionary,
	}
}

// AddCustomItem 动态添加自定义游戏词条
func (t *Translator) AddCustomItem(item GameDictionaryItem) {
	t.items = append(t.items, item)
}

// IsChinese 判断字符串中是否包含汉字字符
func (t *Translator) IsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// ToPinyin 将汉字字符串转换为全拼（无声调、小写拼接）
func (t *Translator) ToPinyin(text string) string {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	a.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	res := pinyin.Pinyin(text, a)
	var sb strings.Builder
	for _, slice := range res {
		for _, py := range slice {
			sb.WriteString(strings.ToLower(py))
		}
	}
	return sb.String()
}

// ToPinyinInitials 将汉字字符串转换为拼音首字母
func (t *Translator) ToPinyinInitials(text string) string {
	a := pinyin.NewArgs()
	a.Style = pinyin.FirstLetter
	a.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}

	res := pinyin.Pinyin(text, a)
	var sb strings.Builder
	for _, slice := range res {
		for _, py := range slice {
			sb.WriteString(strings.ToLower(py))
		}
	}
	return sb.String()
}

// TranslateChineseToEnglish 输入中文名称/别名/流行语，返回匹配的英文游戏名候选集
func (t *Translator) TranslateChineseToEnglish(query string) []string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return nil
	}

	var results []string
	seen := make(map[string]bool)

	for _, item := range t.items {
		// 1. 完全或包含匹配标准中文名
		if strings.Contains(strings.ToLower(item.NameCN), normalized) || strings.Contains(normalized, strings.ToLower(item.NameCN)) {
			if !seen[item.NameEN] {
				results = append(results, item.NameEN)
				seen[item.NameEN] = true
			}
			continue
		}

		// 2. 匹配别名列表
		for _, alias := range item.Aliases {
			if strings.EqualFold(alias, normalized) || strings.Contains(strings.ToLower(alias), normalized) || strings.Contains(normalized, strings.ToLower(alias)) {
				if !seen[item.NameEN] {
					results = append(results, item.NameEN)
					seen[item.NameEN] = true
				}
				break
			}
		}
	}

	return results
}

// TranslateEnglishToChinese 输入英文名称，返回匹配的中文标准名及别名候选集
func (t *Translator) TranslateEnglishToChinese(query string) []string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return nil
	}

	var results []string
	seen := make(map[string]bool)

	for _, item := range t.items {
		if strings.Contains(strings.ToLower(item.NameEN), normalized) || strings.EqualFold(item.Slug, normalized) {
			if !seen[item.NameCN] && item.NameCN != "" {
				results = append(results, item.NameCN)
				seen[item.NameCN] = true
			}
			for _, alias := range item.Aliases {
				if !seen[alias] {
					results = append(results, alias)
					seen[alias] = true
				}
			}
		}
	}

	return results
}

// ExpandQueryTerms 生成搜索关键词扩展集（用于混合检索）
func (t *Translator) ExpandQueryTerms(rawQuery string) []string {
	clean := strings.TrimSpace(rawQuery)
	if clean == "" {
		return nil
	}

	var terms []string
	seen := make(map[string]bool)
	add := func(term string) {
		term = strings.TrimSpace(term)
		if term != "" && !seen[strings.ToLower(term)] {
			seen[strings.ToLower(term)] = true
			terms = append(terms, term)
		}
	}

	// 原始词
	add(clean)

	if t.IsChinese(clean) {
		// 中文关键词扩展：英译结果、全拼、首字母
		enCandidates := t.TranslateChineseToEnglish(clean)
		for _, en := range enCandidates {
			add(en)
		}

		py := t.ToPinyin(clean)
		if py != clean {
			add(py)
		}
		initials := t.ToPinyinInitials(clean)
		if initials != clean {
			add(initials)
		}
	} else {
		// 英文或拼音关键词扩展：中译结果
		cnCandidates := t.TranslateEnglishToChinese(clean)
		for _, cn := range cnCandidates {
			add(cn)
		}

		// 检查是否可能是拼音首字母匹配（如 "sbpk", "aedfh"）
		lower := strings.ToLower(clean)
		for _, item := range t.items {
			initCN := t.ToPinyinInitials(item.NameCN)
			pyCN := t.ToPinyin(item.NameCN)
			if initCN == lower || strings.HasPrefix(initCN, lower) || pyCN == lower || strings.HasPrefix(pyCN, lower) {
				add(item.NameEN)
				add(item.NameCN)
			}
		}
	}

	return terms
}

