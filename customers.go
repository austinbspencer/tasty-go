package tasty

import (
	"fmt"
	"net/http"
)

// Get authenticated customer.
func (c *Client) GetMyCustomerInfo() (Customer, *Error) {
	if c.Session.SessionToken == nil {
		return Customer{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}
	path := "/customers/me"

	type customerResponse struct {
		Customer Customer `json:"data"`
	}

	customersRes := new(customerResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	err := c.request(http.MethodGet, path, header, nil, nil, customersRes)
	if err != nil {
		return Customer{}, err
	}

	return customersRes.Customer, nil
}

// Get a full customer resource.
func (c *Client) GetCustomer(customerID string) (Customer, *Error) {
	if c.Session.SessionToken == nil {
		return Customer{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}
	path := fmt.Sprintf("/customers/%s", customerID)

	type customerResponse struct {
		Customer Customer `json:"data"`
	}

	customersRes := new(customerResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	err := c.request(http.MethodGet, path, header, nil, nil, customersRes)
	if err != nil {
		return Customer{}, err
	}

	return customersRes.Customer, nil
}

// Get a list of all the customer account resources attached to the current customer.
func (c *Client) GetCustomerAccounts(customerID string) ([]Account, *Error) {
	if c.Session.SessionToken == nil {
		return []Account{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}
	path := fmt.Sprintf("/customers/%s/accounts", customerID)

	type customerResponse struct {
		Data struct {
			Items []struct {
				Account Account `json:"account"`
			} `json:"items"`
		} `json:"data"`
	}

	customersRes := new(customerResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	err := c.request(http.MethodGet, path, header, nil, nil, customersRes)
	if err != nil {
		return []Account{}, err
	}

	var accounts []Account

	for _, acct := range customersRes.Data.Items {
		accounts = append(accounts, acct.Account)
	}

	return accounts, nil
}

// Get a full customer account resource.
func (c *Client) GetCustomerAccount(customerID, accountNumber string) (Account, *Error) {
	if c.Session.SessionToken == nil {
		return Account{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}
	path := fmt.Sprintf("/customers/%s/accounts/%s", customerID, accountNumber)

	type customerResponse struct {
		Account Account `json:"data"`
	}

	customersRes := new(customerResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	err := c.request(http.MethodGet, path, header, nil, nil, customersRes)
	if err != nil {
		return Account{}, err
	}

	return customersRes.Account, nil
}

// Get authenticated user's full account resource.
func (c *Client) GetMyAccount(accountNumber string) (Account, *Error) {
	if c.Session.SessionToken == nil {
		return Account{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}
	path := fmt.Sprintf("/customers/me/accounts/%s", accountNumber)

	type customerResponse struct {
		Account Account `json:"data"`
	}

	customersRes := new(customerResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	err := c.request(http.MethodGet, path, header, nil, nil, customersRes)
	if err != nil {
		return Account{}, err
	}

	return customersRes.Account, nil
}

// Returns the appropriate quote streamer endpoint, level and identification token.
// for the current customer to receive market data.
func (c *Client) GetQuoteStreamerTokens() (QuoteStreamerTokenAuthResult, *Error) {
	if c.Session.SessionToken == nil {
		return QuoteStreamerTokenAuthResult{}, &Error{Message: "Session is invalid: Session Token cannot be nil."}
	}
	path := "/quote-streamer-tokens"

	type customerResponse struct {
		Streamer QuoteStreamerTokenAuthResult `json:"data"`
	}

	customersRes := new(customerResponse)

	header := http.Header{}
	header.Add("Authorization", *c.Session.SessionToken)

	err := c.request(http.MethodGet, path, header, nil, nil, customersRes)
	if err != nil {
		return QuoteStreamerTokenAuthResult{}, err
	}

	return customersRes.Streamer, nil
}
