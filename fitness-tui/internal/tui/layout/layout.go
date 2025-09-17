package layout

type Layout struct {
	Width  int
	Height int
}

func NewLayout(width, height int) *Layout {
	return &Layout{
		Width:  width,
		Height: height,
	}
}
