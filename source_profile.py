import json
import os
import time
from datetime import datetime

class SourceProfileManager:
    """Tracks source health, success rates, and dynamic rate-limit recommendations."""
    
    def __init__(self, registry_path="sources_registry.json"):
        self.base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        self.registry_path = os.path.join(self.base_dir, registry_path)

    def _load(self):
        if os.path.exists(self.registry_path):
            with open(self.registry_path, 'r') as f:
                return json.load(f)
        return {}

    def _save(self, data):
        with open(self.registry_path, 'w') as f:
            json.dump(data, f, indent=2)

    def get_profile(self, source_id):
        registry = self._load()
        return registry.get(source_id, {
            "success_rate": 1.0,
            "avg_latency": 0.5,
            "recommended_delay": 2.0,
            "status": "active"
        })

    def update_stats(self, source_id, status_code, latency):
        registry = self._load()
        profile = registry.get(source_id, {"success_rate": 1.0, "recommended_delay": 2.0})
        
        success = 1.0 if status_code == 200 else 0.0
        
        # 1. Update Success Rate (Sliding Window)
        profile["success_rate"] = round((profile["success_rate"] * 0.8) + (success * 0.2), 2)
        
        # 2. Update Latency
        profile["avg_latency"] = round((profile.get("avg_latency", 0.5) * 0.7) + (latency * 0.3), 3)

        # 3. Dynamic Delay Adaptation
        current_delay = profile.get("recommended_delay", 2.0)
        if status_code == 429:
            profile["recommended_delay"] = min(60.0, current_delay * 2.5) # Aggressive backoff
            profile["status"] = "rate_limited"
        elif status_code == 403:
            profile["status"] = "blocked"
            profile["recommended_delay"] = min(120.0, current_delay * 3)
        elif status_code == 200:
            # Gradual recovery: slowly reduce delay if successful
            profile["recommended_delay"] = max(1.0, current_delay * 0.95)
            profile["status"] = "active"

        profile["last_updated"] = datetime.now().isoformat()
        registry[source_id] = profile
        self._save(registry)
        return profile

    def get_source_region_preference(self, source_id):
        # Heuristic: certain sources perform better in specific regions
        preferences = {"opencorporates": "US", "github": "US", "reddit": "EU"}
        return preferences.get(source_id, "ANY")