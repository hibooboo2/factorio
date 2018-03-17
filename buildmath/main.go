//If you need again  go:generate gojson -input export.json  -o factoriotype.go -subStruct -name FactorioData -camelcasefields
//go:generate go-bindata  -pkg main -o export.go data
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

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
		prompt.OptionPrefix("$>  "),
		prompt.OptionMaxSuggestion(20))
	for {
		in := p.Input()
		executor(in)
		time.Sleep(time.Millisecond * 100)
		fmt.Print("\n")
	}
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
	case "items", "blueprint":
		for _, i := range itemKeys {
			if data.Items[i].Name != "" {
				s = append(s, prompt.Suggest{Text: i, Description: data.Items[i].Name})
			}
		}
	case "":
		s = append(s, prompt.Suggest{Text: "items"})
		s = append(s, prompt.Suggest{Text: "build"})
		s = append(s, prompt.Suggest{Text: "recipes"})
		s = append(s, prompt.Suggest{Text: "blueprint"})
		s = append(s, prompt.Suggest{Text: "times"})
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

var bluePrint BluePrint

func executor(in string) {
	// fmt.Println("Your input: " + in)
	args := strings.Split(in, " ")
	switch args[0] {
	case "build":
		if len(args) < 2 {
			fmt.Println("No item passed in.")
			return
		}
		numPersecond := 1
		if len(args) > 2 {
			numPersecond, _ = strconv.Atoi(args[2])
		}
		// HowToMakeItem(args[1], float64(numPersecond))
		toBuild, ok := data.Recipes[args[1]]
		if !ok {
			fmt.Printf("Recipe for %s not found", args[1])
			return
		}

		b, err := BuildItemsPerSecond(int64(numPersecond), toBuild, Assembler{CraftingSpeed: 0.75}, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(b)

		time.Sleep(time.Millisecond * 20)
	case "blueprint":
		bluePrint, err := DecodeBluePrint([]byte(args[1]))
		if err != nil {
			fmt.Println(err)
			return
		}
		out := EncodeBluePrint(*bluePrint)
		fmt.Println(string(out) == args[1])
	case "exit", "e", "quit", "q":
		os.Exit(0)
	case "times":
		for _, i := range data.Recipes {
			fmt.Println(i.Name, i.TimeToCraft, i.Products[0].Amount)
		}
	default:
		log.Println(in)
	}
	// fmt.Println(data.Recipes[in])
	// fmt.Println(data.Assemblers[in])
	// fmt.Println(data.Resources[in])
}

func ctrlC(b *prompt.Buffer) {
}

type Builder struct {
	Product       IngriedentProduct
	Assembler     Assembler
	AssemblerName string
	Num           int
	Level         int
	SubBuilders   []*Builder
	Resources     []*IngriedentProduct
}

func (b Builder) String() string {
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("%s\t%s\t%s %d\n", strings.Repeat("\t", b.Level), b.Product.Name, b.AssemblerName, b.Num))
	for _, b := range b.SubBuilders {
		if b == nil {
			continue
		}
		buff.WriteString(b.String())
	}
	return buff.String()
}

// BuildItemsPerSecond make  a buld for the item that produces n per second.
func BuildItemsPerSecond(itemsPerSecond int64, item interface{}, minAssembler Assembler, level int) (*Builder, error) {
	var toBuild Recipe

	switch val := item.(type) {
	case Item:
		r, ok := data.Recipes[val.Name]
		if !ok {
			_, ok := data.Resources[val.Name]
			if !ok {
				return nil, fmt.Errorf("no recipe for item: %s", val.Name)
			}
			return nil, nil
		}
		toBuild = r
	case Recipe:
		toBuild = val
	case string:
		r, ok := data.Recipes[val]
		if !ok {
			_, ok := data.Resources[val]
			if !ok {
				return nil, fmt.Errorf("no recipe for item: %s", val)
			}
			return nil, nil
		}
		toBuild = r
	default:
		return nil, fmt.Errorf("type %T not supported", val)
	}

	if len(toBuild.Products) > 1 {
		return nil, fmt.Errorf("only 1 product supported right now")
	}

	var b Builder
	b.Assembler, b.AssemblerName = pickAssembler(toBuild, minAssembler)

	b.Level = level

	b.Product = toBuild.Products[0]
	//Need to add productivity
	fmt.Printf("%s %v %v %v %v\n", b.Product.Name, itemsPerSecond, toBuild.TimeToCraft, toBuild.Products[0].Amount, b.Assembler.CraftingSpeed)
	fmt.Printf("Makes: %d each time\n", b.Product.Amount)

	b.Num = int(math.Round((float64(itemsPerSecond) * toBuild.TimeToCraft) / (float64(toBuild.Products[0].Amount) * b.Assembler.CraftingSpeed)))

	for _, i := range toBuild.Ingredients {
		subB, err := BuildItemsPerSecond(itemsPerSecond*i.Amount/b.Product.Amount, i.Name, minAssembler, level+1)
		if err != nil {
			return nil, err
		}
		b.SubBuilders = append(b.SubBuilders, subB)
		b.Resources = append(b.Resources, &i)
	}
	return &b, nil
}

func HowToMakeItem(item string, numPerSecond float64) {
	toBuild, ok := data.Recipes[item]
	if !ok {
		return
	}
	fmt.Println(toBuild.Name)
	var prod IngriedentProduct
	for _, p := range toBuild.Products {
		fmt.Println(p.Name, toBuild.Name)
		if p.Name != toBuild.Name {
			continue
		}
		fmt.Println(p)
		if p.Probability < 1 && p.Probability != 0 {
			panic(fmt.Errorf("%s not supported because probability < 1", p.Name))
		}
		prod = p
		break
	}
	if prod.Name == "" {
		fmt.Println("No products found matching:", toBuild.Name)
		return
	}
	amtPerSecondPerMachine := float64(toBuild.TimeToCraft / float64(prod.Amount))

	numResourcesNeeded := float64(1 / (float64(numPerSecond) / amtPerSecondPerMachine))

	fmt.Println(numResourcesNeeded)
	for r, amt := range ingredientsToString(toBuild.Ingredients, 1, numPerSecond) {
		fmt.Println(r, float64(amt)*numResourcesNeeded)
	}
	fmt.Println("")
}

func ingredientsToString(ingredients []IngriedentProduct, depth int, numPerSecond float64) map[string]int64 {
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
			for r, amt := range ingredientsToString(data.Recipes[i.Name].Ingredients, depth+1, numPerSecond) {
				resources[r] += amt * i.Amount
			}
		}
	}
	return resources
}

func pickAssembler(toBuild Recipe, minAssembler Assembler) (Assembler, string) {
	var usea Assembler
	var name string
	for key, a := range data.Assemblers {
		// fmt.Println(key, a.CraftingSpeed)
		if len(toBuild.Ingredients) > int(a.IngredientCount) {
			continue
		}
		var canUse bool
		for _, c := range a.CraftingCategories {
			if c == toBuild.Category {
				canUse = true
			}
		}
		if !canUse {
			continue
		}
		if minAssembler.CraftingSpeed > a.CraftingSpeed {
			continue
		}

		if usea.IngredientCount == 0 {
			usea = a
			name = key
		}

		if usea.IngredientCount > a.IngredientCount {
			usea = a
			name = key
		}
	}
	return usea, name
}
