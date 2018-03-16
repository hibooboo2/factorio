package main

type FactorioData struct {
	Assemblers map[string]Assembler `json:"assemblers"`
	Items      map[string]Item      `json:"items"`
	Miners     map[string]Miner     `json:"miners"`
	Recipes    map[string]Recipe    `json:"recipes"`
	Resources  map[string]Resource  `json:"resources"`
}

type IngriedentProduct struct {
	Amount      int64   `json:"amount"`
	AmountMax   int64   `json:"amount_max"`
	AmountMin   int64   `json:"amount_min"`
	Name        string  `json:"name"`
	Probability float64 `json:"probability"`
	Type        string  `json:"type"`
}

type Recipe struct {
	Category    string              `json:"category"`
	TimeToCraft float64             `json:"energy"`
	Ingredients []IngriedentProduct `json:"ingredients"`
	Name        string              `json:"name"`
	Products    []IngriedentProduct `json:"products"`
}

type Assembler struct {
	CraftingCategories []string `json:"crafting_categories"`
	CraftingSpeed      float64  `json:"crafting_speed"`
	IngredientCount    int64    `json:"ingredient_count"`
}

type Resource struct {
	FluidAmount       int64               `json:"fluid_amount"`
	Hardness          float64             `json:"hardness"`
	MiningTime        int64               `json:"mining_time"`
	Products          []IngriedentProduct `json:"products"`
	RequiredFluid     string              `json:"required_fluid"`
	ResourceCategorie string              `json:"resource_categorie"`
}

type Miner struct {
	MiningDrillRadius  float64  `json:"mining_drill_radius"`
	MiningPower        float64  `json:"mining_power"`
	MiningSpeed        float64  `json:"mining_speed"`
	ResourceCategories []string `json:"resource_categories"`
}

type Item struct {
	Name      string `json:"name"`
	StackSize int64  `json:"stack_size"`
	Type      string `json:"type"`
}
