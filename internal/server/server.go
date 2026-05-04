package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/caiooliveiraeti/database-deptree/internal/store"
)

type Server struct {
	st     store.Store
	webFS  fs.FS
	port   int
}

func New(st store.Store, webFS fs.FS, port int) *Server {
	return &Server{st: st, webFS: webFS, port: port}
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/graph", s.handleGraph)
	mux.HandleFunc("/api/meta", s.handleMeta)
	mux.HandleFunc("/api/traverse", s.handleTraverse)
	mux.HandleFunc("GET /api/nodes/search", s.handleNodeSearch)
	mux.HandleFunc("GET /api/nodes/{id}", s.handleNodeDetail)
	mux.Handle("/", http.FileServer(http.FS(s.webFS)))

	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{Addr: addr, Handler: mux}

	slog.Info("deptree UI ready", "url", fmt.Sprintf("http://localhost:%d", s.port))

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

// handleGraph returns nodes and edges, optionally filtered.
// Query params: labels=Entity,Table  rels=STORED_IN,DEPENDS_ON
func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	filter := store.GraphFilter{
		Labels:  splitParam(r.URL.Query().Get("labels")),
		Rels:    splitParam(r.URL.Query().Get("rels")),
		Systems: splitParam(r.URL.Query().Get("systems")),
	}

	data, err := s.st.QueryGraph(r.Context(), filter)
	if err != nil {
		slog.Error("graph query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

// handleMeta returns the distinct node labels and relationship types present in the graph.
func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	all, err := s.st.QueryGraph(r.Context(), store.GraphFilter{})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	labels := unique(func() []string {
		out := make([]string, len(all.Nodes))
		for i, n := range all.Nodes {
			out[i] = n.Label
		}
		return out
	}())

	rels := unique(func() []string {
		out := make([]string, len(all.Edges))
		for i, e := range all.Edges {
			out[i] = e.Rel
		}
		return out
	}())

	systems := unique(func() []string {
		out := []string{}
		for _, n := range all.Nodes {
			if n.System != "" {
				out = append(out, n.System)
			}
		}
		return out
	}())

	writeJSON(w, map[string]any{"labels": labels, "rels": rels, "systems": systems})
}

// handleTraverse returns the subgraph reachable from a node within a given depth.
// Query params: from (node ID), depth (1-5), direction (outgoing|incoming|both)
func (s *Server) handleTraverse(w http.ResponseWriter, r *http.Request) {
	q := store.TraversalQuery{
		StartNodeID: r.URL.Query().Get("from"),
		Direction:   r.URL.Query().Get("direction"),
	}
	if q.StartNodeID == "" {
		http.Error(w, "from is required", http.StatusBadRequest)
		return
	}
	if q.Direction == "" {
		q.Direction = "outgoing"
	}
	depth := 2
	if d := r.URL.Query().Get("depth"); d != "" {
		fmt.Sscanf(d, "%d", &depth)
	}
	q.Depth = depth

	data, err := s.st.QueryTraversal(r.Context(), q)
	if err != nil {
		slog.Error("traversal query failed", "from", q.StartNodeID, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, data)
}

// handleNodeSearch returns up to 20 nodes matching the search term (for autocomplete).
// Query params: q (search term)
func (s *Server) handleNodeSearch(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")
	if term == "" {
		writeJSON(w, []store.NodeData{})
		return
	}
	nodes, err := s.st.SearchNodes(r.Context(), term)
	if err != nil {
		slog.Error("node search failed", "term", term, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if nodes == nil {
		nodes = []store.NodeData{}
	}
	writeJSON(w, nodes)
}

func (s *Server) handleNodeDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := s.st.GetNode(r.Context(), id)
	if err != nil {
		slog.Error("get node failed", "id", id, "err", err)
		if strings.Contains(err.Error(), "node not found") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, detail)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("json encode failed", "err", err)
	}
}

func splitParam(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func unique(ss []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
