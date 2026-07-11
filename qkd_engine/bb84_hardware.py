# Copyright 2026 Suresh Krishna R
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys
import os
import json
import argparse
import hashlib
import numpy as np

from qiskit import QuantumCircuit, transpile
from qiskit_ibm_runtime import QiskitRuntimeService, SamplerV2 as Sampler

def run_bb84_hardware(num_bits, ibm_token):
    # Authenticate with IBM Quantum
    print("Authenticating with IBM Quantum Runtime...", file=sys.stderr)
    service = QiskitRuntimeService(channel="ibm_quantum_platform", token=ibm_token)
    
    # Find the least busy real quantum computer (no simulators)
    print("Searching for the least busy quantum backend...", file=sys.stderr)
    backend = service.least_busy(operational=True, simulator=False, min_num_qubits=1)
    print(f"Selected Backend: {backend.name}", file=sys.stderr)
    
    # Alice generates random bits and bases
    alice_bits = np.random.randint(2, size=num_bits)
    alice_bases = np.random.randint(2, size=num_bits)
    
    # Bob generates random bases for measurement
    bob_bases = np.random.randint(2, size=num_bits)
    
    # IMPORTANT: We must batch all circuits into a single job. 
    # Sending 1024 individual jobs to a real quantum computer will take days in the queue!
    circuits = []
    
    print("Building quantum circuits...", file=sys.stderr)
    for i in range(num_bits):
        qc = QuantumCircuit(1, 1) # 1 Qubit, 1 Classical bit
        
        # Alice prepares the qubit
        if alice_bits[i] == 1:
            qc.x(0)
        if alice_bases[i] == 1:
            qc.h(0)
            
        # The Qubits now travel over the real physical microwave/fiber lines inside the IBM QPU!
        
        # Bob measures the qubit
        if bob_bases[i] == 1:
            qc.h(0)
        qc.measure(0, 0)
        
        circuits.append(qc)
        
    print("Transpiling circuits for the specific hardware architecture...", file=sys.stderr)
    # Transpile maps the abstract H and X gates to the specific physical microwave pulses supported by the chosen QPU
    transpiled_circuits = transpile(circuits, backend)
    
    print(f"Submitting {num_bits} circuits to {backend.name} as a single batch job...", file=sys.stderr)
    sampler = Sampler(mode=backend)
    
    # Run the job with shots=1 per circuit (BB84 requires exactly one shot per photon/qubit)
    job = sampler.run(transpiled_circuits, shots=1)
    
    print(f"Job ID: {job.job_id()}", file=sys.stderr)
    print("Waiting for execution (this may take a few minutes depending on the IBM queue)...", file=sys.stderr)
    
    result = job.result()
    print("Results received!", file=sys.stderr)
    
    bob_results = []
    # SamplerV2 returns data in pub results
    for i, pub_result in enumerate(result):
        # Extract the single measured bit from classical register 'c'
        # Memory layout might be bitstrings depending on backend, grab the bit
        # In Qiskit >= 1.0 with SamplerV2, measurement registers are accessed via data
        # Usually named 'c' if not named explicitly, or 'c0' etc. We'll grab the first classical register.
        creg_name = transpiled_circuits[i].cregs[0].name
        bitstring_array = pub_result.data[creg_name].get_bitstrings()
        measured_bit = int(bitstring_array[0])
        bob_results.append(measured_bit)
        
    # Sifting
    alice_key = []
    bob_key = []
    for i in range(num_bits):
        if alice_bases[i] == bob_bases[i]:
            alice_key.append(alice_bits[i])
            bob_key.append(bob_results[i])
            
    # Calculate QBER
    errors = sum(1 for a, b in zip(alice_key, bob_key) if a != b)
    qber = errors / len(alice_key) if len(alice_key) > 0 else 0
    
    # Privacy Amplification (SHA-256)
    sifted_str = ''.join(str(bit) for bit in bob_key)
    pa_hash = hashlib.sha256(sifted_str.encode('utf-8'))
    key_bytes = pa_hash.digest()
    
    # Because real quantum hardware has physical noise (decoherence, thermal relaxation), 
    # the QBER will natively be > 0. A typical real IBM backend might yield 2-8% QBER naturally.
    eavesdropper_detected = qber > 0.11

    return {
        "sifted_key_length": len(alice_key),
        "qber": float(qber),
        "eavesdropper_detected": eavesdropper_detected,
        "symmetric_key_hex": key_bytes.hex(),
        "backend_used": backend.name,
        "job_id": job.job_id()
    }

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="BB84 QKD on Real IBM Quantum Hardware")
    parser.add_argument("--bits", type=int, default=1024, help="Number of qubits to transmit")
    args = parser.parse_args()
    
    token = os.environ.get("IBM_QUANTUM_TOKEN")
    if not token:
        print("ERROR: Please set the IBM_QUANTUM_TOKEN environment variable.", file=sys.stderr)
        sys.exit(1)
        
    try:
        result = run_bb84_hardware(args.bits, token)
        # Print final JSON to stdout so Go could potentially read it
        print(json.dumps(result))
    except Exception as e:
        print(f"Hardware execution failed: {e}", file=sys.stderr)
        sys.exit(1)
