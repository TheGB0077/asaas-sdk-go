package asaas

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/govalues/decimal"
	"github.com/joho/godotenv"
)

func setupTest(t *testing.T) *AsaasApi {
	t.Helper()
	err := godotenv.Load(".env")
	if err != nil {
		t.Skip("Arquivo .env não encontrado, pulando teste")
	}
	token := os.Getenv("ASAAS_ACCESS_TOKEN")
	if token == "" {
		t.Skip("ASAAS_ACCESS_TOKEN não definido, pulando teste")
	}
	return NewAsaasApi(Sandbox, token)
}

func generateValidCnpj(seed int) string {
	base := []int{2, 5, 3, 7, 3, 4, 7, 1, seed / 1000 % 10, seed / 100 % 10, seed / 10 % 10, seed % 10}

	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum1 := 0
	for i := 0; i < 12; i++ {
		sum1 += base[i] * weights1[i]
	}
	d1 := sum1 % 11
	if d1 < 2 {
		d1 = 0
	} else {
		d1 = 11 - d1
	}

	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum2 := 0
	for i := 0; i < 12; i++ {
		sum2 += base[i] * weights2[i]
	}
	sum2 += d1 * weights2[12]
	d2 := sum2 % 11
	if d2 < 2 {
		d2 = 0
	} else {
		d2 = 11 - d2
	}

	result := ""
	for _, d := range base {
		result += fmt.Sprintf("%d", d)
	}
	result += fmt.Sprintf("%d%d", d1, d2)
	return result
}

func createTestCustomerWithCleanup(t *testing.T, api *AsaasApi, name string, cpfCnpj string) *CustomerResponse {
	t.Helper()
	customer, err := api.CreateCustomer(context.Background(), CustomerRequest{
		Name:    name,
		CpfCnpj: cpfCnpj,
	})
	if err != nil {
		t.Fatalf("Falha ao criar cliente de teste: %v", err)
	}

	t.Cleanup(func() {
		_, err := api.DeleteCustomer(context.Background(), customer.Id)
		if err != nil {
			t.Logf("Aviso: falha ao deletar cliente %s: %v", customer.Id, err)
		}
	})

	return customer
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("Recurso não encontrado: %v", err)
	}
	if apiErr, ok := AsAPIError(err); ok {
		t.Fatalf("Erro da API Asaas: %s (Status: %d, Code: %s)", apiErr.Message, apiErr.StatusCode, apiErr.ErrorCode)
	}
	t.Fatalf("Erro inesperado: %v", err)
}

func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func TestSuccessOnCreateCustomer(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(1)
	response, err := api.CreateCustomer(context.Background(), CustomerRequest{
		Name:    "TEST_CREATE_1",
		CpfCnpj: cpfCnpj,
	})

	if err != nil {
		if apiErr, ok := AsAPIError(err); ok {
			t.Errorf("Erro não tratado Asaas: %s (Status: %d)", apiErr.Message, apiErr.StatusCode)
		} else {
			t.Errorf("Erro inesperado: %v", err)
		}
		return
	}

	t.Cleanup(func() {
		api.DeleteCustomer(context.Background(), response.Id)
	})

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnDeleteCustomer(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(2)
	customer := createTestCustomerWithCleanup(t, api, "TEST_DELETE_2", cpfCnpj)

	response, err := api.DeleteCustomer(context.Background(), customer.Id)
	assertNoError(t, err)

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnGetCustomerByIdAsaas(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(3)
	customer := createTestCustomerWithCleanup(t, api, "TEST_GET_BY_ID_3", cpfCnpj)

	response, err := api.GetCustomerByAsaasId(context.Background(), customer.Id)
	assertNoError(t, err)

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnGetCustomerByCpfCnpj(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(4)
	createTestCustomerWithCleanup(t, api, "TEST_GET_BY_CPF_4", cpfCnpj)

	response, err := api.GetCustomerByCpfCnpj(context.Background(), cpfCnpj)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			t.Error("Cliente não encontrado")
		} else if apiErr, ok := AsAPIError(err); ok {
			t.Errorf("Erro não tratado Asaas: %s (Status: %d)", apiErr.Message, apiErr.StatusCode)
		} else {
			t.Errorf("Erro inesperado: %v", err)
		}
		return
	}

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnGetCustomerByName(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(5)
	customerName := "TEST_GET_BY_NAME_5"
	createTestCustomerWithCleanup(t, api, customerName, cpfCnpj)

	response, err := api.GetCustomerByName(context.Background(), customerName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			t.Error("Cliente não encontrado")
		} else if apiErr, ok := AsAPIError(err); ok {
			t.Errorf("Erro não tratado Asaas: %s (Status: %d)", apiErr.Message, apiErr.StatusCode)
		} else {
			t.Errorf("Erro inesperado: %v", err)
		}
		return
	}

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnCreateBilling(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(6)
	customer := createTestCustomerWithCleanup(t, api, "TEST_BILLING_CREATE_6", cpfCnpj)

	response, err := api.CreateBilling(context.Background(), BillingRequest{
		Customer:    customer.Id,
		BillingType: BillingTypeBoleto,
		Value:       decimal.MustParse("101.99"),
		DueDate:     futureDate(7),
	})

	if err != nil {
		if apiErr, ok := AsAPIError(err); ok {
			t.Errorf("Erro não tratado Asaas: %s (Status: %d)", apiErr.Message, apiErr.StatusCode)
		} else {
			t.Errorf("Erro inesperado: %v", err)
		}
		return
	}

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnGetBillingByIdAsaas(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(7)
	customer := createTestCustomerWithCleanup(t, api, "TEST_BILLING_GET_7", cpfCnpj)

	billing, err := api.CreateBilling(context.Background(), BillingRequest{
		Customer:    customer.Id,
		BillingType: BillingTypeBoleto,
		Value:       decimal.MustParse("101.99"),
		DueDate:     futureDate(7),
	})
	if err != nil {
		t.Fatalf("Falha ao criar cobrança de teste: %v", err)
	}

	response, err := api.GetBillingByAsaasId(context.Background(), billing.Id)
	assertNoError(t, err)

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnDeleteBilling(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(8)
	customer := createTestCustomerWithCleanup(t, api, "TEST_BILLING_DELETE_8", cpfCnpj)

	billing, err := api.CreateBilling(context.Background(), BillingRequest{
		Customer:    customer.Id,
		BillingType: BillingTypeBoleto,
		Value:       decimal.MustParse("101.99"),
		DueDate:     futureDate(7),
	})
	if err != nil {
		t.Fatalf("Falha ao criar cobrança de teste: %v", err)
	}

	response, err := api.DeleteBilling(context.Background(), billing.Id)
	assertNoError(t, err)

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnCreateSubscription(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(9)
	customer := createTestCustomerWithCleanup(t, api, "TEST_SUB_CREATE_9", cpfCnpj)

	response, err := api.CreateSubscription(context.Background(), SubscriptionRequest{
		CustomerId:  customer.Id,
		BillingType: BillingTypePix,
		Value:       decimal.MustParse("5.01"),
		NextDueDate: futureDate(7),
		Cycle:       CycleTypeMonthly,
	})

	if err != nil {
		if apiErr, ok := AsAPIError(err); ok {
			t.Errorf("Erro não tratado Asaas: %s (Status: %d)", apiErr.Message, apiErr.StatusCode)
		} else {
			t.Errorf("Erro inesperado: %v", err)
		}
		return
	}

	fmt.Print(response)
	t.Log(response.Id)
}

func TestSuccessOnGetSubscriptionsByCustomerId(t *testing.T) {
	api := setupTest(t)
	t.Parallel()

	cpfCnpj := generateValidCnpj(10)
	customer := createTestCustomerWithCleanup(t, api, "TEST_SUB_GET_10", cpfCnpj)

	_, err := api.CreateSubscription(context.Background(), SubscriptionRequest{
		CustomerId:  customer.Id,
		BillingType: BillingTypePix,
		Value:       decimal.MustParse("5.01"),
		NextDueDate: futureDate(7),
		Cycle:       CycleTypeMonthly,
	})
	if err != nil {
		t.Fatalf("Falha ao criar assinatura de teste: %v", err)
	}

	response, err := api.GetSubscriptionsByCustomerId(context.Background(), customer.Id)
	assertNoError(t, err)

	for _, item := range response {
		fmt.Print(item.Id)
		t.Log(item.Id)
	}
}
