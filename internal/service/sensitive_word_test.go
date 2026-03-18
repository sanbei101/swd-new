package service

import (
	"testing"

	"swd-new/internal/model"
)

func TestSensitiveWordMatchAndReplace(t *testing.T) {
	svc := &sensitiveWordService{
		words: []model.SensitiveWord{
			{Word: "蠢猪", Type: "脏话"},
			{Word: "坏蛋", Type: "辱骂"},
		},
	}

	resp, err := svc.Check("你这个蠢猪真是个坏蛋")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !resp.Contains {
		t.Fatalf("expected contains to be true")
	}

	if resp.FilteredText != "你这个**真是个**" {
		t.Fatalf("unexpected filtered text: %s", resp.FilteredText)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("unexpected match count: %d", len(resp.Matches))
	}

	if resp.Matches[0].Word != "蠢猪" || resp.Matches[0].Category != "脏话" || resp.Matches[0].StartPos != 3 || resp.Matches[0].EndPos != 5 {
		t.Fatalf("unexpected first match: %+v", resp.Matches[0])
	}

	if resp.Matches[1].Word != "坏蛋" || resp.Matches[1].Category != "辱骂" || resp.Matches[1].StartPos != 8 || resp.Matches[1].EndPos != 10 {
		t.Fatalf("unexpected second match: %+v", resp.Matches[1])
	}
}

func BenchmarkCheckSensitiveWords(b *testing.B) {
	svc := &sensitiveWordService{
		words: []model.SensitiveWord{
			{Word: "蠢猪", Type: "脏话"},
			{Word: "坏蛋", Type: "辱骂"},
		},
	}
	for b.Loop() {
		svc.Check("你这个蠢猪真是个坏蛋")
	}
}
