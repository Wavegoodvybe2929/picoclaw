#!/usr/bin/env python3
"""
Phase 3 Tests: Archive & Compression
Comprehensive test suite for archiving functionality.
"""

import sys
import json
from pathlib import Path
from datetime import datetime, timedelta, timezone

sys.path.insert(0, str(Path(__file__).parent.parent))
from archiver import (
    compress_month, 
    search_archive, 
    list_archives, 
    auto_archive_old_months,
    get_archive_stats,
    ARCHIVE_DIR
)
from memory_core import write_event

def test_archive_directory_creation():
    """Test that archive directory is created."""
    print("Testing archive directory creation...")
    assert ARCHIVE_DIR.exists(), "Archive directory should exist"
    assert ARCHIVE_DIR.is_dir(), "Archive path should be a directory"
    print("✓ Archive directory exists")

def test_list_empty_archives():
    """Test listing archives when none exist."""
    print("\nTesting list_archives() with no archives...")
    archives = list_archives()
    assert isinstance(archives, list), "Should return a list"
    print(f"✓ list_archives() returned {len(archives)} archives")

def test_get_archive_stats():
    """Test getting archive statistics."""
    print("\nTesting get_archive_stats()...")
    stats = get_archive_stats()
    assert isinstance(stats, dict), "Should return a dict"
    assert 'total_archives' in stats, "Should have total_archives field"
    assert 'total_size' in stats, "Should have total_size field"
    print(f"✓ get_archive_stats() returned: {stats}")

def test_archive_nonexistent_month():
    """Test archiving a month with no events."""
    print("\nTesting compress_month() for nonexistent month...")
    result = compress_month(2025, 1)  # January 2025 likely has no events
    assert isinstance(result, dict), "Should return a dict"
    assert result.get('status') in ['ok', 'no_events'], "Should have valid status"
    print(f"✓ compress_month() result: {result.get('status')}")

def test_search_nonexistent_archive():
    """Test searching an archive that doesn't exist."""
    print("\nTesting search_archive() for nonexistent archive...")
    results = search_archive("test", 2025, 1)
    assert isinstance(results, list), "Should return a list"
    assert len(results) == 0, "Should return empty list for nonexistent archive"
    print("✓ search_archive() correctly handles nonexistent archive")

def test_auto_archive():
    """Test auto-archive functionality."""
    print("\nTesting auto_archive_old_months()...")
    result = auto_archive_old_months()
    assert isinstance(result, dict), "Should return a dict"
    assert result.get('status') in ['ok', 'no_events'], "Should have valid status"
    assert 'archived' in result, "Should have archived field"
    print(f"✓ auto_archive_old_months() result: {result.get('status')}")

def test_module_imports():
    """Test that all required functions can be imported."""
    print("\nTesting module imports...")
    from archiver import (
        compress_month,
        search_archive,
        list_archives,
        auto_archive_old_months,
        get_archive_stats,
        ARCHIVE_DIR
    )
    print("✓ All required functions imported successfully")

def run_all_tests():
    """Run all Phase 3 tests."""
    print("\n" + "="*60)
    print("🧪 PHASE 3 TESTS: Archive & Compression")
    print("="*60 + "\n")
    
    try:
        test_module_imports()
        test_archive_directory_creation()
        test_list_empty_archives()
        test_get_archive_stats()
        test_archive_nonexistent_month()
        test_search_nonexistent_archive()
        test_auto_archive()
        
        print("\n" + "="*60)
        print("✅ ALL PHASE 3 TESTS PASSED!")
        print("="*60 + "\n")
        return True
        
    except AssertionError as e:
        print(f"\n❌ TEST FAILED: {e}\n")
        return False
    except Exception as e:
        print(f"\n❌ UNEXPECTED ERROR: {e}\n")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    success = run_all_tests()
    sys.exit(0 if success else 1)
