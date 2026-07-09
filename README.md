<div align="center">

# 🛡️ AegisQ Quantum Validator Communication
### Information-Theoretically Secure Post-Quantum Blockchain Infrastructure

*Hybrid Cryptography • Qiskit Hardware Integration • ML-DSA-44 • BB84 Protocol*

</div>

---

## 1. Executive Summary

**AegisQ-QKD** is an advanced conceptual research project derived from the **[AegisQ Platform](https://github.com/sureshKrishna05/Aegisq-Platform)**. It serves as a proof-of-concept for the future of ultra-secure digital trust architectures, combining **Post-Quantum Cryptography (PQC)** and **Quantum Key Distribution (QKD)** to protect decentralized validator networks against both classical and quantum adversaries. 

This repository directly integrates the Go-based blockchain consensus layer with **real IBM Quantum Hardware** via `qiskit-ibm-runtime` to guarantee theoretically impenetrable network transport security.

---

## 2. Hybrid Cryptographic Architecture

To secure a decentralized ledger against Harvest-Now-Decrypt-Later (HNDL) attacks and forged identities, AegisQ splits cryptographic responsibilities across two distinct security layers:

1. **Identity & Consensus Layer (Post-Quantum Cryptography):**
   - Implemented natively in **Go** using `liboqs` (Open Quantum Safe).
   - **Algorithm:** ML-DSA-44 (Dilithium2).
   - **Purpose:** Validators use Dilithium to cryptographically sign block proposals and consensus votes. This prevents malicious nodes from forging votes and ensures non-repudiation.

2. **Network Transport Layer (Quantum Cryptography):**
   - Implemented in **Python** using IBM's **Qiskit** framework and executed on physical superconducting hardware.
   - **Protocol:** BB84 Quantum Key Distribution.
   - **Purpose:** Validators establish an ephemeral, symmetric AES-256-GCM session key to encrypt the transmission of consensus packets over the public internet, completely neutralizing passive interception.

---

## 3. IBM Quantum Hardware Integration (`bb84_hardware.py`)

The QKD engine utilizes `qiskit-ibm-runtime` to submit batched quantum circuits to physical QPU backends (e.g., `ibm_marrakesh`). 

### Protocol Execution Pipeline
1. **Circuit Generation:** 1,024 1-qubit BB84 circuits are built, encoding random classical bits into random bases using $X$ and $H$ gates.
2. **Hardware Transpilation:** Circuits are transpiled into physical microwave pulses specific to the selected IBM hardware topology.
3. **Batch Execution:** To bypass excessive cloud queue times, all 1,024 circuits are bundled into a single batch job using the `SamplerV2` primitive.
4. **Information Reconciliation & Privacy Amplification:** Physical decoherence and thermal relaxation naturally induce a 2-8% Quantum Bit Error Rate (QBER). After sifting the keys, the raw bits are fed through a universal cryptographic hash function (SHA-256) to perform **Privacy Amplification**, mathematically distilling a perfectly secure 256-bit AES key.
5. **Eavesdropper Detection:** The No-Cloning Theorem dictates that any active interception forces a wave function collapse. If the QBER spikes past the 11% theoretical safety threshold, the node immediately aborts the connection.

---

## 4. Prerequisites & Installation

The quantum execution engine is built in Python, while the core blockchain infrastructure is built in Go.

### Step 1: Install Go Dependencies
Ensure you have Go installed, and build the main framework:
```bash
go mod tidy
```

### Step 2: Set Up the Python Virtual Environment
To prevent polluting your global Python environment and to ensure `qiskit-ibm-runtime` functions correctly, use a virtual environment:

```bash
# Move into the QKD engine directory
cd qkd_engine

# Create the virtual environment
python3 -m venv venv

# Activate the virtual environment
source venv/bin/activate

# Install Qiskit and IBM Runtime requirements
pip install -r requirements.txt
```

---

## 5. Running the Blockchain Node on Real Quantum Hardware

To execute the hybrid architecture, run the main AegisQ node. The Go daemon will automatically authenticate with the IBM Quantum cloud, select the least busy physical backend, and submit the batch jobs to entangle the qubits for the 4 local validators.

```bash
# Export your IBM Quantum API key
export IBM_QUANTUM_TOKEN="YOUR_API_KEY_HERE"

# Ensure the Go daemon can access the Qiskit virtual environment and liboqs
export PATH="$(pwd)/qkd_engine/venv/bin:$PATH"
export LD_LIBRARY_PATH="/usr/local/lib64:$LD_LIBRARY_PATH"

# Run the AegisQ node
go run ./cmd/aegisqd
```

> [!NOTE]  
> Because this submits jobs to an actual physical quantum computer, you will be placed in the IBM cloud queue. Establishing the 4 session keys may take anywhere from 10 minutes to an hour of wall-clock time depending on global traffic on the IBM Open Plan.

### Expected Physical Hardware Output
The following is an authentic execution log demonstrating the successful integration between the Go BFT consensus engine and an IBM Quantum physical QPU:

```text
rubberducky@fedora:~/projects/go_project/aegisq-platform-QKD$ go run ./cmd/aegisqd
Validators initialized.
Initializing QKD Engine (IBM Quantum Hardware)...
Establishing Quantum Session Key for Validator 1...
Establishing Quantum Session Key for Validator 2...
Establishing Quantum Session Key for Validator 3...
Establishing Quantum Session Key for Validator 4...
All secure channels established.
2026/07/10 03:14:04 [JOB 1] WAL file aegisq.db/000004.log with log number 000004 stopped reading at offset: 0; replayed 0 keys in 0 batches
Running Crash Recovery & Integrity Verification...
Database integrity verified.
Restored height: 1
Leader selected: validator-3
Generated synthetic storage transactions: 10000
Transaction generation time: 688.916438ms
Block finalize time: 12.462706ms
Proposed block height: 2
[QKD NETWORK] Encrypted PREPARE vote from Validator 1. Ciphertext (first 16 bytes): ebba6f7c9ad093143e9b26d34bb3cc38...
[QKD NETWORK] Encrypted PREPARE vote from Validator 2. Ciphertext (first 16 bytes): 49ed1aedb12eaaf54aff9d184baf5a19...
[QKD NETWORK] Encrypted PREPARE vote from Validator 3. Ciphertext (first 16 bytes): a45fdf02c750d396c8017ebf523eddb6...
[QKD NETWORK] Encrypted PREPARE vote from Validator 4. Ciphertext (first 16 bytes): a7f98ca0d2c3f196577e78cd0a91e4d3...
Prepare quorum reached.
[QKD NETWORK] Encrypted COMMIT vote from Validator 1. Ciphertext (first 16 bytes): 74e9b875067d88769d28abb470ec3858...
[QKD NETWORK] Encrypted COMMIT vote from Validator 2. Ciphertext (first 16 bytes): 0785c778d1ed323af2454c227539c335...
[QKD NETWORK] Encrypted COMMIT vote from Validator 3. Ciphertext (first 16 bytes): e63294d8a20baf82e7395768b19f8918...
[QKD NETWORK] Encrypted COMMIT vote from Validator 4. Ciphertext (first 16 bytes): edc48c56eb554e73955f661013477e89...
Commit quorum reached.
⚡ [EVENT BUS] BlockPersisted published by Storage | Payload: 2
Block committed at height: 2

========= BLOCK SUMMARY =========
Height: 2
Hash: 4b94607f3d41c0d9bcefb1f2de65650d45e8c9d7939e9dbd85fae1554850ccf6
Previous: c5e80e7a7b52dd7e78f23e2d20705b4a2291e69b654b93ddab27e856a7a0b5b5
Total Transactions: 10000
  Tx 1
   Sender: validator-3
   DataHash: [113 30 218 81 115 45 158 204 195 4 92 56 180 132 221 210 20 91 243 234 181 154 153 159 37 2 58 212 5 115 73 138]
  Tx 2
   Sender: validator-3
   DataHash: [250 202 28 61 28 82 87 158 252 171 7 180 141 71 116 191 192 255 16 158 69 60 223 26 165 133 130 71 80 201 63 216]
  Tx 3
   Sender: validator-3
   DataHash: [219 226 49 208 150 70 105 213 17 233 20 164 160 135 207 97 245 225 243 8 80 7 27 26 134 216 176 179 200 160 129 198]
  Tx 4
   Sender: validator-3
   DataHash: [108 117 230 21 186 186 40 121 60 139 151 67 229 73 97 107 138 183 168 106 71 51 184 207 80 116 76 234 167 64 137 234]
  Tx 5
   Sender: validator-3
   DataHash: [66 118 253 218 46 56 210 113 42 246 246 180 43 5 251 186 116 57 127 102 200 9 3 210 127 228 6 122 107 186 245 22]
  ...
API server running on http://localhost:8080
```

---

## 6. Future Scope
- Implementation of the **E91** entanglement-based protocol.
- True **Information Reconciliation** using the Cascade algorithm to handle elevated physical noise natively without key truncation.
- Direct integration into the parent AegisQ Platform governance engine.
