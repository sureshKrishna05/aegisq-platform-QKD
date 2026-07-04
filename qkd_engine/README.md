# AegisQ Framework: Quantum-Secured Validator Communication (QKD Integration)

## 1. Abstract
The **AegisQ Framework** is a deterministic trust-preservation infrastructure designed for digital governance. In this conceptual research integration, we introduce **Quantum Key Distribution (QKD)** via the **BB84 Protocol** to secure the underlying consensus network. This repository demonstrates a highly advanced **Hybrid Cryptographic Architecture**, utilizing post-quantum cryptography (ML-DSA-44) for identity and block signatures alongside QKD for establishing perfectly secure communication channels between Byzantine Fault Tolerant (BFT) validators.

## 2. Hybrid Cryptographic Architecture
To secure a decentralized ledger against both classical and quantum adversaries, AegisQ splits cryptographic responsibilities across two layers:

1. **Identity & Consensus Layer (Post-Quantum Cryptography):**
   - Implemented in **Go** using `liboqs` (Open Quantum Safe).
   - **Algorithm:** ML-DSA-44 (Dilithium2).
   - **Purpose:** Validators use Dilithium to sign block proposals and consensus votes. This ensures cryptographic non-repudiation and prevents malicious nodes from forging votes.

2. **Network Transport Layer (Quantum Cryptography):**
   - Implemented in **Python** using IBM's **Qiskit** framework.
   - **Protocol:** BB84 Quantum Key Distribution.
   - **Purpose:** Validators establish an ephemeral, symmetric AES-256-GCM session key to encrypt the actual transmission of the consensus packets over the public internet, completely neutralizing harvest-now-decrypt-later (HNDL) attacks.

---

## 3. The BB84 Qiskit Simulation (`bb84_sim.py`)
The QKD engine is built entirely on IBM Qiskit's `AerSimulator`. It models the physical transmission of qubits over a quantum channel.

### 3.1 Protocol Execution
1. **State Preparation (Alice):** The sender generates a random string of classical bits and a random string of bases ($Z$-basis or $X$-basis). Qubits are encoded using `qc.x()` and `qc.h()` gates accordingly.
2. **Channel Noise:** The simulator injects random bit-flip errors to emulate decoherence in fiber-optic transmission.
3. **Measurement (Bob):** The receiver measures the qubits in a randomly chosen basis.
4. **Sifting:** Alice and Bob publicly share their measurement bases (via the classical channel) and discard any bits where their bases mismatched.

### 3.2 Eavesdropper Detection (Intercept-Resend)
If a malicious actor (Eve) attempts to intercept the qubits, she must measure them. According to the **No-Cloning Theorem**, this measurement collapses the quantum state. 
- When Eve resends the qubits to Bob, she introduces a statistically significant number of errors.
- The simulator calculates the **Quantum Bit Error Rate (QBER)**. If the QBER exceeds the theoretical safety threshold of **~11%**, the Go bridge detects the eavesdropper, aborts the handshake, and raises a `SECURITY ALERT`.

---

## 4. The Go Bridge & Network Integration
Because the core AegisQ Framework is built in Go for deterministic execution and high throughput, the Python Qiskit simulation acts as an external hardware module.

### 4.1 The Bridge (`core/network/qkd.go`)
- **Execution:** Go spins up the Qiskit engine via a subprocess.
- **Parsing:** It extracts the sifted key and the QBER from the JSON stdout.
- **Enforcement:** If `eavesdropper_detected` is `true`, the node refuses to boot.

### 4.2 The Secure Channel (`core/network/secure_channel.go`)
- The sifted bits are condensed into a 32-byte string.
- This string initializes a standard **AES-256-GCM** cipher block.
- **AQX Serialization:** AegisQ uses a custom deterministic binary serialization format (AQX) to guarantee identical hashing across nodes.
- When a validator votes (`PREPARE` or `COMMIT`), the raw AQX binary payload is passed into the `SecureChannel`, encrypted, and transmitted.

---

## 5. Execution Flow Example
When the `aegisqd` node daemon starts, the following sequence occurs:

1. **Validators initialized.** (Dilithium keypairs generated).
2. **QKD Handshake:** The leader runs BB84 with Validator 2, 3, and 4.
3. **Session Keys Established:** 256-bit symmetric keys are loaded into memory.
4. **Block Proposal:** Leader generates synthetic storage transactions and hashes the block.
5. **Encrypted BFT:**
   - Validator 1 votes `PREPARE`. The vote is serialized to AQX.
   - The AQX payload is encrypted with AES-256 using the QKD key.
   - The ciphertext is sent to the VotePool.
   - The VotePool decrypts the ciphertext, deserializes the vote, and verifies the ML-DSA-44 signature.
6. **Block Finalized:** Quorum is reached and the block is saved to PebbleDB.

## 6. Future Work
- Implementation of **E91** or **B92** protocols in Qiskit.
- Applying formal **Privacy Amplification** and **Information Reconciliation** (e.g., Cascade algorithm) to the sifted keys.
- Deploying the Qiskit circuits on real IBM Quantum Hardware via the IBM Quantum Runtime.
