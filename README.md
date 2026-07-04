<div align="center">

# 🛡️ AegisQ Framework

### Deterministic Trust Preservation Framework for the AegisQ Digital Trust Infrastructure

*Post-Quantum Secure • Deterministic • Consensus-Driven • Storage-Optimized*

</div>

---

## Overview

**AegisQ Framework** is the core **Trust Preservation Layer** of the AegisQ ecosystem. It provides a deterministic blockchain framework that securely preserves trust artifacts using post-quantum cryptography, Byzantine Fault Tolerant (BFT) consensus, immutable storage, and verifiable persistence.

The framework is **not a complete governance platform**. Instead, it serves as the underlying infrastructure responsible for cryptographic integrity, consensus, ledger management, and trusted storage for higher-level governance protocols.

---

## System Architecture

```text
                 Trust Artifact
                       │
                       ▼
              AQX Serialization
                       │
                       ▼
               Cryptographic Hash
                       │
                       ▼
          Post-Quantum Digital Signature
                       │
                       ▼
             Deterministic BFT Consensus
                       │
                       ▼
                Block Finalization
                       │
                       ▼
          PebbleDB Ledger + LRU Cache
                       │
                       ▼
             REST API
```

---

## Current Features

* ✅ Deterministic Byzantine Fault Tolerant (BFT) consensus
* ✅ CRYSTALS-Dilithium post-quantum digital signatures
* ✅ Deterministic AQX serialization
* ✅ Merkle tree based block verification
* ✅ Replay protection using transaction nonces
* ✅ PebbleDB persistent storage engine
* ✅ LRU caching for high-performance reads
* ✅ Database integrity verification and crash recovery
* ✅ Transaction and block separation for efficient storage
* ✅ REST API
* ✅ Comprehensive unit and integration testing

---

## Repository Scope

This repository currently implements the **blockchain infrastructure** of AegisQ, including:

* Identity management for validators
* Transaction lifecycle
* Cryptographic operations
* Block construction and validation
* Consensus engine
* Ledger management
* Persistent storage
* Node APIs

---

## Out of Scope

The following components belong to the broader **AegisQ Platform** and are **not implemented inside this repository**:

* Governance Authorization Protocol (AATP)
* Trust Creation Protocol (AEP)
* Trust Preservation Protocol (ASP)
* Trust Verification Protocol (ARP)
* Governance policy engine
* Digital governance workflows
* Citizen or departmental identity management
* Distributed multi-node networking
* Smart contract or governance execution engine

---

## Technology Stack

| Component     | Technology                         |
| ------------- | ---------------------------------- |
| Language      | Go                                 |
| Consensus     | Deterministic BFT                  |
| Cryptography  | CRYSTALS-Dilithium, Ed25519, ECDSA |
| Serialization | AQX                                |
| Storage       | PebbleDB                           |
| Caching       | LRU Cache                          |
| API           | REST                               |

| Testing       | Go Testing Framework               |

---

## Current Execution Flow

```text
Start Node
      │
      ▼
Initialize Validators
      │
      ▼
Generate / Receive Transactions
      │
      ▼
AQX Serialization
      │
      ▼
Hash & Sign Transaction
      │
      ▼
Leader Proposes Block
      │
      ▼
Prepare Phase
      │
      ▼
Commit Phase
      │
      ▼
Finalize Block
      │
      ▼
Persist to PebbleDB
      │
      ▼
Serve Node APIs
```

---

## Vision

The AegisQ Framework is designed as the foundational trust-preservation engine for the larger **AegisQ Digital Trust Infrastructure**, where higher-level governance protocols will create, preserve, verify, and audit trust artifacts using this framework as the underlying immutable ledger.

Rather than being "another blockchain," AegisQ Framework is intended to become a reusable, post-quantum secure trust infrastructure for digital governance systems.
