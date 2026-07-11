# AegisQKD Interactive Execution Log

The following is an authentic execution log demonstrating the hybrid architecture in action, capturing the QKD physical simulation timing, the Dilithium signing benchmarks, and the granular PBFT network packet encryption tracing.

```text
rubberducky@fedora:~/projects/go_project/aegisq-platform-QKD$ go run ./cmd/aegisqd

--- IDENTITY LAYER ---
[BENCHMARK] ML-DSA-44 (Dilithium2) Keypair generated in 400.554µs
Validators initialized.

===========================================
 Select QKD Engine Execution Mode:
 1) Local Simulation (Fast, no queue)
 2) Real IBM Quantum Hardware (Requires IBM_QUANTUM_TOKEN)
===========================================
Enter choice (1 or 2) [Default: 1]: 1

Initializing QKD Engine (Local Simulation)...
Establishing Quantum Session Key for Validator 1...
[BENCHMARK] QKD Channel Established in 973.978ms | QBER: 0.00%
Establishing Quantum Session Key for Validator 2...
[BENCHMARK] QKD Channel Established in 1.092s | QBER: 0.00%
Establishing Quantum Session Key for Validator 3...
[BENCHMARK] QKD Channel Established in 1.101s | QBER: 0.00%
Establishing Quantum Session Key for Validator 4...
[BENCHMARK] QKD Channel Established in 1.159s | QBER: 0.00%
All secure channels established.
2026/07/12 02:49:24 [JOB 1] WAL file aegisq.db/000110.log with log number 000110 stopped reading at offset: 0; replayed 0 keys in 0 batches
Running Crash Recovery & Integrity Verification...
Database integrity verified.
Restored height: 4
Leader selected: validator-2

--- TRANSACTION LAYER ---
Generated synthetic storage transactions: 10000
[BENCHMARK] ML-DSA-44 Signed 10,000 Transactions in 662.294ms

--- CONSENSUS LAYER ---
Block finalize time: 16.049ms
Proposed block height: 5
[QKD NETWORK] PREPARE Vote Validator 1 | AQX Serialization: 2.212µs | AES-256 Encrypt: 34.869µs | Ciphertext: 9f3c34212583a76081cd1b03a4a6532b...
[QKD NETWORK] DECRYPT Validator 1 | AES-256 Decrypt: 2.068µs | AQX Deserialize: 1.097µs
[QKD NETWORK] PREPARE Vote Validator 2 | AQX Serialization: 137ns | AES-256 Encrypt: 652ns | Ciphertext: 970a156a0182a0dcd3f1d74a8a3f502f...
[QKD NETWORK] DECRYPT Validator 2 | AES-256 Decrypt: 286ns | AQX Deserialize: 110ns
[QKD NETWORK] PREPARE Vote Validator 3 | AQX Serialization: 96ns | AES-256 Encrypt: 1.639µs | Ciphertext: c77b048a73e26f938be8006d125c9274...
[QKD NETWORK] DECRYPT Validator 3 | AES-256 Decrypt: 249ns | AQX Deserialize: 85ns
[QKD NETWORK] PREPARE Vote Validator 4 | AQX Serialization: 109ns | AES-256 Encrypt: 462ns | Ciphertext: 7516e878e9b98deb1e00f53e028ecb3c...
[QKD NETWORK] DECRYPT Validator 4 | AES-256 Decrypt: 339ns | AQX Deserialize: 93ns
Prepare quorum reached.
[QKD NETWORK] COMMIT Vote Validator 1 | AQX Serialization: 279ns | AES-256 Encrypt: 644ns | Ciphertext: 585e66f3d3b879b2ac080bd271d7e4be...
[QKD NETWORK] DECRYPT Validator 1 | AES-256 Decrypt: 1.681µs | AQX Deserialize: 175ns
[QKD NETWORK] COMMIT Vote Validator 2 | AQX Serialization: 134ns | AES-256 Encrypt: 8.818µs | Ciphertext: 0237e303b025ad1ec53b08f9f4382040...
[QKD NETWORK] DECRYPT Validator 2 | AES-256 Decrypt: 468ns | AQX Deserialize: 117ns
[QKD NETWORK] COMMIT Vote Validator 3 | AQX Serialization: 179ns | AES-256 Encrypt: 468ns | Ciphertext: f2c0daac7f971c3f804b9d6a9ecab48b...
[QKD NETWORK] DECRYPT Validator 3 | AES-256 Decrypt: 360ns | AQX Deserialize: 103ns
[QKD NETWORK] COMMIT Vote Validator 4 | AQX Serialization: 98ns | AES-256 Encrypt: 992ns | Ciphertext: 3805563c2bff16d3c58b40473c4c8fb8...
[QKD NETWORK] DECRYPT Validator 4 | AES-256 Decrypt: 484ns | AQX Deserialize: 216ns
Commit quorum reached.
⚡ [EVENT BUS] BlockPersisted published by Storage | Payload: 5
Block committed at height: 5

========= BLOCK SUMMARY =========
Height: 5
Hash: e004e3bf2314d3f8a1511cb84e8434e103244d15218d3ffc04ce9711f46b99ac
Previous: ae016fc6de59355e66a808325b2e68bbbb70bc0d933652744c6fc926ffcfe010
Total Transactions: 10000
  Tx 1
   Sender: validator-2
   DataHash: [58 90 233 130 179 228 45 27 91 123 171 224 114 200 44 82 66 176 12 18 31 189 158 252 16 45 210 170 42 42 218 191]
  Tx 2
   Sender: validator-2
   DataHash: [142 50 113 77 170 13 218 92 41 1 58 96 186 131 77 39 252 26 209 158 99 129 82 7 207 232 50 93 91 144 133 91]
  Tx 3
   Sender: validator-2
   DataHash: [116 8 3 45 182 239 149 201 0 106 40 249 79 10 46 217 85 18 219 173 161 142 112 31 12 62 107 96 82 144 227 24]
  Tx 4
   Sender: validator-2
   DataHash: [168 123 138 141 48 223 70 106 146 139 251 180 100 47 247 230 134 139 188 57 64 163 117 27 53 51 254 137 91 104 187 216]
  Tx 5
   Sender: validator-2
   DataHash: [50 251 92 200 146 110 88 198 226 55 0 120 71 18 204 77 194 127 37 166 117 130 104 30 133 28 232 124 230 234 231 1]
  ...
API server running on http://localhost:8080
```
