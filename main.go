package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

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
	allMoves := getData()
	fmt.Println("Endpoint Hit: returnAllMoves")

	fmt.Fprintf(w, "All moves :-\n")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	enc.Encode(allMoves)
}

func returnSingleMove(w http.ResponseWriter, r *http.Request) {
	allMoves := getData()
	fmt.Println("Endpoint Hit: returnSingleMoves")
	vars := mux.Vars(r)
	key := vars["code"]

	requiredMove, err := getMoveForKey(allMoves, key)
	if err != nil {
		fmt.Fprintf(w, "No move found for the given ID")
	} else {
		fmt.Fprintf(w, "Moves for following code: %s\n", key)
		fmt.Fprintf(w, "Title : %s\n", requiredMove.Title)
		fmt.Fprintf(w, "Moves : %s\n", requiredMove.Moves)
	}
}

func getMoveForKey(moves []Move, key string) (Move, error) {
	for _, move := range moves {
		if move.Code == key {
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

func getData() []Move {
	allMoves := make([]Move, 0)

	doc, err := goquery.NewDocument("https://www.chessgames.com/chessecohelp.html")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("table tr").Each(func(_ int, tr *goquery.Selection) {

		move := Move{}
		tr.Find("td").Each(func(ix int, td *goquery.Selection) {
			if ix == 0 {
				move.Code = td.Text()
			}
			if ix == 1 {
				body := td.Text()
				splitBody := strings.SplitN(body, "\n", 2)
				move.Title = splitBody[0]
				move.Moves = splitBody[1]
			}
		})
		allMoves = append(allMoves, move)
	})
	return allMoves
}
