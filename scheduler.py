import json
import os
import sys
import hashlib
import time
from datetime import datetime
import requests
from adapter import SourceAdapter
from normalizer import normalize
# FieldDetector is used as the intelligence detection engine
from query_builder import QueryBuilder

class NocturneScheduler:
    """Orchestrates event-driven scans based on source priority and content volatility."""
    
    def __init__(self, state_path="state.json"):
        self.base_dir = os.path.dirname(os.path.abspath(__file__))
        self.state_path = os.path.join(self.base_dir, state_path)
        self.adapter = SourceAdapter()
        self.api_base_url = os.getenv("NOCTURNE_API_URL", "http://localhost:8080")
        self.query_builder = QueryBuilder()
        self.state = self._load_state()

    def _load_state(self):
        if os.path.exists(self.state_path):
            try:
                with open(self.state_path, 'r') as f:
                    return json.load(f)
            except:
                return {}
        return {}

    def _save_state(self):
        with open(self.state_path, 'w') as f:
            json.dump(self.state, f, indent=2)

    def _get_content_hash(self, text):
        return hashlib.sha256(text.encode('utf-8')).hexdigest()

    def emit_event(self, topic, target_id, payload):
        """Helper to format and output events for the Go event bus to consume."""
        event = {
            "topic": topic,
            "target_id": target_id,
            "payload": payload,
            "timestamp": datetime.utcnow().isoformat() + "Z"
        }
        
        # Feed to Go API if it's a normalized identity
        if topic == "normalized_identities":
            try:
                endpoint = f"{self.api_base_url}/api/graph"
                requests.post(endpoint, json={"target_id": target_id, "identity": payload}, timeout=5)
            except Exception as e:
                sys.stderr.write(f"[!] API Feed Failed: {e}\n")

        # Logs only to stderr, JSON events to stdout
        print(json.dumps(event))
        sys.stdout.flush()

    def log(self, message):
        sys.stderr.write(f"[*] [SCHEDULER] {datetime.now().strftime('%H:%M:%S')} - {message}\n")

    def run_scheduled_scan(self, query_data):
        """
        1. Prioritize sources with high change frequency.
        2. Execute scans.
        3. Compare hashes and emit events.
        """
        target_name = query_data.get("name") or query_data.get("email")
        self.log(f"Starting prioritized scan for: {target_name}")

        # 1. Prioritize sources based on registry change_frequency
        registry = self.query_builder.registry
        sorted_source_ids = sorted(
            registry.keys(),
            key=lambda k: registry[k].get("change_frequency", 0),
            reverse=True
        )

        # Get valid query URLs
        planned_queries = self.query_builder.build_queries(query_data)
        query_map = {q['source']: q['query_url'] for q in planned_queries if q['valid']}

        for source_id in sorted_source_ids:
            if source_id not in query_map:
                continue
            
            url = query_map[source_id]
            source_cfg = next((s for s in self.adapter.sources if s['id'] == source_id), {})
            
            self.log(f"Polling {source_id} (Frequency: {registry[source_id].get('change_frequency')})")
            
            # Execute scan via adapter logic (simplified call)
            # In production, we'd refactor adapter.py to allow calling _execute_source directly
            # For now, we simulate the execution and capture the result
            start_time = time.time()
            try:
                resp = requests.get(url, headers={"User-Agent": "Nocturne-Scheduler/1.0"}, timeout=15)
                content = resp.text
                status = resp.status_code
            except Exception as e:
                self.log(f"Failed to poll {source_id}: {e}")
                continue

            # 2. Compare content hash with previous run
            new_hash = self._get_content_hash(content)
            state_key = f"{target_name}:{source_id}"
            previous_hash = self.state.get(state_key, {}).get("hash")
            
            # 3. If changed: emit TopicRawData
            if new_hash != previous_hash:
                self.log(f"Change detected in {source_id} content!")
                self.emit_event("raw_data", target_name, {
                    "source_id": source_id,
                    "content": content,
                    "url": url,
                    "status": status
                })
                
                # Update state
                if state_key not in self.state: self.state[state_key] = {"identities": []}
                self.state[state_key]["hash"] = new_hash

                # 4. If new identity detected: emit TopicNormalizedIdentity
                intelligence = FieldIntelligenceDetector.detect(content)
                normalized_identity = IdentityNormalizer.normalize(source_id, target_name, intelligence, content)
                
                known_ids = self.state[state_key].get("identities", [])
                identity_sig = normalized_identity["id"]

                if identity_sig not in known_ids:
                    self.log(f"New Identity structured for {source_id} (Confidence: {normalized_identity['confidence']})")
                    self.emit_event("normalized_identities", target_name, normalized_identity)
                    
                    known_ids.append(identity_sig)
                    self.state[state_key]["identities"] = known_ids
                else:
                    self.log(f"Identity from {source_id} already known, skipping event emission.")
            else:
                self.log(f"No changes detected for {source_id}")

            # Respect rate limit
            time.sleep(source_cfg.get("rate_limit", 1))

        self._save_state()
        self.log("Scan cycle completed.")

if __name__ == "__main__":
    import requests # Required for the simulation inside run_scheduled_scan
    if len(sys.argv) < 2:
        print("Usage: python scheduler.py '{\"name\": \"John Smith\"}'")
        sys.exit(1)
    
    try:
        query = json.loads(sys.argv[1])
        NocturneScheduler().run_scheduled_scan(query)
    except Exception as e:
        sys.stderr.write(f"Critical Scheduler Error: {e}\n")