package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	//"time"
)


// --------------------------------------- TEST THE ENGINE                       ----------------------------------------------------------------

func main() {
	//model := train("big.txt") //pour un chargement du dictionnaire de mots connus depuis un fichier text
	model := train2() //pour un chargement du dictionnaire de mots connus en les entrants à la mains (prototype)

	//startTime := time.Now().Unix() //sur le playground on ne peut pas accéder à l'horloge pr mesurer le temps de calcul

	tests:=[]string{"restaurnt","restornt","chinoi","japoneis","algerien"}
	for _,test:=range tests {
		corrige,ok:=correct(test, model)
		fmt.Println(test," -> ",corrige,ok)
		}


	//fmt.Printf("Time : %v\n", float64(time.Now().Unix() - startTime) / float64(1e9)) //afficherait zero sur le playground
}




// --------------------------------------- FUNCTIONS DOING THE JOB ----------------------------------------------------------------
//this function trains the model with a huge english text file to discover words and compute stats for better accuracy.
func train(training_data string) map[string]int {
	NWORDS := make(map[string]int)
	pattern := regexp.MustCompile("[a-z]+")
	if content, err := ioutil.ReadFile(training_data); err == nil {
 		for _, w := range pattern.FindAllString(strings.ToLower(string(content)), -1) {
			NWORDS[w]++;
		}
	} else {
		panic("Failed loading training data.  Get it from http://norvig.com/big.txt.")
	}
	return NWORDS
}

//static list of known words
func train2() map[string]int {
	NWORDS := make(map[string]int)
	NWORDS["restaurant"]=10
	NWORDS["chinois"]=10
	NWORDS["grillade"]=10
	NWORDS["jap"]=10
	NWORDS["japonais"]=10
	NWORDS["algérien"]=10
	NWORDS["algérienne"]=10
	NWORDS["marocaine"]=10
	NWORDS["tunisienne"]=10
	NWORDS["libanaise"]=10
	NWORDS["resto"]=10
	return NWORDS
}

//calcul toutes les variantes avec une seule variation (un caractère remplacé, supprimé, inversé ou ajouté)
func edits1(word string, ch chan string) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzéèïêà" //on pourrait rajouter les accentuation si nécessaire "éèïêà"
	type Pair struct{a, b string}
	var splits []Pair
	for i := 0; i < len(word) + 1; i++ {
		splits = append(splits, Pair{word[:i], word[i:]}) }

	for _, s := range splits {
		if len(s.b) > 0 { ch <- s.a + s.b[1:] }
		if len(s.b) > 1 { ch <- s.a + string(s.b[1]) + string(s.b[0]) + s.b[2:] }
		for _, c := range alphabet { if len(s.b) > 0 { ch <- s.a + string(c) + s.b[1:] }}
		for _, c := range alphabet { ch <- s.a + string(c) + s.b }
	}
}

//calcul toutes les variantes avec deux variations
func edits2(word string, ch chan string) {
	ch1 := make(chan string, 1024*1024)
	go func() { edits1(word, ch1); ch1 <- "" }()
	for e1 := range ch1 {
		if e1 == "" { break }
		edits1(e1, ch)
	}
}

//calcul toutes les variantes avec trois variations. au-dessus on considère que le mot ne fait pas partie du dictionnaire
func edits3(word string, ch chan string) {
	ch1 := make(chan string, 1024*1024)
	go func() { edits2(word, ch1); ch1 <- "" }()
	for e1 := range ch1 {
		if e1 == "" { break }
		edits1(e1, ch)
	}
}
func edits4(word string, ch chan string) {
	ch1 := make(chan string, 1024*1024)
	go func() { edits3(word, ch1); ch1 <- "" }()
	for e1 := range ch1 {
		if e1 == "" { break }
		edits1(e1, ch)
	}
}
func edits5(word string, ch chan string) {
	ch1 := make(chan string, 1024*1024)
	go func() { edits4(word, ch1); ch1 <- "" }()
	for e1 := range ch1 {
		if e1 == "" { break }
		edits1(e1, ch)
	}
}

func best(word string, edits func(string, chan string), model map[string]int) string {
	ch := make(chan string, 1024*1024)
	go func() { edits(word, ch); ch <- "" }()
	maxFreq := 0
	correction := ""
	for word := range ch {
		if word == "" { break }
		if freq, present := model[word]; present && freq > maxFreq {
			maxFreq, correction = freq, word
		}
	}
	return correction
}

func correct(word string, model map[string]int) (string,string) {
	if _, present := model[word]; present { return word,"-> mot correct" }
	if correction := best(word, edits1, model); correction != "" { return correction, "-> corrigé" }
	//if correction := best(word, edits2, model); correction != "" { return correction, "-> corrigé" }
	//if correction := best(word, edits3, model); correction != "" { return correction, "-> corrigé" }
	return word, "-> pas de correction trouvé"
}
