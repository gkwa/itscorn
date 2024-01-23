package main

import (
	"os"

	"github.com/taylormonacelli/itscorn"
)

func main() {
	code := itscorn.Execute()
	os.Exit(code)
}
