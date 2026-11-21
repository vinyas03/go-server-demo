package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-learn/internal/handler"
	"go-learn/internal/middleware"
	"go-learn/internal/repository"
	"go-learn/internal/service"

	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
)

func main() {
	// 1. CubeAPM Configuration Defaults (can be overridden via environment variables)
	if os.Getenv("ELASTIC_APM_SERVICE_NAME") == "" {
		os.Setenv("ELASTIC_APM_SERVICE_NAME", "go-learn-api")
	}
	if os.Getenv("ELASTIC_APM_SERVER_URL") == "" {
		os.Setenv("ELASTIC_APM_SERVER_URL", "http://localhost:3130")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf(":%s", port)

	// 2. Dependency Injection
	// DI is a design pattern that allows us to inject dependencies into a component rather than having the component create them itself.
	// In this case, we are injecting the repository into the service, and the service into the handler.
	// Explanation of DI:
	// 1. Without DI:
	// In this pattern, the TaskService would create its own instance of the repository, which would make it difficult to test and maintain.
	// Why difficult to test and maintain? Because if the TaskService creates its own instance of the repository, it would be tightly coupled to that specific implementation.
	// This makes it hard to swap out the repository for a mock or a different implementation during testing or if we want to change the data store in the future.

	// 2. With DI:
	// In this pattern, we inject the repository into the TaskService, which makes it easier to test and maintain.
	// In this case, we can easily swap out the repository for a mock or a different implementation during testing or if we want to change the data store in the future.
	// Easier to test because we can inject a mock repository that simulates the behavior of the real repository without actually hitting a database or external service.
	// Easier to maintain because we can change the implementation of the repository without having to change the TaskService or any other components that depend on it.

	// Repository Layer (Data Store)
	repo := repository.NewMemoryTaskRepository()

	// Business Logic Layer (Service)
	taskService := service.NewTaskService(repo)

	// Presentation / Controller Layer (Handler)
	taskHandler := handler.NewTaskHandler(taskService)

	// 3. Routing (standard net/http ServeMux with Go 1.22+ routing)
	mux := http.NewServeMux()

	// Health check route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler.JSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Register domain routes
	taskHandler.RegisterRoutes(mux)

	// 4. Wrap with Middleware
	// Order: Recovery -> RequestContext -> Logger -> CORS
	wrappedHandler := middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.RequestContextMiddleware,
		middleware.Logger,
		middleware.CORS,
	)

	// 5. Wrap with CubeAPM HTTP Tracer
	// apmhttp.Wrap instruments all inbound HTTP requests into CubeAPM transactions,
	// tracks status codes, latency, path, and propagates distributed trace contexts.
	tracedHandler := apmhttp.Wrap(wrappedHandler)

	// 6. HTTP Server Configuration with safe timeouts
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      tracedHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Graceful Shutdown Setup - this line sets up a channel to listen for OS signals (like SIGINT or SIGTERM) to gracefully shut down the server when the application is terminated.
	//without this line, the server would not be able to handle shutdown signals properly, which could lead to abrupt termination of the server and potential loss of in-flight requests or data.
	//example of both scenarios:
	// 1. With Graceful Shutdown:
	// When the server receives a shutdown signal, it stops accepting new requests and waits for existing requests to complete before shutting down.
	// This ensures that all in-flight requests are processed and no data is lost.
	// 2. Without Graceful Shutdown:
	// When the server receives a shutdown signal, it immediately terminates all in-flight requests and shuts down.
	// This can lead to loss of data and incomplete processing of requests.
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on http://localhost%s", serverAddr)
		log.Printf("CubeAPM Tracing enabled -> Service: %s, Collector: %s",
			os.Getenv("ELASTIC_APM_SERVICE_NAME"),
			os.Getenv("ELASTIC_APM_SERVER_URL"),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Block until signal is received
	// This line blocks the main goroutine until a signal is received on the shutdownChan channel.
	sig := <-shutdownChan
	log.Printf("Received signal %s. Shutting down gracefully...", sig)

	// Flush any pending traces to CubeAPM collector
	//this will ensure that any pending traces are sent to the CubeAPM collector before the application exits, preventing loss of tracing data.
	//pending traces are traces that have been recorded but not yet sent to the CubeAPM collector. If the application exits before these traces are sent,
	// they will be lost and will not be available for analysis in the CubeAPM dashboard.
	//this scenario can occur if the application is terminated abruptly (e.g., due to a crash or a forced shutdown) or if the application is shut down gracefully but there are still pending traces that have not been sent to the collector.
	// for example, if the application is processing a long-running request and is terminated before the request completes, any traces associated with that request may be lost if they have not yet been sent to the collector.
	apm.DefaultTracer().Flush(nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown error: %v", err)
	} else {
		log.Println("Server exited cleanly.")
	}
}
