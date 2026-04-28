import json
import os
import sys
from queue import NocturneQueue
from router import JobRouter
from source_profile import SourceProfileManager

class Controller:
    """Generates and dispatches scanning jobs into the distributed pipeline."""
    
    def __init__(self, config_path="sources.json"):
        self.queue = NocturneQueue()
        self.router = JobRouter()
        self.profiles = SourceProfileManager()
        self.base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        self.config_path = os.path.join(self.base_dir, config_path)

    def dispatch_scan(self, target, priority=1.0):
        """Loads sources and creates individual jobs for the workers."""
        if not os.path.exists(self.config_path):
            print(f"[!] Config not found at {self.config_path}")
            return

        with open(self.config_path, 'r') as f:
            sources = json.load(f).get("sources", [])

        print(f"[*] Dispatching {len(sources)} jobs for target: {target}")
        for source in sources:
            source_id = source["id"]
            profile = self.profiles.get_profile(source_id)
            
            # 5. FAILURE HANDLING: Temporary Source Suspension
            if profile.get("success_rate", 1.0) < 0.2:
                print(f"[!] Skipping suspended source: {source_id} (Reliability too low)")
                continue

            # 2. CONTROLLER ROUTING: Multi-Region Assignment
            routing = self.router.get_routing_metadata(source_id, {}, self.profiles)

            job = {
                "target": target,
                "source_id": source_id,
                "url_template": source["url_template"],
                "priority": priority,
                "retries": 0,
                "routing": routing
            }
            self.queue.push_job(job, custom_key=routing["routing_key"])

if __name__ == "__main__":
    if len(sys.argv) < 2:
        sys.exit("Usage: python controller.py <target>")
    Controller().dispatch_scan(sys.argv[1])