package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-staging-ai/api/internal/testutil"
)

func TestStripeWebhookHandler(t *testing.T) {
	// Setup
	e := echo.New()
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name: "valid customer.subscription.created event",
			body: `{
				"id": "evt_test_webhook",
				"object": "event",
				"type": "customer.subscription.created",
				"data": {
					"object": {
						"id": "sub_test",
						"customer": "cus_test",
						"status": "active"
					}
				}
			}`,
			expectedCode: http.StatusOK,
		},
		{
			name: "valid invoice.payment_succeeded event",
			body: `{
				"id": "evt_test_webhook",
				"object": "event",
				"type": "invoice.payment_succeeded",
				"data": {
					"object": {
						"id": "in_test",
						"customer": "cus_test",
						"amount_paid": 2000
					}
				}
			}`,
			expectedCode: http.StatusOK,
		},
		{
			name: "unsupported event type",
			body: `{
				"id": "evt_test_webhook",
				"object": "event",
				"type": "unsupported.event.type",
				"data": {
					"object": {}
				}
			}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid JSON",
			body:         `{"invalid": json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty body",
			body:         "",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhook", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := server.stripeWebhookHandler(c)

			if tc.expectedCode == http.StatusOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestHandleSubscriptionCreated(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":       "sub_test",
		"customer": "cus_test",
		"status":   "active",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "customer.subscription.created",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleSubscriptionCreated(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleSubscriptionUpdated(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":       "sub_test",
		"customer": "cus_test",
		"status":   "past_due",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "customer.subscription.updated",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleSubscriptionUpdated(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleSubscriptionDeleted(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":       "sub_test",
		"customer": "cus_test",
		"status":   "canceled",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "customer.subscription.deleted",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleSubscriptionDeleted(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleInvoicePaymentSucceeded(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":           "in_test",
		"customer":     "cus_test",
		"amount_paid":  float64(2000),
		"subscription": "sub_test",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "invoice.payment_succeeded",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleInvoicePaymentSucceeded(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleInvoicePaymentFailed(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":           "in_test",
		"customer":     "cus_test",
		"amount_due":   float64(2000),
		"subscription": "sub_test",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "invoice.payment_failed",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleInvoicePaymentFailed(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleCustomerCreated(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":    "cus_test",
		"email": "test@example.com",
		"name":  "Test Customer",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "customer.created",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleCustomerCreated(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleCustomerUpdated(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":    "cus_test",
		"email": "updated@example.com",
		"name":  "Updated Customer",
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "customer.updated",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleCustomerUpdated(context.Background(), event)
	assert.NoError(t, err)
}

func TestHandleCustomerDeleted(t *testing.T) {
	// Setup
	mockS3Service := testutil.CreateMockS3Service(t)
	mockImageService := testutil.CreateMockImageService(t)
	server := &Server{s3Service: mockS3Service, imageService: mockImageService}

	eventData := map[string]interface{}{
		"id":      "cus_test",
		"deleted": true,
	}

	// This should not panic (currently just logs)
	event := &StripeEvent{
		ID:   "evt_test",
		Type: "customer.deleted",
		Data: map[string]interface{}{
			"object": eventData,
		},
	}
	err := server.handleCustomerDeleted(context.Background(), event)
	assert.NoError(t, err)
}
