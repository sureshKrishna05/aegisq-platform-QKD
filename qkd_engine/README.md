# AegisQ Quantum Validator Communication (QKD Integration)

This repository serves as a conceptual sub-project of the **[AegisQ Platform](https://github.com/sureshKrishna05/Aegisq-Platform)**. 

It demonstrates how **Quantum Key Distribution (QKD)** via the **BB84 Protocol** can be used to secure the underlying consensus network of a blockchain framework. It utilizes a **Hybrid Cryptographic Architecture**, combining post-quantum cryptography (ML-DSA-44 / Dilithium) for identity signatures in Go, alongside QKD for establishing perfectly secure communication channels between BFT validators.

---

## 1. How It Works

To secure the network against both classical and quantum adversaries:
1. **Network Transport Layer (Quantum Cryptography):** Validators execute a Python-based Qiskit simulation of the BB84 protocol to generate a symmetric AES-256 session key. To prevent partial knowledge extraction, the sifted bits undergo Privacy Amplification using SHA-256.
2. **Consensus Layer (Post-Quantum Cryptography):** Once the secure channel is established, the Go-based AegisQ framework encrypts BFT consensus votes (`PREPARE`, `COMMIT`) and transmits them. The votes themselves remain cryptographically signed by ML-DSA-44 to prove identity.

If an eavesdropper (Eve) attempts to intercept the quantum channel during the key exchange, the **Quantum Bit Error Rate (QBER)** spikes above the theoretical safety threshold of ~11%. The Go bridge detects this anomaly and aborts the node connection.

---

## 2. Prerequisites & Installation

The quantum simulation is built in Python using IBM's Qiskit, while the blockchain consensus is built in Go. 

### Step 1: Install Go Dependencies
Ensure you have Go installed, and build the main framework:
```bash
go mod tidy
```

### Step 2: Set Up the Python Virtual Environment
To prevent polluting your global Python system, we use a virtual environment for the Qiskit dependencies. Run the following commands from the project root:

```bash
# Move into the QKD engine directory
cd qkd_engine

# Create the virtual environment
python3 -m venv venv

# Activate the virtual environment
source venv/bin/activate

# Install Qiskit and required packages
pip install -r requirements.txt
```

---

## 3. Running the QKD Simulator (Python)

You can test the quantum key generation in isolation. Make sure your virtual environment is activated (`source qkd_engine/venv/bin/activate`).

### Ideal Channel Simulation
Run the BB84 protocol without noise or eavesdroppers to generate a perfect 256-bit AES key:
```bash
python qkd_engine/bb84_sim.py --bits 1024
```
**Expected Output:**
```json
{"sifted_key_length": 495, "qber": 0.0, "eavesdropper_detected": false, "symmetric_key_hex": "5103d325..."}
```

### Eavesdropper Simulation (Intercept-Resend Attack)
Trigger an active eavesdropper on the quantum channel:
```bash
python qkd_engine/bb84_sim.py --bits 1024 --eavesdrop
```
Because of the No-Cloning Theorem, Eve's measurements collapse the qubits, causing the QBER to spike to ~25%. The simulator will successfully flag `"eavesdropper_detected": true`.

---

## 4. Running the Full Blockchain Node (Go + Python)

To see the hybrid architecture in action, run the main AegisQ node simulation. The Go application will automatically spin up the Python simulator in the background to establish secure channels between 4 local validators before proposing a block.

```bash
# Make sure your Python venv path is accessible to Go, and liboqs is linked
export PATH="$(pwd)/qkd_engine/venv/bin:$PATH"
export LD_LIBRARY_PATH="/usr/local/lib64:$LD_LIBRARY_PATH"

# Run the AegisQ node
go run ./cmd/aegisqd
```

**Expected Console Output:**
You will see the node boot up, generate the quantum keys, and begin encrypting the consensus votes:
```text
Validators initialized.
Initializing QKD Engine (BB84)...
Establishing Quantum Session Key for Validator 1...
...
[QKD NETWORK] Encrypted PREPARE vote from Validator 1. Ciphertext: afefcf12...
[QKD NETWORK] Encrypted COMMIT vote from Validator 1. Ciphertext: 5c44fe94...
...
Block committed at height: 1
```
