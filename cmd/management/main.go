package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

const (
	defaultSocketPath = "/var/run/proxy/management.sock"
	templatePath      = "/app/nginx.template"
	configPath        = "/etc/nginx/conf.d/default.conf"
	streamConfigPath  = "/etc/nginx/stream.d/default.conf"
	defaultConfigPath = "/app/default.conf"
	streamDelimiter   = "### STREAM_CONFIG ###"
)

func main() {
	socketPath := os.Getenv("MANAGEMENT_SOCKET")
	if socketPath == "" {
		socketPath = defaultSocketPath
	}

	// Clean up stale socket from previous run
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", socketPath, err)
	}
	defer listener.Close()

	// Make socket accessible to other containers sharing the volume
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Printf("Warning: could not chmod socket: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /template", handleGetTemplate)
	mux.HandleFunc("POST /config", handlePostConfig)
	mux.HandleFunc("GET /health", handleHealth)

	server := &http.Server{Handler: mux}

	// Graceful shutdown on signals
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Management server shutting down...")
		server.Close()
	}()

	log.Printf("Management server listening on %s", socketPath)
	if err := server.Serve(listener); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		http.Error(w, "Failed to read template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func handlePostConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		// Empty body means reset to default configuration
		log.Println("Resetting configuration to default")
		defaultData, err := os.ReadFile(defaultConfigPath)
		if err != nil {
			http.Error(w, "Failed to read default config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		body = defaultData

		// Remove stream config on reset
		os.Remove(streamConfigPath)
	} else {
		log.Println("Writing new configuration")
	}

	// Split HTTP and stream config on the delimiter
	httpConfig, streamConfig := splitConfig(string(body))

	if err := os.WriteFile(configPath, []byte(httpConfig), 0644); err != nil {
		http.Error(w, "Failed to write HTTP config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Write or remove stream config
	if streamConfig != "" {
		log.Println("Writing stream configuration")
		if err := os.WriteFile(streamConfigPath, []byte(streamConfig), 0644); err != nil {
			http.Error(w, "Failed to write stream config: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		os.Remove(streamConfigPath)
	}

	// Validate config before reloading
	testCmd := exec.Command("nginx", "-t")
	if testOutput, err := testCmd.CombinedOutput(); err != nil {
		http.Error(w, "Config validation failed: "+err.Error()+"\n"+string(testOutput), http.StatusBadRequest)
		return
	}

	// Reload nginx
	log.Println("Reloading nginx")
	reloadCmd := exec.Command("nginx", "-s", "reload")
	if output, err := reloadCmd.CombinedOutput(); err != nil {
		http.Error(w, "Failed to reload nginx: "+err.Error()+"\n"+string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration updated and nginx reloaded\n"))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

// splitConfig separates the rendered template output into HTTP and stream
// portions using the well-known delimiter. Returns (httpConfig, streamConfig).
func splitConfig(config string) (string, string) {
	parts := strings.SplitN(config, streamDelimiter, 2)
	httpConfig := strings.TrimSpace(parts[0]) + "\n"

	streamConfig := ""
	if len(parts) > 1 {
		streamConfig = strings.TrimSpace(parts[1])
		if streamConfig != "" {
			streamConfig += "\n"
		}
	}

	return httpConfig, streamConfig
}
