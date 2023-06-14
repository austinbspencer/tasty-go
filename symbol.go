package tasty

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/austinbspencer/tasty-go/models"
)

// Get the position limit
func (c *Client) SymbolSearch(symbol string) ([]models.SymbolData, *Error) {
	if c.Session.SessionToken == nil {
		return []models.SymbolData{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}

	// url escape required for instances where "/" exists in symbol i.e. BRK/B
	symbol = url.QueryEscape(symbol)

	path := fmt.Sprintf("//%s/symbols/search/%s", c.baseHost, symbol)

	type symbolResponse struct {
		Data struct {
			SymbolData []models.SymbolData `json:"items"`
		} `json:"data"`
	}

	symbolRes := new(symbolResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	// customGet required for instances where "/" exists in symbol i.e. BRK/B
	err := c.customRequest(http.MethodGet, path, header, nil, nil, symbolRes)
	if err != nil {
		return []models.SymbolData{}, err
	}

	return symbolRes.Data.SymbolData, nil
}