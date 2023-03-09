package sonolus

type LevelData struct {
	Entities []LevelDataEntity `json:"entities"`
}

type LevelDataEntity struct {
	Archetype int `json:"archetype"`
	Data      struct {
		Values []float64 `json:"values"`
	} `json:"data"`
}

