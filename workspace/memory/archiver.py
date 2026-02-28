#!/usr/bin/env python3
"""
Memory Archiving & Compression
Compress old events to save space.
"""

import json
import gzip
import tarfile
from pathlib import Path
from typing import Dict, Any
from datetime import datetime, timedelta, timezone
from memory_core import MEMORY_DIR, EVENTS_LOG, get_db_connection

ARCHIVE_DIR = MEMORY_DIR / "log" / "archive"
ARCHIVE_DIR.mkdir(parents=True, exist_ok=True)

def compress_month(year: int, month: int) -> Dict[str, Any]:
    """
    Compress events from a specific month into archive.
    
    Args:
        year: Year (e.g., 2026)
        month: Month (1-12)
    
    Returns:
        Archive metadata
    """
    month_key = f"{year}-{month:02d}"
    print(f"📦 Archiving events for {month_key}...")
    
    # Read all events
    if not EVENTS_LOG.exists():
        return {"status": "no_events", "month": month_key}
    
    month_events = []
    
    with open(EVENTS_LOG, 'r') as f:
        for line in f:
            if not line.strip():
                continue
            try:
                event = json.loads(line)
                timestamp = event.get('timestamp', '')
                # Check if event is from target month
                if timestamp.startswith(month_key):
                    month_events.append(event)
            except json.JSONDecodeError:
                continue
    
    if not month_events:
        print(f"⚠️  No events found for {month_key}")
        return {"status": "no_events", "month": month_key}
    
    # Write to compressed archive
    archive_file = ARCHIVE_DIR / f"{month_key}.ndjson.gz"
    
    with gzip.open(archive_file, 'wt', encoding='utf-8') as f:
        for event in month_events:
            f.write(json.dumps(event) + '\n')
    
    # Calculate compression ratio
    original_size = sum(len(json.dumps(e)) for e in month_events)
    compressed_size = archive_file.stat().st_size
    ratio = (1 - compressed_size / original_size) * 100 if original_size > 0 else 0
    
    print(f"✓ Archived {len(month_events)} events")
    print(f"  Original: {original_size:,} bytes")
    print(f"  Compressed: {compressed_size:,} bytes")
    print(f"  Compression: {ratio:.1f}%")
    print(f"  File: {archive_file}")
    
    return {
        "status": "ok",
        "month": month_key,
        "event_count": len(month_events),
        "original_size": original_size,
        "compressed_size": compressed_size,
        "compression_ratio": ratio,
        "archive_file": str(archive_file)
    }

def search_archive(query: str, year: int, month: int) -> list:
    """Search compressed archive for specific month."""
    month_key = f"{year}-{month:02d}"
    archive_file = ARCHIVE_DIR / f"{month_key}.ndjson.gz"
    
    if not archive_file.exists():
        return []
    
    results = []
    query_lower = query.lower()
    
    with gzip.open(archive_file, 'rt', encoding='utf-8') as f:
        for line in f:
            if not line.strip():
                continue
            try:
                event = json.loads(line)
                # Search in content and metadata
                searchable = (
                    event.get('content', '') + ' ' +
                    json.dumps(event.get('metadata', {}))
                )
                if query_lower in searchable.lower():
                    results.append(event)
            except json.JSONDecodeError:
                continue
    
    return results

def list_archives() -> list:
    """List all available archives."""
    archives = []
    
    for archive_file in sorted(ARCHIVE_DIR.glob("*.ndjson.gz")):
        stat = archive_file.stat()
        archives.append({
            "month": archive_file.stem.replace('.ndjson', ''),
            "file": str(archive_file),
            "size": stat.st_size,
            "modified": datetime.fromtimestamp(stat.st_mtime, tz=timezone.utc).isoformat()
        })
    
    return archives

def auto_archive_old_months() -> Dict[str, Any]:
    """
    Automatically archive months older than 30 days.
    
    Returns:
        Summary of archiving operations
    """
    print("🔄 Auto-archiving months older than 30 days...")
    
    if not EVENTS_LOG.exists():
        return {"status": "no_events", "archived": []}
    
    # Get unique months from events
    months_to_archive = set()
    thirty_days_ago = (datetime.now(timezone.utc) - timedelta(days=30)).replace(day=1)
    
    with open(EVENTS_LOG, 'r') as f:
        for line in f:
            if not line.strip():
                continue
            try:
                event = json.loads(line)
                timestamp_str = event.get('timestamp', '')
                if timestamp_str:
                    event_date = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
                    # Only archive complete months older than 30 days
                    if event_date < thirty_days_ago:
                        month_key = event_date.strftime('%Y-%m')
                        months_to_archive.add(month_key)
            except (json.JSONDecodeError, ValueError):
                continue
    
    # Archive each month
    archived = []
    for month_key in sorted(months_to_archive):
        # Check if already archived
        archive_file = ARCHIVE_DIR / f"{month_key}.ndjson.gz"
        if archive_file.exists():
            print(f"  ⏭️  {month_key} already archived, skipping")
            continue
        
        year, month = month_key.split('-')
        result = compress_month(int(year), int(month))
        if result.get('status') == 'ok':
            archived.append(month_key)
    
    print(f"\n✅ Auto-archive complete: {len(archived)} months archived")
    return {
        "status": "ok",
        "archived": archived,
        "total_archives": len(list_archives())
    }

def get_archive_stats() -> Dict[str, Any]:
    """Get statistics about all archives."""
    archives = list_archives()
    
    if not archives:
        return {
            "total_archives": 0,
            "total_size": 0,
            "oldest_month": None,
            "newest_month": None
        }
    
    total_size = sum(a['size'] for a in archives)
    months = [a['month'] for a in archives]
    
    return {
        "total_archives": len(archives),
        "total_size": total_size,
        "total_size_mb": total_size / 1024 / 1024,
        "oldest_month": min(months) if months else None,
        "newest_month": max(months) if months else None,
        "archives": archives
    }
