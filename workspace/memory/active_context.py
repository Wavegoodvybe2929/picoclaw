#!/usr/bin/env python3
"""
Active Context Cache (Tier 1)
In-memory hot storage for recent and frequently accessed memories.
"""

import json
from pathlib import Path
from typing import Dict, List, Any
from datetime import datetime, timedelta, timezone
from memory_core import MEMORY_DIR, get_db_connection

CACHE_FILE = MEMORY_DIR / "index" / "active_context.json"
MAX_CACHE_ITEMS = 100  # Limit to prevent unbounded growth

class ActiveContext:
    """Manages Tier 1 active context cache."""
    
    def __init__(self):
        self.cache = self._load_cache()
    
    def _load_cache(self) -> Dict[str, Any]:
        """Load cache from disk or create new."""
        if CACHE_FILE.exists():
            try:
                with open(CACHE_FILE, 'r') as f:
                    return json.load(f)
            except Exception:
                pass
        return {
            "pinned": [],
            "recent": [],
            "frequent": [],
            "last_updated": None
        }
    
    def _save_cache(self):
        """Persist cache to disk."""
        self.cache["last_updated"] = datetime.now(timezone.utc).isoformat()
        with open(CACHE_FILE, 'w') as f:
            json.dump(self.cache, f, indent=2)
    
    def refresh(self):
        """Rebuild cache from database."""
        conn = get_db_connection()
        cursor = conn.cursor()
        
        # Get pinned memories (priority >= 8)
        cursor.execute('''
            SELECT memory_id, content, priority, created_at
            FROM memories
            WHERE priority >= 8
            ORDER BY priority DESC, created_at DESC
            LIMIT 20
        ''')
        self.cache["pinned"] = [dict(row) for row in cursor.fetchall()]
        
        # Get recent documents (last 7 days, tier 1)
        seven_days_ago = (datetime.now(timezone.utc) - timedelta(days=7)).isoformat()
        cursor.execute('''
            SELECT doc_id, content, timestamp, role
            FROM documents
            WHERE timestamp >= ? AND tier = 1
            ORDER BY timestamp DESC
            LIMIT 30
        ''', (seven_days_ago,))
        self.cache["recent"] = [dict(row) for row in cursor.fetchall()]
        
        # Get frequently accessed (access_count > 5)
        cursor.execute('''
            SELECT doc_id, content, access_count, last_access
            FROM documents
            WHERE access_count > 5
            ORDER BY access_count DESC, last_access DESC
            LIMIT 20
        ''')
        self.cache["frequent"] = [dict(row) for row in cursor.fetchall()]
        
        conn.close()
        self._save_cache()
    
    def get_context(self, max_items: int = 50) -> List[Dict[str, Any]]:
        """Get active context items for recall."""
        all_items = (
            self.cache["pinned"] +
            self.cache["recent"] +
            self.cache["frequent"]
        )
        # Deduplicate by doc_id/memory_id
        seen = set()
        unique = []
        for item in all_items:
            key = item.get('doc_id') or item.get('memory_id')
            if key and key not in seen:
                seen.add(key)
                unique.append(item)
        return unique[:max_items]
    
    def record_access(self, doc_id: str):
        """Record that a document was accessed (increment count)."""
        conn = get_db_connection()
        cursor = conn.cursor()
        cursor.execute('''
            UPDATE documents
            SET access_count = access_count + 1,
                last_access = ?
            WHERE doc_id = ?
        ''', (datetime.now(timezone.utc).isoformat(), doc_id))
        conn.commit()
        conn.close()
