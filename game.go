package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

func main() {

	g := Game{
		PlayerOne: &HumanPlayer{},
		PlayerTwo: &AiPlayer{},
	}
	g.Run()
}

type Game struct {
	PlayerOne Player
	PlayerTwo Player

	boardOne, boardTwo Board
}

type Board [][]int

// 0 - no information
// 1 - player has a ship here
// 2 - op shot, but missed
// 3 - player has a ship here, but it's hit
// Shooting is simply `curval + 2`

func NewBoard() Board {
	out := make([][]int, 10)
	for idx, _ := range out {
		out[idx] = make([]int, 10)
	}
	return out
}

func (b Board) HidePlayerShips() Board {
	out := NewBoard()
	for rIdx, row := range b {
		for cIdx, v := range row {
			if v == 1 {
				continue
			}
			out[rIdx][cIdx] = v
		}
	}
	return out
}

// returns if this was already shot at, and whether this was a hit
func (b Board) Shoot(r, c int) (bool, bool) {
	hit := b[r][c]&1 == 1

	if b[r][c]&2 == 2 {
		return true, hit
	}

	b[r][c] |= 2
	return false, hit
}

func (b Board) LocationsAlive() int {
	out := 0
	for _, row := range b {
		for _, v := range row {
			if v == 1 {
				out += 1
			}
		}
	}
	return out
}

func (b Board) String() string {
	// 0 - no information - -
	// 1 - player has a ship here - S
	// 2 - op shot, but missed - O
	// 3 - player has a ship here, but it's hit - X
	// Shooting is simply `curval + 2`

	out := "  "
	c := "A"[0]
	for i := 0; i < len(b); i++ {
		uI := uint8(i)
		out += string(c+uI) + " "
	}
	out += "\n"

	for idx, row := range b {
		out += strconv.Itoa(idx) + " "
		for _, c := range row {
			switch c {
			case 0:
				out += "- "
			case 1:
				out += "S "
			case 2:
				out += "O "
			case 3:
				out += "X "
			}
		}
		out += "\n"
	}
	return out
}

// returns true if the ship could be written. False if it can't.
//
// False can occur due to a collision or falling off the edge
func (b Board) WriteShip(length int, r, c int, d Direction) bool {
	endR, endC := r, c
	switch d {
	case Up:
		endR -= length - 1
	case Down:
		endR += length - 1
	case Left:
		endC -= length - 1
	case Right:
		endC += length - 1
	}

	if endR < 0 || endR >= len(b) || endC < 0 || endC >= len(b[0]) {
		return false
	}

	if r > endR {
		tmp := endR
		endR = r
		r = tmp
	}

	if c > endC {
		tmp := endC
		endC = c
		c = tmp
	}

	for rIdx := r; rIdx <= endR; rIdx++ {
		for cIdx := c; cIdx <= endC; cIdx++ {
			if b[rIdx][cIdx] != 0 {
				return false
			}
		}
	}

	for rIdx := r; rIdx <= endR; rIdx++ {
		for cIdx := c; cIdx <= endC; cIdx++ {
			b[rIdx][cIdx] = 1
		}
	}
	return true
}

func (g *Game) Run() {
	if g.PlayerOne == nil || g.PlayerTwo == nil {
		panic("players must be set before game is run")
	}

	g.boardOne = g.PlayerOne.InitializeBoard()
	g.boardTwo = g.PlayerTwo.InitializeBoard()

	for true {
		for true {
			x, y := g.PlayerOne.Play(g.boardOne, g.boardTwo.HidePlayerShips())
			repeat, hit := g.boardTwo.Shoot(x, y)
			if !repeat {
				fmt.Println("Can't repeat a play!")
				break
			}
			g.PlayerOne.ReportPlayResult(x, y, hit)
		}

		if g.boardTwo.LocationsAlive() == 0 {
			fmt.Println(g.PlayerOne.Name() + " has won!")
			break
		}

		for true {
			x, y := g.PlayerTwo.Play(g.boardTwo, g.boardOne.HidePlayerShips())
			repeat, hit := g.boardOne.Shoot(x, y)
			if !repeat {
				break
			}
			g.PlayerTwo.ReportPlayResult(x, y, hit)
		}

		if g.boardOne.LocationsAlive() == 0 {
			fmt.Println(g.PlayerTwo.Name() + " has won!")
			break
		}
	}
}

type Player interface {
	InitializeBoard() Board // used to setup the game
	Play(playerBoard, opBoard Board) (int, int)
	ReportPlayResult(r, c int, hit bool)
	Name() string
}

var shipLengths = [...]int{5, 4, 3, 3, 2}

var _ Player = &HumanPlayer{}

type HumanPlayer struct{}

func getValidPos() (int, int) {
	var col int = -1
	var row int = -1
	for true {
		var input string

		fmt.Printf("(enter ColRow, e.g. 'A1') ")
		n, err := fmt.Scanln(&input)
		if err != nil || n < 1 || len(input) < 2 {
			fmt.Println("Invalid input")
			continue
		}

		rowChar := int(input[1])
		colChar := int(input[0])

		row = rowChar - int("0"[0])
		col = colChar - int("A"[0])

		if col < 0 || col > 9 || row < 0 || row > 9 {
			fmt.Println("Invalid input")
		} else {
			break
		}
	}

	return row, col
}

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
	Invalid
)

func getValidDirection() Direction {
	for true {
		fmt.Println("What orientation?")
		fmt.Print("(U/D/L/R) ")

		var outStr string
		_, err := fmt.Scanln(&outStr)
		if err != nil || len(outStr) < 1 {
			fmt.Println("Invalid input")
			continue
		}

		switch outStr[0] {
		case 'U':
			return Up
		case 'D':
			return Down
		case 'L':
			return Left
		case 'R':
			return Right
		}

		fmt.Println("Invalid input")
	}
	panic("not reachable")
}

func (p *HumanPlayer) InitializeBoard() Board {
	board := NewBoard()

	shipIdx := 0
	for shipIdx < len(shipLengths) {
		ship := shipLengths[shipIdx]
		fmt.Println(board)
		fmt.Printf("Where do you want to place a ship of length %d? \n", ship)
		row, col := getValidPos()
		dir := getValidDirection()

		if board.WriteShip(ship, row, col, dir) {
			shipIdx++
		} else {
			fmt.Println("Invalid placement.")
		}
	}
	return board
}

func (p *HumanPlayer) Play(myBoard, opBoard Board) (int, int) {
	fmt.Println(opBoard)
	fmt.Println(myBoard)
	fmt.Println("Where would you like to shoot?")
	return getValidPos()
}

func (p *HumanPlayer) ReportPlayResult(r, c int, hit bool) {
	if hit {
		fmt.Println("That was a hit!")
	} else {
		fmt.Println("That was a miss!")
	}
	fmt.Print("Enter to continue...")
	fmt.Scanln()
}

func (p *HumanPlayer) Name() string {
	return "Human"
}

var _ Player = &AiPlayer{}

type AiPlayer struct{}

func (p *AiPlayer) InitializeBoard() Board {
	board := NewBoard()

	shipIdx := 0
	for shipIdx < len(shipLengths) {
		r, c := rand.Intn(10), rand.Intn(10)
		d := Direction(rand.Intn(5)) // TODO: is there a better way to write this?

		if board.WriteShip(shipLengths[shipIdx], r, c, d) {
			shipIdx++
		}
	}

	return board
}

func (p *AiPlayer) Play(myBoard, opBoard Board) (int, int) {
	for true {
		r, c := rand.Intn(10), rand.Intn(10)
		if opBoard[r][c] == 0 {
			fmt.Printf("Computer shot at %d,%d\n", r, c)
			return r, c
		}
	}
	panic("unreachable")
}

func (p *AiPlayer) ReportPlayResult(r, c int, hit bool) {
	// do nothing
}

func (p *AiPlayer) Name() string {
	return "Computer"
}
