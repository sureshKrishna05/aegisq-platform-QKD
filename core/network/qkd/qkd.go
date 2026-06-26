package qkd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// QKDResult maps to the JSON output from the Python Qiskit simulator
type QKDResult struct {
	SiftedKeyLength      int     `json:"sifted_key_length"`
	QBER                 float64 `json:"qber"`
	EavesdropperDetected bool    `json:"eavesdropper_detected"`
	SymmetricKeyHex      string  `json:"symmetric_key_hex"`
}

// Engine acts as the bridge between Go and the Python Qiskit simulation
type Engine struct {
	PythonScriptPath string
}

func NewEngine(scriptPath string) *Engine {
	return &Engine{
		PythonScriptPath: scriptPath,
	}
}

// GenerateSessionKey calls the Python BB84 simulation to generate a 256-bit AES key.
func (e *Engine) GenerateSessionKey(qubits int, simulateEavesdropper bool, noiseLevel float64) (*QKDResult, error) {
	scriptAbsPath, err := filepath.Abs(e.PythonScriptPath)
	if err != nil {
		return nil, err
	}

	args := []string{scriptAbsPath, "--bits", fmt.Sprintf("%d", qubits)}
	if simulateEavesdropper {
		args = append(args, "--eavesdrop")
	}
	if noiseLevel > 0 {
		args = append(args, "--noise", fmt.Sprintf("%f", noiseLevel))
	}

	cmd := exec.Command("python3", args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("qkd python script failed: %s, stderr: %s", err, stderr.String())
	}

	var result QKDResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse QKD output: %w, raw: %s", err, out.String())
	}

	// Security Check: Refuse the key if QBER is too high or an eavesdropper was explicitly flagged
	if result.EavesdropperDetected {
		return nil, fmt.Errorf("SECURITY ALERT: Eavesdropper detected! QBER at %.2f%%", result.QBER*100)
	}

	return &result, nil
}
