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

type sensitiveWordTrieNode struct {
	children map[rune]*sensitiveWordTrieNode
	word     string
	category string
	terminal bool
}

type sensitiveWordService struct {
	*Service
	root       *sensitiveWordTrieNode
	maxWordLen int
}

func NewSensitiveWordService(service *Service, sensitiveWordRepository repository.SensitiveWordRepository) (SensitiveWordService, error) {
	words, err := sensitiveWordRepository.List(context.Background())
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, errors.New("no sensitive words loaded from postgres")
	}

	root, maxWordLen := buildSensitiveWordTrie(words)

	return &sensitiveWordService{
		Service:    service,
		root:       root,
		maxWordLen: maxWordLen,
	}, nil
}

func (s *sensitiveWordService) Check(text string) (*SensitiveWordCheckResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text must not be empty")
	}

	textRunes := []rune(text)
	matches := s.matchAll(textRunes)
	return &SensitiveWordCheckResult{
		Contains:     len(matches) > 0,
		FilteredText: replaceWithAsterisk(textRunes, matches),
		Matches:      matches,
	}, nil
}

func buildSensitiveWordTrie(words []model.SensitiveWord) (*sensitiveWordTrieNode, int) {
	root := &sensitiveWordTrieNode{}
	maxWordLen := 0

	for _, word := range words {
		wordRunes := []rune(word.Word)
		if len(wordRunes) == 0 {
			continue
		}
		if len(wordRunes) > maxWordLen {
			maxWordLen = len(wordRunes)
		}

		node := root
		for _, r := range wordRunes {
			if node.children == nil {
				node.children = make(map[rune]*sensitiveWordTrieNode)
			}
			child := node.children[r]
			if child == nil {
				child = &sensitiveWordTrieNode{}
				node.children[r] = child
			}
			node = child
		}

		node.terminal = true
		node.word = word.Word
		node.category = word.Type
	}

	return root, maxWordLen
}

func (s *sensitiveWordService) matchAll(textRunes []rune) []SensitiveWordMatch {
	if len(textRunes) == 0 || s.root == nil || s.maxWordLen == 0 {
		return []SensitiveWordMatch{}
	}

	candidates := make([]SensitiveWordMatch, 0)
	for start := range textRunes {
		node := s.root
		limit := len(textRunes)
		if maxEnd := start + s.maxWordLen; maxEnd < limit {
			limit = maxEnd
		}

		for end := start; end < limit; end++ {
			node = node.children[textRunes[end]]
			if node == nil {
				break
			}
			if node.terminal {
				candidates = append(candidates, SensitiveWordMatch{
					Word:     node.word,
					Category: node.category,
					StartPos: start,
					EndPos:   end + 1,
				})
			}
		}
	}

	if len(candidates) == 0 {
		return []SensitiveWordMatch{}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		leftLen := candidates[i].EndPos - candidates[i].StartPos
		rightLen := candidates[j].EndPos - candidates[j].StartPos
		if leftLen != rightLen {
			return leftLen > rightLen
		}
		if candidates[i].StartPos != candidates[j].StartPos {
			return candidates[i].StartPos < candidates[j].StartPos
		}
		return candidates[i].EndPos < candidates[j].EndPos
	})

	matches := make([]SensitiveWordMatch, 0, len(candidates))
	occupied := make([]bool, len(textRunes))
	for _, candidate := range candidates {
		overlap := false
		for i := candidate.StartPos; i < candidate.EndPos; i++ {
			if occupied[i] {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}

		for i := candidate.StartPos; i < candidate.EndPos; i++ {
			occupied[i] = true
		}
		matches = append(matches, candidate)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].StartPos != matches[j].StartPos {
			return matches[i].StartPos < matches[j].StartPos
		}
		return matches[i].EndPos < matches[j].EndPos
	})

	return matches
}

func replaceWithAsterisk(textRunes []rune, matches []SensitiveWordMatch) string {
	runes := append([]rune(nil), textRunes...)
	for _, match := range matches {
		for i := match.StartPos; i < match.EndPos && i < len(runes); i++ {
			runes[i] = '*'
		}
	}
	return string(runes)
}
