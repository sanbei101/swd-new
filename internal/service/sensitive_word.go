package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"swd-new/internal/model"
	"swd-new/internal/repository"
)

type SensitiveWordMatch struct {
	Word     string `json:"word"`
	Category string `json:"category"`
	StartPos int    `json:"start_pos"`
	EndPos   int    `json:"end_pos"`
}

type SensitiveWordCheckResult struct {
	Contains     bool                 `json:"contains"`
	FilteredText string               `json:"filtered_text"`
	Matches      []SensitiveWordMatch `json:"matches"`
}

type SensitiveWordService interface {
	Check(text string) (*SensitiveWordCheckResult, error)
}

type sensitiveWordService struct {
	*Service
	words []model.SensitiveWord
}

func NewSensitiveWordService(service *Service, sensitiveWordRepository repository.SensitiveWordRepository) (SensitiveWordService, error) {
	words, err := sensitiveWordRepository.List(context.Background())
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, errors.New("no sensitive words loaded from postgres")
	}

	sort.SliceStable(words, func(i, j int) bool {
		return len([]rune(words[i].Word)) > len([]rune(words[j].Word))
	})

	return &sensitiveWordService{
		Service: service,
		words:   words,
	}, nil
}

func (s *sensitiveWordService) Check(text string) (*SensitiveWordCheckResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text must not be empty")
	}

	matches := s.matchAll(text)
	return &SensitiveWordCheckResult{
		Contains:     len(matches) > 0,
		FilteredText: replaceWithAsterisk(text, matches),
		Matches:      matches,
	}, nil
}

func (s *sensitiveWordService) matchAll(text string) []SensitiveWordMatch {
	textRunes := []rune(text)
	if len(textRunes) == 0 || len(s.words) == 0 {
		return []SensitiveWordMatch{}
	}

	matches := make([]SensitiveWordMatch, 0)
	occupied := make([]bool, len(textRunes))

	for _, word := range s.words {
		wordRunes := []rune(word.Word)
		if len(wordRunes) == 0 || len(wordRunes) > len(textRunes) {
			continue
		}

		for start := 0; start <= len(textRunes)-len(wordRunes); start++ {
			skip := false
			for i := start; i < start+len(wordRunes); i++ {
				if occupied[i] {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			ok := true
			for i := range wordRunes {
				if textRunes[start+i] != wordRunes[i] {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}

			for i := start; i < start+len(wordRunes); i++ {
				occupied[i] = true
			}

			matches = append(matches, SensitiveWordMatch{
				Word:     word.Word,
				Category: word.Type,
				StartPos: start,
				EndPos:   start + len(wordRunes),
			})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].StartPos != matches[j].StartPos {
			return matches[i].StartPos < matches[j].StartPos
		}
		return matches[i].EndPos < matches[j].EndPos
	})

	return matches
}

func replaceWithAsterisk(text string, matches []SensitiveWordMatch) string {
	runes := []rune(text)
	for _, match := range matches {
		for i := match.StartPos; i < match.EndPos && i < len(runes); i++ {
			runes[i] = '*'
		}
	}
	return string(runes)
}
