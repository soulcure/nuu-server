package paypal

import (
	"fmt"
)

//GetPayment 发起支付
func (c *Client) GetPayment(paymentID string) (*PaymentResponse, error) {
	payment := &PaymentResponse{}
	url := fmt.Sprintf("%s%s%s", c.APIBase, "/v1/payments/payment/", paymentID)
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return payment, err
	}

	if err = c.SendWithAuth(req, payment); err != nil {
		return payment, err
	}

	return payment, nil
}
