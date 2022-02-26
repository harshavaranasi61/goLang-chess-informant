package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ReneKroon/ttlcache/v2"
	"github.com/gorilla/mux"
)

var cache ttlcache.SimpleCache = ttlcache.NewCache()

type Move struct {
	Code  string `json:"code"`
	Title string `json:"title"`
	Moves string `json:"moves"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func returnAllMoves(w http.ResponseWriter, r *http.Request) {
	allMoves, err := cache.Get("All moves")
	if err != nil {
		allMoves = getData()
	}
	fmt.Println("Endpoint Hit: returnAllMoves")

	fmt.Fprintf(w, "All moves :-\n")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	enc.Encode(allMoves.([]Move))
}

func returnSingleMove(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnSingleMoves")
	vars := mux.Vars(r)
	key := vars["code"]

	value, err := cache.Get(key)
	if err != nil && value == nil {
		requiredMove, err := getMoveForKey(key)
		if err != nil {
			fmt.Fprintf(w, "No move found for the given ID")
		} else {
			fmt.Fprintf(w, "From api call: Moves for following code: %s\n", key)
			fmt.Fprintf(w, "Title : %s\n", requiredMove.Title)
			fmt.Fprintf(w, "Moves : %s\n", requiredMove.Moves)
		}
	} else {
		fmt.Fprintf(w, "From cache: Moves for following code: %s\n", key)
		fmt.Fprintf(w, "Title : %s\n", value.(Move).Title)
		fmt.Fprintf(w, "Moves : %s\n", value.(Move).Moves)
	}
}

func getMoveForKey(key string) (Move, error) {

	allMoves, err := cache.Get("All moves")
	if err != nil {
		allMoves = getData()
	}

	for _, move := range allMoves.([]Move) {
		if move.Code == key {
			cache.SetWithTTL(key, move, time.Duration(3*time.Second))
			return move, nil
		}
	}
	return Move{}, errors.New("no move found for the given ID")
}

func handleRequests() {

	port := os.Getenv("PORT")

	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/home", homePage)
	myRouter.HandleFunc("/", returnAllMoves)
	myRouter.HandleFunc("/{code}", returnSingleMove)

	log.Fatal(http.ListenAndServe(":"+port, myRouter))
}

func main() {
	handleRequests()
}

type Coordinate struct {
	yCoordinate int
	zCoordinate int
}

type Config struct {
	url            string
	locationConfig map[string]Coordinate
}

func getData() []Move {
	allMoves := make([]Move, 0)

	config := Config{
		"https://www.chessgames.com/chessecohelp.html",
		map[string]Coordinate{
			"codeConfig":  {0, 0},
			"nameConfig":  {1, 0},
			"movesConfig": {1, 1},
		},
	}
	doc, err := goquery.NewDocument(config.url)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table tr").Each(func(_ int, tr *goquery.Selection) {

		move := Move{}
		var data [][]string
		tr.Find("td").Each(func(ix int, td *goquery.Selection) {
			data = append(data, strings.Split(td.Text(), "\n"))
		})

		move.Code = data[config.locationConfig["codeConfig"].yCoordinate][config.locationConfig["codeConfig"].zCoordinate]
		move.Title = data[config.locationConfig["nameConfig"].yCoordinate][config.locationConfig["nameConfig"].zCoordinate]
		move.Moves = data[config.locationConfig["movesConfig"].yCoordinate][config.locationConfig["movesConfig"].zCoordinate]
		allMoves = append(allMoves, move)
	})
	cache.SetWithTTL("All Moves", allMoves, time.Minute*30)
	return allMoves
}
