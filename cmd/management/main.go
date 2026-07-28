package main

import (
	"fmt"
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

	if status, err := applyConfig(configPath, streamConfigPath, httpConfig, streamConfig, nginxTest); err != nil {
		http.Error(w, err.Error(), status)
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

// nginxTest is the validation applyConfig uses in production: nginx checking its own
// configuration, which is the only thing that can say authoritatively whether it loads.
func nginxTest() ([]byte, error) {
	return exec.Command("nginx", "-t").CombinedOutput()
}

// applyConfig writes the configuration to disk, validates it, and puts back whatever was
// there before if validation fails.
//
// Refusing to reload protects the RUNNING nginx, but the rejected files used to stay on
// disk — so the next container restart loaded them and nginx would not come up at all.
// The proxy was then down until someone worked out that a config posted hours earlier had
// been rejected. A config bad enough to refuse is bad enough not to leave behind.
//
// Returns the HTTP status to report alongside the error, or 0 and nil on success.
func applyConfig(httpPath, streamPath, httpConfig, streamConfig string, validate func() ([]byte, error)) (int, error) {
	prevHTTP, prevHTTPErr := os.ReadFile(httpPath)
	prevStream, prevStreamErr := os.ReadFile(streamPath)
	restore := func() {
		// Absent before means absent after. Writing nothing back would leave the
		// rejected file exactly where it was, which is the bug this prevents.
		if prevHTTPErr == nil {
			_ = os.WriteFile(httpPath, prevHTTP, 0644)
		} else {
			_ = os.Remove(httpPath)
		}
		if prevStreamErr == nil {
			_ = os.WriteFile(streamPath, prevStream, 0644)
		} else {
			_ = os.Remove(streamPath)
		}
	}

	if err := os.WriteFile(httpPath, []byte(httpConfig), 0644); err != nil {
		restore()
		return http.StatusInternalServerError, fmt.Errorf("failed to write HTTP config: %w", err)
	}

	if streamConfig != "" {
		log.Println("Writing stream configuration")
		if err := os.WriteFile(streamPath, []byte(streamConfig), 0644); err != nil {
			restore()
			return http.StatusInternalServerError, fmt.Errorf("failed to write stream config: %w", err)
		}
	} else {
		os.Remove(streamPath)
	}

	if out, err := validate(); err != nil {
		restore()
		log.Printf("Config validation failed, restored the previous configuration: %v", err)
		return http.StatusBadRequest, fmt.Errorf("config validation failed: %w\n%s", err, out)
	}

	return 0, nil
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
