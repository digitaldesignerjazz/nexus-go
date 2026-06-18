package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// NexusGoVersion is the current version of the orchestrator
const NexusGoVersion = "0.1.0"

// systemReady indicates whether the Nexus orchestrator has successfully completed startup
var (
	systemReady     bool
	systemReadyLock sync.RWMutex
	startTime       = time.Now()
)

func setSystemReady(ready bool) {
	systemReadyLock.Lock()
	systemReady = ready
	systemReadyLock.Unlock()
}

func isSystemReady() bool {
	systemReadyLock.RLock()
	defer systemReadyLock.RUnlock()
	return systemReady
}

func getUptime() time.Duration {
	return time.Since(startTime)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		printVersion()
	case "doctor":
		runDoctor()
	case "start":
		runStart()
	case "serve":
		runServe()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`nexus-go - Go-based Nexus Ecosystem Orchestrator

Usage:
  nexus-go <command> [flags]

Commands:
  help      Show this help message
  version   Print version information
  doctor    Check environment prerequisites and Nexus readiness
  start     Orchestrate startup of Nexus components (mesh, blockchain, ai, prototypes)
  serve     Start HTTP server with health and readiness endpoints

Flags (for start):
  --component string   Components to start (default "all")
  --dry-run bool       Preview commands without executing (default true)
  --execute            Perform live execution (marks system as ready)
  --force              Skip some safety prompts

Flags (for serve):
  --port string        Port to listen on (default "8080")

Examples:
  nexus-go doctor
  nexus-go start --component=all --execute
  nexus-go serve
  nexus-go serve --port=9090

For full documentation see README.md
`)
}

func printVersion() {
	fmt.Printf("nexus-go version %s\n", NexusGoVersion)
	fmt.Printf("Go version: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Println("Part of Esslinger & Co. Nexus Ecosystem - June 2026")
}

// runServe starts a simple HTTP server with health and readiness endpoints
func runServe() {
	port := "8080"

	for i := 2; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "--port=") {
			port = strings.TrimPrefix(os.Args[i], "--port=")
		}
	}

	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/readyz", readyHandler)
	http.HandleFunc("/ready", readyHandler)

	fmt.Printf("Starting nexus-go HTTP server on :%s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  /healthz , /health   - Liveness probe (always returns 200 if server is running)")
	fmt.Println("  /readyz  , /ready    - Readiness probe (reflects successful start --execute)")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	uptime := getUptime().Round(time.Second)

	fmt.Fprintf(w, `{
  "status": "ok",
  "service": "nexus-go",
  "version": "%s",
  "uptime": "%s",
  "time": "%s"
}
`, NexusGoVersion, uptime, time.Now().UTC().Format(time.RFC3339))
}

// readyHandler returns 200 if system is ready, 503 otherwise
func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if isSystemReady() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"nexus-go","version":"%s","time":"%s"}\n`,
			NexusGoVersion, time.Now().UTC().Format(time.RFC3339))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","service":"nexus-go","version":"%s","time":"%s"}\n`,
			NexusGoVersion, time.Now().UTC().Format(time.RFC3339))
	}
}

func runDoctor() {
	fmt.Println("=== Nexus Go Doctor - Environment & Readiness Check ===")
	fmt.Println("Timestamp:", time.Now().Format(time.RFC3339))
	fmt.Printf("Platform: %s/%s | Go: %s\n\n", runtime.GOOS, runtime.GOARCH, runtime.Version())

	checks := []struct {
		name string
		cmd  string
		args []string
	}{
		{"Go toolchain", "go", []string{"version"}},
		{"Docker", "docker", []string{"--version"}},
		{"Docker Compose", "docker", []string{"compose", "version"}},
		{"Yggdrasil", "yggdrasil", []string{"--version"}},
		{"Git", "git", []string{"--version"}},
	}

	for _, c := range checks {
		fmt.Printf("Checking %s... ", c.name)
		cmd := exec.Command(c.cmd, c.args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("MISSING or ERROR\n")
			if c.name == "Yggdrasil" {
				fmt.Println("  -> Install: https://yggdrasil-network.github.io/ or via package manager")
			}
			continue
		}
		fmt.Printf("OK\n  %s", strings.TrimSpace(string(output)))
		fmt.Println()
	}

	fmt.Println("\n=== Nexus-Specific Checks ===")
	fmt.Println("- Yggdrasil config: (future) check ~/.config/yggdrasil.conf or systemd unit")
	fmt.Println("- Mesh peer bootstrap list: (future) validate diversity and uptime")
	fmt.Println("- Blockchain node: (future) XCoin/QCoin binary or Docker image present")
	fmt.Println("- AI/Prototype hooks: (future) Grok Launcher path, sensor drivers")

	fmt.Println("\nDoctor complete. Address any MISSING items before full start.")
	fmt.Println("Next: nexus-go start --component=all  (review dry-run output first)")
}

func runStart() {
	component := "all"
	dryRun := true
	execute := false
	force := false
	verbose := false

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "--component="):
			component = strings.TrimPrefix(arg, "--component=")
		case arg == "--dry-run":
			dryRun = true
		case arg == "--execute":
			dryRun = false
			execute = true
		case arg == "--force":
			force = true
		case arg == "--verbose":
			verbose = true
	}

	if execute {
		dryRun = false
	}

	fmt.Printf("=== Nexus Go Start Orchestrator ===\n")
	fmt.Printf("Component(s): %s | Dry-run: %v | Force: %v | Verbose: %v\n\n", component, dryRun, force, verbose)

	if dryRun {
		fmt.Println("[DRY-RUN MODE] No changes will be made. Reviewing planned actions...\n")
	}

	switch component {
	case "all", "mesh":
		startMesh(dryRun, verbose)
		fallthrough
	case "blockchain":
		if component == "all" || component == "blockchain" {
			startBlockchain(dryRun, verbose)
		}
	case "ai", "ai-swarm":
		startAISwarm(dryRun, verbose)
	case "prototypes":
		startPrototypes(dryRun, verbose)
	case "grok", "grok-launcher":
		startGrokLauncher(dryRun, verbose)
	default:
		fmt.Printf("Unknown component: %s. Valid: all, mesh, blockchain, ai, prototypes, grok\n", component)
		return
	}

	if component == "all" {
		startBlockchain(dryRun, verbose)
		startAISwarm(dryRun, verbose)
		startPrototypes(dryRun, verbose)
		startGrokLauncher(dryRun, verbose)
	}

	fmt.Println("\n=== Startup orchestration complete ===")

	if execute {
		setSystemReady(true)
		fmt.Println("[READY] System marked as ready for traffic (readiness probes will now pass).")
	} else {
		fmt.Println("[INFO] Dry-run completed. Use --execute to mark system as ready.")
	}

	if dryRun {
		fmt.Println("Review the above commands and implications carefully.")
		fmt.Println("Re-run with --execute --force to perform live actions (after backups and review).")
	}
}

func startMesh(dryRun, verbose bool) {
	fmt.Println("--- [1/5] MESH LAYER: xMesh / NovaNet / QNET / Yggdrasil ---")
	fmt.Println("Purpose: Self-healing IPv6 overlay for all Nexus traffic (blockchain gossip, AI messages, prototype telemetry). Privacy-respecting, censorship resistant.")
	fmt.Println("Implications: Foundation for global reach from Hannover base; enables decentralized operation beyond single jurisdiction.")

	commands := []string{
		"# Generate fresh Yggdrasil config (if none exists)",
		"yggdrasil -genconf > ~/.config/yggdrasil.conf",
		"# Review/edit config: set Peers to diverse bootstrap nodes, enable TunnelRouting if needed",
		"# Start Yggdrasil (user service or foreground)",
		"systemctl --user enable --now yggdrasil || yggdrasil -useconffile ~/.config/yggdrasil.conf",
		"# Docker networking note: use host or macvlan for Yggdrasil to access real interfaces",
		"# Optional: Tor/I2P for sensitive subsets - route select traffic through privacy layers",
	}

	for _, cmd := range commands {
		fmt.Println("  ", cmd)
	}

	if !dryRun {
		fmt.Println("  [LIVE] Executing mesh startup sequence... (placeholder - implement exec in v0.2)")
	}

	fmt.Println("Edge cases: Partition risk if <3 good peers; Tenda Nova WiFi tuning (channel, TX power vs privacy leak); config drift on restarts.")
	fmt.Println("Recovery: Re-bootstrap from known good peers list; monitor with yggdrasilctl or future Go status command.")
	fmt.Println()
}

func startBlockchain(dryRun, verbose bool) {
	fmt.Println("--- [2/5] BLOCKCHAIN LAYER: XCoin / QCoin / QNET runes ---")
	fmt.Println("Purpose: Economic incentives for mesh relaying & AI compute; on-chain governance; Wizard Q rune scripting; oracle data from prototypes.")
	fmt.Println("Implications: Aligns participation with value; Delaware C-Corp structure for liability protection & scaling; regulatory considerations (MiCA, tax on token movements).")

	commands := []string{
		"# Ensure QCoin/XCoin node or light client binary / Docker image available",
		"# docker pull nexus/qcoin-node:latest || ./qcoin-node --config ~/.qcoin/config.toml --sync",
		"# Start node (sync may take time on first run)",
		"./qcoin-node start --config ~/.qcoin/config.toml",
		"# Future: Go-native light client or RPC client for balance, rune ops, arbitrage",
		"# Wizard Q rune deployment for custom automation/incentives",
	}

	for _, cmd := range commands {
		fmt.Println("  ", cmd)
	}

	if !dryRun {
		fmt.Println("  [LIVE] Blockchain startup placeholder...")
	}

	fmt.Println("Edge cases: Chain reorgs (design agents tolerant); token volatility (hedging or stable mechanisms); key management (never commit seed phrases).")
	fmt.Println("Nuance: Tokenomics reward mesh uptime + data quality from Soilnova etc.; arbitrage opportunities across QNET.")
	fmt.Println()
}

func startAISwarm(dryRun, verbose bool) {
	fmt.Println("--- [3/5] AI AGENT SWARM LAYER: Self-improving emotional AI (Ara etc.) ---")
	fmt.Println("Purpose: Autonomous task execution, recursive self-improvement, emotional intelligence, decentralized decision-making over mesh.")
	fmt.Println("Implications: Reduces human ops load; enables 24/7 global coordination; prompt hygiene critical to prevent drift/hallucination cascades in swarms.")

	commands := []string{
		"# Launch Grok Launcher (Rust + egui) for local prototyping/UI if available",
		"# cd ../grok-launcher && cargo run --release  (or prebuilt binary)",
		"# Start agent swarm coordinator (future: Go runtime or call external Python/Rust agents)",
		"# Example: ./agent-swarm --mesh ygg0 --blockchain qcoin-rpc --emotional-model ara",
		"# Heartbeat monitors + state sync via Yggdrasil pub/sub or QNET",
	}

	for _, cmd := range commands {
		fmt.Println("  ", cmd)
	}

	if !dryRun {
		fmt.Println("  [LIVE] AI swarm startup placeholder...")
	}

	fmt.Println("Edge cases: Agent drift, emotional model instability, prompt injection via mesh; self-improvement proposals require human or on-chain approval gate.")
	fmt.Println("Nuance: Go concurrency ideal for parallel agent monitors and gossip handlers.")
	fmt.Println()
}

func startPrototypes(dryRun, verbose bool) {
	fmt.Println("--- [4/5] PROTOTYPES LAYER: Soilnova, Vista Nova, York Autotype, Lumia ---")
	fmt.Println("Purpose: Physical/digital grounding — sensors (soil/env), visualization, automation, adaptive lighting. Feed trusted oracles into blockchain/AI.")
	fmt.Println("Implications: Closes the loop from digital orchestration to real-world action & data; hardware failures require robust recovery in orchestrator.")

	commands := []string{
		"# Soilnova: I2C/SPI sensors on Raspberry/Arduino - calibrate, read, sign, publish to mesh",
		"# go run cmd/soilnova/main.go  (future native Go driver)",
		"# Vista Nova / Lumia: Mesh-controlled scenes, low-power modes, LED/ display drivers",
		"# York Autotype: Workflow automation triggered by on-chain events or agent decisions",
		"# Docker Compose for sensor gateways if applicable: docker compose up -d soilnova-gateway",
	}

	for _, cmd := range commands {
		fmt.Println("  ", cmd)
	}

	if !dryRun {
		fmt.Println("  [LIVE] Prototypes startup placeholder...")
	}

	fmt.Println("Edge cases: Sensor drift/calibration loss, power outage (last-known-good state recovery), hardware supply chain (verify firmware).")
	fmt.Println("Nuance: Go syscall or pure Go drivers (e.g. periph.io) for direct hardware access without heavy deps.")
	fmt.Println()
}

func startGrokLauncher(dryRun, verbose bool) {
	fmt.Println("--- [5/5] GROK LAUNCHER (Rust + egui) ---")
	fmt.Println("Purpose: Local AI prototyping UI, inference sandbox, integration point for Nexus agent experiments and creative tools.")
	fmt.Println("Implications: Bridges xAI Grok capabilities with local sovereign stack; supports rapid iteration on agent behaviors before mesh deployment.")

	commands := []string{
		"# Build/Run Grok Launcher (see its repo: rust + egui)",
		"cd ../grok-launcher && cargo build --release",
		"./target/release/grok-launcher",
		"# Or Dockerized if available. Connects to local mesh for swarm testing.",
	}

	for _, cmd := range commands {
		fmt.Println("  ", cmd)
	}

	if !dryRun {
		fmt.Println("  [LIVE] Grok Launcher startup placeholder...")
	}

	fmt.Println("Edge cases: Rust toolchain missing (install via rustup); egui GPU deps on some platforms; sync state with Go orchestrator via files/IPC.")
	fmt.Println()
}
