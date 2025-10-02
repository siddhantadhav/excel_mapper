package excel_mapper

type HeaderMapping struct {
	InputHeader string
	OutputHeader OutputHeader
}

type OutputHeader struct {
	Name string
	Type string
}
