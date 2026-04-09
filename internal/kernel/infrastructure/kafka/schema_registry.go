package kafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// SchemaRegistry - client for Confluent Schema Registry.
// Manages Protobuf schema versions for Kafka topics.
//
// Schema compatibility modes:
//   - BACKWARD (default): new schema can read old data
//   - FORWARD: old consumers can read new data
//   - FULL: both directions
type SchemaRegistry struct {
	baseURL    string
	httpClient *http.Client
}

// SchemaInfo - response from Schema Registry
type SchemaInfo struct {
	Subject    string `json:"subject"`
	Version    int    `json:"version"`
	ID         int    `json:"id"`
	SchemaType string `json:"schemaType"`
	Schema     string `json:"schema"`
}

// NewSchemaRegistry creates a Schema Registry client
func NewSchemaRegistry(baseURL string) *SchemaRegistry {
	return &SchemaRegistry{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterSchema - register a Protobuf schema for a subject (topic-value).
// Returns the schema ID assigned by the registry.
func (sr *SchemaRegistry) RegisterSchema(subject, protoSchema string) (int, error) {
	payload := map[string]string{
		"schemaType": "PROTOBUF",
		"schema":     protoSchema,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal schema payload: %w", err)
	}

	url := fmt.Sprintf("%s/subjects/%s/versions", sr.baseURL, subject)
	resp, err := sr.httpClient.Post(url, "application/vnd.schemaregistry.v1+json", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("schema registry request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("schema registry returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("unmarshal response: %w", err)
	}

	logger.Log.Info("Schema registered",
		zap.String("subject", subject),
		zap.Int("id", result.ID),
	)
	return result.ID, nil
}

// GetLatestSchema - retrieve the latest schema version for a subject
func (sr *SchemaRegistry) GetLatestSchema(subject string) (*SchemaInfo, error) {
	url := fmt.Sprintf("%s/subjects/%s/versions/latest", sr.baseURL, subject)
	resp, err := sr.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("schema registry request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // no schema registered yet
	}

	var info SchemaInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &info, nil
}

// SetCompatibility - set compatibility mode for a subject
// Modes: BACKWARD, FORWARD, FULL, NONE
func (sr *SchemaRegistry) SetCompatibility(subject, mode string) error {
	payload := map[string]string{"compatibility": mode}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/config/%s", sr.baseURL, subject)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	resp, err := sr.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("set compatibility failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("schema registry returned %d: %s", resp.StatusCode, string(respBody))
	}

	logger.Log.Info("Schema compatibility set",
		zap.String("subject", subject),
		zap.String("mode", mode),
	)
	return nil
}

// CheckCompatibility - test if a new schema is compatible with the latest version
func (sr *SchemaRegistry) CheckCompatibility(subject, protoSchema string) (bool, error) {
	payload := map[string]string{
		"schemaType": "PROTOBUF",
		"schema":     protoSchema,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/compatibility/subjects/%s/versions/latest", sr.baseURL, subject)
	resp, err := sr.httpClient.Post(url, "application/vnd.schemaregistry.v1+json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("compatibility check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return true, nil // no existing schema = always compatible
	}

	var result struct {
		IsCompatible bool `json:"is_compatible"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}
	return result.IsCompatible, nil
}

// RegisterAllSchemas - registers all XBank Protobuf schemas.
// Call this at startup to ensure schemas are registered.
func (sr *SchemaRegistry) RegisterAllSchemas() {
	schemas := map[string]string{
		"xbank.accounts.opened-value":          accountOpenedProto,
		"xbank.accounts.credited-value":        accountCreditedProto,
		"xbank.accounts.debited-value":         accountDebitedProto,
		"xbank.accounts.frozen-value":          accountFrozenProto,
		"xbank.accounts.closed-value":          accountClosedProto,
		"xbank.transfers.created-value":        transferCreatedProto,
		"xbank.transfers.completed-value":      transferCompletedProto,
		"xbank.transfers.failed-value":         transferFailedProto,
		"xbank.cards.issued-value":             cardIssuedProto,
		"xbank.cards.blocked-value":            cardBlockedProto,
		"xbank.cards.activated-value":          cardActivatedProto,
		"xbank.kyc.submitted-value":            kycSubmittedProto,
		"xbank.kyc.approved-value":             kycApprovedProto,
		"xbank.kyc.rejected-value":             kycRejectedProto,
		"xbank.notifications.requested-value":  notificationRequestedProto,
	}

	for subject, schema := range schemas {
		if _, err := sr.RegisterSchema(subject, schema); err != nil {
			logger.Log.Warn("Failed to register schema (registry may be unavailable)",
				zap.String("subject", subject),
				zap.Error(err),
			)
		}
	}
}

// Embedded Protobuf schemas for registration.
// These match the .proto files in the proto/ directory.
const (
	metadataProto = `syntax = "proto3";
package xbank.common;
import "google/protobuf/timestamp.proto";
message Metadata {
  string event_id = 1;
  string correlation_id = 2;
  string user_id = 3;
  google.protobuf.Timestamp timestamp = 4;
  int32 retry_count = 5;
  string source = 6;
}`

	accountOpenedProto = `syntax = "proto3";
package xbank.accounts;
message AccountOpened {
  string account_id = 2;
  string user_id = 3;
  string account_number = 4;
  string currency = 5;
}`

	accountCreditedProto = `syntax = "proto3";
package xbank.accounts;
message AccountCredited {
  string account_id = 2;
  int64 amount = 3;
  string currency = 4;
  int64 balance = 5;
}`

	accountDebitedProto = `syntax = "proto3";
package xbank.accounts;
message AccountDebited {
  string account_id = 2;
  int64 amount = 3;
  string currency = 4;
  int64 balance = 5;
}`

	accountFrozenProto = `syntax = "proto3";
package xbank.accounts;
message AccountFrozen {
  string account_id = 2;
}`

	accountClosedProto = `syntax = "proto3";
package xbank.accounts;
message AccountClosed {
  string account_id = 2;
}`

	transferCreatedProto = `syntax = "proto3";
package xbank.transfers;
message TransferCreated {
  string transfer_id = 2;
  string from_account_id = 3;
  string to_account_id = 4;
  int64 amount = 5;
  string currency = 6;
  string description = 7;
}`

	transferCompletedProto = `syntax = "proto3";
package xbank.transfers;
message TransferCompleted {
  string transfer_id = 2;
  string from_account_id = 3;
  string to_account_id = 4;
  int64 amount = 5;
  string currency = 6;
}`

	transferFailedProto = `syntax = "proto3";
package xbank.transfers;
message TransferFailed {
  string transfer_id = 2;
  string from_account_id = 3;
  string to_account_id = 4;
  int64 amount = 5;
  string currency = 6;
  string reason = 7;
}`

	cardIssuedProto = `syntax = "proto3";
package xbank.cards;
message CardIssued {
  string card_id = 2;
  string account_id = 3;
  string user_id = 4;
  string card_type = 5;
  string masked_pan = 6;
}`

	cardBlockedProto = `syntax = "proto3";
package xbank.cards;
message CardBlocked {
  string card_id = 2;
  string user_id = 3;
  string reason = 4;
}`

	cardActivatedProto = `syntax = "proto3";
package xbank.cards;
message CardActivated {
  string card_id = 2;
  string user_id = 3;
}`

	kycSubmittedProto = `syntax = "proto3";
package xbank.kyc;
message KYCSubmitted {
  string verification_id = 2;
  string user_id = 3;
  string document_type = 4;
}`

	kycApprovedProto = `syntax = "proto3";
package xbank.kyc;
message KYCApproved {
  string verification_id = 2;
  string user_id = 3;
  string level = 4;
}`

	kycRejectedProto = `syntax = "proto3";
package xbank.kyc;
message KYCRejected {
  string verification_id = 2;
  string user_id = 3;
  string reason = 4;
}`

	notificationRequestedProto = `syntax = "proto3";
package xbank.notifications;
message NotificationRequested {
  string notification_id = 2;
  string user_id = 3;
  string type = 4;
  string title = 5;
  string body = 6;
  string channel = 7;
  map<string, string> data = 8;
}`
)
