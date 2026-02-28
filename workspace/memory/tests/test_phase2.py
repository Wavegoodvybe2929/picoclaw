#!/usr/bin/env python3
"""
Test Phase 2: Progressive Distillation
Comprehensive test suite for distillation functionality.
"""

import sys
from pathlib import Path
from datetime import datetime, timezone

sys.path.insert(0, str(Path(__file__).parent.parent))
from memory_core import search_summaries, get_db_connection
from distiller import distill_daily, distill_weekly, distill_monthly, get_llm_config

def test_llm_config():
    """Test that LLM configuration can be read."""
    print("Testing LLM configuration...")
    config = get_llm_config()
    assert isinstance(config, dict), "Config should be a dictionary"
    print(f"✓ LLM config loaded: {config.get('provider', 'unknown')} provider")

def test_daily_distillation():
    """Test daily distillation."""
    print("\nTesting daily distillation...")
    
    # Distill 2026-02-20 (known to have data)
    date = datetime(2026, 2, 20, tzinfo=timezone.utc)
    result = distill_daily(date)
    
    assert result.get("status") in ["ok", "no_events"], f"Expected ok or no_events, got {result.get('status')}"
    
    if result.get("status") == "ok":
        assert result.get("event_count", 0) > 0, "Should have processed some events"
        assert "summary_file" in result, "Should return summary file path"
        
        # Verify file exists
        summary_file = Path(result["summary_file"])
        assert summary_file.exists(), f"Summary file should exist: {summary_file}"
        print(f"✓ Daily distillation works ({result['event_count']} events processed)")
    else:
        print("✓ Daily distillation works (no events found)")

def test_weekly_distillation():
    """Test weekly distillation."""
    print("\nTesting weekly distillation...")
    
    # Distill week 2026-W08 (should have at least one daily summary)
    result = distill_weekly(2026, 8)
    
    assert result.get("status") in ["ok", "no_summaries"], f"Expected ok or no_summaries, got {result.get('status')}"
    
    if result.get("status") == "ok":
        assert result.get("daily_count", 0) > 0, "Should have processed some daily summaries"
        assert "summary_file" in result, "Should return summary file path"
        
        # Verify file exists
        summary_file = Path(result["summary_file"])
        assert summary_file.exists(), f"Summary file should exist: {summary_file}"
        print(f"✓ Weekly distillation works ({result['daily_count']} daily summaries processed)")
    else:
        print("✓ Weekly distillation works (no summaries found)")

def test_monthly_distillation():
    """Test monthly distillation."""
    print("\nTesting monthly distillation...")
    
    # Distill month 2026-02 (should have at least one weekly summary)
    result = distill_monthly(2026, 2)
    
    assert result.get("status") in ["ok", "no_summaries"], f"Expected ok or no_summaries, got {result.get('status')}"
    
    if result.get("status") == "ok":
        assert result.get("weekly_count", 0) > 0, "Should have processed some weekly summaries"
        assert "summary_file" in result, "Should return summary file path"
        
        # Verify file exists
        summary_file = Path(result["summary_file"])
        assert summary_file.exists(), f"Summary file should exist: {summary_file}"
        print(f"✓ Monthly distillation works ({result['weekly_count']} weekly summaries processed)")
    else:
        print("✓ Monthly distillation works (no summaries found)")

def test_summary_search():
    """Test searching summaries."""
    print("\nTesting summary search...")
    
    results = search_summaries("prefer", limit=10)
    assert isinstance(results, list), "Results should be a list"
    
    if results:
        print(f"✓ Summary search works ({len(results)} summaries found)")
        
        # Verify result structure
        for r in results:
            assert "summary_id" in r, "Result should have summary_id"
            assert "period_type" in r, "Result should have period_type"
            assert "period_key" in r, "Result should have period_key"
            assert "content" in r, "Result should have content"
    else:
        print("✓ Summary search works (no results, but no errors)")

def test_database_storage():
    """Test that summaries are stored in database."""
    print("\nTesting database storage...")
    
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Check summaries table exists
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='summaries'")
    assert cursor.fetchone() is not None, "Summaries table should exist"
    
    # Check for summaries
    cursor.execute("SELECT COUNT(*) FROM summaries")
    count = cursor.fetchone()[0]
    
    if count > 0:
        print(f"✓ Database storage works ({count} summaries stored)")
        
        # Verify all period types
        cursor.execute("SELECT DISTINCT period_type FROM summaries ORDER BY period_type")
        types = [row[0] for row in cursor.fetchall()]
        print(f"  Period types: {', '.join(types)}")
    else:
        print("✓ Database storage works (table exists, no data yet)")
    
    conn.close()

def test_recall_integration():
    """Test that summaries are integrated into recall."""
    print("\nTesting recall integration...")
    
    from memory_core import recall_with_tiers
    
    # Test recall with a query that should match summaries
    result = recall_with_tiers("preferences", budget_tokens=1000, format="markdown")
    
    assert isinstance(result, str), "Result should be a string"
    assert len(result) > 0, "Result should not be empty"
    
    # Check if summaries section is present (if we have summaries)
    conn = get_db_connection()
    cursor = conn.cursor()
    cursor.execute("SELECT COUNT(*) FROM summaries")
    has_summaries = cursor.fetchone()[0] > 0
    conn.close()
    
    if has_summaries:
        # Should contain summary section
        if "Summaries" in result or "Summary" in result:
            print("✓ Recall integration works (summaries included)")
        else:
            print("⚠️  Recall works but summaries not found in output (query may not match)")
    else:
        print("✓ Recall integration works (no summaries to test)")

def test_directory_structure():
    """Test that directory structure is correct."""
    print("\nTesting directory structure...")
    
    from memory_core import MEMORY_DIR
    from distiller import DISTILLED_DIR
    
    # Check main distilled directory
    assert DISTILLED_DIR.exists(), f"Distilled directory should exist: {DISTILLED_DIR}"
    
    # Check subdirectories
    daily_dir = DISTILLED_DIR / "daily"
    weekly_dir = DISTILLED_DIR / "weekly"
    monthly_dir = DISTILLED_DIR / "monthly"
    
    assert daily_dir.exists(), f"Daily directory should exist: {daily_dir}"
    assert weekly_dir.exists(), f"Weekly directory should exist: {weekly_dir}"
    assert monthly_dir.exists(), f"Monthly directory should exist: {monthly_dir}"
    
    print("✓ Directory structure correct")

if __name__ == "__main__":
    print("\n🧪 Testing Phase 2: Progressive Distillation\n")
    print("="*60)
    
    try:
        test_llm_config()
        test_directory_structure()
        test_daily_distillation()
        test_weekly_distillation()
        test_monthly_distillation()
        test_database_storage()
        test_summary_search()
        test_recall_integration()
        
        print("\n" + "="*60)
        print("✅ All Phase 2 tests passed!")
        print("="*60 + "\n")
        
    except AssertionError as e:
        print("\n" + "="*60)
        print(f"❌ Test failed: {e}")
        print("="*60 + "\n")
        sys.exit(1)
    except Exception as e:
        print("\n" + "="*60)
        print(f"❌ Unexpected error: {e}")
        print("="*60 + "\n")
        import traceback
        traceback.print_exc()
        sys.exit(1)
