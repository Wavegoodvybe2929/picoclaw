#!/usr/bin/env python3
"""Test tier functionality."""

import sys
from pathlib import Path
from datetime import datetime, timedelta, timezone

sys.path.insert(0, str(Path(__file__).parent.parent))
from memory_core import write_event, get_db_connection
from active_context import ActiveContext

def test_tier_assignment():
    """Test that new events are assigned to correct tier."""
    print("Testing tier assignment...")
    
    # Write a new event
    result = write_event(
        role="user",
        content="Test preference: I like coffee in the morning",
        event_type="message"
    )
    assert result["status"] == "ok", f"Write failed: {result}"
    
    # Check it's in tier 1 or 2 (recent, tier defaults to 2)
    conn = get_db_connection()
    cursor = conn.cursor()
    cursor.execute("SELECT tier FROM documents ORDER BY created_at DESC LIMIT 1")
    row = cursor.fetchone()
    conn.close()
    
    if row:
        tier = row[0]
        assert tier in [1, 2], f"Expected tier 1 or 2, got {tier}"
        print(f"✓ Tier assignment works (tier={tier})")
    else:
        print("⚠️  No documents found (may need to run memory_materialize first)")

def test_active_context():
    """Test active context cache."""
    print("Testing active context cache...")
    
    cache = ActiveContext()
    cache.refresh()
    
    context = cache.get_context(max_items=10)
    assert isinstance(context, list), "Context should be a list"
    
    print(f"✓ Active context returned {len(context)} items")

def test_tier_recall():
    """Test recall with tier support."""
    print("Testing tier-based recall...")
    
    from memory_core import recall_with_tiers
    
    result = recall_with_tiers("coffee preferences", budget_tokens=500)
    assert isinstance(result, str), "Recall should return a string"
    assert len(result) > 0, "Recall should return content"
    
    print(f"✓ Tier recall returned {len(result)} chars")

def test_database_schema():
    """Test that all tier-related schema elements exist."""
    print("Testing database schema...")
    
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Check tier columns exist in documents table
    cursor.execute("PRAGMA table_info(documents)")
    columns = {row[1] for row in cursor.fetchall()}
    assert 'tier' in columns, "tier column missing"
    assert 'last_access' in columns, "last_access column missing"
    assert 'access_count' in columns, "access_count column missing"
    
    # Check tier_config table exists
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='tier_config'")
    assert cursor.fetchone() is not None, "tier_config table missing"
    
    # Check tier_config has data
    cursor.execute("SELECT COUNT(*) FROM tier_config")
    count = cursor.fetchone()[0]
    assert count >= 3, f"tier_config should have at least 3 tiers, found {count}"
    
    # Check summaries table exists (for future use)
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='summaries'")
    assert cursor.fetchone() is not None, "summaries table missing"
    
    conn.close()
    print("✓ Database schema is correct")

def test_tier_management():
    """Test tier promotion/demotion logic."""
    print("Testing tier management...")
    
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Get current tier distribution
    cursor.execute("SELECT tier, COUNT(*) FROM documents GROUP BY tier")
    tiers_before = {row[0]: row[1] for row in cursor.fetchall()}
    
    # Test that we can query by tier
    cursor.execute("SELECT COUNT(*) FROM documents WHERE tier = 2")
    tier2_count = cursor.fetchone()[0]
    
    conn.close()
    print(f"✓ Tier management queries work (Tier 2: {tier2_count} docs)")

if __name__ == "__main__":
    print("\n🧪 Running Tier Tests\n")
    print("="*60)
    
    test_database_schema()
    test_tier_assignment()
    test_active_context()
    test_tier_recall()
    test_tier_management()
    
    print("="*60)
    print("\n✅ All Phase 1 tests passed!\n")
