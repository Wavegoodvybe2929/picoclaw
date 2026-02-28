#!/usr/bin/env python3
"""
PicoClaw Memory Core
Low-RAM, reliable, plug-and-play personal memory system.
"""

import json
import hashlib
import sqlite3
import os
import time
from pathlib import Path
from typing import Optional, List, Dict, Any
from datetime import datetime, timezone

# Workspace paths
WORKSPACE = Path(__file__).parent.parent
MEMORY_DIR = WORKSPACE / "memory"
LOG_DIR = MEMORY_DIR / "log"
INDEX_DIR = MEMORY_DIR / "index"
EVENTS_LOG = LOG_DIR / "events.ndjson"
MEMORY_DB = INDEX_DIR / "memory.db"
STATUS_FILE = INDEX_DIR / "status.json"

# Ensure directories exist
LOG_DIR.mkdir(parents=True, exist_ok=True)
INDEX_DIR.mkdir(parents=True, exist_ok=True)


def get_event_hash(event: Dict[str, Any]) -> str:
    """Generate deterministic hash for an event."""
    # Create canonical representation (sorted keys, no hash fields)
    canonical = {k: v for k, v in sorted(event.items()) if k not in ['hash', 'prev_hash']}
    content = json.dumps(canonical, sort_keys=True, separators=(',', ':'))
    return hashlib.sha256(content.encode('utf-8')).hexdigest()[:16]


def get_last_event_hash() -> Optional[str]:
    """Get hash of the last event in the log (for chain integrity)."""
    if not EVENTS_LOG.exists():
        return None
    
    try:
        with open(EVENTS_LOG, 'r') as f:
            lines = f.readlines()
            if not lines:
                return None
            last_line = lines[-1].strip()
            if last_line:
                event = json.loads(last_line)
                return event.get('hash')
    except Exception:
        return None
    
    return None


def write_event(
    role: str,
    content: str,
    event_type: str = "message",
    conversation_id: Optional[str] = None,
    thread_id: Optional[str] = None,
    attachments: Optional[List[str]] = None,
    metadata: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Write an event to the append-only log (guaranteed storage).
    
    Returns the event with hash and receipt confirmation.
    """
    timestamp = datetime.now(timezone.utc).isoformat()
    
    # Get previous hash for chain
    prev_hash = get_last_event_hash()
    
    # Build event
    event = {
        "event_id": hashlib.sha256(f"{timestamp}{role}{content}".encode()).hexdigest()[:12],
        "timestamp": timestamp,
        "event_type": event_type,
        "role": role,
        "content": content,
        "conversation_id": conversation_id or "default",
        "thread_id": thread_id,
        "attachments": attachments or [],
        "metadata": metadata or {},
        "prev_hash": prev_hash
    }
    
    # Add hash
    event["hash"] = get_event_hash(event)
    
    # Atomic write (append)
    try:
        with open(EVENTS_LOG, 'a') as f:
            f.write(json.dumps(event, separators=(',', ':')) + '\n')
            f.flush()
            os.fsync(f.fileno())  # Force write to disk
        
        # Update status
        update_status({"last_write": timestamp, "last_event_id": event["event_id"]})
        
        return {"status": "ok", "event": event, "receipt": event["hash"]}
    except Exception as e:
        return {"status": "error", "error": str(e)}


def get_db_connection():
    """Get SQLite connection with optional vector extension."""
    conn = sqlite3.connect(str(MEMORY_DB))
    conn.row_factory = sqlite3.Row
    
    # Try to load sqlite-vec if available
    try:
        conn.enable_load_extension(True)
        # Common paths for sqlite-vec
        vec_paths = [
            "/usr/local/lib/vec0.so",
            "/opt/homebrew/lib/vec0.so",
            str(WORKSPACE / "lib" / "vec0.so")
        ]
        for path in vec_paths:
            if os.path.exists(path):
                conn.load_extension(path)
                break
    except Exception:
        pass  # Vector search will be unavailable, but we can still use SQLite
    
    return conn


def init_memory_db():
    """Initialize memory database schema."""
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Documents table (chunked content with metadata)
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS documents (
            doc_id TEXT PRIMARY KEY,
            event_id TEXT,
            chunk_index INTEGER,
            content TEXT,
            role TEXT,
            timestamp TEXT,
            conversation_id TEXT,
            metadata TEXT,
            created_at TEXT DEFAULT CURRENT_TIMESTAMP,
            tier INTEGER DEFAULT 2,
            last_access TEXT,
            access_count INTEGER DEFAULT 0
        )
    ''')
    
    # Memories table (pinned/procedural/semantic items)
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS memories (
            memory_id TEXT PRIMARY KEY,
            memory_type TEXT,
            content TEXT,
            priority INTEGER DEFAULT 0,
            source_event_id TEXT,
            created_at TEXT DEFAULT CURRENT_TIMESTAMP,
            last_used_at TEXT,
            use_count INTEGER DEFAULT 0
        )
    ''')
    
    # Embeddings table (if vector extension available)
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS embeddings (
            doc_id TEXT PRIMARY KEY,
            embedding BLOB,
            model TEXT,
            created_at TEXT DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (doc_id) REFERENCES documents(doc_id)
        )
    ''')
    
    # Feedback table (learning signals)
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS feedback (
            feedback_id INTEGER PRIMARY KEY AUTOINCREMENT,
            doc_id TEXT,
            query_hash TEXT,
            reward REAL,
            timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (doc_id) REFERENCES documents(doc_id)
        )
    ''')
    
    # FTS for keyword search (deterministic fallback)
    cursor.execute('''
        CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
            doc_id UNINDEXED,
            content,
            metadata
        )
    ''')
    
    # Migrate existing documents table to add tier columns
    try:
        cursor.execute("SELECT tier FROM documents LIMIT 1")
    except sqlite3.OperationalError:
        # Column doesn't exist, add it
        cursor.execute("ALTER TABLE documents ADD COLUMN tier INTEGER DEFAULT 2")
        cursor.execute("ALTER TABLE documents ADD COLUMN last_access TEXT")
        cursor.execute("ALTER TABLE documents ADD COLUMN access_count INTEGER DEFAULT 0")
        conn.commit()
    
    # Tier configuration table
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS tier_config (
            tier INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            retention_days INTEGER,
            max_size_mb INTEGER,
            description TEXT
        )
    ''')
    
    # Insert default tier config
    cursor.execute('''
        INSERT OR IGNORE INTO tier_config (tier, name, retention_days, max_size_mb, description)
        VALUES 
            (1, 'active', 7, 10, 'Hot storage - last 7 days, frequently accessed'),
            (2, 'working', 30, 100, 'Warm storage - last 30 days, moderate access'),
            (3, 'archive', -1, -1, 'Cold storage - compressed archives, unlimited retention')
    ''')
    
    # Summaries table (for future distillation feature)
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS summaries (
            summary_id TEXT PRIMARY KEY,
            period_type TEXT NOT NULL,
            period_key TEXT NOT NULL,
            content TEXT NOT NULL,
            event_count INTEGER,
            created_at TEXT DEFAULT CURRENT_TIMESTAMP,
            model TEXT
        )
    ''')
    
    # Indexes
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_documents_event ON documents(event_id)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_documents_timestamp ON documents(timestamp)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_documents_tier ON documents(tier)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_documents_last_access ON documents(last_access)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_documents_access_count ON documents(access_count)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(memory_type)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_memories_priority ON memories(priority DESC)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_summaries_period ON summaries(period_type, period_key)')
    
    conn.commit()
    conn.close()


def update_status(data: Dict[str, Any]):
    """Update status file with latest processing info."""
    status = {}
    if STATUS_FILE.exists():
        try:
            with open(STATUS_FILE, 'r') as f:
                status = json.load(f)
        except Exception:
            pass
    
    status.update(data)
    status['updated_at'] = datetime.now(timezone.utc).isoformat()
    
    with open(STATUS_FILE, 'w') as f:
        json.dump(status, f, indent=2)


def get_status() -> Dict[str, Any]:
    """Get current memory system status."""
    if not STATUS_FILE.exists():
        return {"status": "uninitialized"}
    
    try:
        with open(STATUS_FILE, 'r') as f:
            return json.load(f)
    except Exception as e:
        return {"status": "error", "error": str(e)}


def count_events() -> int:
    """Count total events in the log."""
    if not EVENTS_LOG.exists():
        return 0
    
    try:
        with open(EVENTS_LOG, 'r') as f:
            return sum(1 for line in f if line.strip())
    except Exception:
        return 0


def read_events_since(last_event_id: Optional[str] = None) -> List[Dict[str, Any]]:
    """Read events from log since a specific event_id."""
    if not EVENTS_LOG.exists():
        return []
    
    events = []
    found_marker = last_event_id is None
    
    try:
        with open(EVENTS_LOG, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                
                event = json.loads(line)
                
                if not found_marker:
                    if event.get('event_id') == last_event_id:
                        found_marker = True
                    continue
                
                events.append(event)
    except Exception:
        pass
    
    return events


def chunk_text(text: str, max_chunk_size: int = 500) -> List[str]:
    """Split text into chunks for embedding (simple sentence-aware chunking)."""
    if len(text) <= max_chunk_size:
        return [text]
    
    chunks = []
    current_chunk = ""
    
    # Split by sentence boundaries
    sentences = text.replace('. ', '.|').replace('? ', '?|').replace('! ', '!|').split('|')
    
    for sentence in sentences:
        if len(current_chunk) + len(sentence) <= max_chunk_size:
            current_chunk += sentence
        else:
            if current_chunk:
                chunks.append(current_chunk.strip())
            current_chunk = sentence
    
    if current_chunk:
        chunks.append(current_chunk.strip())
    
    return chunks if chunks else [text[:max_chunk_size]]


def search_summaries(query: str, limit: int = 3) -> List[Dict[str, Any]]:
    """Search distilled summaries for relevant context."""
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Search summaries (prioritize recent, then weekly, then monthly)
    cursor.execute('''
        SELECT summary_id, period_type, period_key, content, created_at
        FROM summaries
        WHERE content LIKE ?
        ORDER BY 
            CASE period_type
                WHEN 'daily' THEN 1
                WHEN 'weekly' THEN 2
                WHEN 'monthly' THEN 3
            END,
            period_key DESC
        LIMIT ?
    ''', (f'%{query}%', limit))
    
    results = []
    for row in cursor.fetchall():
        results.append({
            'summary_id': row[0],
            'period_type': row[1],
            'period_key': row[2],
            'content': row[3],
            'created_at': row[4],
            'source': 'summary'
        })
    
    conn.close()
    return results


def recall_with_tiers(query: str, budget_tokens: int = 2000, format: str = "markdown") -> str:
    """
    Enhanced recall using hierarchical tiers.
    
    Search order:
    1. Tier 1 (active context) - always included
    2. Tier 2 (working memory) - semantic + keyword search
    3. Tier 3 (archive) - if still under budget
    """
    from active_context import ActiveContext
    
    results = []
    token_count = 0
    
    # 1. Get Tier 1 active context (always included, ~500 tokens)
    cache = ActiveContext()
    active = cache.get_context(max_items=20)
    
    for item in active:
        content = item.get('content', '')
        results.append({
            'content': content,
            'source': 'tier_1_active',
            'doc_id': item.get('doc_id') or item.get('memory_id'),
            'timestamp': item.get('timestamp') or item.get('created_at'),
            'priority': item.get('priority', 0)
        })
        token_count += len(content.split()) * 1.3  # Rough token estimate
        
        if token_count >= 500:
            break
    
    # 1.5. Check summaries (compressed knowledge)
    summaries = search_summaries(query, limit=3)
    for item in summaries:
        content = f"[{item['period_type'].title()} Summary: {item['period_key']}]\n{item['content']}"
        results.append({
            'content': content,
            'source': 'summary',
            'period': item['period_key'],
            'type': item['period_type']
        })
        token_count += len(content.split()) * 1.3
        
        if token_count >= budget_tokens * 0.4:  # Reserve 40% for summaries max
            break
    
    # 2. Search Tier 2 (working memory) if we have budget
    if token_count < budget_tokens:
        conn = get_db_connection()
        cursor = conn.cursor()
        
        # Keyword search in Tier 2 using FTS
        try:
            cursor.execute('''
                SELECT d.doc_id, d.content, d.timestamp, d.role, d.access_count
                FROM documents_fts fts
                JOIN documents d ON fts.doc_id = d.doc_id
                WHERE fts.documents_fts MATCH ? AND d.tier = 2
                ORDER BY d.access_count DESC, d.timestamp DESC
                LIMIT 10
            ''', (query,))
        except Exception:
            # Fallback to basic LIKE search if FTS fails
            cursor.execute('''
                SELECT doc_id, content, timestamp, role, access_count
                FROM documents
                WHERE content LIKE ? AND tier = 2
                ORDER BY access_count DESC, timestamp DESC
                LIMIT 10
            ''', (f'%{query}%',))
        
        for row in cursor.fetchall():
            content = row[1]
            results.append({
                'content': content,
                'source': 'tier_2_working',
                'doc_id': row[0],
                'timestamp': row[2],
                'role': row[3],
                'access_count': row[4]
            })
            token_count += len(content.split()) * 1.3
            
            # Record access
            cache.record_access(row[0])
            
            if token_count >= budget_tokens:
                break
        
        conn.close()
    
    # 3. Search Tier 3 (archive) if still under budget - placeholder for Phase 3
    # This will be implemented in Phase 3
    
    # Format results
    if format == "json":
        return json.dumps(results, indent=2)
    else:
        # Markdown format
        packet = "# Memory Context\n\n"
        
        # Group by source
        tier1 = [r for r in results if r['source'] == 'tier_1_active']
        tier2 = [r for r in results if r['source'] == 'tier_2_working']
        summaries_list = [r for r in results if r['source'] == 'summary']
        
        if tier1:
            packet += "## Recent Context (Tier 1)\n\n"
            for item in tier1[:10]:
                packet += f"- {item['content']}\n"
            packet += "\n"
        
        if summaries_list:
            packet += "## Summaries\n\n"
            for item in summaries_list[:5]:
                packet += f"### {item.get('type', 'summary').title()}: {item.get('period', '')}\n"
                packet += f"{item['content']}\n\n"
        
        if tier2:
            packet += "## Relevant Memories (Tier 2)\n\n"
            for item in tier2[:10]:
                packet += f"- {item['content']} _(accessed {item.get('access_count', 0)} times)_\n"
            packet += "\n"
        
        packet += f"\n---\n_Retrieved {len(results)} items (~{int(token_count)} tokens)_\n"
        
        return packet


# Initialize DB on import
if __name__ != "__main__":
    init_memory_db()
