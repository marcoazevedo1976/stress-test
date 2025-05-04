package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// Config representa a configuração do teste de carga
type Config struct {
	URL         string
	Requests    int
	Concurrency int
}

// Result representa o resultado de um único request HTTP
type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

// Report contém as estatísticas do teste de carga
type Report struct {
	TotalTime     time.Duration
	TotalRequests int
	StatusCodes   map[int]int
	Errors        int
}

func main() {
	// Definir flags da linha de comando
	url := flag.String("url", "", "URL do serviço a ser testado")
	requests := flag.Int("requests", 0, "Número total de requests")
	concurrency := flag.Int("concurrency", 0, "Número de chamadas simultâneas")

	flag.Parse()

	// Validar parâmetros
	if *url == "" || *requests <= 0 || *concurrency <= 0 {
		fmt.Println("Uso: stress-test --url=URL --requests=N --concurrency=M")
		os.Exit(1)
	}

	config := Config{
		URL:         *url,
		Requests:    *requests,
		Concurrency: *concurrency,
	}

	// Executar o teste de carga
	report := runLoadTest(config)

	// Imprimir relatório
	printReport(report)
}

// runLoadTest executa o teste de carga com base na configuração fornecida
func runLoadTest(config Config) Report {
	startTime := time.Now()

	// Canal para coletar resultados
	results := make(chan Result, config.Requests)

	// Distribuir trabalho
	var wg sync.WaitGroup

	// Limitar o número de workers ao número de requisições se for menor que a concorrência
	workers := config.Concurrency
	if config.Requests < workers {
		workers = config.Requests
	}

	// Distribuir requisições entre workers
	requestsPerWorker := config.Requests / workers
	remainingRequests := config.Requests % workers

	for i := 0; i < workers; i++ {
		// Calcular número de requisições para este worker
		workerRequests := requestsPerWorker
		if i < remainingRequests {
			workerRequests++
		}

		wg.Add(1)
		go func(requests int) {
			defer wg.Done()
			for j := 0; j < requests; j++ {
				results <- makeRequest(config.URL)
			}
		}(workerRequests)
	}

	// Esperar em uma goroutine separada para fechar o canal quando todos os workers terminarem
	go func() {
		wg.Wait()
		close(results)
	}()

	// Processar resultados
	report := Report{
		StatusCodes: make(map[int]int),
	}

	for result := range results {
		report.TotalRequests++

		if result.Error != nil {
			report.Errors++
		} else {
			report.StatusCodes[result.StatusCode]++
		}
	}

	report.TotalTime = time.Since(startTime)

	return report
}

// makeRequest realiza uma única requisição HTTP e retorna o resultado
func makeRequest(url string) Result {
	startTime := time.Now()

	resp, err := http.Get(url)

	result := Result{
		Duration: time.Since(startTime),
		Error:    err,
	}

	if err == nil {
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	return result
}

// printReport imprime o relatório do teste de carga
func printReport(report Report) {
	fmt.Println("\n--- Relatório do Teste de Carga ---")
	fmt.Printf("Tempo total: %v\n", report.TotalTime)
	fmt.Printf("Total de requests: %d\n", report.TotalRequests)

	// Contar requests com status 200
	successRequests := report.StatusCodes[http.StatusOK]
	fmt.Printf("Requests com status 200: %d\n", successRequests)

	// Distribuição de outros códigos de status
	fmt.Println("Distribuição de códigos de status:")
	for code, count := range report.StatusCodes {
		if code != http.StatusOK {
			fmt.Printf("  %d: %d\n", code, count)
		}
	}

	if report.Errors > 0 {
		fmt.Printf("Requests com erro: %d\n", report.Errors)
	}
}
