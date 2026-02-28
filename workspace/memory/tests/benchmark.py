#!/usr/bin/env python3
"""Benchmark memory system performance."""

import sys
import time
from pathlib import Path
from datetime import datetime, timezone

sys.path.insert(0, str(Path(__file__).parent.parent))
from memory_core import write_event, recall_with_tiers
from active_context import ActiveContext

def benchmark_write(iterations=100):
    """Benchmark event writing."""
    print(f"Benchmarking write_event ({iterations} iterations)...")
    
    start = time.time()
    for i in range(iterations):
        write_event(
            role="user",
            content=f"Test benchmark event {i}",
            event_type="message"
        )
    elapsed = time.time() - start
    
    ops_per_sec = iterations / elapsed
    print(f"  ✓ {elapsed:.2f}s total ({ops_per_sec:.1f} writes/sec)")
    return ops_per_sec

def benchmark_recall(iterations=50):
    """Benchmark memory recall."""
    print(f"Benchmarking recall_with_tiers ({iterations} iterations)...")
    
    queries = [
        "preferences",
        "decisions made",
        "important facts",
        "calendar events",
        "tasks completed"
    ]
    
    start = time.time()
    for i in range(iterations):
        query = queries[i % len(queries)]
        recall_with_tiers(query, budget_tokens=1000)
    elapsed = time.time() - start
    
    ops_per_sec = iterations / elapsed
    avg_latency = (elapsed / iterations) * 1000
    print(f"  ✓ {elapsed:.2f}s total ({ops_per_sec:.1f} recalls/sec, {avg_latency:.1f}ms avg)")
    return ops_per_sec

def benchmark_cache_refresh():
    """Benchmark active context cache refresh."""
    print("Benchmarking active context cache refresh...")
    
    cache = ActiveContext()
    
    start = time.time()
    cache.refresh()
    elapsed = time.time() - start
    
    print(f"  ✓ {elapsed*1000:.1f}ms")
    return elapsed

def run_benchmarks():
    """Run all benchmarks."""
    print("\n🔬 MEMORY SYSTEM BENCHMARKS\n")
    print("="*60)
    
    write_ops = benchmark_write(100)
    print()
    recall_ops = benchmark_recall(50)
    print()
    cache_time = benchmark_cache_refresh()
    
    print("\n" + "="*60)
    print("SUMMARY")
    print("="*60)
    print(f"Write throughput:     {write_ops:>8.1f} ops/sec")
    print(f"Recall throughput:    {recall_ops:>8.1f} ops/sec")
    print(f"Cache refresh:        {cache_time*1000:>8.1f} ms")
    print()
    
    # Performance targets
    print("TARGETS vs ACTUAL")
    print("-"*60)
    targets = {
        "Write latency": (10, 1000/write_ops, "ms"),
        "Recall latency": (100, 1000/recall_ops, "ms"),
        "Cache refresh": (50, cache_time*1000, "ms")
    }
    
    for metric, (target, actual, unit) in targets.items():
        status = "✅ PASS" if actual <= target else "❌ FAIL"
        print(f"{metric:20} {status}  (target: <{target}{unit}, actual: {actual:.1f}{unit})")
    
    print("="*60 + "\n")

if __name__ == "__main__":
    run_benchmarks()
