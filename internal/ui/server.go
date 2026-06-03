package ui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sort"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/dispatch"
	"github.com/prometheus/common/model"
)

//go:embed index.html
var indexHTML []byte

// Server serves the interactive route-testing UI.
type Server struct {
	cfg  *config.Config
	addr string
}

type receiverInfo struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}

type configResponse struct {
	Receivers       []receiverInfo `json:"receivers"`
	DefaultReceiver string         `json:"default_receiver,omitempty"`
	SubRouteCount   int            `json:"sub_route_count"`
}

type matchRequest struct {
	Labels map[string]string `json:"labels"`
}

type matchResponse struct {
	MatchedReceivers []string `json:"matched_receivers"`
}

func NewServer(cfg *config.Config, port int) (*Server, error) {
	if port == 0 {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, fmt.Errorf("finding free port: %w", err)
		}
		port = l.Addr().(*net.TCPAddr).Port
		l.Close()
	}
	return &Server{cfg: cfg, addr: fmt.Sprintf(":%d", port)}, nil
}

func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost%s", s.addr)
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/match", s.handleMatch)
	slog.Info("UI server listening", "url", s.URL())
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	var receivers []receiverInfo
	for _, rcv := range s.cfg.Receivers {
		receivers = append(receivers, receiverInfo{
			Name:  rcv.Name,
			Types: receiverTypes(rcv),
		})
	}
	sort.Slice(receivers, func(i, j int) bool { return receivers[i].Name < receivers[j].Name })

	defaultReceiver := ""
	subRouteCount := 0
	if s.cfg.Route != nil {
		defaultReceiver = s.cfg.Route.Receiver
		subRouteCount = len(s.cfg.Route.Routes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configResponse{
		Receivers:       receivers,
		DefaultReceiver: defaultReceiver,
		SubRouteCount:   subRouteCount,
	})
}

func (s *Server) handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req matchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	labelSet := make(model.LabelSet, len(req.Labels))
	for k, v := range req.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}

	routeTree := dispatch.NewRoute(s.cfg.Route, nil)
	routes := routeTree.Match(labelSet)

	seen := make(map[string]bool)
	var matched []string
	for _, route := range routes {
		recv := route.RouteOpts.Receiver
		if !seen[recv] {
			seen[recv] = true
			matched = append(matched, recv)
		}
		if !route.Continue {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchResponse{MatchedReceivers: matched})
}

func receiverTypes(r config.Receiver) []string {
	var types []string
	if len(r.DiscordConfigs) > 0 {
		types = append(types, "discord")
	}
	if len(r.EmailConfigs) > 0 {
		types = append(types, "email")
	}
	if len(r.IncidentioConfigs) > 0 {
		types = append(types, "incidentio")
	}
	if len(r.JiraConfigs) > 0 {
		types = append(types, "jira")
	}
	if len(r.MattermostConfigs) > 0 {
		types = append(types, "mattermost")
	}
	if len(r.MSTeamsConfigs) > 0 {
		types = append(types, "msteams")
	}
	if len(r.MSTeamsV2Configs) > 0 {
		types = append(types, "msteamsv2")
	}
	if len(r.OpsGenieConfigs) > 0 {
		types = append(types, "opsgenie")
	}
	if len(r.PagerdutyConfigs) > 0 {
		types = append(types, "pagerduty")
	}
	if len(r.PushoverConfigs) > 0 {
		types = append(types, "pushover")
	}
	if len(r.RocketchatConfigs) > 0 {
		types = append(types, "rocketchat")
	}
	if len(r.SlackConfigs) > 0 {
		types = append(types, "slack")
	}
	if len(r.SNSConfigs) > 0 {
		types = append(types, "sns")
	}
	if len(r.TelegramConfigs) > 0 {
		types = append(types, "telegram")
	}
	if len(r.VictorOpsConfigs) > 0 {
		types = append(types, "victorops")
	}
	if len(r.WebexConfigs) > 0 {
		types = append(types, "webex")
	}
	if len(r.WebhookConfigs) > 0 {
		types = append(types, "webhook")
	}
	if len(r.WechatConfigs) > 0 {
		types = append(types, "wechat")
	}
	return types
}

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		slog.Warn("cannot open browser automatically on this platform", "url", url)
		return
	}
	if err := exec.Command(cmd, append(args, url)...).Start(); err != nil {
		slog.Warn("failed to open browser", "error", err, "url", url)
	}
}
