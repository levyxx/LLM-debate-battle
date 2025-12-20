package openai

// ディベートテーマ生成用のスキーマ
var DebateTopicSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"topic": map[string]any{
			"type":        "string",
			"description": "ディベートのテーマ（議論の対象となる命題）",
		},
		"pro_position": map[string]any{
			"type":        "string",
			"description": "賛成側の立場の説明",
		},
		"con_position": map[string]any{
			"type":        "string",
			"description": "反対側の立場の説明",
		},
		"background": map[string]any{
			"type":        "string",
			"description": "このテーマの背景や重要性の説明",
		},
	},
	"required":             []string{"topic", "pro_position", "con_position", "background"},
	"additionalProperties": false,
}

// ディベート引数生成用のスキーマ
var DebateArgumentSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"argument": map[string]any{
			"type":        "string",
			"description": "メインの主張・反論",
		},
		"key_points": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
			"description": "主張を支える重要なポイント",
		},
		"counterpoint": map[string]any{
			"type":        "string",
			"description": "相手への反論ポイント（あれば）",
		},
	},
	"required":             []string{"argument", "key_points", "counterpoint"},
	"additionalProperties": false,
}

// 審査結果用のスキーマ
var JudgeResultSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"winner": map[string]any{
			"type":        "string",
			"enum":        []string{"pro", "con", "draw"},
			"description": "勝者（pro=賛成側, con=反対側, draw=引き分け）",
		},
		"score": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pro": map[string]any{
					"type":        "integer",
					"description": "賛成側のスコア（0-100）",
				},
				"con": map[string]any{
					"type":        "integer",
					"description": "反対側のスコア（0-100）",
				},
			},
			"required":             []string{"pro", "con"},
			"additionalProperties": false,
		},
		"reasoning": map[string]any{
			"type":        "string",
			"description": "判定理由の詳細説明",
		},
		"pro_strengths": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
			"description": "賛成側の良かった点",
		},
		"pro_weaknesses": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
			"description": "賛成側の改善点",
		},
		"con_strengths": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
			"description": "反対側の良かった点",
		},
		"con_weaknesses": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
			"description": "反対側の改善点",
		},
		"final_comment": map[string]any{
			"type":        "string",
			"description": "審査員からの総評コメント",
		},
	},
	"required":             []string{"winner", "score", "reasoning", "pro_strengths", "pro_weaknesses", "con_strengths", "con_weaknesses", "final_comment"},
	"additionalProperties": false,
}

// LLM同士のディベート継続判定用スキーマ
var DebateContinueSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"should_continue": map[string]any{
			"type":        "boolean",
			"description": "ディベートを続けるべきかどうか",
		},
		"reason": map[string]any{
			"type":        "string",
			"description": "判断理由",
		},
	},
	"required":             []string{"should_continue", "reason"},
	"additionalProperties": false,
}
