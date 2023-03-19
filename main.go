package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type Pokemon struct {
	Generation int    `json:"generation"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Types      []Type `json:"types"`
	Sprites    struct {
		FrontDefault string `json:"front_default"`
	} `json:"sprites"`
}

type Type struct {
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

func namepokedex(pokemonName string) (Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)

	resp, err := http.Get(url)
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()

	var pokemon Pokemon
	err = json.NewDecoder(resp.Body).Decode(&pokemon)
	if err != nil {
		return Pokemon{}, err
	}

	return pokemon, nil
}

func idpokedex(pokemonID int) (Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d", pokemonID)

	resp, err := http.Get(url)
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()

	var pokemon Pokemon
	err = json.NewDecoder(resp.Body).Decode(&pokemon)
	if err != nil {
		return Pokemon{}, err
	}

	return pokemon, nil
}

func servePokedexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/pokedex.html"))

	var pokemon Pokemon
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		pokedexID := r.FormValue("pokedexID")
		id, err := strconv.Atoi(pokedexID)
		if err != nil {
			pokemon, err = namepokedex(pokedexID)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		} else {
			pokemon, err = idpokedex(id)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		err = tmpl.Execute(w, pokemon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tmpl.Execute(w, nil)
	}
}

func serveHomePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func generationPokedex(generation int) ([]Pokemon, error) {
	var pokemonList []Pokemon

	var startID, endID int
	switch generation {
	case 1:
		startID = 1
		endID = 151
	case 2:
		startID = 152
		endID = 251
	case 3:
		startID = 252
		endID = 386
	case 4:
		startID = 387
		endID = 493
	case 5:
		startID = 494
		endID = 649
	case 6:
		startID = 650
		endID = 721
	case 7:
		startID = 722
		endID = 809
	case 8:
		startID = 810
		endID = 1008
	default:
		return nil, fmt.Errorf("invalid generation number")
	}

	for i := startID; i <= endID; i++ {
		pokemon, err := idpokedex(i)
		if err != nil {
			return nil, err
		}
		pokemonList = append(pokemonList, pokemon)
	}

	return pokemonList, nil
}

func serveGenerationPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/generation.html"))

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		generationStr := r.FormValue("generation")
		generation, err := strconv.Atoi(generationStr)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		pokemonList, err := generationPokedex(generation)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, pokemonList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tmpl.Execute(w, nil)
	}
}

func serveStartersPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/starters.html"))

	if r.Method == http.MethodGet {
		pokemonList, err := startersPokedex()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, pokemonList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func startersPokedex() ([]Pokemon, error) {
	var pokemonList []Pokemon

	// List of generations and their corresponding starter IDs
	generations := map[int]int{
		1:  1,
		2:  4,
		3:  7,
		4:  152,
		5:  155,
		6:  158,
		7:  252,
		8:  255,
		9:  258,
		10: 387,
		11: 390,
		12: 393,
		13: 495,
		14: 498,
		15: 501,
		16: 650,
		17: 653,
		18: 656,
		19: 722,
		20: 725,
		21: 728,
		22: 810,
		23: 813,
		24: 816,
		25: 909,
		26: 912,
		27: 906,
	}

	for generation, startID := range generations {
		pokemon, err := idpokedex(startID)
		if err != nil {
			return nil, err
		}
		pokemon.Generation = generation
		pokemonList = append(pokemonList, pokemon)
	}

	return pokemonList, nil
}

func main() {

	static := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", static))

	http.HandleFunc("/", serveHomePage)
	http.HandleFunc("/pokedex", servePokedexPage)
	http.HandleFunc("/generation", serveGenerationPage)
	http.HandleFunc("/starters", serveStartersPage)

	http.ListenAndServe(":8080", nil)
}
