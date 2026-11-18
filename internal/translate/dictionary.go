package translate

// GameDictionaryItem 游戏词典词条
type GameDictionaryItem struct {
	NameEN   string   // 英文标准名称
	NameCN   string   // 中文标准名称
	Aliases  []string // 中文别名、简称、网络流行昵称
	Slug     string   // 规范标识符
}

// BuiltinDictionary 内置常用热门 PC 游戏词典
var BuiltinDictionary = []GameDictionaryItem{
	{
		NameEN:  "Cyberpunk 2077",
		NameCN:  "赛博朋克2077",
		Aliases: []string{"赛博朋克", "2077", "夜之城"},
		Slug:    "cyberpunk-2077",
	},
	{
		NameEN:  "ELDEN RING",
		NameCN:  "艾尔登法环",
		Aliases: []string{"老头环", "法环", "黄金树幽影", "黄金树之影"},
		Slug:    "elden-ring",
	},
	{
		NameEN:  "Red Dead Redemption 2",
		NameCN:  "荒野大镖客：救赎2",
		Aliases: []string{"荒野大镖客2", "大镖客2", "救赎2", "RDR2"},
		Slug:    "red-dead-redemption-2",
	},
	{
		NameEN:  "Black Myth: Wukong",
		NameCN:  "黑神话：悟空",
		Aliases: []string{"黑神话悟空", "黑猴", "黑神话", "悟空"},
		Slug:    "black-myth-wukong",
	},
	{
		NameEN:  "Baldur's Gate 3",
		NameCN:  "博德之门3",
		Aliases: []string{"博德之门", "BG3"},
		Slug:    "baldurs-gate-3",
	},
	{
		NameEN:  "Resident Evil 4",
		NameCN:  "生化危机4",
		Aliases: []string{"生化危机4重制版", "生化4", "RE4", "恶灵古堡4"},
		Slug:    "resident-evil-4",
	},
	{
		NameEN:  "Monster Hunter: World",
		NameCN:  "怪物猎人：世界",
		Aliases: []string{"怪猎世界", "怪猎", "怪物猎人", "MHW"},
		Slug:    "monster-hunter-world",
	},
	{
		NameEN:  "Monster Hunter Wilds",
		NameCN:  "怪物猎人：荒野",
		Aliases: []string{"怪猎荒野", "怪物猎人荒野"},
		Slug:    "monster-hunter-wilds",
	},
	{
		NameEN:  "The Witcher 3: Wild Hunt",
		NameCN:  "巫师3：狂猎",
		Aliases: []string{"巫师3", "猎魔人3", "白狼"},
		Slug:    "the-witcher-3-wild-hunt",
	},
	{
		NameEN:  "Grand Theft Auto V",
		NameCN:  "侠盗猎车手5",
		Aliases: []string{"GTA5", "横行霸道5", "洛圣都"},
		Slug:    "grand-theft-auto-v",
	},
	{
		NameEN:  "Hogwarts Legacy",
		NameCN:  "霍格沃茨之遗",
		Aliases: []string{"霍格沃茨的遗产", "霍格沃茨", "哈利波特"},
		Slug:    "hogwarts-legacy",
	},
	{
		NameEN:  "Palworld",
		NameCN:  "幻兽帕鲁",
		Aliases: []string{"帕鲁"},
		Slug:    "palworld",
	},
	{
		NameEN:  "God of War",
		NameCN:  "战神4",
		Aliases: []string{"战神", "战神PC版", "奎托斯"},
		Slug:    "god-of-war",
	},
	{
		NameEN:  "God of War Ragnarok",
		NameCN:  "战神：诸神黄昏",
		Aliases: []string{"诸神黄昏", "战神5"},
		Slug:    "god-of-war-ragnarok",
	},
	{
		NameEN:  "Dark Souls III",
		NameCN:  "黑暗之魂3",
		Aliases: []string{"黑魂3", "传火"},
		Slug:    "dark-souls-iii",
	},
	{
		NameEN:  "Sekiro: Shadows Die Twice",
		NameCN:  "只狼：影逝二度",
		Aliases: []string{"只狼", "打铁", "影逝二度"},
		Slug:    "sekiro-shadows-die-twice",
	},
	{
		NameEN:  "Starfield",
		NameCN:  "星空",
		Aliases: []string{"星空之地"},
		Slug:    "starfield",
	},
	{
		NameEN:  "ARMORED CORE VI FIRES OF RUBICON",
		NameCN:  "装甲核心6：境界天火",
		Aliases: []string{"装甲核心6", "机战佣兵6", "AC6"},
		Slug:    "armored-core-vi-fires-of-rubicon",
	},
	{
		NameEN:  "Fallout 4",
		NameCN:  "辐射4",
		Aliases: []string{"辐射4年度版", "垃圾王4"},
		Slug:    "fallout-4",
	},
	{
		NameEN:  "The Elder Scrolls V: Skyrim",
		NameCN:  "上古卷轴5：天际",
		Aliases: []string{"上古卷轴5", "老滚5", "天际"},
		Slug:    "the-elder-scrolls-v-skyrim",
	},
	{
		NameEN:  "Lies of P",
		NameCN:  "匹诺曹的谎言",
		Aliases: []string{"P的谎言", "匹诺曹"},
		Slug:    "lies-of-p",
	},
	{
		NameEN:  "Helldivers 2",
		NameCN:  "绝地潜兵2",
		Aliases: []string{"地狱潜者2", "绝地潜兵", "地狱潜者"},
		Slug:    "helldivers-2",
	},
	{
		NameEN:  "Dragon's Dogma 2",
		NameCN:  "龙之信条2",
		Aliases: []string{"龙信2"},
		Slug:    "dragons-dogma-2",
	},
	{
		NameEN:  "Ghost of Tsushima",
		NameCN:  "对马岛之魂",
		Aliases: []string{"对马岛之魂导演剪辑版", "对马岛"},
		Slug:    "ghost-of-tsushima",
	},
	{
		NameEN:  "Final Fantasy VII Remake",
		NameCN:  "最终幻想7：重制版",
		Aliases: []string{"最终幻想7重制版", "FF7R", "最终幻想7"},
		Slug:    "final-fantasy-vii-remake",
	},
	{
		NameEN:  "Wo Long: Fallen Dynasty",
		NameCN:  "卧龙：苍天陨落",
		Aliases: []string{"卧龙苍天陨落", "卧龙"},
		Slug:    "wo-long-fallen-dynasty",
	},
	{
		NameEN:  "Death Stranding",
		NameCN:  "死亡搁浅",
		Aliases: []string{"送快递", "小岛秀夫"},
		Slug:    "death-stranding",
	},
	{
		NameEN:  "Horizon Zero Dawn",
		NameCN:  "地平线：零之曙光",
		Aliases: []string{"地平线零之曙光", "地平线"},
		Slug:    "horizon-zero-dawn",
	},
	{
		NameEN:  "Dave the Diver",
		NameCN:  "潜水员戴夫",
		Aliases: []string{"潜水员戴夫", "捕鱼戴夫"},
		Slug:    "dave-the-diver",
	},
	{
		NameEN:  "Remnant II",
		NameCN:  "遗迹2",
		Aliases: []string{"遗迹2来自灰烬"},
		Slug:    "remnant-ii",
	},
}

