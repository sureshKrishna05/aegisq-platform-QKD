<div align="center">

# 🛡️ AegisQKD

An experimental framework demonstrating the integration of Post-Quantum Cryptography (PQC) and Quantum Key Distribution (QKD) into a distributed blockchain architecture using Go, Qiskit, and IBM Quantum.

</div>

> [!IMPORTANT]
> AegisQKD is an experimental research framework intended for education, prototyping, and research. It is **not** intended for production deployment or real-world cryptographic infrastructure.

*AegisQKD is an open-source framework that demonstrates how quantum key distribution protocols implemented with Qiskit can be integrated into modern distributed systems. The project aims to serve as an educational and research reference for developers exploring the intersection of quantum computing, post-quantum cryptography, and secure networked applications.*

## Research Scope

AegisQKD is intended as an educational and experimental framework for exploring the integration of Quantum Key Distribution (QKD) and Post-Quantum Cryptography (PQC) into distributed systems.

The project demonstrates:
- Integration of Qiskit with Go-based applications.
- BB84-based quantum key establishment using Qiskit Aer and IBM Quantum.
- Secure communication concepts within a PBFT-style consensus architecture.

It does not claim to provide production-ready quantum networking or unconditional security.

---

## Architecture Overview

```text
                   +------------------------+
                   |     Validator Node     |
                   +-----------+------------+
                               |
                 ML-DSA Authentication
                               |
                               v
                 +--------------------------+
                 | Secure Channel Manager   |
                 +------------+-------------+
                              |
                 Quantum Key Establishment
                              |
          +-------------------+------------------+
          |                                      |
          v                                      v
  Aer Simulator                         IBM Quantum Hardware
          |                                      |
          +-------------------+------------------+
                              |
                     256-bit Session Key
                              |
                              v
                   AES-256-GCM Transport
                              |
                              v
                 PBFT Consensus Messages
```

## Features

- **BB84 Quantum Key Distribution** (via Qiskit)
- **IBM Quantum Runtime** integration
- **Aer simulation** support
- **ML-DSA-44** (Dilithium2) authentication
- **AES-256-GCM** secure transport layer
- **AQX binary serialization**
- **PBFT validator communication**
- **Go ↔ Python interoperability**
- **Modular backend architecture**

---

## Why Go?

While Qiskit is natively implemented in Python, distributed systems (like blockchain nodes and consensus engines) often require high-performance, concurrent network services. 

AegisQKD demonstrates how robust Go services can invoke complex Qiskit workloads using a language-agnostic interface, effectively keeping the quantum cryptographic layer strictly isolated from the high-throughput PBFT consensus engine.

---

## Supported Backends

AegisQKD is designed to be modular. It currently supports:

- [x] **Local Simulation:** (via Qiskit Aer) for rapid prototyping and testing without queue times.
- [x] **IBM Quantum Hardware:** (via `qiskit-ibm-runtime`) for true physical entanglement on superconducting QPUs.

*Future Backends:*
- [ ] Other Qiskit-compatible providers (IonQ, Quantinuum, etc.)
- [ ] Additional localized simulators

---

## What is AQX?

**AQX (AegisQ Exchange format)** is a custom, deterministic binary serialization protocol built specifically for this framework. Traditional serialization methods (like JSON or Protobuf) can produce varying byte arrays for the exact same data structure depending on map ordering or language implementations. This breaks cryptographic hashing. AQX guarantees that a transaction or consensus vote will *always* serialize into the exact same byte slice across every node, ensuring PBFT block hashes are perfectly consistent across the decentralized network before they are encrypted by the QKD AES keys.

---

## Repository Structure

```text
core/
    consensus/     # PBFT consensus engine logic and block finalization
    crypto/        # ML-DSA-44 post-quantum implementations
    network/qkd/   # Secure transport layer and Go-to-Python bridge
    storage/       # PebbleDB persistent ledger logic

qkd_engine/
    bb84_sim.py       # BB84 implementation using Aer simulator
    bb84_hardware.py  # BB84 implementation using IBM Quantum Runtime

docs/
    benchmark.md      # Full interactive execution logs and traces
```

---

## Performance Benchmarks

The hybrid architecture yields highly performant local execution while offloading key generation to the cloud QPUs.

| Component | Result |
|-----------|---------|
| ML-DSA KeyGen | ~400 µs |
| 10k Signatures | ~660 ms |
| QKD Session | ~900 ms (Aer Simulation) |
| Block Finalization | ~16 ms |
| AES Encryption | <35 µs |

For a complete, interactive execution trace of a block proposal utilizing both layers, please see the [Detailed Benchmark Logs](docs/benchmark.md).

---

## Setup & Installation

### Step 1: Go Infrastructure
```bash
go mod tidy
```

### Step 2: Qiskit Environment

**Linux / macOS**
```bash
cd qkd_engine
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

**Windows (PowerShell)**
```powershell
cd qkd_engine
python -m venv venv
.\venv\Scripts\Activate.ps1
pip install -r requirements.txt
```

---

## Execution

The AegisQKD daemon provides an interactive CLI at startup, allowing you to dynamically select your quantum backend.

**Linux / macOS**
```bash
export IBM_QUANTUM_TOKEN="YOUR_API_KEY_HERE"
export PATH="$(pwd)/qkd_engine/venv/bin:$PATH"
export LD_LIBRARY_PATH="/usr/local/lib64:$LD_LIBRARY_PATH"

go run ./cmd/aegisqd
```

**Windows (PowerShell)**
```powershell
$env:IBM_QUANTUM_TOKEN="YOUR_API_KEY_HERE"
$env:PATH = "$(Get-Location)\qkd_engine\venv\Scripts;" + $env:PATH
# Note: Windows users must ensure liboqs is compiled and present in their system PATH

go run ./cmd/aegisqd
```

---

## Roadmap

- [x] BB84 Protocol
- [x] IBM Quantum Runtime Integration
- [x] Aer Simulator Backend
- [ ] E91 Entanglement-based Protocol
- [ ] B92 Protocol
- [ ] Six-State Protocol
- [ ] Cascade Information Reconciliation
- [ ] LDPC Reconciliation
- [ ] Multi-node distributed deployment

---

## Citation

If you use AegisQKD in your research or educational materials, please cite this repository:

```text
@misc{aegisqkd2026,
  author       = {Suresh Krishna R},
  title        = {AegisQKD: Experimental Implementation of PQC and QKD into Distributed Blockchain System},
  year         = {2026},
  howpublished = {GitHub repository},
  note         = {Available at: https://github.com/sureshKrishna05/aegisq-platform-QKD (Accessed: 2026-07-12)}
}
```
