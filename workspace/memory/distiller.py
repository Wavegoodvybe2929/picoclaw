#!/usr/bin/env python3
"""
Memory Distillation Module
LLM-based summarization of events into compressed memories.
"""

import json
import hashlib
import urllib.request
import urllib.error
from pathlib import Path
from typing import List, Dict, Any, Optional
from datetime import datetime, timedelta, timezone
from memory_core import get_db_connection, MEMORY_DIR

DISTILLED_DIR = MEMORY_DIR / "distilled"
DISTILLED_DIR.mkdir(exist_ok=True)
(DISTILLED_DIR / "daily").mkdir(exist_ok=True)
(DISTILLED_DIR / "weekly").mkdir(exist_ok=True)
(DISTILLED_DIR / "monthly").mkdir(exist_ok=True)


def get_llm_config() -> Dict[str, str]:
    """Get LLM configuration from PicoClaw config."""
    config_path = Path.home() / ".picoclaw" / "config.json"
    if not config_path.exists():
        return {}
    
    try:
        with open(config_path) as f:
            config = json.load(f)
            
        # Get default provider and model
        defaults = config.get("agents", {}).get("defaults", {})
        provider = defaults.get("provider", "openai")
        model = defaults.get("model", "gpt-3.5-turbo")
        
        # Get provider configuration
        provider_config = config.get("providers", {}).get(provider, {})
        api_key = provider_config.get("api_key", "")
        api_base = provider_config.get("api_base", "")
        
        # Extract model name (remove provider prefix if present)
        if "/" in model:
            model = model.split("/", 1)[1]
        
        return {
            "api_key": api_key,
            "api_base": api_base,
            "model": model,
            "provider": provider
        }
    except Exception as e:
        print(f"Warning: Could not read LLM config: {e}")
        return {}


def call_llm(prompt: str, max_tokens: int = 500) -> str:
    """Call LLM API for summarization."""
    config = get_llm_config()
    
    if not config or not config.get("api_base"):
        # Fallback to simple extraction if no LLM available
        print("Warning: LLM not configured, using simple extraction fallback")
        lines = prompt.split('\n')
        key_lines = [l for l in lines if any(keyword in l.lower() for keyword in 
                     ['prefer', 'decision', 'important', 'remember', 'always', 'never', 
                      'like', 'need', 'want', 'should', 'must', 'will'])]
        return "\n".join(key_lines[:20]) if key_lines else "No significant patterns found."
    
    try:
        # Use OpenAI-compatible API
        api_base = config["api_base"].rstrip('/')
        endpoint = f"{api_base}/chat/completions"
        
        headers = {
            "Content-Type": "application/json"
        }
        
        # Add API key if not using LM Studio (which uses "lm-studio" as placeholder)
        if config.get("api_key") and config["api_key"] != "lm-studio":
            headers["Authorization"] = f"Bearer {config['api_key']}"
        
        payload = {
            "model": config["model"],
            "messages": [
                {
                    "role": "system",
                    "content": "You are a helpful assistant that creates concise summaries of conversation logs. Focus on extracting actionable information, decisions, preferences, and key facts."
                },
                {
                    "role": "user",
                    "content": prompt
                }
            ],
            "max_tokens": max_tokens,
            "temperature": 0.3  # Lower temperature for more focused summaries
        }
        
        # Make request using urllib
        req = urllib.request.Request(
            endpoint,
            data=json.dumps(payload).encode('utf-8'),
            headers=headers,
            method='POST'
        )
        
        with urllib.request.urlopen(req, timeout=30) as response:
            result = json.loads(response.read().decode('utf-8'))
            return result["choices"][0]["message"]["content"].strip()
        
    except Exception as e:
        print(f"Warning: LLM call failed ({e}), using fallback extraction")
        # Fallback to simple extraction
        lines = prompt.split('\n')
        key_lines = [l for l in lines if any(keyword in l.lower() for keyword in 
                     ['prefer', 'decision', 'important', 'remember', 'always', 'never',
                      'like', 'need', 'want', 'should', 'must', 'will'])]
        return "\n".join(key_lines[:20]) if key_lines else "No significant patterns found."


def distill_daily(date: Optional[datetime] = None) -> Dict[str, Any]:
    """
    Distill events from a single day into a summary.
    
    Args:
        date: Date to distill (defaults to yesterday)
    
    Returns:
        Summary metadata
    """
    if date is None:
        date = datetime.now(timezone.utc) - timedelta(days=1)
    
    date_str = date.strftime("%Y-%m-%d")
    print(f"📝 Distilling events for {date_str}...")
    
    # Get events from that day
    conn = get_db_connection()
    cursor = conn.cursor()
    
    start_time = date.replace(hour=0, minute=0, second=0, microsecond=0).isoformat()
    end_time = date.replace(hour=23, minute=59, second=59, microsecond=999999).isoformat()
    
    cursor.execute('''
        SELECT doc_id, content, role, timestamp
        FROM documents
        WHERE timestamp >= ? AND timestamp <= ?
        ORDER BY timestamp ASC
    ''', (start_time, end_time))
    
    events = cursor.fetchall()
    event_count = len(events)
    
    if event_count == 0:
        print(f"⚠️  No events found for {date_str}")
        conn.close()
        return {"status": "no_events", "date": date_str}
    
    # Build prompt for LLM
    events_text = "\n\n".join([
        f"[{row[3]}] {row[2]}: {row[1]}"
        for row in events
    ])
    
    prompt = f"""Analyze these conversation events from {date_str} and create a concise summary.

Events:
{events_text}

Extract and summarize:
1. Key decisions made
2. Important facts learned
3. Preferences stated
4. Actions taken
5. Notable patterns

Be concise. Focus only on actionable or memorable information.
Format as bullet points."""
    
    # Get LLM summary
    summary = call_llm(prompt, max_tokens=500)
    
    # Save to file
    output_file = DISTILLED_DIR / "daily" / f"{date_str}.md"
    with open(output_file, 'w') as f:
        f.write(f"# Daily Summary: {date_str}\n\n")
        f.write(f"**Events processed**: {event_count}\n\n")
        f.write("## Summary\n\n")
        f.write(summary)
        f.write(f"\n\n---\n_Generated: {datetime.now(timezone.utc).isoformat()}_\n")
    
    # Store in summaries table
    summary_id = hashlib.sha256(f"daily-{date_str}".encode()).hexdigest()[:12]
    cursor.execute('''
        INSERT OR REPLACE INTO summaries (summary_id, period_type, period_key, content, event_count, created_at, model)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    ''', (summary_id, 'daily', date_str, summary, event_count, 
          datetime.now(timezone.utc).isoformat(), 'auto'))
    
    conn.commit()
    conn.close()
    
    print(f"✓ Created daily summary: {output_file}")
    return {
        "status": "ok",
        "date": date_str,
        "event_count": event_count,
        "summary_file": str(output_file),
        "summary_id": summary_id
    }


def distill_weekly(year: int, week: int) -> Dict[str, Any]:
    """Distill a week's worth of daily summaries into a weekly summary."""
    week_key = f"{year}-W{week:02d}"
    print(f"📝 Distilling weekly summary for {week_key}...")
    
    # Get daily summaries for that week
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Calculate date range for this week
    # ISO week date: week 1 is the week with the first Thursday
    from datetime import date
    jan4 = date(year, 1, 4)
    week_start = jan4 - timedelta(days=jan4.isoweekday() - 1) + timedelta(weeks=week - 1)
    week_end = week_start + timedelta(days=6)
    
    start_str = week_start.strftime("%Y-%m-%d")
    end_str = week_end.strftime("%Y-%m-%d")
    
    cursor.execute('''
        SELECT content, period_key
        FROM summaries
        WHERE period_type = 'daily' 
          AND period_key >= ? 
          AND period_key <= ?
        ORDER BY period_key
    ''', (start_str, end_str))
    
    dailies = cursor.fetchall()
    
    if not dailies:
        print(f"⚠️  No daily summaries found for {week_key}")
        conn.close()
        return {"status": "no_summaries", "week": week_key}
    
    # Combine daily summaries
    combined = "\n\n".join([f"**{row[1]}**\n{row[0]}" for row in dailies])
    
    prompt = f"""Synthesize these daily summaries from week {week_key} into a concise weekly overview.

Daily summaries:
{combined}

Create a weekly summary that:
1. Highlights key themes and patterns
2. Lists important decisions
3. Notes significant changes in preferences or behavior
4. Identifies recurring topics

Be very concise - aim for 200 words or less."""
    
    summary = call_llm(prompt, max_tokens=300)
    
    # Save to file
    output_file = DISTILLED_DIR / "weekly" / f"{week_key}.md"
    with open(output_file, 'w') as f:
        f.write(f"# Weekly Summary: {week_key}\n\n")
        f.write(f"**Daily summaries processed**: {len(dailies)}\n\n")
        f.write("## Summary\n\n")
        f.write(summary)
        f.write(f"\n\n---\n_Generated: {datetime.now(timezone.utc).isoformat()}_\n")
    
    # Store in database
    summary_id = hashlib.sha256(f"weekly-{week_key}".encode()).hexdigest()[:12]
    cursor.execute('''
        INSERT OR REPLACE INTO summaries (summary_id, period_type, period_key, content, event_count, created_at, model)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    ''', (summary_id, 'weekly', week_key, summary, len(dailies),
          datetime.now(timezone.utc).isoformat(), 'auto'))
    
    conn.commit()
    conn.close()
    
    print(f"✓ Created weekly summary: {output_file}")
    return {
        "status": "ok",
        "week": week_key,
        "daily_count": len(dailies),
        "summary_file": str(output_file)
    }


def distill_monthly(year: int, month: int) -> Dict[str, Any]:
    """Distill a month's worth of weekly summaries into a monthly summary."""
    month_key = f"{year}-{month:02d}"
    print(f"📝 Distilling monthly summary for {month_key}...")
    
    # Get weekly summaries for that month
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Get all weekly summaries that overlap with this month
    cursor.execute('''
        SELECT content, period_key
        FROM summaries
        WHERE period_type = 'weekly' 
          AND period_key LIKE ?
        ORDER BY period_key
    ''', (f"{year}-W%",))
    
    all_weeklies = cursor.fetchall()
    
    # Filter to only weeks that are primarily in this month
    from datetime import date
    weeklies = []
    for row in all_weeklies:
        week_key = row[1]
        try:
            # Parse week key (YYYY-Www)
            parts = week_key.split('-W')
            w_year = int(parts[0])
            w_week = int(parts[1])
            
            # Calculate week's midpoint
            jan4 = date(w_year, 1, 4)
            week_start = jan4 - timedelta(days=jan4.isoweekday() - 1) + timedelta(weeks=w_week - 1)
            week_mid = week_start + timedelta(days=3)
            
            # Check if midpoint is in this month
            if week_mid.year == year and week_mid.month == month:
                weeklies.append(row)
        except Exception:
            continue
    
    if not weeklies:
        print(f"⚠️  No weekly summaries found for {month_key}")
        conn.close()
        return {"status": "no_summaries", "month": month_key}
    
    # Combine weekly summaries
    combined = "\n\n".join([f"**{row[1]}**\n{row[0]}" for row in weeklies])
    
    prompt = f"""Synthesize these weekly summaries from {month_key} into a concise monthly overview.

Weekly summaries:
{combined}

Create a monthly summary that:
1. Captures major themes and accomplishments
2. Notes significant decisions and changes
3. Identifies long-term patterns
4. Lists key learnings

Be very concise - aim for 300 words or less."""
    
    summary = call_llm(prompt, max_tokens=400)
    
    # Save to file
    output_file = DISTILLED_DIR / "monthly" / f"{month_key}.md"
    with open(output_file, 'w') as f:
        f.write(f"# Monthly Summary: {month_key}\n\n")
        f.write(f"**Weekly summaries processed**: {len(weeklies)}\n\n")
        f.write("## Summary\n\n")
        f.write(summary)
        f.write(f"\n\n---\n_Generated: {datetime.now(timezone.utc).isoformat()}_\n")
    
    # Store in database
    summary_id = hashlib.sha256(f"monthly-{month_key}".encode()).hexdigest()[:12]
    cursor.execute('''
        INSERT OR REPLACE INTO summaries (summary_id, period_type, period_key, content, event_count, created_at, model)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    ''', (summary_id, 'monthly', month_key, summary, len(weeklies),
          datetime.now(timezone.utc).isoformat(), 'auto'))
    
    conn.commit()
    conn.close()
    
    print(f"✓ Created monthly summary: {output_file}")
    return {
        "status": "ok",
        "month": month_key,
        "weekly_count": len(weeklies),
        "summary_file": str(output_file)
    }
