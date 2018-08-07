package main

import (
	"fmt"
	"github.com/namsral/flag"
	"github.com/pmdcosta/treasure-coin/ost"
	"strconv"
	"sync"
	"time"
)

func main() {
	var (
		ostUrl     = flag.String("ost-url", "", "Choose the OST API base url.")
		ostKey     = flag.String("ost-key", "", "Choose the OST API key.")
		ostSecret  = flag.String("ost-secret", "", "Choose the OST API secret.")
		ostCompany = flag.String("ost-company", "", "Choose the OST API company ID.")
		games      = flag.Int("games", 100, "Choose the number of games to be created")
		treasures  = flag.Int("treasures", 10, "Choose the number of treasures per game")
		players    = flag.Int("players", 10, "Choose the number of players")
	)
	flag.Parse()

	// get ost config.
	config := ost.Config{}
	config.LoadCred(".env", *ostUrl, *ostKey, *ostSecret, *ostCompany)

	// instantiate the ost client service.
	st := ost.NewClient(config)
	// create gm
	gm, err := st.CreateUser("GameMaster")
	if err != nil {
		fmt.Println(err)
		return
	}

	// deposit sufficient funds to gm
	a := float64(*games**treasures) * 0.1
	err = st.Airdrop(gm, a)
	if err != nil {
		fmt.Println(err)
		return
	}

	// credit isn't immediate, hacky way to wait for it to be processed
	for {
		s, err := st.GetUserBalance(gm)
		if err != nil {
			fmt.Println(err)
			return
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		if v >= a {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// creates all games with 'treasure' amounts of tokens
	for i := 0; i < *games; i++ {
		err = st.MakePayment(gm, *treasures)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// create all players
	var p []string
	for i := 0; i < *players; i++ {
		player, err := st.CreateUser(fmt.Sprintf("TreasureHunter%d", i))
		if err != nil {
			fmt.Println(err)
			return
		}
		p = append(p, player)
	}

	// find all treasures
	var wg sync.WaitGroup
	wg.Add(*games * *treasures)

	for i := 0; i < *games**treasures; i++ {
		go Reward(st, p[i%len(p)], &wg)
		if i%5 == 0 {
			time.Sleep(250 * time.Millisecond)
		}
	}
	wg.Wait()

	time.Sleep(10 * time.Second)
	// remove tokens from all players
	wg.Add(*players)
	for i := 0; i < *players; i++ {
		go RemoveTokens(st, p[i], &wg)
		if i%5 == 0 {
			time.Sleep(250 * time.Millisecond)
		}
	}
	wg.Wait()
}

func Reward(st *ost.Client, s string, wg *sync.WaitGroup) {
	defer wg.Done()

	err := st.GetRewarded(s)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func RemoveTokens(st *ost.Client, s string, wg *sync.WaitGroup) {
	defer wg.Done()

	err := st.RemoveTokens(s)
	if err != nil {
		fmt.Println(err)
		return
	}
}
