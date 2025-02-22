package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"math/rand/v2"
	"os"
	"sort"
)

// Constants representing different colors.
// The values range from 1 to 10, starting with Red as 1.
var colors = []string{"Red", "Yellow", "Purple", "Orange", "Green", "Cyan", "Pink", "Blue", "Brown", "Magenta"}

// Constants representing different event types.
// The values are assigned using iota, starting from 0.
const (
	eventLuckyColor = iota
	eventOnePair
	eventLuckyStrike
	eventAllDifferent
	eventClear
)

// eventDesc is a slice of strings that contains the descriptions of different events in the game.
// The index of each description corresponds to an event type, which is typically represented by an integer constant.
// This slice is used to provide a human-readable description of the events when printing or displaying event information.
var eventDesc = []string{"Lucky Color", "One Pair", "Lucky Strike", "Family Portrait", "Clear The Board"}

type ev struct {
	acquired map[int]int
	event    int
}

// eventAcquired is a map that defines the reward values for different events.
// The keys represent specific event types (identified by event constants),
// and the values represent the number of toys acquired as a result of that event.
// This map is used to track the rewards associated with each event in the game.
var eventAcquired = map[int]int{
	eventLuckyColor:  0,
	eventOnePair:     2,
	eventLuckyStrike: 3,
}

// eventRewardRules is a map that defines the reward points for different events.
// The keys represent specific event types (identified by event constants),
// and the values represent the points awarded for that event.
// This map is used to track how many reward points each event gives to the player.
var eventRewardRules = map[int]int{
	eventLuckyColor:   1,
	eventOnePair:      1,
	eventLuckyStrike:  3,
	eventAllDifferent: 5,
	eventClear:        5,
}

// tripleCombination defines a 2D slice where each inner slice represents
// a combination of three indices that form a "triple combination" in a game or puzzle.
var tripleCombination = [][]int{
	// The first set of combinations (vertical lines in a 3x3 grid).
	{0, 3, 6},
	{1, 4, 7},
	{2, 5, 8},

	// The second set of combinations (horizontal lines in a 3x3 grid).
	{0, 1, 2},
	{3, 4, 5},
	{6, 7, 8},

	// The third set of combinations (diagonals in a 3x3 grid).
	{0, 4, 8},
	{2, 3, 6},
}

// packages is a slice that represents the number of toys in different packs.
// Each integer corresponds to a specific pack size, for example, 9, 18, and 35 toys per pack.
var packages = []int{9, 18, 30}

var initialOrderedSlots = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}

// die is a utility function that prints an error message and exits the program with a non-zero status.
// The msg parameter is a formatted string, and args are the arguments to format the string.
func die(msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

func main() {
	interactive()
}

// interactive function runs the main loop of the game, guiding the user through the entire gameplay process.
// It starts the game, selects the lucky color, selects the toy package, and then continuously places toys on the board,
// checks for events, and handles acquired items. The loop continues until all the remaining toys are placed.
func interactive() {
	startGame()
	luckColor := selectLuckColor()
	remaining := selectPackageType()
	board := make([]int, 9)
	acquired := make([]int, len(colors))
	orderedEmptySlots := initialOrderedSlots
	for remaining > 0 {
		events := make([]ev, 0)
		remaining, events, orderedEmptySlots = placeInSlot(board, orderedEmptySlots, events, remaining, luckColor)
		printBoard(board)
		events, orderedEmptySlots = checkBoard(board, orderedEmptySlots, events)
		printEvents(events)
		remaining = handleEvents(events, acquired, remaining)
		printAcquired(acquired, false)
		fmt.Printf("Remaining: %d\n", remaining)
		next()
	}
	for _, v := range board {
		if v > 0 {
			acquired[v-1] += 1
		}
	}
	printAcquired(acquired, true)
}

// placeInSlot function randomly places colors into empty slots on the board
// and generates events for lucky color occurrences during the process.
func placeInSlot(board, orderedEmptySlots []int, events []ev, remaining, luckyColor int) (int, []ev, []int) {
	for len(orderedEmptySlots) > 0 {
		if remaining <= 0 {
			break
		}
		remaining -= 1
		randColor := rand.IntN(cap(board)) + 1
		if randColor == luckyColor {
			events = append(events, ev{map[int]int{randColor: eventAcquired[eventLuckyColor]}, eventLuckyColor})
		}
		board[orderedEmptySlots[0]] = randColor
		orderedEmptySlots = orderedEmptySlots[1:]
	}
	return remaining, events, orderedEmptySlots
}

// checkBoard function checks the current state of the board for specific combinations and updates the board, empty slots, and events accordingly.
func checkBoard(board, orderedEmptySlots []int, events []ev) ([]ev, []int) {
	for _, comb := range tripleCombination {
		if board[comb[0]] != 0 && board[comb[0]] == board[comb[1]] && board[comb[0]] == board[comb[2]] {
			events = append(events, ev{map[int]int{board[comb[0]]: eventAcquired[eventLuckyStrike]}, eventLuckyStrike})
			orderedEmptySlots = append(orderedEmptySlots, comb...)
			board[comb[0]] = 0
			board[comb[1]] = 0
			board[comb[2]] = 0
		}
	}
	rt := make(map[int]int)
	for k, v := range board {
		if v > 0 {
			if pos, ok := rt[v]; ok {
				events = append(events, ev{map[int]int{board[k]: eventAcquired[eventOnePair]}, eventOnePair})
				board[pos] = 0
				board[k] = 0
				orderedEmptySlots = append(orderedEmptySlots, pos, k)
				delete(rt, v)
			} else {
				rt[v] = k
			}
		}
	}
	if len(orderedEmptySlots) == cap(board) {
		events = append(events, ev{map[int]int{}, eventClear})
	}
	if len(orderedEmptySlots) == 0 {
		acq := map[int]int{}
		for _, v := range board {
			acq[v] = 1
		}
		board = make([]int, 9)
		orderedEmptySlots = initialOrderedSlots
		events = append(events, ev{acq, eventAllDifferent})
	}
	sort.Slice(orderedEmptySlots, func(i, j int) bool {
		return orderedEmptySlots[i] < orderedEmptySlots[j]
	})
	return events, orderedEmptySlots
}

// handleEvents function processes a list of events and updates the acquired rewards for each event.
// It calculates the total reward based on the event rules and updates the acquired rewards for specific items.
func handleEvents(events []ev, acq []int, remaining int) int {
	n := 0
	for _, e := range events {
		n += eventRewardRules[e.event]
		for k, v := range e.acquired {
			acq[k-1] += v
		}
	}
	return n + remaining
}

// printEvents function prints the details of each event in the provided events list.
// It displays the event description and the associated reward for each event.
func printEvents(events []ev) {
	if len(events) != 0 {
		fmt.Println("========== events ==========")
	}
	for _, e := range events {
		fmt.Printf("Event: %-20s +%d\n", eventDesc[e.event], eventRewardRules[e.event])
	}
}

// printAcquired function prints the list of acquired items (e.g., toys) along with their quantities.
// If the `finish` flag is set to true, it also prints the total number of acquired items.
func printAcquired(acq []int, finish bool) {
	fmt.Println("========== acquired ==========")
	n := 0
	for k, v := range acq {
		fmt.Printf("%s: %d; ", colors[k], v)
		n += v
	}
	if finish {
		fmt.Printf("\nYou have received %d toys\n", n)
	}
}

// printBoard function prints the current state of the board, showing the items (e.g., colors) placed in each slot.
// If a slot is empty, it prints "Empty" for that slot. The board is printed in a grid format, with 3 items per row.
func printBoard(board []int) {
	fmt.Println("========== board ==========")
	for i, v := range board {
		if v <= 0 {
			fmt.Printf("%-10s ", "Empty")
		} else {
			fmt.Printf("%-10s ", colors[v-1])
		}
		if i%3 == 2 {
			fmt.Print("\n")
		}
	}
}

// next function prompts the user to press "Enter" to continue the game.
// It displays a prompt with the label "Please type enter to continue game" and waits for the user to press the Enter key.
func next() {
	prompt := promptui.Prompt{
		Label: "Please type enter to continue game",
	}
	_, _ = prompt.Run()
}

// startGame function displays a brief introduction to the game, listing the rewards for various events,
// and then prompts the user to press "Enter" to start the game.
// It provides an overview of the game rules and waits for the user to continue before starting the game.
func startGame() {
	description := `Game Introduction
1. Lucky Color +1
2. One Pair +1
3. Lucky Strike +3
4. Family Portrait +5
5. Clear The Board +5`
	fmt.Println(description)
	prompt := promptui.Prompt{
		Label: "Please type enter to start game",
	}
	_, _ = prompt.Run()
}

// selectPackageType function prompts the user to select a toy package from a list of available packages.
// It displays a list of packages, with each item showing the number of toys included in the package, and then waits for the user to choose one.
// After the user makes a selection, the function prints the selected package and returns the number of toys in the selected package.
func selectPackageType() int {
	items := make([]string, 0)
	for _, v := range packages {
		items = append(items, fmt.Sprintf("%d toys", v))
	}
	prompt := promptui.Select{
		Label: "Select your toy package",
		Items: items,
	}
	packIdx, _, err := prompt.Run()
	if err != nil {
		die("choose toy package failed, %v\n", err)
	}
	fmt.Printf("You choose %s \n", items[packIdx])
	return packages[packIdx]
}

// selectLuckColor function prompts the user to select their lucky color from a list of available colors.
// It displays a list of colors and waits for the user to choose one. After the user makes a selection,
// the function prints the selected color and returns the index of the chosen color (1-based).
func selectLuckColor() int {
	prompt := promptui.Select{
		Label: "Select your lucky color",
		Items: colors,
	}
	colorIdx, _, err := prompt.Run()
	if err != nil {
		die("choose lucky color failed, %v\n", err)
	}
	fmt.Printf("You choose %s \n", colors[colorIdx])
	return colorIdx + 1
}
