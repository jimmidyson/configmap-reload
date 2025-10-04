package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	fsnotify "github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "configmap_reload"

var (
	volumeDirs        volumeDirsFlag
	webhook           webhookFlag
	webhookMethod     = flag.String("webhook-method", "POST", "the HTTP method url to use to send the webhook (only applies if --target-process-name is not set)")
	webhookStatusCode = flag.Int("webhook-status-code", 200, "the HTTP status code indicating successful triggering of reload (only applies if --target-process-name is not set)")
	webhookRetries    = flag.Int("webhook-retries", 1, "the amount of times to retry the webhook reload request (only applies if --target-process-name is not set)")
	listenAddress     = flag.String("web.listen-address", ":9533", "Address to listen on for web interface and telemetry.")
	metricPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	// targetProcessNames is now a custom flag type to accept multiple values
	targetProcessNames processNamesFlag

	lastReloadError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "last_reload_error",
		Help:      "Whether the last reload resulted in an error (1 for error, 0 for success)",
	}, []string{"webhook"})
	requestDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "last_request_duration_seconds",
		Help:      "Duration of last webhook request",
	}, []string{"webhook"})
	successReloads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "success_reloads_total",
		Help:      "Total success reload calls",
	}, []string{"webhook"})
	requestErrorsByReason = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "request_errors_total",
		Help:      "Total request errors by reason",
	}, []string{"webhook", "reason"})
	watcherErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "watcher_errors_total",
		Help:      "Total filesystem watcher errors",
	})
	requestsByStatusCode = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "requests_total",
		Help:      "Total requests by response status code",
	}, []string{"webhook", "status_code"})

	// New Prometheus metrics for SIGHUP operations
	sighupLastReloadError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sighup_last_reload_error",
		Help:      "Whether the last SIGHUP reload resulted in an error (1 for error, 0 for success)",
	}, []string{"process_name"})
	sighupSuccesses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "sighup_success_reloads_total",
		Help:      "Total success SIGHUP reload calls",
	}, []string{"process_name"})
	sighupErrorsByReason = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "sighup_request_errors_total",
		Help:      "Total SIGHUP request errors by reason",
	}, []string{"process_name", "reason"})
)

func init() {
	prometheus.MustRegister(lastReloadError)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(successReloads)
	prometheus.MustRegister(requestErrorsByReason)
	prometheus.MustRegister(watcherErrors)
	prometheus.MustRegister(requestsByStatusCode)
	// Register new SIGHUP metrics
	prometheus.MustRegister(sighupLastReloadError)
	prometheus.MustRegister(sighupSuccesses)
	prometheus.MustRegister(sighupErrorsByReason)
}

func main() {
	flag.Var(&volumeDirs, "volume-dir", "the config map volume directory to watch for updates; may be used multiple times")
	flag.Var(&webhook, "webhook-url", "the url to send a request to when the specified config map volume directory has been updated (ignored if --target-process-name is set)")
	// Register the new targetProcessNames flag
	flag.Var(&targetProcessNames, "target-process-name", "Name of the target application process to send SIGHUP to when a filesystem change is detected; may be used multiple times. If empty, a webhook will be triggered instead.")
	flag.Parse()

	if len(volumeDirs) < 1 {
		log.Println("Missing volume-dir")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}

	// Validate that either webhook-url OR target-process-name is set, but not both
	webhookSet := len(webhook) > 0
	targetProcessNameSet := len(targetProcessNames) > 0

	if !webhookSet && !targetProcessNameSet {
		log.Println("Error: Either --webhook-url or --target-process-name must be specified.")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}
	if webhookSet && targetProcessNameSet {
		log.Println("Error: Cannot specify both --webhook-url and --target-process-name. Choose one strategy.")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// Handle filesystem events
				if !isValidEvent(event) {
					continue
				}
				log.Println("Config map updated via filesystem event.")

				if targetProcessNameSet { // Check if any target process names are set
					// If target-process-name(s) are set, find PIDs and send SIGHUP
					sendSIGHUPToTargetProcesses(targetProcessNames)
				} else {
					// If no target-process-name is set, trigger the webhook reload
					log.Println("Triggering webhook reload...")
					triggerWebhookReload()
				}
			case err := <-watcher.Errors:
				// Handle watcher errors
				watcherErrors.Inc()
				log.Println("error:", err)
			}
		}
	}()

	for _, d := range volumeDirs {
		log.Printf("Watching directory: %q", d)
		err = watcher.Add(d)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Fatal(serverMetrics(*listenAddress, *metricPath))
}

// sendSIGHUPToTargetProcesses iterates through a list of process names,
// finds their PIDs, and sends a SIGHUP signal to each found process.
func sendSIGHUPToTargetProcesses(processNames []string) {
	for _, procName := range processNames {
		pid, err := findPIDbyName(procName)
		if err != nil {
			log.Printf("Error finding PID for process '%s': %v", procName, err)
			setSIGHUPFailureMetrics(procName, "find_pid_error")
		} else if pid == 0 {
			log.Printf("Process '%s' not found.", procName)
			setSIGHUPFailureMetrics(procName, "process_not_found")
		} else {
			log.Printf("Attempting to send SIGHUP to target process '%s' with PID: %d", procName, pid)
			proc, err := os.FindProcess(pid)
			if err != nil {
				log.Printf("Error finding process object for PID %d: %v", pid, err)
				setSIGHUPFailureMetrics(procName, "find_proc_object_error")
			} else {
				if err := proc.Signal(syscall.SIGHUP); err != nil {
					log.Printf("Error sending SIGHUP to target process (PID %d): %v", pid, err)
					setSIGHUPFailureMetrics(procName, "send_sighup_error")
				} else {
					log.Printf("Successfully sent SIGHUP to target process '%s' with PID: %d", procName, pid)
					setSIGHUPSuccessMetrics(procName)
				}
			}
		}
	}
}

// findPIDbyName attempts to find the PID of a process by its name.
// It iterates through /proc/<pid>/comm to find a matching process.
// Returns the first matching PID found, or 0 if not found, along with an error.
func findPIDbyName(processName string) (int, error) {
	// Read the /proc directory
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc directory: %w", err)
	}

	for _, entry := range entries {
		// Check if the entry name is a PID (a number)
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID directory
		}

		// Read the 'comm' file for the process (command name)
		commPath := filepath.Join("/proc", entry.Name(), "comm")
		commBytes, err := os.ReadFile(commPath)
		if err != nil {
			// This can happen for processes that exit or permission issues, just skip
			continue
		}

		// Trim whitespace (especially newline) and compare
		commName := strings.TrimSpace(string(commBytes))
		if commName == processName {
			return pid, nil // Found a match
		}
	}

	return 0, nil // Process not found
}

// triggerWebhookReload encapsulates the logic for sending webhook requests
func triggerWebhookReload() {
	for _, h := range webhook {
		begun := time.Now()
		req, err := http.NewRequest(*webhookMethod, h.String(), nil)
		if err != nil {
			setFailureMetrics(h.String(), "client_request_create")
			log.Println("error creating webhook request:", err)
			continue
		}
		userInfo := h.User
		if userInfo != nil {
			if password, passwordSet := userInfo.Password(); passwordSet {
				req.SetBasicAuth(userInfo.Username(), password)
			}
		}

		successfulReloadWebhook := false

		for retries := *webhookRetries; retries != 0; retries-- {
			log.Printf("Performing webhook request to %s (%d/%d)", h.String(), *webhookRetries-retries+1, *webhookRetries)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				setFailureMetrics(h.String(), "client_request_do")
				log.Println("error performing webhook request:", err)
				time.Sleep(time.Second * 10) // Wait before retrying
				continue
			}
			resp.Body.Close()
			requestsByStatusCode.WithLabelValues(h.String(), strconv.Itoa(resp.StatusCode)).Inc()
			if resp.StatusCode != *webhookStatusCode {
				setFailureMetrics(h.String(), "client_response")
				log.Printf("error: Received response code %d from %s, expected %d", resp.StatusCode, h.String(), *webhookStatusCode)
				time.Sleep(time.Second * 10) // Wait before retrying
				continue
			}

			setSuccessMetrics(h.String(), begun)
			log.Printf("Successfully triggered reload for %s", h.String())
			successfulReloadWebhook = true
			break // Break out of retry loop on success
		}

		if !successfulReloadWebhook {
			setFailureMetrics(h.String(), "retries_exhausted")
			log.Printf("error: Webhook reload retries exhausted for %s", h.String())
		}
	}
}

// setSIGHUPFailureMetrics updates Prometheus metrics for failed SIGHUP attempts.
func setSIGHUPFailureMetrics(procName, reason string) {
	sighupErrorsByReason.WithLabelValues(procName, reason).Inc()
	sighupLastReloadError.WithLabelValues(procName).Set(1.0)
}

// setSIGHUPSuccessMetrics updates Prometheus metrics for successful SIGHUP attempts.
func setSIGHUPSuccessMetrics(procName string) {
	sighupSuccesses.WithLabelValues(procName).Inc()
	sighupLastReloadError.WithLabelValues(procName).Set(0.0)
}

func setFailureMetrics(h, reason string) {
	requestErrorsByReason.WithLabelValues(h, reason).Inc()
	lastReloadError.WithLabelValues(h).Set(1.0)
}

func setSuccessMetrics(h string, begun time.Time) {
	requestDuration.WithLabelValues(h).Set(time.Since(begun).Seconds())
	successReloads.WithLabelValues(h).Inc()
	lastReloadError.WithLabelValues(h).Set(0.0)
}

func isValidEvent(event fsnotify.Event) bool {
	// Only trigger on create events for the "..data" symlink
	if event.Op&fsnotify.Create != fsnotify.Create {
		return false
	}
	if filepath.Base(event.Name) != "..data" {
		return false
	}
	return true
}

func serverMetrics(listenAddress, metricsPath string) error {
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>ConfigMap Reload Metrics</title></head>
			<body>
			<h1>ConfigMap Reload</h1>
			<p><a href='` + metricsPath + `'>Metrics</a></p>
			</body>
			</html>
		`))
	})
	return http.ListenAndServe(listenAddress, nil)
}

// processNamesFlag is a custom flag type to allow multiple --target-process-name flags
type processNamesFlag []string

func (p *processNamesFlag) Set(value string) error {
	*p = append(*p, value)
	return nil
}

func (p *processNamesFlag) String() string {
	return fmt.Sprint(*p)
}

type volumeDirsFlag []string

func (v *volumeDirsFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func (v *volumeDirsFlag) String() string {
	return fmt.Sprint(*v)
}

type webhookFlag []*url.URL

func (v *webhookFlag) Set(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	*v = append(*v, u)
	return nil
}

func (v *webhookFlag) String() string {
	return fmt.Sprint(*v)
}
