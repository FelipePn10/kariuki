package autocomplete

import (
	"fmt"

	input_autocomplete "github.com/JoaoDanielRufino/go-input-autocomplete"
)

// A useful input that can autocomplete users path to directories or files when tab key is pressed.
// The purpose is to be similar to bash/cmd native autocompletion.

func InputAutocomplete() {
	path, err := input_autocomplete.Read("Path: ")
	if err != nil {
		panic(err)
	}
	fmt.Println(path)
}
