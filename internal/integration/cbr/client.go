package cbr

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/beevik/etree"
)

// Client represents a CBR SOAP API client
type Client struct {
	config     *config.CBRConfig
	httpClient *http.Client
}

// NewClient creates a new CBR client
func NewClient(config *config.CBRConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetKeyRate retrieves the current key rate from CBR
func (c *Client) GetKeyRate() (float64, error) {
	// Build SOAP request
	soapRequest := c.buildKeyRateRequest()

	// Send request
	resp, err := c.sendRequest(soapRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}

	// Parse response
	rate, err := c.parseKeyRateResponse(resp)
	if err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return rate, nil
}

// buildKeyRateRequest creates a SOAP request for key rate
func (c *Client) buildKeyRateRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
		<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
			<soap12:Body>
				<KeyRate xmlns="http://web.cbr.ru/">
					<fromDate>%s</fromDate>
					<ToDate>%s</ToDate>
				</KeyRate>
			</soap12:Body>
		</soap12:Envelope>`, fromDate, toDate)
}

// sendRequest sends a SOAP request to CBR
func (c *Client) sendRequest(soapRequest string) ([]byte, error) {
	req, err := http.NewRequest(
		"POST",
		c.config.BaseURL+c.config.RateEndpoint,
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	// Send request with retries
	var resp *http.Response
	var lastErr error

	for i := 0; i <= c.config.RetryCount; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(c.config.RetryDelay)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", c.config.RetryCount, lastErr)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// parseKeyRateResponse parses the SOAP response to extract the key rate
func (c *Client) parseKeyRateResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find rate elements
	elements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(elements) == 0 {
		return 0, fmt.Errorf("no rate data found in response")
	}

	// Get the latest rate
	latestElement := elements[0]
	rateElement := latestElement.FindElement("./Rate")
	if rateElement == nil {
		return 0, fmt.Errorf("rate element not found")
	}

	// Parse rate value
	var rate float64
	if _, err := fmt.Sscanf(rateElement.Text(), "%f", &rate); err != nil {
		return 0, fmt.Errorf("failed to parse rate value: %w", err)
	}

	return rate, nil
}

// KeyRateResponse represents the CBR key rate response
type KeyRateResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		KeyRateResponse struct {
			KeyRateResult struct {
				Schema struct {
					Element struct {
						ComplexType struct {
							Choice struct {
								Element struct {
									ComplexType struct {
										Sequence struct {
											Element []struct {
												Name     string `xml:"name,attr"`
												Type     string `xml:"type,attr"`
												MaxValue string `xml:"maxValue,attr,omitempty"`
											} `xml:"element"`
										} `xml:"sequence"`
									} `xml:"complexType"`
								} `xml:"element"`
							} `xml:"choice"`
						} `xml:"complexType"`
					} `xml:"element"`
				} `xml:"schema"`
				DiffGram struct {
					KeyRate struct {
						KR []struct {
							Rate    float64   `xml:"Rate"`
							Date    time.Time `xml:"Date"`
							DateEnd time.Time `xml:"DateEnd"`
						} `xml:"KR"`
					} `xml:"KeyRate"`
				} `xml:"diffgram"`
			} `xml:"KeyRateResult"`
		} `xml:"KeyRateResponse"`
	} `xml:"Body"`
}
