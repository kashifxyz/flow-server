package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration/models"
	"github.com/kashifxyz/flow-server/internal/realtime"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

// presenceThrottle bounds how often one connection's cursor/selection churn
// is written to presence_states and rebroadcast — it is high-frequency,
// ephemeral state (PROJECT-INFO.md §98), not something every keystroke
// needs to persist.
const presenceThrottle = time.Second

// wsEnvelope is the JSON control-frame protocol. Binary frames carry raw Yjs
// update bytes and are never wrapped in this envelope — the server relays
// and persists them without decoding (SCHEMA.md §8).
type wsEnvelope struct {
	Type      string          `json:"type"`
	ClientID  string          `json:"client_id,omitempty"`
	Selection json.RawMessage `json:"selection,omitempty"`
	Cursor    json.RawMessage `json:"cursor,omitempty"`
	Viewport  json.RawMessage `json:"viewport,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	Snapshot  *string         `json:"snapshot,omitempty"`
	Updates   []string        `json:"updates,omitempty"`
}

// ServeWS upgrades to a WebSocket for one node: presence updates and Yjs
// document updates share the connection, distinguished by WS frame type
// (text JSON control messages vs. binary Yjs updates).
//
// Authorization happens before Accept: once upgraded there is no clean way
// to send an HTTP error, so access.RequireNodeMember must pass first
// (PROJECT-INFO.md §97).
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	userID := actor(r).ID
	workspaceID, err := access.RequireNodeMember(r.Context(), h.Svc.DB, userID, nodeID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.Origins})
	if err != nil {
		return // Accept already wrote the HTTP response for a failed handshake
	}
	conn := realtime.NewConn(wsConn)
	defer conn.Close()
	defer h.Hub.Forget(conn)

	ctx := r.Context()
	topic := "doc:" + nodeID.String()
	h.Hub.Subscribe(topic, conn)
	defer h.Hub.Unsubscribe(topic, conn)

	go conn.WritePump(ctx)

	if err := h.Svc.EnsureDocument(ctx, nodeID, workspaceID); err != nil {
		h.Log.Error().Err(err).Msg("ensure document")
		return
	}
	snapshot, updates, err := h.Svc.LoadSync(ctx, nodeID)
	if err != nil {
		h.Log.Error().Err(err).Msg("load document sync")
		return
	}
	conn.Send(realtime.Message{Kind: realtime.Text, Payload: encodeSync(snapshot, updates)})

	var clientID string
	var lastPresence time.Time
	for {
		typ, data, err := wsConn.Read(ctx)
		if err != nil {
			return
		}
		switch typ {
		case websocket.MessageBinary:
			if clientID == "" {
				continue // client must send {"type":"hello"} before streaming updates
			}
			if err := h.Svc.AppendUpdate(ctx, nodeID, workspaceID, userID, clientID, data); err != nil {
				h.Log.Error().Err(err).Msg("append document update")
				continue
			}
			h.Hub.Broadcast(topic, realtime.Message{Kind: realtime.Binary, Payload: data}, conn)

		case websocket.MessageText:
			var env wsEnvelope
			if err := json.Unmarshal(data, &env); err != nil {
				continue
			}
			switch env.Type {
			case "hello":
				clientID = env.ClientID
			case "presence":
				if time.Since(lastPresence) < presenceThrottle {
					continue
				}
				lastPresence = time.Now()
				if _, err := h.Svc.Put(ctx, nodeID, userID, models.PutPresenceRequest{
					Selection: env.Selection,
					Cursor:    env.Cursor,
					Viewport:  env.Viewport,
				}); err != nil {
					h.Log.Error().Err(err).Msg("put presence")
					continue
				}
				out, err := json.Marshal(wsEnvelope{
					Type: "presence", UserID: userID.String(),
					Selection: env.Selection, Cursor: env.Cursor, Viewport: env.Viewport,
				})
				if err != nil {
					continue
				}
				h.Hub.Broadcast(topic, realtime.Message{Kind: realtime.Text, Payload: out}, conn)
			}
		}
	}
}

func encodeSync(snapshot []byte, updates [][]byte) []byte {
	env := wsEnvelope{Type: "sync", Updates: make([]string, len(updates))}
	if snapshot != nil {
		s := base64.StdEncoding.EncodeToString(snapshot)
		env.Snapshot = &s
	}
	for i, u := range updates {
		env.Updates[i] = base64.StdEncoding.EncodeToString(u)
	}
	b, _ := json.Marshal(env)
	return b
}
