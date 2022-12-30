package chains

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type Request struct {
	height   int
	page     int
	per_page int
}

type BinanceResponse struct {
	Total      int `json:"total"`
	Validators []struct {
		Validator             string  `json:"validator"`
		ValName               string  `json:"valName"`
		Proposer_priority     string  `json:"proposer_priority"`
		VotingPower           float64 `json:"votingPower"`
		VotingPowerProportion float64 `json:"votingPowerProportion"`
	} `json:"validators"`
}

type BinanceErrorResponse struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Error   string `json:"error"`
}

// https://docs.binance.org/api-reference/dex-api/staking.html#v1stakingchainschain-idvalidators
func Binance() (int, error) {
	votingPowers := make([]float64, 0, 200)
	pageLimit, pageOffset := 50, 0
	url := ""
	for true {
		url = fmt.Sprintf("https://api.binance.org/v1/staking/chains/bsc/validators?limit=%d&offset=%d", pageLimit, pageOffset)
		resp, err := http.Get(url)
		if err != nil {
			errBody, _ := ioutil.ReadAll(resp.Body)
			var errResp BinanceErrorResponse
			json.Unmarshal(errBody, &errResp)
			log.Println(errResp.Error)
			return -1, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, err
		}

		var response BinanceResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return -1, err
		}

		// break if no more entries left
		if len(response.Validators) == 0 {
			break
		}

		// loop through the validators voting power proportions
		for _, ele := range response.Validators {
			votingPowers = append(votingPowers, ele.VotingPowerProportion)
		}

		// increment counters
		pageOffset += pageLimit
	}

	// NOTE: NO need to calculate total voting power as the API response already
	// sends us proportional shares of each validator

	// now we're ready to calculate the nakomoto coefficient
	nakamotoCoefficient := calcNakamotoCoefficient(votingPowers)
	fmt.Println("The Nakamoto coefficient for binance Chain is", nakamotoCoefficient)

	return nakamotoCoefficient, nil
}

func calcNakamotoCoefficient(votingPowers []float64) int {
	var cumulativePercent, thresholdPercent float64 = 0.00, utils.THRESHOLD_PERCENT
	nakamotoCoefficient := 0
	for _, vpp := range votingPowers {
		// directly multiply voting power proportion with 100 to get
		// the actual voting percentage
		cumulativePercent += vpp * 100
		nakamotoCoefficient += 1
		if cumulativePercent >= thresholdPercent {
			break
		}
	}
	return nakamotoCoefficient
}
