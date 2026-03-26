package i18n

var locales = map[string]map[string]string{
	"en": {
		"context": "Context", "elapsed": "Elapsed", "ratelimit": "Remaining", "cost": "Cost",
		"api": "API", "in": "In", "out": "Out",
	},
	"ko": {
		"context": "컨텍스트", "elapsed": "경과", "ratelimit": "잔여", "cost": "비용",
		"api": "API 대기", "in": "입력", "out": "출력",
	},
	"ja": {
		"context": "コンテキスト", "elapsed": "経過", "ratelimit": "残り", "cost": "費用",
		"api": "API待機", "in": "入力", "out": "出力",
	},
	"zh": {
		"context": "上下文", "elapsed": "已用", "ratelimit": "剩余", "cost": "费用",
		"api": "API等待", "in": "输入", "out": "输出",
	},
}

type Locale struct {
	lang string
}

func New(lang string) Locale {
	if _, ok := locales[lang]; !ok {
		lang = "en"
	}
	return Locale{lang: lang}
}

func (l Locale) Get(key string) string {
	if m, ok := locales[l.lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := locales["en"]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

func (l Locale) Lang() string { return l.lang }
