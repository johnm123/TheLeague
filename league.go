package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

type fixture struct {
	PlayerOne string
	PlayerTwo string
}

type result struct {
	PlayerOne string
	PlayerTwo string
	Winner    string
}

type ranking struct {
	PlayerName string
	Points     int
}

func main() {

	results := getResultsFromFile()

	// This ensures we save the results on exit.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		saveResultsToFile(results)
		os.Exit(0)
	}()

	fmt.Println("Running Server")

	http.HandleFunc("/rankings", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(200)
		if req.Method == "GET" {
			rankings := generateRankings(results)
			rankingsJSON, _ := json.Marshal(rankings)
			fmt.Fprintf(w, string(rankingsJSON[:]))
		}
	})

	http.HandleFunc("/results", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if req.Method == "GET" {
			w.WriteHeader(200)
			resultsJSON, _ := json.Marshal(results)
			fmt.Fprintf(w, string(resultsJSON[:]))
		}
		if req.Method == "POST" {
			w.WriteHeader(201)
			bodyBytes, _ := ioutil.ReadAll(req.Body)
			var r result
			json.Unmarshal(bodyBytes, &r)
			results = append(results, r)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./static")
	})

	http.ListenAndServe(":8080", nil)
}

func getResultsFromFile() []result {
	file, _ := os.Open("results.txt")
	var results []result
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		splitString := strings.Split(scanner.Text(), ",")
		newResult := result{PlayerOne: splitString[0], PlayerTwo: splitString[1], Winner: splitString[2]}
		results = append(results, newResult)
	}

	return results
}

func saveResultsToFile(results []result) {
	f, _ := os.Create("results.txt")
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, r := range results {
		w.WriteString(r.PlayerOne + "," + r.PlayerTwo + "," + r.Winner + "\n")
	}
	w.Flush()
}

func generateRankings(results []result) []ranking {

	var playerNames []string
	for _, r := range results {
		playerOneFound := false
		playerTwoFound := false
		for _, p := range playerNames {

			if r.PlayerOne == p {
				playerOneFound = true
			}

			if r.PlayerTwo == p {
				playerTwoFound = true
			}

			if r.PlayerOne == r.PlayerTwo {
				playerTwoFound = true
			}
		}

		if playerOneFound == false {
			playerNames = append(playerNames, r.PlayerOne)
		}
		if playerTwoFound == false {
			playerNames = append(playerNames, r.PlayerTwo)
		}
	}

	var rankings = []ranking{}

	// Initialise
	for _, p := range playerNames {
		r := ranking{PlayerName: p, Points: 0}
		rankings = append(rankings, r)
	}

	for _, r := range results {
		for i := 0; i < len(rankings); i++ {
			if rankings[i].PlayerName == r.Winner {
				rankings[i].Points += 3
			}
		}
	}

	// Sort rankings
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].Points > rankings[j].Points
	})

	return rankings
}

func generateRandomFixtures(people []string) []fixture {

	fixtureExists := func(fixture fixture, fixtures []fixture) bool {

		fixtureExists := false

		for _, f := range fixtures {
			if f.PlayerOne == fixture.PlayerOne && f.PlayerTwo == fixture.PlayerTwo {
				fixtureExists = true
				break
			}

			if f.PlayerOne == fixture.PlayerTwo && f.PlayerTwo == fixture.PlayerOne {
				fixtureExists = true
				break
			}
		}

		return fixtureExists
	}

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomPlayerTwos := make([]string, len(people))

	for i, v := range generator.Perm(len(people)) {
		randomPlayerTwos[v] = people[i]
	}

	var fixtures []fixture
	for _, k := range people {
		for _, l := range randomPlayerTwos {
			fixture := fixture{PlayerOne: k, PlayerTwo: l}

			if fixtureExists(fixture, fixtures) == false && k != l {
				fixtures = append(fixtures, fixture)
			}
		}
	}

	temp := make([]fixture, len(fixtures))
	for i, v := range generator.Perm(len(fixtures)) {
		temp[v] = fixtures[i]
	}

	return temp
}
