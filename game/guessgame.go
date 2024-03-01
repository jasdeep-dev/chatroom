// game/game.go
package game

import (
	"errors"
)

type GameState struct {
	WordToGuess string
	Guesses     []bool
}

func NewGame(word string) *GameState {
	return &GameState{
		WordToGuess: word,
		Guesses:     make([]bool, len(word)),
	}
}

func (g *GameState) GetTheWord() []int {
	text := g.WordToGuess
	indexes := StringToIndexes(text)

	return indexes
}

func StringToIndexes(s string) []int {
	var indexes []int
	for i, _ := range s {
		indexes = append(indexes, i)
	}
	return indexes
}

func (g *GameState) GuessLetter(letter rune) (data []string, err error) {
	if len(g.WordToGuess) == 0 {
		return nil, errors.New("no word set to guess")
	}

	var state []string

	for i, w := range g.WordToGuess {
		if w == letter && !g.Guesses[i] {
			g.Guesses[i] = true
		}

		if g.Guesses[i] == true {
			s := rune(w)
			state = append(state, string(s))
		} else {
			state = append(state, string('_'))
		}
	}
	return state, nil
}

func (g *GameState) AllGuessed() bool {
	for _, guessed := range g.Guesses {
		if !guessed {
			return false
		}
	}
	return true
}

func (g *GameState) CurrentWordState() string {
	result := ""
	for i, letter := range g.WordToGuess {
		if g.Guesses[i] {
			result += string(letter)
		} else {
			result += "_"
		}
	}
	return result
}
