import json
import requests
import time
import os
import random
import sys
import re
import hashlib
from classifier import SourceClassifier
from parser import AdaptiveParser
# FieldDetector and SourceProfiler are currently imported as stubs if files are missing
try:
    from core.detector import FieldDetector
    from core.profiler import SourceProfiler
except ImportError:
    class FieldDetector:
        @staticmethod
        def detect(s): return []
    class SourceProfiler:
        @staticmethod
        def profile_source(a, b, c): return {}

from datetime import datetime

class SourceHealthMonitor:
    """Analyzes source response health and calculates success rates."""
    def __init__(self, registry_path):
        self.path = registry_path

    def analyze(self, source_id, content, status, resp_time, profile, entry):
        success_rate = entry.get("reliability_score", 1.0)
        is_success = status == 200 and len(content) > 0
        
        # Adjust score: +0.05 for success, -0.10 for failure
        new_score = round(max(0.0, min(1.0, success_rate + (0.05 if is_success else -0.10))), 2)
        
        metrics = {"success_rate": new_score, "latency": resp_time}
        new_status = "active" if is_success else "degraded"
        return new_status, metrics

# Rotated User-Agents to prevent basic fingerprinting
USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
]

class SourceAdapter:
    def __init__(self, config_path="sources.json"):
        self.base_dir = os.path.dirname(os.path.abspath(__file__))
        self.config_path = os.path.join(self.base_dir, config_path)
        self.raw_dir = os.path.join(self.base_dir, "raw")
        self.registry_path = os.path.join(self.base_dir, "sources_registry.json")
        self.health_monitor = SourceHealthMonitor(self.registry_path)
        self.sources = self._load_sources()

    def _load_sources(self):
        try:
            with open(self.config_path, 'r') as f:
                return json.load(f)
        except Exception as e:
            sys.stderr.write(f"Error loading config: {e}\n")
            return []

    def _adjust_reliability(self, current_score, success):
        """Adjust score: +0.05 for success, -0.10 for failure. Clamped 0.0-1.0."""
        adjustment = 0.05 if success else -0.10
        return round(max(0.0, min(1.0, current_score + adjustment)), 2)

    def _update_registry(self, source_id, status_code, profile=None, source_type="scrape"):
        """Update the persistent source capability registry with health and scores."""
        try:
            registry = {}
            if os.path.exists(self.registry_path):
                with open(self.registry_path, 'r') as f:
                    registry = json.load(f)
            
            # 1. Get/Initialize entry
            entry = registry.get(source_id, {
                "id": source_id,
                "category": "identity",
                "type": source_type,
                "reliability_score": 1.0,
                "capabilities": {"supports_name": True, "supports_email": False},
                "change_frequency": 0.5
            })

            # 2. Run Health Forensics
            # Note: response_time is passed from _execute_source context
            resp_time = getattr(self, '_last_resp_time', "0.000s")
            raw_content = getattr(self, '_last_raw_content', "")
            
            new_status, health_metrics = self.health_monitor.analyze(
                source_id, raw_content, status_code, resp_time, profile, entry
            )

            # 3. Update Registry Entry
            entry["last_crawled"] = datetime.now().isoformat()
            entry["status"] = new_status
            entry["health_metrics"] = health_metrics
            entry["reliability_score"] = health_metrics["success_rate"]

            # 4. Merge profiled intelligence if available
            if profile:
                entry.update({k: v for k, v in profile.items() if v is not None})

            registry[source_id] = entry
            with open(self.registry_path, 'w') as f:
                json.dump(registry, f, indent=2)
        except Exception as e:
            sys.stderr.write(f"Registry update failed: {e}\n")

    def _save_raw(self, source_id, content, timestamp):
        path = os.path.join(self.raw_dir, source_id)
        os.makedirs(path, exist_ok=True)
        filename = os.path.join(path, f"{timestamp}.html")
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(content)

    def ingest(self, query):
        """Iterate through all sources and execute the query."""
        for source in self.sources:
            self._execute_source(source, query)

    def _execute_source(self, source, query):
        source_id = source['id']
        url = source['url_template'].format(query=query)
        method = source.get('method', 'GET')
        headers = source.get('headers', {})
        
        # Inject rotated User-Agent
        headers['User-Agent'] = random.choice(USER_AGENTS)
        
        start_time = time.time()
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        self._last_raw_content = "" # Reset
        
        try:
            response = requests.request(
                method=method,
                url=url,
                headers=headers,
                timeout=15 # Standard OSINT timeout
            )
            status = response.status_code
            raw_content = response.text
            self._last_raw_content = raw_content
            
            # Handle common blocking/rate-limit errors
            if status == 429:
                sys.stderr.write(f"WARN: Rate limited by {source_id}\n")
            elif status == 403:
                sys.stderr.write(f"WARN: Forbidden access to {source_id}\n")
                
        except requests.exceptions.Timeout:
            status = "timeout"
            raw_content = ""
        except Exception as e:
            status = f"error: {str(e)}"
            raw_content = ""

        response_time = time.time() - start_time
        self._last_resp_time = f"{response_time:.3f}s"

        profile = None
        intelligence = []
        if raw_content:
            self._save_raw(source_id, raw_content, timestamp)
            profile = SourceProfiler.profile_source(source_id, raw_content, source.get('url_template', ''))
            
            # Standardize logic: Classify -> Parse -> Detect
            classification = SourceClassifier.classify(source_id, raw_content)
            signals = AdaptiveParser.parse(raw_content, classification["content_type"])
            intelligence = FieldDetector.detect(signals)

        # Update registry with results of this run
        self._update_registry(source_id, status, profile, source.get('type', 'scrape'))

        # 4. Send results to stdout as JSON
        output = {
            "raw_response": raw_content,
            "metadata": {
                "source_id": source_id,
                "status": status,
                "response_time": f"{response_time:.3f}s",
                "timestamp": timestamp,
                "discovered_intelligence": intelligence
            }
        }
        print(json.dumps(output))
        sys.stdout.flush() # Ensure Go system receives the event immediately

        # 2. Respect rate limiting
        time.sleep(source.get('rate_limit', 1))

if __name__ == "__main__":
    if len(sys.argv) < 2:
        sys.exit("Usage: python adapter.py <query>")
    
    adapter = SourceAdapter()
    adapter.ingest(sys.argv[1])