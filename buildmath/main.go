//If you need again  go:generate gojson -input export.json  -o factoriotype.go -subStruct -name FactorioData -camelcasefields
//go:generate go-bindata  -pkg main -o export.go data
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
)

var (
	data     FactorioData
	itemKeys []string
)

func main() {
	err := json.Unmarshal(MustAsset("data/export.json"), &data)
	if err != nil {
		panic(err)
	}
	for key := range data.Recipes {
		if data.Items[key].Name == "" {
			itemKeys = append(itemKeys, key)
		}
	}
	for key := range data.Items {
		itemKeys = append(itemKeys, key)
	}
	sort.Strings(itemKeys)

	p := prompt.New(executor, completer,
		prompt.OptionTitle("Factorio Builds: "),
		prompt.OptionPrefix("Pick item: "),
		prompt.OptionMaxSuggestion(20))
	p.Run()
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}

	var cmd string

	prev := d.FindStartOfPreviousWord()
	if prev != 0 {
		cmd = strings.Split(d.CurrentLineBeforeCursor()[:], " ")[0]
	}

	switch cmd {
	case "build":
		for _, i := range itemKeys {
			if data.Recipes[i].Name != "" {
				s = append(s, prompt.Suggest{Text: i, Description: data.Recipes[i].Name})
			}
		}
	case "items":
		for _, i := range itemKeys {
			if data.Items[i].Name != "" {
				s = append(s, prompt.Suggest{Text: i, Description: data.Items[i].Name})
			}
		}
	case "":
		s = append(s, prompt.Suggest{Text: "items"})
		s = append(s, prompt.Suggest{Text: "build"})
		s = append(s, prompt.Suggest{Text: "recipes"})
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executor(in string) {
	// fmt.Println("Your input: " + in)
	args := strings.Split(in, " ")
	switch args[0] {
	case "build":
		HowToMakeItem(args[1])
	case "exit", "e", "quit", "q":
		os.Exit(0)
	default:
		log.Println(in)
	}
	// fmt.Println(data.Recipes[in])
	// fmt.Println(data.Assemblers[in])
	// fmt.Println(data.Resources[in])
}

func ctrlC(b *prompt.Buffer) {
}

type BuildPlan struct {
}

func BuildItemsPerSecond(ItemsPerSecond int, itemToBuild Item) BuildPlan {
	//Tree of how to build
	//AssemblersNeeded

	return BuildPlan{}
}

func HowToMakeItem(item string) {
	toBuild, ok := data.Recipes[item]
	if !ok {
		return
	}
	fmt.Println(toBuild.Name)
	for r, amt := range ingredientsToString(toBuild.Ingredients, 1) {
		fmt.Println(r, amt)
	}
	fmt.Println("")
}

func ingredientsToString(ingredients []IngriedentProduct, depth int) map[string]int64 {
	var indent string
	for i := 0; i < depth; i++ {
		indent += "\t"
	}
	resources := make(map[string]int64)
	for _, i := range ingredients {
		fmt.Printf("%s%s %s %d\n", indent, i.Name, i.Type, i.Amount)
		if _, ok := data.Resources[i.Name]; ok {
			resources[i.Name] += i.Amount
		} else {
			for r, amt := range ingredientsToString(data.Recipes[i.Name].Ingredients, depth+1) {
				resources[r] += amt * i.Amount
			}
		}
	}
	return resources
}
