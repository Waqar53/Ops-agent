package billing
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)
var (
	ErrInvalidPlan        = errors.New("invalid subscription plan")
	ErrNoPaymentMethod    = errors.New("no payment method attached")
	ErrSubscriptionExists = errors.New("subscription already exists")
)
type SubscriptionPlan struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	DisplayName          string                 `json:"display_name"`
	PriceMonthly         float64                `json:"price_monthly"`
	PriceYearly          float64                `json:"price_yearly"`
	StripePriceIDMonthly string                 `json:"stripe_price_id_monthly,omitempty"`
	StripePriceIDYearly  string                 `json:"stripe_price_id_yearly,omitempty"`
	Features             map[string]interface{} `json:"features"`
	Limits               map[string]interface{} `json:"limits"`
}
type UsageRecord struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	ResourceType   string                 `json:"resource_type"`
	Amount         float64                `json:"amount"`
	Unit           string                 `json:"unit"`
	Cost           float64                `json:"cost"`
	Metadata       map[string]interface{} `json:"metadata"`
	RecordedAt     time.Time              `json:"recorded_at"`
}
type Invoice struct {
	ID              string     `json:"id"`
	OrganizationID  string     `json:"organization_id"`
	StripeInvoiceID string     `json:"stripe_invoice_id,omitempty"`
	AmountDue       float64    `json:"amount_due"`
	AmountPaid      float64    `json:"amount_paid"`
	Currency        string     `json:"currency"`
	Status          string     `json:"status"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	InvoicePDFURL   string     `json:"invoice_pdf_url,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
type BillingService struct {
	db            *sql.DB
	webhookSecret string
}
func NewBillingService(db *sql.DB, stripeKey, webhookSecret string) *BillingService {
	stripe.Key = stripeKey
	return &BillingService{
		db:            db,
		webhookSecret: webhookSecret,
	}
}
func (bs *BillingService) GetPlans() ([]SubscriptionPlan, error) {
	rows, err := bs.db.Query(`
		SELECT id, name, display_name, price_monthly, price_yearly, 
		       stripe_price_id_monthly, stripe_price_id_yearly, features, limits
		FROM subscription_plans
		ORDER BY price_monthly ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []SubscriptionPlan
	for rows.Next() {
		var plan SubscriptionPlan
		var featuresJSON, limitsJSON []byte
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.DisplayName,
			&plan.PriceMonthly, &plan.PriceYearly,
			&plan.StripePriceIDMonthly, &plan.StripePriceIDYearly,
			&featuresJSON, &limitsJSON,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(featuresJSON, &plan.Features)
		json.Unmarshal(limitsJSON, &plan.Limits)
		plans = append(plans, plan)
	}
	return plans, nil
}
func (bs *BillingService) CreateCustomer(orgID, email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
		Metadata: map[string]string{
			"org_id": orgID,
		},
	}
	c, err := customer.New(params)
	if err != nil {
		return "", err
	}
	_, err = bs.db.Exec(`
		UPDATE organizations
		SET stripe_customer_id = $1
		WHERE id = $2
	`, c.ID, orgID)
	return c.ID, err
}
func (bs *BillingService) CreateSubscription(orgID, planName string, yearly bool) error {
	var stripeCustomerID string
	err := bs.db.QueryRow(`
		SELECT stripe_customer_id FROM organizations WHERE id = $1
	`, orgID).Scan(&stripeCustomerID)
	if err != nil {
		return err
	}
	if stripeCustomerID == "" {
		return errors.New("no stripe customer ID")
	}
	var priceID string
	if yearly {
		err = bs.db.QueryRow(`
			SELECT stripe_price_id_yearly FROM subscription_plans WHERE name = $1
		`, planName).Scan(&priceID)
	} else {
		err = bs.db.QueryRow(`
			SELECT stripe_price_id_monthly FROM subscription_plans WHERE name = $1
		`, planName).Scan(&priceID)
	}
	if err != nil {
		return ErrInvalidPlan
	}
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(priceID)},
		},
		Metadata: map[string]string{
			"org_id": orgID,
		},
	}
	sub, err := subscription.New(params)
	if err != nil {
		return err
	}
	_, err = bs.db.Exec(`
		UPDATE organizations
		SET plan = $1, stripe_subscription_id = $2
		WHERE id = $3
	`, planName, sub.ID, orgID)
	return err
}
func (bs *BillingService) TrackUsage(orgID, resourceType string, amount float64, unit string, metadata map[string]interface{}) error {
	cost := bs.calculateCost(resourceType, amount, unit)
	metadataJSON, _ := json.Marshal(metadata)
	_, err := bs.db.Exec(`
		INSERT INTO usage_records (organization_id, resource_type, amount, unit, cost, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, orgID, resourceType, amount, unit, cost, metadataJSON)
	return err
}
func (bs *BillingService) calculateCost(resourceType string, amount float64, unit string) float64 {
	rates := map[string]float64{
		"compute_hours": 0.10,
		"storage_gb":    0.02,
		"bandwidth_gb":  0.05,
		"builds":        0.01,
	}
	key := resourceType + "_" + unit
	if rate, ok := rates[key]; ok {
		return amount * rate
	}
	return 0
}
func (bs *BillingService) GetUsage(orgID string, start, end time.Time) ([]UsageRecord, error) {
	rows, err := bs.db.Query(`
		SELECT id, organization_id, resource_type, amount, unit, cost, metadata, recorded_at
		FROM usage_records
		WHERE organization_id = $1 AND recorded_at BETWEEN $2 AND $3
		ORDER BY recorded_at DESC
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []UsageRecord
	for rows.Next() {
		var record UsageRecord
		var metadataJSON []byte
		err := rows.Scan(
			&record.ID, &record.OrganizationID, &record.ResourceType,
			&record.Amount, &record.Unit, &record.Cost,
			&metadataJSON, &record.RecordedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(metadataJSON, &record.Metadata)
		records = append(records, record)
	}
	return records, nil
}
func (bs *BillingService) HandleWebhook(payload []byte, signature string) error {
	event, err := webhook.ConstructEvent(payload, signature, bs.webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}
	bs.logBillingEvent(event)
	switch event.Type {
	case "customer.subscription.created":
		return bs.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		return bs.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return bs.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		return bs.handlePaymentSucceeded(event)
	case "invoice.payment_failed":
		return bs.handlePaymentFailed(event)
	}
	return nil
}
func (bs *BillingService) logBillingEvent(event stripe.Event) {
	dataJSON, _ := json.Marshal(event.Data)
	bs.db.Exec(`
		INSERT INTO billing_events (event_type, stripe_event_id, data)
		VALUES ($1, $2, $3)
	`, event.Type, event.ID, dataJSON)
}
func (bs *BillingService) handleSubscriptionCreated(event stripe.Event) error {
	return nil
}
func (bs *BillingService) handleSubscriptionUpdated(event stripe.Event) error {
	return nil
}
func (bs *BillingService) handleSubscriptionDeleted(event stripe.Event) error {
	return nil
}
func (bs *BillingService) handlePaymentSucceeded(event stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return err
	}
	_, err := bs.db.Exec(`
		INSERT INTO invoices (stripe_invoice_id, amount_due, amount_paid, currency, status, paid_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, inv.ID, float64(inv.AmountDue)/100, float64(inv.AmountPaid)/100, inv.Currency, "paid")
	return err
}
func (bs *BillingService) handlePaymentFailed(event stripe.Event) error {
	return nil
}
