package texas_algorithm

import "fmt"

// Poke value constants (card ranks), aligned with Java's Poke.
const (
	PokeValue2  byte = 2
	PokeValue3  byte = 3
	PokeValue4  byte = 4
	PokeValue5  byte = 5
	PokeValue6  byte = 6
	PokeValue7  byte = 7
	PokeValue8  byte = 8
	PokeValue9  byte = 9
	PokeValue10 byte = 10
	PokeValueJ  byte = 11
	PokeValueQ  byte = 12
	PokeValueK  byte = 13
	PokeValueA  byte = 14
)

// PokeValues lists all ranks in ascending order.
var PokeValues = []byte{
	PokeValue2, PokeValue3, PokeValue4, PokeValue5, PokeValue6, PokeValue7,
	PokeValue8, PokeValue9, PokeValue10, PokeValueJ, PokeValueQ, PokeValueK, PokeValueA,
}

// Poke color constants, aligned with Java's Poke.
const (
	PokeColorFang byte = 0 // 方 diamonds
	PokeColorMei  byte = 1 // 梅 clubs
	PokeColorHong byte = 2 // 红 hearts
	PokeColorHei  byte = 3 // 黑 spades
)

// GUI is the wild card (ghost card), aligned with Java's Poke.GUI.
var GUI = Poke{Color: 5, Value: 8}

var huaseName = [...]string{"方", "梅", "红", "黑"}

var valueName = [...]string{"", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

// Poke is a single card: a color (suit) plus a value (rank).
type Poke struct {
	Color byte
	Value byte
}

// NewPoke builds a card from its color and value.
func NewPoke(color, value byte) Poke {
	return Poke{Color: color, Value: value}
}

// NewPokeFromByte decodes a card byte (high nibble color, low nibble value),
// the same encoding as Java's new Poke(byteValue).
func NewPokeFromByte(byteValue byte) Poke {
	return Poke{Color: byteValue >> 4, Value: byteValue % 16}
}

// IsGuiByte reports whether the card byte is a wild card,
// mirroring Java's static Poke.isGui(int).
func IsGuiByte(i int) bool {
	color := byte(i) >> 4
	value := byte(i) % 16
	return value == GUI.Value && color == GUI.Color
}

// IsGui reports whether the card is a wild card.
func (p Poke) IsGui() bool {
	return p.Value == GUI.Value && p.Color == GUI.Color
}

// ToByte encodes the card into a byte (color<<4 | value).
func (p Poke) ToByte() byte {
	return p.Color<<4 | p.Value
}

// String renders the card the same way as Java's Poke.toString:
// a Chinese suit name followed by the rank, e.g. "方A", "鬼".
func (p Poke) String() string {
	if p.IsGui() {
		return "鬼"
	}
	return fmt.Sprintf("%s%s", huaseName[p.Color], valueName[p.Value])
}
