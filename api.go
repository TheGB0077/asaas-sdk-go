package asaas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Olimi-org/asaas-sdk-go/internal/request"
)

var ErrNotFound = errors.New("recurso não encontrado")

type environment string

const Production environment = "production"
const Sandbox environment = "sandbox"

// Constantes da API
const (
	// Chaves dos cabeçalhos (Headers)
	HeaderAccessToken = "access_token"

	// Caminhos da API
	PathVersion       = "/v3"
	PathCustomers     = PathVersion + "/customers"
	PathPayments      = PathVersion + "/payments"
	PathSubscriptions = PathVersion + "/subscriptions"

	// URLs dos ambientes
	BaseURLProduction = "https://api.asaas.com"
	BaseURLSandbox    = "https://sandbox.asaas.com/api"

	// Códigos de status HTTP
	StatusClientErrorMin = 400
)

// Config contém as opções de configuração para o cliente Asaas
type Config struct {
	Token      string
	BaseURL    string
	HTTPClient *http.Client // Opcional: o padrão será criado
}

type AsaasApi struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
}

// NewClient cria uma nova instância do cliente Asaas com a configuração fornecida
func NewClient(cfg Config) (*AsaasApi, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("token é obrigatório")
	}

	// Se nenhum cliente HTTP for fornecido, cria um com configurações padrão
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
				MaxConnsPerHost:     10,
			},
		}
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = BaseURLSandbox
	}

	return &AsaasApi{
		BaseURL:    baseURL,
		Token:      cfg.Token,
		httpClient: httpClient,
	}, nil
}

// NewAsaasApi cria uma nova instância do cliente Asaas (mantido para compatibilidade)
func NewAsaasApi(environment environment, token string) *AsaasApi {
	client, _ := NewClient(Config{
		Token:   token,
		BaseURL: getBaseURL(environment),
	})
	return client
}

// CreateCustomer é o método responsável por realizar a criação do Cliente
func (a *AsaasApi) CreateCustomer(ctx context.Context, customerRequest CustomerRequest) (*CustomerResponse, error) {

	params := request.Params{
		Method:  "POST",
		Body:    customerRequest,
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathCustomers,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("CreateCustomer: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("CreateCustomer: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("CreateCustomer: %w", apiErr)
	}

	var customerResponse CustomerResponse
	if err := json.Unmarshal(response.RawBody, &customerResponse); err != nil {
		return nil, fmt.Errorf("CreateCustomer: falha ao fazer parse da resposta: %w", err)
	}
	return &customerResponse, nil
}

// GetCustomerByAsaasId é o método responsável por buscar um cliente pelo ID Asaas
func (a *AsaasApi) GetCustomerByAsaasId(ctx context.Context, customerId string) (*CustomerResponse, error) {

	params := request.Params{
		Method:  "GET",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathCustomers + "/" + customerId,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("GetCustomerByAsaasId: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("GetCustomerByAsaasId: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("GetCustomerByAsaasId: %w", apiErr)
	}

	var customerResponse CustomerResponse
	if err := json.Unmarshal(response.RawBody, &customerResponse); err != nil {
		return nil, fmt.Errorf("GetCustomerByAsaasId: falha ao fazer parse da resposta: %w", err)
	}
	return &customerResponse, nil
}

// GetCustomerByCpfCnpj é o método responsável por buscar um cliente pelo CPF/CNPJ
func (a *AsaasApi) GetCustomerByCpfCnpj(ctx context.Context, customerCpfCnpj string) (*CustomerResponse, error) {

	params := request.Params{
		Method:  "GET",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathCustomers,
		QueryParams: map[string]any{
			"cpfCnpj": customerCpfCnpj,
		},
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("GetCustomerByCpfCnpj: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("GetCustomerByCpfCnpj: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("GetCustomerByCpfCnpj: %w", apiErr)
	}

	var customerResponse ListCustomerResponse
	if err := json.Unmarshal(response.RawBody, &customerResponse); err != nil {
		return nil, fmt.Errorf("GetCustomerByCpfCnpj: falha ao fazer parse da resposta: %w", err)
	}

	if len(customerResponse.Data) > 0 {
		return &customerResponse.Data[0], nil
	}

	return nil, fmt.Errorf("GetCustomerByCpfCnpj: %w", ErrNotFound)
}

// GetCustomerByName é o método responsável por buscar um cliente pelo nome
func (a *AsaasApi) GetCustomerByName(ctx context.Context, customerName string) (*CustomerResponse, error) {

	params := request.Params{
		Method:  "GET",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathCustomers,
		QueryParams: map[string]any{
			"name": customerName,
		},
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("GetCustomerByName: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("GetCustomerByName: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("GetCustomerByName: %w", apiErr)
	}

	var customerResponse ListCustomerResponse
	if err := json.Unmarshal(response.RawBody, &customerResponse); err != nil {
		return nil, fmt.Errorf("GetCustomerByName: falha ao fazer parse da resposta: %w", err)
	}

	if len(customerResponse.Data) > 0 {
		return &customerResponse.Data[0], nil
	}

	return nil, fmt.Errorf("GetCustomerByName: %w", ErrNotFound)
}

// DeleteCustomer é o método responsável por deletar um cliente pelo ID Asaas
func (a *AsaasApi) DeleteCustomer(ctx context.Context, customerId string) (*DeleteCustomerResponse, error) {

	params := request.Params{
		Method:  "DELETE",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathCustomers + "/" + customerId,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("DeleteCustomer: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("DeleteCustomer: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("DeleteCustomer: %w", apiErr)
	}

	var deleteCustomerResponse DeleteCustomerResponse
	if err := json.Unmarshal(response.RawBody, &deleteCustomerResponse); err != nil {
		return nil, fmt.Errorf("DeleteCustomer: falha ao fazer parse da resposta: %w", err)
	}
	return &deleteCustomerResponse, nil
}

// CreateBilling é o método responsável por realizar a criação de uma Cobrança
func (a *AsaasApi) CreateBilling(ctx context.Context, billingRequest BillingRequest) (*BillingResponse, error) {

	params := request.Params{
		Method:  "POST",
		Body:    billingRequest,
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathPayments,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("CreateBilling: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("CreateBilling: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("CreateBilling: %w", apiErr)
	}

	var billingResponse BillingResponse
	if err := json.Unmarshal(response.RawBody, &billingResponse); err != nil {
		return nil, fmt.Errorf("CreateBilling: falha ao fazer parse da resposta: %w", err)
	}
	return &billingResponse, nil
}

// GetBillingByAsaasId é o método responsável por buscar uma cobrança pelo ID Asaas
func (a *AsaasApi) GetBillingByAsaasId(ctx context.Context, billingId string) (*BillingResponse, error) {

	params := request.Params{
		Method:  "GET",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathPayments + "/" + billingId,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("GetBillingByAsaasId: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("GetBillingByAsaasId: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("GetBillingByAsaasId: %w", apiErr)
	}

	var billingResponse BillingResponse
	if err := json.Unmarshal(response.RawBody, &billingResponse); err != nil {
		return nil, fmt.Errorf("GetBillingByAsaasId: falha ao fazer parse da resposta: %w", err)
	}
	return &billingResponse, nil
}

// DeleteBilling é o método responsável por deletar uma cobrança pelo ID Asaas
func (a *AsaasApi) DeleteBilling(ctx context.Context, billingId string) (*DeleteBillingResponse, error) {

	params := request.Params{
		Method:  "DELETE",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathPayments + "/" + billingId,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("DeleteBilling: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("DeleteBilling: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("DeleteBilling: %w", apiErr)
	}

	var deleteBillingResponse DeleteBillingResponse
	if err := json.Unmarshal(response.RawBody, &deleteBillingResponse); err != nil {
		return nil, fmt.Errorf("DeleteBilling: falha ao fazer parse da resposta: %w", err)
	}
	return &deleteBillingResponse, nil
}

// CreateSubscription é o método responsável por realizar a criação de uma assinatura para um cliente
func (a *AsaasApi) CreateSubscription(ctx context.Context, subscriptionRequest SubscriptionRequest) (*SubscriptionResponse, error) {

	params := request.Params{
		Method:  "POST",
		Body:    subscriptionRequest,
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathSubscriptions,
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("CreateSubscription: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("CreateSubscription: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("CreateSubscription: %w", apiErr)
	}

	var subscriptionResponse SubscriptionResponse
	if err := json.Unmarshal(response.RawBody, &subscriptionResponse); err != nil {
		return nil, fmt.Errorf("CreateSubscription: falha ao fazer parse da resposta: %w", err)
	}
	return &subscriptionResponse, nil
}

// GetSubscriptionsByCustomerId é o método responsável por buscar as assinaturas de um cliente
func (a *AsaasApi) GetSubscriptionsByCustomerId(ctx context.Context, customerId string) ([]SubscriptionResponse, error) {

	params := request.Params{
		Method:  "GET",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathSubscriptions,
		QueryParams: map[string]any{
			"customer": customerId,
		},
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("GetSubscriptionsByCustomerId: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("GetSubscriptionsByCustomerId: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("GetSubscriptionsByCustomerId: %w", apiErr)
	}

	var subscriptionResponse ListSubscriptionResponse
	if err := json.Unmarshal(response.RawBody, &subscriptionResponse); err != nil {
		return nil, fmt.Errorf("GetSubscriptionsByCustomerId: falha ao fazer parse da resposta: %w", err)
	}
	return subscriptionResponse.Data, nil
}

// GetSubscriptionsPayments é o método responsável por buscar os pagamentos de uma a assinatura de um cliente
func (a *AsaasApi) GetSubscriptionsPayments(ctx context.Context, subscriptionId string) ([]BillingResponse, error) {

	params := request.Params{
		Method:  "GET",
		Headers: map[string]string{HeaderAccessToken: a.Token},
		URL:     a.BaseURL + PathSubscriptions + "/" + subscriptionId + "/payments",
		Context: ctx,
		Client:  a.httpClient,
	}

	response, err := request.New(params)
	if err != nil {
		return nil, fmt.Errorf("GetSubscriptionsPayments: %w", err)
	}

	if response.StatusCode >= StatusClientErrorMin {
		apiErr, err := parseError(response.RawBody)
		if err != nil {
			return nil, fmt.Errorf("GetSubscriptionsPayments: API error (status %d, falha ao fazer parse): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("GetSubscriptionsPayments: %w", apiErr)
	}

	var billingsResponse ListBillingResponse
	if err := json.Unmarshal(response.RawBody, &billingsResponse); err != nil {
		return nil, fmt.Errorf("GetSubscriptionsPayments: falha ao fazer parse da resposta: %w", err)
	}
	return billingsResponse.Data, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getAccessToken é a função responsável por retornar o AccessToken do Asaas.
// Caso tenha sido passado um token por parâmetro pegamos o token passado por parâmetro, se não pegamos da variável de ambiente ASAAS_ACCESS_TOKEN.
func getAccessToken(asaasAccessToken ...string) string {
	if len(asaasAccessToken) > 0 {
		return asaasAccessToken[0]
	} else {
		return os.Getenv("ASAAS_ACCESS_TOKEN")
	}
}

// getBaseURL é a função responsável por validar o ambiente e a URL base.
func getBaseURL(environment environment) string {
	if environment == Production {
		return BaseURLProduction
	}
	return BaseURLSandbox
}

// parseError é a função que pega os dados do erro do Asaas e retorna em formato de APIError.
func parseError(body []byte) (*APIError, error) {
	var errResponse ErrorResponse
	if err := json.Unmarshal(body, &errResponse); err != nil {
		return nil, fmt.Errorf("falha ao fazer parse da resposta de erro: %w", err)
	}

	statusCode := 0
	if errResponse.Status != 0 {
		statusCode = errResponse.Status
	}

	return &APIError{
		StatusCode: statusCode,
		ErrorCode:  errResponse.Error,
		Message:    errResponse.Message,
		Err:        nil,
	}, nil
}
