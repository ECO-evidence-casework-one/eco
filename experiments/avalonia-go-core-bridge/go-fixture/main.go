package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	protocolVersion = 1
	maxFrameBytes   = 8 * 1024 * 1024
)

type request struct {
	ProtocolVersion int    `json:"protocol_version"`
	RequestID       string `json:"request_id"`
	Kind            string `json:"kind"`
	MatterID        string `json:"matter_id,omitempty"`
}

type response struct {
	ProtocolVersion int         `json:"protocol_version"`
	RequestID       string      `json:"request_id"`
	Status          string      `json:"status"`
	ErrorCode       string      `json:"error_code,omitempty"`
	UserMessage     string      `json:"user_message,omitempty"`
	Projection      *projection `json:"projection,omitempty"`
}

type projection struct {
	MatterID    string   `json:"matter_id"`
	Revision    string   `json:"revision"`
	GeneratedAt string   `json:"generated_at"`
	Identity    identity `json:"identity"`
	Evidence    evidence `json:"evidence"`
}

type identity struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

type evidence struct {
	Records    int `json:"records"`
	Readable   int `json:"readable"`
	Unresolved int `json:"unresolved"`
}

func readFrame(r io.Reader, dst any) error {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	sz := int(binary.BigEndian.Uint32(header[:]))
	if sz <= 0 || sz > maxFrameBytes {
		return fmt.Errorf("frame size %d outside allowed range", sz)
	}
	payload := make([]byte, sz)
	if _, err := io.ReadFull(r, payload); err != nil {
		return err
	}
	return json.Unmarshal(payload, dst)
}

func writeFrame(w io.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(payload) == 0 || len(payload) > maxFrameBytes {
		return fmt.Errorf("frame size %d outside allowed range", len(payload))
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if _, err := w.Write(header[:]); err != nil {
		return err
	}
	_, err = w.Write(payload)
	return err
}

func handle(req request) response {
	out := response{ProtocolVersion: protocolVersion, RequestID: req.RequestID, Status: "failed"}
	if req.ProtocolVersion != protocolVersion {
		out.ErrorCode = "protocol_version_mismatch"
		out.UserMessage = "The ECO interface and core use different protocol versions."
		return out
	}
	if req.RequestID == "" {
		out.ErrorCode = "missing_request_id"
		out.UserMessage = "The request could not be correlated safely."
		return out
	}
	switch req.Kind {
	case "ping":
		out.Status = "succeeded"
		out.UserMessage = "ECO core ready."
		return out
	case "project_matter":
		if req.MatterID != "MAT-SYNTHETIC" {
			out.ErrorCode = "projection_failed"
			out.UserMessage = "ECO could not refresh the Matter state."
			return out
		}
		out.Status = "succeeded"
		out.Projection = &projection{
			MatterID:    "MAT-SYNTHETIC",
			Revision:    "REV-SYNTH-1",
			GeneratedAt: "2026-09-03T14:00:00Z",
			Identity:    identity{Title: "Synthetic Bridge Matter", Status: "Open"},
			Evidence:    evidence{Records: 3, Readable: 2, Unresolved: 1},
		}
		return out
	default:
		out.ErrorCode = "unsupported_request"
		out.UserMessage = "This ECO core request is not supported."
		return out
	}
}

func main() {
	for {
		var req request
		err := readFrame(os.Stdin, &req)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				if os.Getenv("ECO_FIXTURE_IGNORE_EOF") == "1" {
					for {
						time.Sleep(time.Hour)
					}
				}
				return
			}
			fmt.Fprintln(os.Stderr, "fixture:", err)
			os.Exit(2)
		}
		if err := writeFrame(os.Stdout, handle(req)); err != nil {
			fmt.Fprintln(os.Stderr, "fixture:", err)
			os.Exit(3)
		}
	}
}
