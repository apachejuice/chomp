package chomp

// Represents a chess piece
type Piece int8

const (
	PieceNone = -1

	PieceWhitePawn   = 0
	PieceWhiteRook   = 1
	PieceWhiteKnight = 2
	PieceWhiteBishop = 3
	PieceWhiteQueen  = 4
	PieceWhiteKing   = 5

	PieceBlackPawn   = 6
	PieceBlackRook   = 7
	PieceBlackKnight = 8
	PieceBlackBishop = 9
	PieceBlackQueen  = 10
	PieceBlackKing   = 11
)

var pieceNames = map[Piece]string{
	PieceNone: "<none>",

	PieceWhitePawn:   "white pawn",
	PieceWhiteRook:   "white rook",
	PieceWhiteKnight: "white knight",
	PieceWhiteBishop: "white bishop",
	PieceWhiteQueen:  "white queen",
	PieceWhiteKing:   "white king",

	PieceBlackPawn:   "black pawn",
	PieceBlackRook:   "black rook",
	PieceBlackKnight: "black knight",
	PieceBlackBishop: "black bishop",
	PieceBlackQueen:  "black queen",
	PieceBlackKing:   "black king",
}

func (p Piece) Name() string {
	name, ok := pieceNames[p]
	if !ok {
		return "<invalid>"
	}

	return name
}

func (p Piece) String() string {
	return p.Name()
}

// Represents a color
type Color int8

const (
	ColorNone  = -1
	ColorWhite = 0
	ColorBlack = 2
)

// Represents a chess board
type Board struct {
	// The grid is represented as an array of numbers. Each number is a constant representing a piece.
	// The grid can be accessed by first accessing the X and then the Y coordinate.
	Grid [8][8]Piece `json:"grid"`
	// The color that is currently the victim of a check.
	Checked Color `json:"checked"`
	// Checkmate. If this value is anything else than ColorNone, the game has ended, and the opposite of this color won.
	Checkmate Color `json:"checkmate"`
}
