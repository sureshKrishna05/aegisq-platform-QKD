import sys
import json
import argparse
import hashlib
import numpy as np
from qiskit import QuantumCircuit
from qiskit_aer import AerSimulator

def run_bb84(num_bits, eavesdrop=False, noise_level=0.0):
    simulator = AerSimulator()
    
    # Alice generates random bits and random bases (0: Z-basis, 1: X-basis)
    alice_bits = np.random.randint(2, size=num_bits)
    alice_bases = np.random.randint(2, size=num_bits)
    
    # Bob generates random bases for measurement
    bob_bases = np.random.randint(2, size=num_bits)
    
    eve_bases = []
    if eavesdrop:
        eve_bases = np.random.randint(2, size=num_bits)

    bob_results = []
    
    for i in range(num_bits):
        # 1 Qubit, 2 Classical bits (c[0] for Eve, c[1] for Bob)
        qc = QuantumCircuit(1, 2)
        
        # 1. Alice prepares the qubit
        if alice_bits[i] == 1:
            qc.x(0)
        if alice_bases[i] == 1:
            qc.h(0)
            
        # 2. Channel Noise (Simulated as random bit-flips in the channel)
        if noise_level > 0:
            if np.random.rand() < noise_level:
                qc.x(0)
                
        # 3. Eve intercepts and resends
        if eavesdrop:
            # Eve measures in her random basis
            if eve_bases[i] == 1:
                qc.h(0)
            qc.measure(0, 0)
            
            # Eve prepares the state again to send to Bob
            # (In Qiskit, the measurement already collapsed the state in the current basis.
            # We apply H again if Eve used X-basis to return it to the correct orientation for Bob)
            if eve_bases[i] == 1:
                qc.h(0)
        
        # 4. Bob measures the qubit
        if bob_bases[i] == 1:
            qc.h(0)
        qc.measure(0, 1)
        
        # Run circuit
        job = simulator.run(qc, shots=1, memory=True)
        result = job.result()
        memory_str = result.get_memory()[0]  # format: "c1 c0" where c1 is left-most
        
        # Extract Bob's measurement (which is in classical register 1, so left-most character)
        measured_bit = int(memory_str[0])
        bob_results.append(measured_bit)
        
    # 5. Sifting: Alice and Bob discard bits where their bases didn't match
    alice_key = []
    bob_key = []
    for i in range(num_bits):
        if alice_bases[i] == bob_bases[i]:
            alice_key.append(alice_bits[i])
            bob_key.append(bob_results[i])
            
    # 6. Calculate Quantum Bit Error Rate (QBER)
    errors = sum(1 for a, b in zip(alice_key, bob_key) if a != b)
    qber = errors / len(alice_key) if len(alice_key) > 0 else 0
    
    # 7. Final Key Derivation (Privacy Amplification)
    # To eliminate any partial information Eve might have gained (assuming QBER < 11%),
    # we apply a cryptographic hash function (SHA-256) as a universal hash for privacy amplification.
    # This securely distills the raw sifted key into a perfectly uniform 256-bit (32-byte) AES key.
    
    sifted_str = ''.join(str(bit) for bit in bob_key)
    pa_hash = hashlib.sha256(sifted_str.encode('utf-8'))
    key_bytes = pa_hash.digest()  # Exactly 32 bytes (256-bit)
    
    # QBER > 11% typically indicates the presence of an eavesdropper in BB84
    eavesdropper_detected = qber > 0.11

    return {
        "sifted_key_length": len(alice_key),
        "qber": float(qber),
        "eavesdropper_detected": eavesdropper_detected,
        "symmetric_key_hex": key_bytes.hex(),
    }

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="BB84 QKD Simulator for AegisQ")
    parser.add_argument("--bits", type=int, default=1024, help="Number of qubits to transmit (default 1024)")
    parser.add_argument("--eavesdrop", action="store_true", help="Enable Eve intercept-resend attack")
    parser.add_argument("--noise", type=float, default=0.0, help="Quantum channel noise level (0.0 to 1.0)")
    args = parser.parse_args()
    
    result = run_bb84(args.bits, eavesdrop=args.eavesdrop, noise_level=args.noise)
    print(json.dumps(result))
