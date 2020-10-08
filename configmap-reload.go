package main

import (
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/process"
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
	signalHook        signalHookFlag
	webhookMethod     = flag.String("webhook-method", "POST", "the HTTP method url to use to send the webhook")
	webhookStatusCode = flag.Int("webhook-status-code", 200, "the HTTP status code indicating successful triggering of reload")
	webhookRetries    = flag.Int("webhook-retries", 1, "the amount of times to retry the webhook reload request")
	listenAddress     = flag.String("web.listen-address", ":9533", "Address to listen on for web interface and telemetry.")
	metricPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")

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
)

func init() {
	prometheus.MustRegister(lastReloadError)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(successReloads)
	prometheus.MustRegister(requestErrorsByReason)
	prometheus.MustRegister(watcherErrors)
	prometheus.MustRegister(requestsByStatusCode)
}

func main() {
	flag.Var(&volumeDirs, "volume-dir", "the config map volume directory to watch for updates; may be used multiple times")
	flag.Var(&webhook, "webhook-url", "the url to send a request to when the specified config map volume directory has been updated")
	flag.Var(&signalHook, "signalHook", "the signal params to send a signal to target process when the specified config map volume directory has been updated, format:signalNumber:processName, eg: send nginx process with SIGHUP: 1:nginx ")
	flag.Parse()

	if len(volumeDirs) < 1 {
		log.Println("Missing volume-dir")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}

	if len(webhook) < 1 && len(signalHook) < 1 {
		log.Println("Missing webhook-url or signalHook")
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
				if !isValidEvent(event) {
					continue
				}
				log.Println("config map updated")
				for _, h := range webhook {
					begun := time.Now()
					req, err := http.NewRequest(*webhookMethod, h.String(), nil)
					if err != nil {
						setFailureMetrics(h.String(), "client_request_create")
						log.Println("error:", err)
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
						log.Printf("performing webhook request (%d/%d)", retries, *webhookRetries)
						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							setFailureMetrics(h.String(), "client_request_do")
							log.Println("error:", err)
							time.Sleep(time.Second * 10)
							continue
						}
						resp.Body.Close()
						requestsByStatusCode.WithLabelValues(h.String(), strconv.Itoa(resp.StatusCode)).Inc()
						if resp.StatusCode != *webhookStatusCode {
							setFailureMetrics(h.String(), "client_response")
							log.Println("error:", "Received response code", resp.StatusCode, ", expected", *webhookStatusCode)
							time.Sleep(time.Second * 10)
							continue
						}

						setSuccessMetrics(h.String(), begun)
						log.Println("successfully triggered reload")
						successfulReloadWebhook = true
						break
					}

					if !successfulReloadWebhook {
						setFailureMetrics(h.String(), "retries_exhausted")
						log.Println("error:", "Webhook reload retries exhausted")
					}
				}
				for _, h := range signalHook {
					begun := time.Now()
					err := signalHookExec(h)
					if err != nil {
						setFailureMetrics(h.processName, "signal_hook_do")
						log.Println("error:", err)
						continue
					}
					successfulReloadWebhook := false
					for retries := *webhookRetries; retries != 0; retries-- {
						log.Printf("performing webhook request (%d/%d)", retries, *webhookRetries)
						err := signalHookExec(h)
						if err != nil {
							setFailureMetrics(h.processName, "signal_hook_do")
							log.Println("error:", err)
							time.Sleep(time.Second * 10)
							continue
						}

						setSuccessMetrics(h.processName, begun)
						log.Println("successfully triggered reload")
						successfulReloadWebhook = true
						break
					}

					if !successfulReloadWebhook {
						setFailureMetrics(h.processName, "retries_exhausted")
						log.Println("error:", "Webhook reload retries exhausted")
					}
				}
			case err := <-watcher.Errors:
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

func findProcessByName(name string) (*os.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	for _, process := range processes {
		processName, err := process.Name()
		if err != nil {
			log.Printf("find process by name %s: get process name %d err: %s", name, process.Pid, err)
			continue
		}
		cmdline, err := process.Cmdline()
		if err != nil {
			log.Printf("find process by name %s: get process cmdline %d err: %s", name, process.Pid, err)
			continue
		}
		log.Printf("process name (target name: %s) %d : %s : %s", name, process.Pid, processName, cmdline)

		if cmdline == name {
			p, err := os.FindProcess(int(process.Pid))
			if err != nil {
				return nil, fmt.Errorf("find process by name %s: tran pid: %d to os process err: %s", name, process.Pid, err)
			}
			return p, nil
		}
	}
	return nil, fmt.Errorf("not find pid with process name %s", name)
}

func signalHookExec(h *signalHookParam) error {
	signal := syscall.Signal(h.signalNumber)
	log.Printf("signaling process ( %s) -> %d (%s)\n", h.processName, h.signalNumber, signal.String())
	process, err := findProcessByName(h.processName)
	if err != nil {
		return fmt.Errorf(" find process with name %s err: %s", h.processName, err)
	}
	log.Printf("signaling process (%s) pid: %d -> %d (%s)\n", h.processName, process.Pid, h.signalNumber, signal.String())
	return process.Signal(signal)
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

type signalHookParam struct {
	processName  string
	signalNumber int
}

type signalHookFlag []*signalHookParam

func (v *signalHookFlag) Set(value string) error {
	strings := strings.SplitN(value, ":", 2)
	if len(strings) != 2 {
		return fmt.Errorf("invalid signalHook: %v, format: signalNumber:processName,eg: 1:nginx ", value)
	}
	signalNumber, err := strconv.Atoi(strings[0])
	if err != nil {
		return fmt.Errorf("invalid signalHook: %v, format: signalNumber:processName,eg: 1:nginx ", value)
	}
	*v = append(*v, &signalHookParam{
		processName:  strings[1],
		signalNumber: signalNumber,
	})
	return nil
}

func (v *signalHookFlag) String() string {
	return fmt.Sprint(*v)
}
