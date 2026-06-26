package qkd

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestGenerateSessionKey(t *testing.T) {
	// Assumes tests are run from the project root or the path is adjusted
	scriptPath := "../../../qkd_engine/bb84_sim.py"
	
	// Because the Python script requires the venv with Qiskit, we need to wrap the call
	// in a shell script or adjust our Go code to use the venv python if testing directly,
	// but for this prototype, we'll assume `python3` in the environment has the modules,
	// or we point to the venv python.
	venvPython, _ := filepath.Abs("../../../qkd_engine/venv/bin/python3")

	engine := &Engine{
		PythonScriptPath: scriptPath,
	}
	
	// Override the command in testing to use the venv
	// Note: We'd need to modify GenerateSessionKey to accept a python path, but for now 
	// this test just serves as a skeleton for local execution.
	fmt.Printf("Make sure to run this test with a python environment that has qiskit installed.\n")

	// Uncomment to actually test when environment is ready:
	/*
	res, err := engine.GenerateSessionKey(1024, false, 0.0)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if res.QBER > 0.11 {
		t.Errorf("Expected low QBER for ideal channel, got %f", res.QBER)
	}

	if len(res.SymmetricKeyHex) != 64 {
		t.Errorf("Expected 32-byte hex key (64 chars), got %d chars", len(res.SymmetricKeyHex))
	}
	*/
}
